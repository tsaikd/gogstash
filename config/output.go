package config

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/inject"
	"github.com/tsaikd/KDGoLib/errutil"
	"github.com/tsaikd/KDGoLib/injectutil"
	"github.com/tsaikd/gogstash/config/logevent"
)

type TypeOutputConfig interface {
	TypeConfig
	Event(event logevent.LogEvent) (err error)
}

type OutputConfig struct {
	CommonConfig
}

type OutputHandler interface{}

var (
	mapOutputHandler = map[string]OutputHandler{}
)

func RegistOutputHandler(name string, handler OutputHandler) {
	mapOutputHandler[name] = handler
}

func (t *Config) RunOutputs() (err error) {
	_, err = t.Invoke(t.runOutputs)
	return
}

func (t *Config) runOutputs(evchan chan logevent.LogEvent, logger *logrus.Logger) (err error) {
	outputs, err := t.getOutputs()
	if err != nil {
		return errutil.New("get config output failed", err)
	}
	go func() {
		for {
			select {
			case event := <-evchan:
				for _, output := range outputs {
					if err = output.Event(event); err != nil {
						logger.Errorf("output failed: %v", err)
					}
				}
			}
		}
	}()
	return
}

func (t *Config) getOutputs() (outputs []TypeOutputConfig, err error) {
	for _, confraw := range t.OutputRaw {
		handler, ok := mapOutputHandler[confraw["type"].(string)]
		if !ok {
			err = fmt.Errorf("unknown output config type: %q", confraw["type"])
			return
		}

		inj := inject.New()
		inj.SetParent(t)
		inj.Map(&confraw)
		refvs, err := injectutil.Invoke(inj, handler)
		if err != nil {
			err = errutil.NewErrorSlice(fmt.Errorf("handle output config failed: %q", confraw), err)
			return []TypeOutputConfig{}, err
		}

		for _, refv := range refvs {
			if !refv.CanInterface() {
				continue
			}
			if conf, ok := refv.Interface().(TypeOutputConfig); ok {
				conf.SetInjector(inj)
				outputs = append(outputs, conf)
			}
		}
	}
	return
}
