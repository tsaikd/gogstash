package inputexec

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"

	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
)

const (
	ModuleName = "exec"
)

type InputConfig struct {
	config.InputConfig
	Command   string   `json:"command"`                  // Command to run. e.g. “uptime”
	Args      []string `json:"args,omitempty"`           // Arguments of command
	Interval  int      `json:"interval,omitempty"`       // Second, default: 60
	MsgTrim   string   `json:"message_trim,omitempty"`   // default: " \t\r\n"
	MsgPrefix string   `json:"message_prefix,omitempty"` // e.g. "%{@timestamp} [uptime] "

	hostname string `json:"-"`
}

func DefaultInputConfig() InputConfig {
	return InputConfig{
		InputConfig: config.InputConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		Interval: 60,
		MsgTrim:  " \t\r\n",
	}
}

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

func (self *InputConfig) Start() {
	startChan := make(chan bool) // startup tick
	ticker := time.NewTicker(time.Duration(self.Interval) * time.Second)

	go func() {
		startChan <- true
	}()

	for {
		select {
		case <-startChan:
			self.Invoke(self.Exec)
		case <-ticker.C:
			self.Invoke(self.Exec)
		}
	}
}

func (self *InputConfig) Exec(evchan chan logevent.LogEvent, logger *logrus.Logger) {
	data, err := self.doExec()

	event := logevent.LogEvent{
		Timestamp: time.Now(),
		Message:   data,
		Extra: map[string]interface{}{
			"host": self.hostname,
		},
	}

	event.Message = event.Format(self.MsgPrefix) + event.Message

	if err != nil {
		event.AddTag("inputexec_failed")
		event.Extra["error"] = err.Error()
	}

	logger.Debugf("%+v", event)
	evchan <- event

	return
}

func (self *InputConfig) doExec() (data string, err error) {
	var (
		buferr bytes.Buffer
		raw    []byte
		cmd    *exec.Cmd
	)
	cmd = exec.Command(self.Command, self.Args...)
	cmd.Stderr = &buferr
	if raw, err = cmd.Output(); err != nil {
		return
	}
	data = string(raw)
	data = strings.Trim(data, self.MsgTrim)
	if buferr.Len() > 0 {
		err = errors.New(buferr.String())
	}
	return
}
