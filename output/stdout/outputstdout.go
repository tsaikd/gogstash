package outputstdout

import (
	"fmt"

	"github.com/tsaikd/gogstash/config"
)

const (
	ModuleName = "stdout"
)

type OutputConfig struct {
	config.CommonConfig
}

func DefaultOutputConfig() OutputConfig {
	return OutputConfig{
		CommonConfig: config.CommonConfig{
			Type: ModuleName,
		},
	}
}

func init() {
	config.RegistOutputHandler(ModuleName, func(mapraw map[string]interface{}) (retconf config.TypeOutputConfig, err error) {
		conf := DefaultOutputConfig()
		if err = config.ReflectConfig(mapraw, &conf); err != nil {
			return
		}

		retconf = &conf
		return
	})
}

func (t *OutputConfig) Event(event config.LogEvent) (err error) {
	raw, err := event.MarshalIndent()
	if err != nil {
		return
	}

	fmt.Println(string(raw))
	return
}
