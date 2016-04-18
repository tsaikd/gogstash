package config

import (
	"fmt"

	"github.com/codegangsta/inject"
	"github.com/tsaikd/KDGoLib/errutil"
	"github.com/tsaikd/KDGoLib/injectutil"
	"github.com/tsaikd/gogstash/config/logevent"
)

type TypeInputConfig interface {
	TypeConfig
	Start()
}

type InputConfig struct {
	CommonConfig
}

type InputHandler interface{}

var (
	mapInputHandler = map[string]InputHandler{}
)

func RegistInputHandler(name string, handler InputHandler) {
	mapInputHandler[name] = handler
}

func (t *Config) RunInputs() (err error) {
	_, err = t.Invoke(t.runInputs)
	return
}

func (t *Config) runInputs(evchan chan logevent.LogEvent) (err error) {
	inputs, err := t.getInputs(evchan)
	if err != nil {
		return errutil.New("get config inputs failed", err)
	}
	for _, input := range inputs {
		go input.Start()
	}
	return
}

func (t *Config) getInputs(evchan chan logevent.LogEvent) (inputs []TypeInputConfig, err error) {
	for _, confraw := range t.InputRaw {
		handler, ok := mapInputHandler[confraw["type"].(string)]
		if !ok {
			err = fmt.Errorf("unknown input config type: %q", confraw["type"])
			return
		}

		inj := inject.New()
		inj.SetParent(t)
		inj.Map(&confraw)
		inj.Map(evchan)
		refvs, err := injectutil.Invoke(inj, handler)
		if err != nil {
			err = errutil.NewErrors(fmt.Errorf("handle input config failed: %q", confraw), err)
			return []TypeInputConfig{}, err
		}

		for _, refv := range refvs {
			if !refv.CanInterface() {
				continue
			}
			if conf, ok := refv.Interface().(TypeInputConfig); ok {
				conf.SetInjector(inj)
				inputs = append(inputs, conf)
			}
		}
	}
	return
}
