package inputdockerstats

import (
	"context"
	"os"
	"regexp"
	"time"

	"github.com/fsouza/go-dockerclient"
	"github.com/tsaikd/KDGoLib/errutil"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
	"github.com/tsaikd/gogstash/input/dockerlog/dockertool"
	"golang.org/x/sync/errgroup"
)

// ModuleName is the name used in config file
const ModuleName = "dockerstats"

// InputConfig holds the configuration json fields and internal objects
type InputConfig struct {
	config.InputConfig
	DockerURL               string   `json:"dockerurl"`
	IncludePatterns         []string `json:"include_patterns"`
	ExcludePatterns         []string `json:"exclude_patterns"`
	StatInterval            int      `json:"stat_interval"`
	ConnectionRetryInterval int      `json:"connection_retry_interval,omitempty"`
	LogMode                 Mode     `json:"log_mode,omitempty"`

	sincemap
	containerExist dockertool.StringExist
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
		StatInterval:            15,
		ConnectionRetryInterval: 10,
		LogMode:                 ModeFull,

		sincemap:       newSinceMap(),
		containerExist: dockertool.NewStringExist(),
	}
}

// errors
var (
	ErrorPingFailed              = errutil.NewFactory("ping docker server failed")
	ErrorListContainerFailed     = errutil.NewFactory("list docker container failed")
	ErrorListenDockerEventFailed = errutil.NewFactory("listen docker event failed")
	ErrorContainerLoopRunning1   = errutil.NewFactory("container log loop running: %s")
	ErrorGetContainerInfoFailed  = errutil.NewFactory("get container info failed")
)

// InitHandler initialize the input plugin
func InitHandler(ctx context.Context, raw *config.ConfigRaw) (config.TypeInputConfig, error) {
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
	if conf.hostname, err = os.Hostname(); err != nil {
		return nil, err
	}
	if conf.client, err = docker.NewClient(conf.DockerURL); err != nil {
		return nil, err
	}
	if err = conf.client.Ping(); err != nil {
		return nil, ErrorPingFailed.New(err)
	}

	// This is really a "reference" codec instance, with each Stream getting their own copy.
	//  copying codec instances is needed to allow codecs to do sequential processing, such as milti-line logs with proper isolation.
	conf.Codec, err = config.GetCodecOrDefault(ctx, *raw)

	return &conf, nil
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
			since := t.sincemap.ensure(container.ID)
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
					container, err := t.client.InspectContainer(dockerEvent.ID)
					if err != nil {
						return errutil.New("inspect container failed", err)
					}
					if !t.isValidContainer([]string{container.Name}) {
						return errutil.New("invalid container name " + container.Name)
					}
					since := t.sincemap.ensure(container.ID)
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
	if len(t.includes) > 0 {
		return false
	}
	return true
}
