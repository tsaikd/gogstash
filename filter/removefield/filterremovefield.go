package filterremovefield

import (
	"strings"

	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
)

// ModuleName is the name used in config file
const ModuleName = "remove_field"

// FilterConfig holds the configuration json fields and internal objects
type FilterConfig struct {
	config.FilterConfig

	Fields []string `json:"fields"`
}

// DefaultFilterConfig returns an FilterConfig struct with default values
func DefaultFilterConfig() FilterConfig {
	return FilterConfig{
		FilterConfig: config.FilterConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		Fields: []string{},
	}
}

// InitHandler initialize the filter plugin
func InitHandler(confraw *config.ConfigRaw) (retconf config.TypeFilterConfig, err error) {
	conf := DefaultFilterConfig()
	if err = config.ReflectConfig(confraw, &conf); err != nil {
		return
	}

	if len(conf.Fields) < 1 {
		config.Logger.Warn("filter remove_field config empty fields")
	}

	retconf = &conf
	return
}

// Event the main filter event
func (f *FilterConfig) Event(event logevent.LogEvent) logevent.LogEvent {
	if event.Extra == nil {
		event.Extra = map[string]interface{}{}
	}

	for _, field := range f.Fields {
		removeField(event.Extra, field)
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
