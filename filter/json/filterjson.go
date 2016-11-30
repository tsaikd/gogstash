package filterjson

import (
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
	"encoding/json"
	"time"
)

const (
	ModuleName = "json"
)

type FilterConfig struct {
	config.FilterConfig
	Msgfield   string `json:"message"`
	Tsfield string `json:"timestamp"`
	Tsformat string `json:"timeformat"`
}

func DefaultFilterConfig() FilterConfig {
	return FilterConfig{
		FilterConfig: config.FilterConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
	}
}

func InitHandler(confraw *config.ConfigRaw) (retconf config.TypeFilterConfig, err error) {
	conf := DefaultFilterConfig()
	if err = config.ReflectConfig(confraw, &conf); err != nil {
		return
	}

	retconf = &conf
	return
}

func (f *FilterConfig) Event(event logevent.LogEvent) logevent.LogEvent {

	var parsedMessage map[string]interface{}
	if err := json.Unmarshal([]byte(event.Message), &parsedMessage); err != nil {
		return event
	}

	if event.Extra == nil {
		event.Extra = make(map[string]interface{})
	}

	for key, value := range parsedMessage {
		switch key {
		case f.Msgfield:
			event.Message = value.(string)
		case f.Tsfield:
			if ts, err := time.Parse(f.Tsformat, value.(string)); err == nil {
				event.Timestamp = ts
			}
		default:
			event.Extra[key] = value
		}
	}

	return event
}

