package filtercond

import (
	"context"
	"math/rand"
	"reflect"
	"strings"

	"github.com/Knetic/govaluate"
	"github.com/tsaikd/KDGoLib/errutil"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
)

// ModuleName is the name used in config file
const ModuleName = "cond"

// ErrorTag tag added to event when process geoip2 failed
const ErrorTag = "gogstash_filter_cond_error"

// built-in functions
var (
	ErrorBuiltInFunctionParameters1 = errutil.NewFactory("Built-in function '%s' parameters error")
	BuiltInFunctions                = map[string]govaluate.ExpressionFunction{
		"empty": func(args ...interface{}) (interface{}, error) {
			if len(args) > 1 {
				return nil, ErrorBuiltInFunctionParameters1.New(nil, "empty")
			} else if len(args) == 0 {
				return true, nil
			}
			return args[0] == nil, nil
		},
		"strlen": func(args ...interface{}) (interface{}, error) {
			if len(args) > 1 {
				return nil, ErrorBuiltInFunctionParameters1.New(nil, "strlen")
			} else if len(args) == 0 {
				return float64(0), nil
			}
			length := len(args[0].(string))
			return (float64)(length), nil
		},
		"map": func(args ...interface{}) (interface{}, error) {
			if len(args) > 1 {
				return nil, ErrorBuiltInFunctionParameters1.New(nil, "map")
			} else if len(args) == 0 {
				return []interface{}{}, nil
			}

			s := reflect.ValueOf(args[0])
			if s.Kind() != reflect.Slice {
				return nil, ErrorBuiltInFunctionParameters1.New(nil, "map")
			}

			ret := make([]interface{}, s.Len())

			for i := 0; i < s.Len(); i++ {
				ret[i] = s.Index(i).Interface()
			}

			return ret, nil
		},
		"rand": func(args ...interface{}) (interface{}, error) {
			if len(args) > 0 {
				return nil, ErrorBuiltInFunctionParameters1.New(nil, "rand")
			}
			return rand.Float64(), nil
		},
	}
)

// FilterConfig holds the configuration json fields and internal objects
type FilterConfig struct {
	config.FilterConfig

	Condition     string             `json:"condition"`   // condition need to be satisfied
	FilterRaw     []config.ConfigRaw `json:"filter"`      // filters when satisfy the condition
	ElseFilterRaw []config.ConfigRaw `json:"else_filter"` // filters when does not met the condition
	filters       []config.TypeFilterConfig
	elseFilters   []config.TypeFilterConfig
	expression    *govaluate.EvaluableExpression
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
	v, _ := ep.Event.GetValue(field)
	return v, nil
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
		goglog.Logger.Warn("filter cond config condition empty, ignored")
		return &conf, nil
	}
	conf.filters, err = config.GetFilters(ctx, conf.FilterRaw)
	if err != nil {
		return nil, err
	}
	if len(conf.filters) <= 0 {
		goglog.Logger.Warn("filter cond config filters empty, ignored")
		return &conf, nil
	}
	if len(conf.ElseFilterRaw) > 0 {
		conf.elseFilters, err = config.GetFilters(ctx, conf.ElseFilterRaw)
		if err != nil {
			return nil, err
		}
	}
	conf.expression, err = govaluate.NewEvaluableExpressionWithFunctions(conf.Condition, BuiltInFunctions)
	return &conf, nil
}

// Event the main filter event
func (f *FilterConfig) Event(ctx context.Context, event logevent.LogEvent) logevent.LogEvent {
	if f.expression != nil {
		ep := EventParameters{Event: &event}
		ret, err := f.expression.Eval(&ep)
		if err != nil {
			goglog.Logger.Error(err)
			event.AddTag(ErrorTag)
			return event
		}
		if r, ok := ret.(bool); ok {
			if r {
				for _, filter := range f.filters {
					event = filter.Event(ctx, event)
				}
			} else {
				for _, filter := range f.elseFilters {
					event = filter.Event(ctx, event)
				}
			}
		} else {
			goglog.Logger.Warn("filter cond condition returns not a boolean, ignored")
		}
	}
	return event
}
