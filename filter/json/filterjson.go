package filterjson

import (
	"context"
	"encoding/json"
	"time"

	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
)

// ModuleName is the name used in config file
const ModuleName = "json"

// ErrorTag tag added to event when process module failed
const ErrorTag = "gogstash_filter_json_error"

// FilterConfig holds the configuration json fields and internal objects
type FilterConfig struct {
	config.FilterConfig
	Msgfield string `json:"message"`
	Tsfield  string `json:"timestamp"`
	Tsformat string `json:"timeformat"`
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
	var parsedMessage map[string]interface{}
	if err := json.Unmarshal([]byte(event.Message), &parsedMessage); err != nil {
		event.AddTag(ErrorTag)
		config.Logger.Error(err)
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
