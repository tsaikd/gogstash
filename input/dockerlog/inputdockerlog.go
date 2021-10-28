package inputdockerlog

import (
	"context"
	"os"
	"regexp"
	"time"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/tsaikd/KDGoLib/errutil"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
	"github.com/tsaikd/gogstash/input/dockerlog/dockertool"
	"golang.org/x/sync/errgroup"
)

// ModuleName is the name used in config file
const ModuleName = "dockerlog"

// InputConfig holds the configuration json fields and internal objects
type InputConfig struct {
	config.InputConfig
	DockerURL               string   `json:"dockerurl"`
	IncludePatterns         []string `json:"include_patterns"`
	ExcludePatterns         []string `json:"exclude_patterns"`
	SincePath               string   `json:"sincepath"`
	StartPos                string   `json:"start_position,omitempty"` // one of ["beginning", "end"]
	ConnectionRetryInterval int      `json:"connection_retry_interval,omitempty"`
	TLSCert                 string   `json:"tls_cert,omitempty"`
	TLSCertKey              string   `json:"tls_cert_key,omitempty"`
	TLSCaCert               string   `json:"tls_ca_cert,omitempty"`

	containerExist dockertool.StringExist
	sincedb        *SinceDB
	includes       []*regexp.Regexp
	excludes       []*regexp.Regexp
	hostname       string
	client         *docker.Client
}

// DefaultInputConfig returns an InputConfig struct with default values
func DefaultInputConfig() InputConfig {
	return InputConfig{
		InputConfig: config.InputConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		DockerURL:               "unix:///var/run/docker.sock",
		ConnectionRetryInterval: 10,
		ExcludePatterns:         []string{"gogstash"},
		SincePath:               "sincedb-%{HOSTNAME}",
		StartPos:                "beginning",

		containerExist: dockertool.NewStringExist(),
	}
}

// errors
var (
	ErrorPingFailed              = errutil.NewFactory("ping docker server failed")
	ErrorListContainerFailed     = errutil.NewFactory("list docker container failed")
	ErrorListenDockerEventFailed = errutil.NewFactory("listen docker event failed")
	ErrorInspectContainerFailed  = errutil.NewFactory("inspect container failed")
	ErrorContainerLoopRunning1   = errutil.NewFactory("container log loop running: %s")
	ErrorGetContainerInfoFailed  = errutil.NewFactory("get container info failed")
)

// InitHandler initialize the input plugin
func InitHandler(ctx context.Context, raw config.ConfigRaw) (config.TypeInputConfig, error) {
	conf := DefaultInputConfig()
	err := config.ReflectConfig(raw, &conf)
	if err != nil {
		return nil, err
	}

	for _, pattern := range conf.IncludePatterns {
		conf.includes = append(conf.includes, regexp.MustCompile(pattern))
	}
	for _, pattern := range conf.ExcludePatterns {
		conf.excludes = append(conf.excludes, regexp.MustCompile(pattern))
	}
	if conf.sincedb, err = NewSinceDB(conf.SincePath); err != nil {
		return nil, err
	}
	if conf.hostname, err = os.Hostname(); err != nil {
		return nil, err
	}
	if conf.TLSCert != "" && conf.TLSCaCert != "" && conf.TLSCertKey != "" {
		if conf.client, err = docker.NewTLSClient(conf.DockerURL, conf.TLSCert, conf.TLSCertKey, conf.TLSCaCert); err != nil {
			return nil, err
		}
	} else {
		if conf.client, err = docker.NewClient(conf.DockerURL); err != nil {
			return nil, err
		}
	}
	if err = conf.client.Ping(); err != nil {
		return nil, ErrorPingFailed.New(err)
	}

	// This is really a "reference" codec instance, with each Stream getting their own copy.
	//  copying codec instances is needed to allow codecs to do sequential processing, such as multi-line logs with proper isolation.
	conf.Codec, err = config.GetCodecOrDefault(ctx, raw["codec"])
	if err != nil {
		return nil, err
	}

	return &conf, err
}

// Start wraps the actual function starting the plugin
func (t *InputConfig) Start(ctx context.Context, msgChan chan<- logevent.LogEvent) error {
	eg, ctx := errgroup.WithContext(ctx)

	// listen running containers
	eg.Go(func() error {
		containers, err := t.client.ListContainers(docker.ListContainersOptions{})
		if err != nil {
			return ErrorListContainerFailed.New(err)
		}

		for _, container := range containers {
			if !t.isValidContainer(container.Names) {
				continue
			}
			since, err2 := t.getSince(container.ID)
			if err2 != nil {
				return err2
			}
			func(container interface{}, since *time.Time) {
				eg.Go(func() error {
					return t.containerLogLoop(ctx, container, since, msgChan)
				})
			}(container, since)
		}

		return nil
	})

	// listen for running in future containers
	eg.Go(func() error {
		dockerEventChan := make(chan *docker.APIEvents)

		if err := t.client.AddEventListener(dockerEventChan); err != nil {
			return ErrorListenDockerEventFailed.New(err)
		}

		for {
			select {
			case <-ctx.Done():
				return nil
			case dockerEvent := <-dockerEventChan:
				if dockerEvent.Status == "start" {
					container, err := t.client.InspectContainerWithOptions(
						docker.InspectContainerOptions{ID: dockerEvent.ID})
					if err != nil {
						return ErrorInspectContainerFailed.New(err)
					}
					if !t.isValidContainer([]string{container.Name}) {
						continue
					}
					since, err := t.getSince(dockerEvent.ID)
					if err != nil {
						return err
					}
					func(container interface{}, since *time.Time) {
						eg.Go(func() error {
							return t.containerLogLoop(ctx, container, since, msgChan)
						})
					}(container, since)
				}
			}
		}
	})

	return eg.Wait()
}

func (t *InputConfig) getSince(containerID string) (since *time.Time, err error) {
	since, err = t.sincedb.Get(containerID)
	if err != nil {
		return nil, errutil.New("get sincedb failed", err)
	}
	if since.IsZero() && t.StartPos == "end" {
		now := time.Now()
		since = &now
	}
	return
}

func (t *InputConfig) isValidContainer(names []string) bool {
	for _, name := range names {
		for _, re := range t.excludes {
			if re.MatchString(name) {
				return false
			}
		}
		for _, re := range t.includes {
			if re.MatchString(name) {
				return true
			}
		}
	}

	return len(t.includes) < 1
}
