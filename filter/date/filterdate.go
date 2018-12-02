package filterdate

import (
	"context"
	"time"

	"github.com/tengattack/jodatime"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
)

// ModuleName is the name used in config file
const ModuleName = "date"

// FilterConfig holds the configuration json fields and internal objects
type FilterConfig struct {
	config.FilterConfig

	Format       []string `json:"format"`         // date parse format
	Source       string   `json:"source"`         // source message field name
	Joda         bool     `json:"joda"`           // whether using joda time format
	TagOnFailure []string `json:"tag_on_failure"` // tags to append on failure
	Target       string   `json:"target"`         // target field

	timeParser func(layout, value string) (time.Time, error)
}

// DefaultFilterConfig returns an FilterConfig struct with default values
func DefaultFilterConfig() FilterConfig {
	return FilterConfig{
		FilterConfig: config.FilterConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		Format:       []string{time.RFC3339Nano},
		Source:       "message",
		Target:       "@timestamp",
		TagOnFailure: []string{"gogstash_filter_date_error", "_dateparsefailure"},
	}
}

// InitHandler initialize the filter plugin
func InitHandler(ctx context.Context, raw *config.ConfigRaw) (config.TypeFilterConfig, error) {
	conf := DefaultFilterConfig()
	if err := config.ReflectConfig(raw, &conf); err != nil {
		return nil, err
	}

	if conf.Joda {
		conf.timeParser = jodatime.Parse
	} else {
		conf.timeParser = time.Parse
	}

	return &conf, nil
}

// Event the main filter event
func (f *FilterConfig) Event(ctx context.Context, event logevent.LogEvent) logevent.LogEvent {
	var (
		timestamp time.Time
		err       error
	)
	for _, thisFormat := range f.Format {
		timestamp, err = f.timeParser(thisFormat, event.GetString(f.Source))
		if err == nil {
			break
		}
	}

	if err != nil {
		for _, t := range f.TagOnFailure {
			event.AddTag(t)
		}
		goglog.Logger.Error(err)
		return event
	}

	switch f.Target {
	case "@timestamp":
		event.Timestamp = timestamp.UTC()
		break
	default:
		event.SetValue(f.Target, timestamp.UTC())
		break
	}
	return event
}
