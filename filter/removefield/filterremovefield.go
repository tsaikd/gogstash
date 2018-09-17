package filterremovefield

import (
	"context"
	"strings"

	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
)

// ModuleName is the name used in config file
const ModuleName = "remove_field"

// FilterConfig holds the configuration json fields and internal objects
type FilterConfig struct {
	config.FilterConfig

	// list all fields to remove
	Fields []string `json:"fields"`
	// remove event origin message field, not in extra
	RemoveMessage bool `json:"remove_message"`
}

// DefaultFilterConfig returns an FilterConfig struct with default values
func DefaultFilterConfig() FilterConfig {
	return FilterConfig{
		FilterConfig: config.FilterConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		Fields:        []string{},
		RemoveMessage: false,
	}
}

// InitHandler initialize the filter plugin
func InitHandler(ctx context.Context, raw *config.ConfigRaw) (config.TypeFilterConfig, error) {
	conf := DefaultFilterConfig()
	if err := config.ReflectConfig(raw, &conf); err != nil {
		return nil, err
	}

	if len(conf.Fields) < 1 {
		goglog.Logger.Warn("filter remove_field config empty fields")
	}

	return &conf, nil
}

// Event the main filter event
func (f *FilterConfig) Event(ctx context.Context, event logevent.LogEvent) logevent.LogEvent {
	if event.Extra == nil {
		event.Extra = map[string]interface{}{}
	}

	for _, field := range f.Fields {
		removeField(event.Extra, field)
	}

	if f.RemoveMessage {
		event.Message = ""
	}

	return event
}

func removeField(obj map[string]interface{}, field string) {
	fieldSplits := strings.Split(field, ".")
	if len(fieldSplits) < 2 {
		delete(obj, field)
		return
	}

	switch child := obj[fieldSplits[0]].(type) {
	case map[string]interface{}:
		removeField(child, strings.Join(fieldSplits[1:], "."))
	}
}
