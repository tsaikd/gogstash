package outputstdout

import (
	"context"
	"fmt"

	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
)

// ModuleName is the name used in config file
const ModuleName = "stdout"

// OutputConfig holds the configuration json fields and internal objects
type OutputConfig struct {
	config.OutputConfig

	msg   chan []byte            // channel to push message from codec to
	codec config.TypeCodecConfig // the codec we will use
	ctx   context.Context
}

// DefaultOutputConfig returns an OutputConfig struct with default values
func DefaultOutputConfig() OutputConfig {
	return OutputConfig{
		OutputConfig: config.OutputConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		msg: make(chan []byte),
	}
}

// InitHandler initialize the output plugin
func InitHandler(ctx context.Context, raw config.ConfigRaw) (config.TypeOutputConfig, error) {
	conf := DefaultOutputConfig()
	err := config.ReflectConfig(raw, &conf)
	if err != nil {
		return nil, err
	}

	conf.ctx = ctx
	conf.codec, err = config.GetCodecOrDefault(ctx, raw["codec"])
	if err != nil {
		return nil, err
	}

	go conf.backgroundtask()

	return &conf, nil
}

// backgroundtask receives messages and prints them to stdout
func (t *OutputConfig) backgroundtask() {
	for {
		select {
		case <-t.ctx.Done():
			return
		case msg := <-t.msg:
			fmt.Println(string(msg))
		}
	}
}

// Output event
func (t *OutputConfig) Output(ctx context.Context, event logevent.LogEvent) (err error) {
	_, err = t.codec.Encode(ctx, event, t.msg)

	return
}
