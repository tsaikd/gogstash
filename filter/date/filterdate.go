package filterdate

import (
	"time"

	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
)

// ModuleName is the name used in config file
const ModuleName = "date"

// FilterConfig holds the configuration json fields and internal objects
type FilterConfig struct {
	config.FilterConfig

	Format string `json:"format"` // date parse format
	Source string `json:"source"` // source message field name
}

// DefaultFilterConfig returns an FilterConfig struct with default values
func DefaultFilterConfig() FilterConfig {
	return FilterConfig{
		FilterConfig: config.FilterConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		Format: time.RFC3339Nano,
		Source: "message",
	}
}

// InitHandler initialize the filter plugin
func InitHandler(confraw *config.ConfigRaw) (retconf config.TypeFilterConfig, err error) {
	conf := DefaultFilterConfig()
	if err = config.ReflectConfig(confraw, &conf); err != nil {
		return
	}

	retconf = &conf
	return
}

// Event the main filter event
func (f *FilterConfig) Event(event logevent.LogEvent) logevent.LogEvent {
	if event.Extra == nil {
		event.Extra = map[string]interface{}{}
	}

	timestamp, err := time.Parse(f.Format, event.GetString(f.Source))
	if err != nil {
		config.Logger.Error(err)
		return event
	}

	event.Timestamp = timestamp.UTC()
	return event
}
