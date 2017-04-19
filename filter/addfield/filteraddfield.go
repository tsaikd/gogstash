package filteraddfield

import (
	"context"

	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
)

// ModuleName is the name used in config file
const ModuleName = "add_field"

// FilterConfig holds the configuration json fields and internal objects
type FilterConfig struct {
	config.FilterConfig
	Key   string `json:"key"`
	Value string `json:"value"`
}

// DefaultFilterConfig returns an FilterConfig struct with default values
func DefaultFilterConfig() FilterConfig {
	return FilterConfig{
		FilterConfig: config.FilterConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
	}
}

// InitHandler initialize the filter plugin
func InitHandler(ctx context.Context, raw *config.ConfigRaw) (config.TypeFilterConfig, error) {
	conf := DefaultFilterConfig()
	if err := config.ReflectConfig(raw, &conf); err != nil {
		return nil, err
	}

	return &conf, nil
}

// Event the main filter event
func (f *FilterConfig) Event(ctx context.Context, event logevent.LogEvent) logevent.LogEvent {
	if _, ok := event.Extra[f.Key]; ok {
		return event
	}
	if event.Extra == nil {
		event.Extra = make(map[string]interface{})
	}
	event.Extra[f.Key] = event.Format(f.Value)
	return event
}
