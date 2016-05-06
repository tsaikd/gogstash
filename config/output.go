package config

import (
	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/inject"
	"github.com/tsaikd/KDGoLib/errutil"
	"github.com/tsaikd/KDGoLib/injectutil"
	"github.com/tsaikd/gogstash/config/logevent"
)

// errors
var (
	ErrorUnknownOutputType1 = errutil.NewFactory("unknown output config type: %q")
	ErrorRunOutput1         = errutil.NewFactory("run output module failed: %q")
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
	return t.InvokeSimple(t.runOutputs)
}

func (t *Config) runOutputs(outchan OutChan, logger *logrus.Logger) (err error) {
	outputs, err := t.getOutputs()
	if err != nil {
		return
	}
	go func() {
		for {
			select {
			case event := <-outchan:
				for _, output := range outputs {
					if err = output.Event(event); err != nil {
						logger.Errorf("output failed: %v\n", err)
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
			err = ErrorUnknownOutputType1.New(nil, confraw["type"])
			return
		}

		inj := inject.New()
		inj.SetParent(t)
		inj.Map(&confraw)
		refvs, err := injectutil.Invoke(inj, handler)
		if err != nil {
			err = ErrorRunOutput1.New(err, confraw)
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
