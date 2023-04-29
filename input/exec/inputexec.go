package inputexec

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/tsaikd/KDGoLib/errutil"

	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
)

// ModuleName is the name used in config file
const ModuleName = "exec"

// ErrorTag tag added to event when process module failed
const ErrorTag = "gogstash_input_exec_error"

// InputConfig holds the configuration json fields and internal objects
type InputConfig struct {
	config.InputConfig
	Command   string   `json:"command"`                  // Command to run. e.g. “uptime”
	Args      []string `json:"args,omitempty"`           // Arguments of command
	Interval  int      `json:"interval,omitempty"`       // Second, default: 60
	MsgTrim   string   `json:"message_trim,omitempty"`   // default: " \t\r\n"
	MsgPrefix string   `json:"message_prefix,omitempty"` // only in text type, e.g. "%{@timestamp} [uptime] "
	MsgType   MsgType  `json:"message_type,omitempty"`   // default: "text"

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
		Interval: 60,
		MsgTrim:  " \t\r\n",
		MsgType:  MsgTypeText,
	}
}

// errors
var (
	ErrorExecCommandFailed1 = errutil.NewFactory("run exec failed: %q")
)

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
			t.exec(msgChan)
		case <-ticker.C:
			t.exec(msgChan)
		}
	}
}

func (t *InputConfig) exec(msgChan chan<- logevent.LogEvent) {
	errs := []error{}

	message, err := t.doExecCommand()
	if err != nil {
		errs = append(errs, err)
	}
	extra := map[string]any{
		"host": t.hostname,
	}

	switch t.MsgType {
	case MsgTypeJson:
		if err = jsoniter.Unmarshal([]byte(message), &extra); err != nil {
			errs = append(errs, err)
		} else {
			message = ""
		}
	}

	event := logevent.LogEvent{
		Timestamp: time.Now(),
		Message:   message,
		Extra:     extra,
	}

	switch t.MsgType {
	case MsgTypeText:
		event.Message = event.Format(t.MsgPrefix) + event.Message
	}

	if len(errs) > 0 {
		event.AddTag(ErrorTag)
		event.Extra["error"] = errutil.NewErrors(errs...).Error()
	}

	msgChan <- event
}

func (t *InputConfig) doExecCommand() (data string, err error) {
	buferr := &bytes.Buffer{}
	cmd := exec.Command(t.Command, t.Args...)
	cmd.Stderr = buferr

	raw, err := cmd.Output()
	if err != nil {
		return
	}
	data = string(raw)
	data = strings.Trim(data, t.MsgTrim)
	if buferr.Len() > 0 {
		err = ErrorExecCommandFailed1.New(nil, buferr.String())
	}
	return
}
