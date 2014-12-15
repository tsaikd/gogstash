package outputstdout

import (
	"encoding/json"
	"fmt"

	log "github.com/Sirupsen/logrus"

	"github.com/tsaikd/gogstash/config"
)

type OutputConfig struct {
	config.CommonConfig
}

func DefaultOutputConfig() OutputConfig {
	return OutputConfig{
		CommonConfig: config.CommonConfig{
			Type: "stdout",
		},
	}
}

func init() {
	config.RegistOutputHandler("stdout", func(mapraw map[string]interface{}) (conf config.TypeOutputConfig, err error) {
		var (
			raw []byte
		)
		if raw, err = json.Marshal(mapraw); err != nil {
			log.Error(err)
			return
		}
		defconf := DefaultOutputConfig()
		conf = &defconf
		if err = json.Unmarshal(raw, &conf); err != nil {
			log.Error(err)
			return
		}
		return
	})
}

func (self *OutputConfig) Type() string {
	return self.CommonConfig.Type
}

func (self *OutputConfig) Event(event config.LogEvent) (err error) {
	var (
		raw []byte
	)
	if raw, err = event.MarshalIndent(); err != nil {
		return
	}
	fmt.Print(string(raw))
	return
}
