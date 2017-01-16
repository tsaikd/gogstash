package inputhttp

import (
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"

	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
)

// ModuleName is the name used in config file
const ModuleName = "http"

// InputConfig holds the configuration json fields and internal objects
type InputConfig struct {
	config.InputConfig
	Method   string `json:"method,omitempty"` // one of ["HEAD", "GET"]
	Url      string `json:"url"`
	Interval int    `json:"interval,omitempty"`

	hostname string
}

// DefaultInputConfig returns an InputConfig struct with default values
func DefaultInputConfig() InputConfig {
	return InputConfig{
		InputConfig: config.InputConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		Method:   "GET",
		Interval: 60,
	}
}

// InitHandler initialize the input plugin
func InitHandler(confraw *config.ConfigRaw) (retconf config.TypeInputConfig, err error) {
	conf := DefaultInputConfig()
	if err = config.ReflectConfig(confraw, &conf); err != nil {
		return
	}

	if conf.hostname, err = os.Hostname(); err != nil {
		return
	}

	retconf = &conf
	return
}

// Start wraps the actual function starting the plugin
func (t *InputConfig) Start() {
	t.Invoke(t.start)
}

func (t *InputConfig) start(logger *logrus.Logger, inchan config.InChan) {
	startChan := make(chan bool) // startup tick
	ticker := time.NewTicker(time.Duration(t.Interval) * time.Second)

	go func() {
		startChan <- true
	}()

	for {
		select {
		case <-startChan:
			t.Request(logger, inchan)
		case <-ticker.C:
			t.Request(logger, inchan)
		}
	}
}

func (t *InputConfig) Request(logger *logrus.Logger, inchan config.InChan) {
	data, err := t.SendRequest()

	event := logevent.LogEvent{
		Timestamp: time.Now(),
		Message:   data,
		Extra: map[string]interface{}{
			"host": t.hostname,
			"url":  t.Url,
		},
	}

	if err != nil {
		event.AddTag("inputhttp_failed")
	}

	logger.Debugf("%v", event)
	inchan <- event

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
