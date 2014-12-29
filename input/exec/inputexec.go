package inputexec

import (
	"bytes"
	"encoding/json"
	"errors"
	"os/exec"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/tsaikd/gogstash/config"
)

const (
	ModuleName = "exec"
)

type InputConfig struct {
	config.CommonConfig
	Command  string   `json:"command"`        // Command to run. e.g. “uptime”
	Args     []string `json:"args,omitempty"` // Arguments of command
	Interval int      `json:"interval,omitempty"`

	EventChan chan config.LogEvent `json:"-"`
}

func DefaultInputConfig() InputConfig {
	return InputConfig{
		CommonConfig: config.CommonConfig{
			Type: ModuleName,
		},
		Interval: 60,
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
		startChan = make(chan bool) // startup tick
		ticker    = time.NewTicker(time.Duration(self.Interval) * time.Second)
	)

	go func() {
		startChan <- true
	}()

	for {
		select {
		case <-startChan:
			self.Exec()
		case <-ticker.C:
			self.Exec()
		}
	}

	return
}

func (self *InputConfig) Exec() {
	var (
		err  error
		data string
	)

	data, err = self.doExec()

	event := config.LogEvent{
		Timestamp: time.Now(),
		Message:   data,
		Extra:     map[string]interface{}{
		//"url": self.Url,
		},
	}

	if err != nil {
		event.AddTag("intputexec_failed")
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
	if buferr.Len() > 0 {
		err = errors.New(buferr.String())
	}
	return
}
