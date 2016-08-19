package inputdockerstats

import (
	"os"
	"regexp"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
	"github.com/tsaikd/KDGoLib/errutil"
	"github.com/tsaikd/gogstash/config"
)

const (
	ModuleName = "dockerstats"
)

type InputConfig struct {
	config.InputConfig
	DockerURL               string   `json:"dockerurl"`
	IncludePatterns         []string `json:"include_patterns"`
	ExcludePatterns         []string `json:"exclude_patterns"`
	StatInterval            int      `json:"stat_interval"`
	ConnectionRetryInterval int      `json:"connection_retry_interval,omitempty"`

	LogMode Mode `json:"log_mode,omitempty"`

	sincemap map[string]*time.Time
	includes []*regexp.Regexp
	excludes []*regexp.Regexp
	hostname string
	client   *docker.Client
}

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

		sincemap: map[string]*time.Time{},
	}
}

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
		since, ok := t.sincemap[container.ID]
		if !ok || since == nil {
			since = &time.Time{}
			t.sincemap[container.ID] = since
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
				since, ok := t.sincemap[container.ID]
				if !ok || since == nil {
					since = &time.Time{}
					t.sincemap[container.ID] = since
				}
				go t.containerLogLoop(container, since, inchan, logger)
			}
		}
	}
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
