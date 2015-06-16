package inputhttp

import (
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/tsaikd/gogstash/config"
)

const (
	ModuleName = "http"
)

type InputConfig struct {
	config.CommonConfig
	Method   string `json:"method,omitempty"` // one of ["HEAD", "GET"]
	Url      string `json:"url"`
	Interval int    `json:"interval,omitempty"`

	EventChan chan config.LogEvent `json:"-"`
}

func DefaultInputConfig() InputConfig {
	return InputConfig{
		CommonConfig: config.CommonConfig{
			Type: ModuleName,
		},
		Method:   "GET",
		Interval: 60,
	}
}

func init() {
	config.RegistInputHandler(ModuleName, func(mapraw map[string]interface{}) (retconf config.TypeInputConfig, err error) {
		conf := DefaultInputConfig()
		if err = config.ReflectConfig(mapraw, &conf); err != nil {
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

	go self.RequestLoop()

	return
}

func (self *InputConfig) RequestLoop() {
	var (
		hostname  string
		err       error
		startChan = make(chan bool) // startup tick
		ticker    = time.NewTicker(time.Duration(self.Interval) * time.Second)
	)

	if hostname, err = os.Hostname(); err != nil {
		log.Errorf("Get hostname failed: %v", err)
	}

	go func() {
		startChan <- true
	}()

	for {
		select {
		case <-startChan:
			self.Request(hostname)
		case <-ticker.C:
			self.Request(hostname)
		}
	}

	return
}

func (self *InputConfig) Request(hostname string) {
	var (
		err  error
		data string
	)

	data, err = self.SendRequest()

	event := config.LogEvent{
		Timestamp: time.Now(),
		Message:   data,
		Extra: map[string]interface{}{
			"host": hostname,
			"url":  self.Url,
		},
	}

	if err != nil {
		event.AddTag("inputhttp_failed")
	}

	log.Debugf("%v", event)
	self.EventChan <- event

	return
}

func (self *InputConfig) SendRequest() (data string, err error) {
	var (
		res *http.Response
		raw []byte
	)
	switch self.Method {
	case "HEAD":
		res, err = http.Head(self.Url)
	case "GET":
		res, err = http.Get(self.Url)
	default:
		err = errors.New("Unknown method")
	}

	if err != nil {
		return
	}

	defer res.Body.Close()
	if raw, err = ioutil.ReadAll(res.Body); err != nil {
		return
	}
	data = string(raw)
	data = strings.TrimSpace(data)

	return
}
