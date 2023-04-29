package inputhttp

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
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

	control  config.Control
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
func InitHandler(
	ctx context.Context,
	raw config.ConfigRaw,
	control config.Control,
) (config.TypeInputConfig, error) {
	conf := DefaultInputConfig()
	err := config.ReflectConfig(raw, &conf)
	if err != nil {
		return nil, err
	}

	conf.control = control
	if conf.hostname, err = os.Hostname(); err != nil {
		return nil, err
	}

	conf.Codec, err = config.GetCodecOrDefault(ctx, raw["codec"])
	if err != nil {
		return nil, err
	}

	return &conf, err
}

// Start wraps the actual function starting the plugin
func (t *InputConfig) Start(
	ctx context.Context,
	msgChan chan<- logevent.LogEvent,
) (err error) {
	startChan := make(chan bool, 1) // startup tick
	ticker := time.NewTicker(time.Duration(t.Interval) * time.Second)
	defer ticker.Stop()

	startChan <- true
	isPaused := false

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-startChan:
			t.Request(ctx, msgChan)
		case <-t.control.PauseSignal():
			goglog.Logger.Info("pause received")
			isPaused = true
		case <-t.control.ResumeSignal():
			goglog.Logger.Info("resume received")
			isPaused = false
		case <-ticker.C:
			if !isPaused {
				t.Request(ctx, msgChan)
			}
		}
	}
}

func (t *InputConfig) Request(ctx context.Context, msgChan chan<- logevent.LogEvent) {
	data, err := t.SendRequest()
	extra := map[string]any{
		"host": t.hostname,
		"url":  t.URL,
	}
	tags := []string{}
	if err != nil {
		tags = append(tags, ErrorTag)
	}
	_, err = t.Codec.Decode(ctx, []byte(data),
		extra,
		tags,
		msgChan)

	if err != nil {
		goglog.Logger.Errorf("%v", err)
	}
}

func (t *InputConfig) SendRequest() (data []byte, err error) {
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
	if raw, err = io.ReadAll(res.Body); err != nil {
		return
	}
	data = bytes.TrimSpace(raw)

	return
}
