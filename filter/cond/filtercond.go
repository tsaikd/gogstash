package filtercond

import (
	"context"
	"strings"

	"github.com/Knetic/govaluate"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
)

// ModuleName is the name used in config file
const ModuleName = "cond"

// ErrorTag tag added to event when process geoip2 failed
const ErrorTag = "gogstash_filter_cond_error"

// FilterConfig holds the configuration json fields and internal objects
type FilterConfig struct {
	config.FilterConfig

	Condition  string             `json:"condition"` // condition need to be satisfied
	FilterRaw  []config.ConfigRaw `json:"filter"`    // filters when satisfy the condition
	filters    []config.TypeFilterConfig
	expression *govaluate.EvaluableExpression
}

// EventParameters pack event's parameters by member function `Get` access
type EventParameters struct {
	Event *logevent.LogEvent
}

// Get obtaining value from event's specified field recursively
func (ep *EventParameters) Get(field string) (interface{}, error) {
	if strings.IndexRune(field, '.') < 0 {
		// no nest fields
		return ep.Event.Get(field), nil
	}
	return config.GetFromObject(ep.Event.Extra, field), nil
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
	err := config.ReflectConfig(raw, &conf)
	if err != nil {
		return nil, err
	}
	if conf.Condition == "" {
		config.Logger.Warn("filter cond config condition empty, ignored")
		return &conf, nil
	}
	conf.filters, err = config.GetFilters(ctx, conf.FilterRaw)
	if err != nil {
		return nil, err
	}
	if len(conf.filters) <= 0 {
		config.Logger.Warn("filter cond config filters empty, ignored")
		return &conf, nil
	}
	conf.expression, err = govaluate.NewEvaluableExpression(conf.Condition)
	return &conf, nil
}

// Event the main filter event
func (f *FilterConfig) Event(ctx context.Context, event logevent.LogEvent) logevent.LogEvent {
	if f.expression != nil {
		ep := EventParameters{Event: &event}
		ret, err := f.expression.Eval(&ep)
		if err != nil {
			config.Logger.Error(err)
			event.AddTag(ErrorTag)
			return event
		}
		if ok, _ := ret.(bool); ok {
			for _, filter := range f.filters {
				event = filter.Event(ctx, event)
			}
		}
	}
	return event
}
