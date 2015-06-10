package inputexec

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/tsaikd/gogstash/config"
)

const (
	ModuleName = "exec"
)

type InputConfig struct {
	config.CommonConfig
	Command   string   `json:"command"`                  // Command to run. e.g. “uptime”
	Args      []string `json:"args,omitempty"`           // Arguments of command
	Interval  int      `json:"interval,omitempty"`       // Second, default: 60
	MsgTrim   string   `json:"message_trim,omitempty"`   // default: " \t\r\n"
	MsgPrefix string   `json:"message_prefix,omitempty"` // e.g. "%{@timestamp} [uptime] "

	EventChan chan config.LogEvent `json:"-"`
}

func DefaultInputConfig() InputConfig {
	return InputConfig{
		CommonConfig: config.CommonConfig{
			Type: ModuleName,
		},
		Interval: 60,
		MsgTrim:  " \t\r\n",
	}
}

func init() {
	config.RegistInputHandler(ModuleName, func(mapraw map[string]interface{}) (conf config.TypeInputConfig, err error) {
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
			self.Exec(hostname)
		case <-ticker.C:
			self.Exec(hostname)
		}
	}

	return
}

func (self *InputConfig) Exec(hostname string) {
	var (
		err  error
		data string
	)

	data, err = self.doExec()

	event := config.LogEvent{
		Timestamp: time.Now(),
		Message:   data,
		Extra: map[string]interface{}{
			"host": hostname,
		},
	}

	event.Message = event.Format(self.MsgPrefix) + event.Message

	if err != nil {
		event.AddTag("inputexec_failed")
		event.Extra["error"] = err.Error()
	}

	log.Debugf("%v", event)
	self.EventChan <- event

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
