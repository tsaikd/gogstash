package inputdocker

import (
	"encoding/json"
	"errors"
	"os"
	"regexp"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/fsouza/go-dockerclient"
	"github.com/tsaikd/gogstash/config"
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
}

func DefaultInputConfig() InputConfig {
	return InputConfig{
		CommonConfig: config.CommonConfig{
			Type: "docker",
		},
		Host: "unix:///var/run/docker.sock",
		ConnectionRetryInterval: 10,
		ExcludePatterns:         []string{"gogstash"},
		SincePath:               "sincedb",
	}
}

func init() {
	config.RegistInputHandler("docker", func(mapraw map[string]interface{}) (conf config.TypeInputConfig, err error) {
		var (
			raw []byte
		)
		if raw, err = json.Marshal(mapraw); err != nil {
			log.Error(err)
			return
		}
		defconf := DefaultInputConfig()
		conf = &defconf
		if err = json.Unmarshal(raw, &conf); err != nil {
			log.Error(err)
			return
		}
		for _, pattern := range defconf.IncludePatterns {
			defconf.includes = append(defconf.includes, regexp.MustCompile(pattern))
		}
		for _, pattern := range defconf.ExcludePatterns {
			defconf.excludes = append(defconf.excludes, regexp.MustCompile(pattern))
		}
		if defconf.sincedb, err = NewSinceDB(defconf.SincePath); err != nil {
			log.Error(err)
			return
		}
		return
	})
}

func (self *InputConfig) Type() string {
	return self.CommonConfig.Type
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

func (self *InputConfig) Loop() {
	hostname, err := os.Hostname()
	if err != nil {
		log.Errorf("Get hostname failed: %v", err)
	}

	eventExtra := map[string]interface{}{
		"host": hostname,
	}

	client, err := docker.NewClient(self.Host)
	if err != nil {
		log.Fatal("create docker client failed", err)
		return
	}

	containers, err := client.ListContainers(docker.ListContainersOptions{})
	if err != nil {
		log.Fatal("list docker container failed", err)
		return
	}

	for _, container := range containers {
		if !self.isValidContainer(container.Names) {
			continue
		}
		since, err := self.sincedb.Get(container.ID)
		if err != nil {
			log.Fatal("get sincedb failed", err)
			return
		}
		go containerLogLoop(client, container.ID, self.EventChan, eventExtra, since)
	}

	dockerEventChan := make(chan *docker.APIEvents)

	if err = client.AddEventListener(dockerEventChan); err != nil {
		log.Fatal("listen docker event failed", err)
		return
	}

	for {
		select {
		case dockerEvevt := <-dockerEventChan:
			if dockerEvevt.Status == "start" {
				container, err := client.InspectContainer(dockerEvevt.ID)
				if err != nil {
					log.Fatal("inspect container failed", err)
					return
				}
				if !self.isValidContainer([]string{container.Name}) {
					return
				}
				since, err := self.sincedb.Get(dockerEvevt.ID)
				if err != nil {
					log.Fatal("get sincedb failed", err)
					return
				}
				go containerLogLoop(client, dockerEvevt.ID, self.EventChan, eventExtra, since)
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

func (self *InputConfig) sendEvent(data string, hostname string, err error) {
	event := config.LogEvent{
		Timestamp: time.Now(),
		Message:   data,
		Extra: map[string]interface{}{
			"host": hostname,
		},
	}

	if err != nil {
		event.AddTag("inputdocker_failed")
	}

	log.Debugf("%v", event)
	self.EventChan <- event
}
