package config

import (
	"context"

	"github.com/tsaikd/KDGoLib/errutil"
	"github.com/tsaikd/gogstash/config/logevent"
)

// errors
var (
	ErrorUnknownInputType1 = errutil.NewFactory("unknown input config type: %q")
	ErrorInitInputFailed1  = errutil.NewFactory("initialize input module failed: %v")
)

// TypeInputConfig is interface of input module
type TypeInputConfig interface {
	TypeCommonConfig
	Start(ctx context.Context, msgChan chan<- logevent.LogEvent) (err error)
}

// InputConfig is basic input config struct
type InputConfig struct {
	CommonConfig
	Codec TypeCodecConfig `json:"-"`
}

// InputHandler is a handler to regist input module
type InputHandler func(ctx context.Context, raw *ConfigRaw) (TypeInputConfig, error)

var (
	mapInputHandler = map[string]InputHandler{}
)

// RegistInputHandler regist a input handler
func RegistInputHandler(name string, handler InputHandler) {
	mapInputHandler[name] = handler
}

func (t *Config) getInputs() (inputs []TypeInputConfig, err error) {
	var input TypeInputConfig
	for _, raw := range t.InputRaw {
		handler, ok := mapInputHandler[raw["type"].(string)]
		if !ok {
			return inputs, ErrorUnknownInputType1.New(nil, raw["type"])
		}

		if input, err = handler(t.ctx, &raw); err != nil {
			return inputs, ErrorInitInputFailed1.New(err, raw)
		}

		inputs = append(inputs, input)
	}
	return
}

func (t *Config) startInputs() (err error) {
	inputs, err := t.getInputs()
	if err != nil {
		return
	}

	for _, input := range inputs {
		func(input TypeInputConfig) {
			t.eg.Go(func() error {
				return input.Start(t.ctx, t.chInFilter)
			})
		}(input)
	}

	return
}
