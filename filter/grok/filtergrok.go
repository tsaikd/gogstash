package filtergrok

import (
	"context"

	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
	"github.com/vjeantet/grok"
)

// ModuleName is the name used in config file
const ModuleName = "grok"

// ErrorTag tag added to event when process module failed
const ErrorTag = "gogstash_filter_grok_error"

// FilterConfig holds the configuration json fields and internal objects
type FilterConfig struct {
	config.FilterConfig

	PatternsPath      string            `json:"patterns_path"`       // path to patterns file
	Patterns          map[string]string `json:"patterns"`            // pattern definitions
	Match             []string          `json:"match"`               // match pattern
	Source            string            `json:"source"`              // source message field name
	RemoveEmptyValues bool              `json:"remove_empty_values"` // remove empty values

	grk *grok.Grok
}

// DefaultFilterConfig returns an FilterConfig struct with default values
func DefaultFilterConfig() FilterConfig {
	return FilterConfig{
		FilterConfig: config.FilterConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		PatternsPath:      "",
		Patterns:          nil,
		Match:             []string{"%{COMMONAPACHELOG}"},
		Source:            "message",
		RemoveEmptyValues: true,
	}
}

// InitHandler initialize the filter plugin
func InitHandler(ctx context.Context, raw *config.ConfigRaw) (config.TypeFilterConfig, error) {
	conf := DefaultFilterConfig()
	err := config.ReflectConfig(raw, &conf)
	if err != nil {
		return nil, err
	}

	g, err := grok.NewWithConfig(&grok.Config{
		NamedCapturesOnly: true,
		RemoveEmptyValues: conf.RemoveEmptyValues,
	})
	if err != nil {
		return nil, err
	}
	if conf.PatternsPath != "" {
		g.AddPatternsFromPath(conf.PatternsPath)
	}
	if conf.Patterns != nil {
		g.AddPatternsFromMap(conf.Patterns)
	}

	conf.grk = g

	return &conf, nil
}

// Event the main filter event
func (f *FilterConfig) Event(ctx context.Context, event logevent.LogEvent) logevent.LogEvent {
	message := event.GetString(f.Source)
	found := false
	for _, thisMatch := range f.Match {
		// grok Parse will success even it doesn't match
		values, err := f.grk.ParseTyped(thisMatch, message)
		if err == nil && len(values) > 0 {
			found = true
			for key, value := range values {
				switch v := value.(type) {
				case string:
					event.SetValue(key, event.Format(v))
				case nil:
					// pass
				default:
					event.SetValue(key, value)
				}
			}
			break
		}
	}

	if !found {
		event.AddTag(ErrorTag)
		goglog.Logger.Errorf("grok: no matches for %q", message)
	}

	return event
}
