package inputdocker

import (
	"errors"
	"os"
	"regexp"

	log "github.com/Sirupsen/logrus"

	"github.com/fsouza/go-dockerclient"
	"github.com/tsaikd/KDGoLib/errutil"
	"github.com/tsaikd/gogstash/config"
)

const (
	ModuleName = "docker"
)

type InputConfig struct {
	config.CommonConfig
	Host                    string   `json:"host"`
	IncludePatterns         []string `json:"include_patterns"`
	ExcludePatterns         []string `json:"exclude_patterns"`
	SincePath               string   `json:"sincepath"`
	ConnectionRetryInterval int      `json:"connection_retry_interval,omitempty"`

	EventChan chan config.LogEvent `json:"-"`
	sincedb   *SinceDB             `json:"-"`
	includes  []*regexp.Regexp     `json:"-"`
	excludes  []*regexp.Regexp     `json:"-"`
	hostname  string               `json:"-"`
	client    *docker.Client       `json:"-"`
}

func DefaultInputConfig() InputConfig {
	return InputConfig{
		CommonConfig: config.CommonConfig{
			Type: ModuleName,
		},
		Host: "unix:///var/run/docker.sock",
		ConnectionRetryInterval: 10,
		ExcludePatterns:         []string{"gogstash"},
		SincePath:               "sincedb",
	}
}

func init() {
	config.RegistInputHandler(ModuleName, func(mapraw map[string]interface{}) (retconf config.TypeInputConfig, err error) {
		conf := DefaultInputConfig()
		if err = config.ReflectConfig(mapraw, &conf); err != nil {
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
		if conf.client, err = docker.NewClient(conf.Host); err != nil {
			err = errutil.New("create docker client failed", err)
			return
		}

		retconf = &conf
		return
	})
}

func (self *InputConfig) Event(eventChan chan config.LogEvent) (err error) {
	if self.EventChan != nil {
		err = errors.New("Event chan already inited")
		log.Error(err)
		return
	}
	self.EventChan = eventChan

	go self.Loop()

	return
}

func (t *InputConfig) Loop() {
	containers, err := t.client.ListContainers(docker.ListContainersOptions{})
	if err != nil {
		log.Fatal("list docker container failed", err)
		return
	}

	for _, container := range containers {
		if !t.isValidContainer(container.Names) {
			continue
		}
		since, err := t.sincedb.Get(container.ID)
		if err != nil {
			log.Fatal("get sincedb failed", err)
			return
		}
		go t.containerLogLoop(container, since)
	}

	dockerEventChan := make(chan *docker.APIEvents)

	if err = t.client.AddEventListener(dockerEventChan); err != nil {
		log.Fatal("listen docker event failed", err)
		return
	}

	for {
		select {
		case dockerEvent := <-dockerEventChan:
			if dockerEvent.Status == "start" {
				container, err := t.client.InspectContainer(dockerEvent.ID)
				if err != nil {
					log.Fatal("inspect container failed", err)
					return
				}
				if !t.isValidContainer([]string{container.Name}) {
					return
				}
				since, err := t.sincedb.Get(dockerEvent.ID)
				if err != nil {
					log.Fatal("get sincedb failed", err)
					return
				}
				go t.containerLogLoop(container, since)
			}
		}
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
	} else {
		return true
	}
}
