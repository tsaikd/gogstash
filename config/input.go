package config

import (
	"github.com/codegangsta/inject"
	"github.com/tsaikd/KDGoLib/errutil"
	"github.com/tsaikd/KDGoLib/injectutil"
)

// errors
var (
	ErrorUnknownInputType1 = errutil.NewFactory("unknown input config type: %q")
	ErrorRunInput1         = errutil.NewFactory("run input module failed: %q")
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
	return t.InvokeSimple(t.runInputs)
}

func (t *Config) runInputs(inchan InChan) (err error) {
	inputs, err := t.getInputs(inchan)
	if err != nil {
		return
	}
	for _, input := range inputs {
		go input.Start()
	}
	return
}

func (t *Config) getInputs(inchan InChan) (inputs []TypeInputConfig, err error) {
	for _, confraw := range t.InputRaw {
		handler, ok := mapInputHandler[confraw["type"].(string)]
		if !ok {
			err = ErrorUnknownInputType1.New(nil, confraw["type"])
			return
		}

		inj := inject.New()
		inj.SetParent(t)
		inj.Map(&confraw)
		inj.Map(inchan)
		refvs, err := injectutil.Invoke(inj, handler)
		if err != nil {
			err = ErrorRunInput1.New(err, confraw)
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
