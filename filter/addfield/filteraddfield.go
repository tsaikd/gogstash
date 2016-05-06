package filteraddfield

import (
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
)

const (
	ModuleName = "add_field"
)

type FilterConfig struct {
	config.FilterConfig
	Key   string `json:"key"`
	Value string `json:"value"`
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
	if _, ok := event.Extra[f.Key]; ok {
		return event
	}
	if event.Extra == nil {
		event.Extra = make(map[string]interface{})
	}
	event.Extra[f.Key] = event.Format(f.Value)
	return event
}
