package inputdockerlog

import (
	"os"
	"regexp"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
	"github.com/tsaikd/KDGoLib/errutil"
	"github.com/tsaikd/gogstash/config"
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

	sincedb  *SinceDB
	includes []*regexp.Regexp
	excludes []*regexp.Regexp
	hostname string
	client   *docker.Client
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
	}
}

// InitHandler initialize the input plugin
func InitHandler(confraw *config.ConfigRaw) (retconf config.TypeInputConfig, err error) {
	conf := DefaultInputConfig()
	if err = config.ReflectConfig(confraw, &conf); err != nil {
		return
	}

	for _, pattern := range conf.IncludePatterns {
		conf.includes = append(conf.includes, regexp.MustCompile(pattern))
	}
	for _, pattern := range conf.ExcludePatterns {
		conf.excludes = append(conf.excludes, regexp.MustCompile(pattern))
	}
	if conf.sincedb, err = NewSinceDB(conf.SincePath); err != nil {
		return
	}
	if conf.hostname, err = os.Hostname(); err != nil {
		err = errutil.New("get hostname failed", err)
		return
	}
	if conf.client, err = docker.NewClient(conf.DockerURL); err != nil {
		err = errutil.New("create docker client failed", err)
		return
	}

	retconf = &conf
	return
}

// Start wraps the actual function starting the plugin
func (t *InputConfig) Start() {
	t.Invoke(t.start)
}

func (t *InputConfig) start(logger *logrus.Logger, inchan config.InChan) (err error) {
	defer func() {
		if err != nil {
			logger.Errorln(err)
		}
	}()

	containers, err := t.client.ListContainers(docker.ListContainersOptions{})
	if err != nil {
		return errutil.New("list docker container failed", err)
	}

	for _, container := range containers {
		if !t.isValidContainer(container.Names) {
			continue
		}
		since, err := t.getSince(container.ID)
		if err != nil {
			return err
		}
		go t.containerLogLoop(container, since, inchan, logger)
	}

	dockerEventChan := make(chan *docker.APIEvents)

	if err = t.client.AddEventListener(dockerEventChan); err != nil {
		return errutil.New("listen docker event failed", err)
	}

	for {
		select {
		case dockerEvent := <-dockerEventChan:
			if dockerEvent.Status == "start" {
				container, err := t.client.InspectContainer(dockerEvent.ID)
				if err != nil {
					return errutil.New("inspect container failed", err)
				}
				if !t.isValidContainer([]string{container.Name}) {
					return errutil.New("invalid container name " + container.Name)
				}
				since, err := t.getSince(dockerEvent.ID)
				if err != nil {
					return err
				}
				go t.containerLogLoop(container, since, inchan, logger)
			}
		}
	}
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

	if len(t.includes) > 0 {
		return false
	}
	return true
}
