package filterjson

import (
	"context"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
)

// ModuleName is the name used in config file
const ModuleName = "json"

// ErrorTag tag added to event when process module failed
const ErrorTag = "gogstash_filter_json_error"

// FilterConfig holds the configuration json fields and internal objects
type FilterConfig struct {
	config.FilterConfig
	Msgfield  string `json:"message"`
	Appendkey string `json:"appendkey"`
	Tsfield   string `json:"timestamp"`
	Tsformat  string `json:"timeformat"`
	Source    string `json:"source"`
	AddTag            []string          `json:"add_tag"`             // tags to add on filter success
	TagOnFailure      []string          `json:"tag_on_failure"`      // tags to add on filter failure
}

// DefaultFilterConfig returns an FilterConfig struct with default values
func DefaultFilterConfig() FilterConfig {
	return FilterConfig{
		FilterConfig: config.FilterConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		Source:"message",
		TagOnFailure: []string{ErrorTag},
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
	source := event.GetString(f.Source)
	if err := jsoniter.Unmarshal([]byte(source), &parsedMessage); err != nil {
		event.AddTag(f.TagOnFailure...)
		goglog.Logger.Errorf("json error when reading %v %v\n", source, err)
		return event
	}

	event.AddTag(f.AddTag...)
	if f.Appendkey != "" {
		event.SetValue(f.Appendkey, parsedMessage)
	} else {
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
			case logevent.TagsField:
				event.ParseTags(value)
			default:
				event.Extra[key] = value
			}
		}
	}

	return event
}
