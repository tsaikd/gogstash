package inputhttp

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
)

// ModuleName is the name used in config file
const ModuleName = "http"

// ErrorTag tag added to event when process module failed
const ErrorTag = "gogstash_input_http_error"

// InputConfig holds the configuration json fields and internal objects
type InputConfig struct {
	config.InputConfig
	Method   string `json:"method,omitempty"` // one of ["HEAD", "GET"]
	URL      string `json:"url"`
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
func InitHandler(ctx context.Context, raw *config.ConfigRaw) (config.TypeInputConfig, error) {
	conf := DefaultInputConfig()
	err := config.ReflectConfig(raw, &conf)
	if err != nil {
		return nil, err
	}

	if conf.hostname, err = os.Hostname(); err != nil {
		return nil, err
	}

	return &conf, nil
}

// Start wraps the actual function starting the plugin
func (t *InputConfig) Start(ctx context.Context, msgChan chan<- logevent.LogEvent) (err error) {
	startChan := make(chan bool, 1) // startup tick
	ticker := time.NewTicker(time.Duration(t.Interval) * time.Second)
	defer ticker.Stop()

	startChan <- true

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-startChan:
			t.Request(config.Logger, msgChan)
		case <-ticker.C:
			t.Request(config.Logger, msgChan)
		}
	}
}

func (t *InputConfig) Request(logger *logrus.Logger, msgChan chan<- logevent.LogEvent) {
	data, err := t.SendRequest()

	event := logevent.LogEvent{
		Timestamp: time.Now(),
		Message:   data,
		Extra: map[string]interface{}{
			"host": t.hostname,
			"url":  t.URL,
		},
	}

	if err != nil {
		event.AddTag(ErrorTag)
	}

	logger.Debugf("%v", event)
	msgChan <- event

	return
}

func (t *InputConfig) SendRequest() (data string, err error) {
	var (
		res *http.Response
		raw []byte
	)
	switch t.Method {
	case "HEAD":
		res, err = http.Head(t.URL)
	case "GET":
		res, err = http.Get(t.URL)
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
