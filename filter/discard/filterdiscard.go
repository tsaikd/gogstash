package discard

import (
	"context"
	"errors"
	"github.com/Knetic/govaluate"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
	filtercond "github.com/tsaikd/gogstash/filter/cond"
	"sync/atomic"
)

const (
	ModuleName = "discard" // the name used in config file
	ErrorTag   = "discard_error"
)

// FilterConfig holds the configuration json fields and internal objects
type FilterConfig struct {
	config.FilterConfig
	// filters data
	control     config.Control
	ctx         context.Context
	expressions []*govaluate.EvaluableExpression
	hasPressure uint64 // 0=no pressure, any other value is pressure - access with atomic package
	// configuration
	DiscardIfBackpressure bool     `json:"discard_if_backpressure" yaml:"discard_if_backpressure"` // if true an event will always be discarded if there is some backpressure
	Conditions            []string `yaml:"conditions" json:"conditions"`                           // our conditions
	NegateConditions      bool     `yaml:"negate_conditions" json:"negate_conditions"`             // if true we will discard if we otherwise would not
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
func InitHandler(
	ctx context.Context,
	raw config.ConfigRaw,
	control config.Control,
) (config.TypeFilterConfig, error) {
	conf := DefaultFilterConfig()
	conf.control = control
	conf.ctx = ctx

	if err := config.ReflectConfig(raw, &conf); err != nil {
		return nil, err
	}

	// parse conditions
	for _, condition := range conf.Conditions {
		cond, err := govaluate.NewEvaluableExpressionWithFunctions(condition, filtercond.BuiltInFunctions)
		if err == nil {
			conf.expressions = append(conf.expressions, cond)
		} else {
			goglog.Logger.Errorf("%s: %s", ModuleName, err.Error())
		}
	}
	if len(conf.expressions) == 0 && !conf.DiscardIfBackpressure {
		return nil, errors.New("no conditions for discard filter")
	}

	if conf.DiscardIfBackpressure {
		go conf.backgroundtask()
	}
	return &conf, nil
}

// Event the main filter event
func (f *FilterConfig) Event(
	ctx context.Context,
	event logevent.LogEvent,
) (logevent.LogEvent, bool) {
	// check for backpressure drop
	if f.DiscardIfBackpressure && atomic.LoadUint64(&f.hasPressure) > 0 {
		event.FilterPos = logevent.DiscardEvent
		return event, false
	}
	// handle conditions
	var shouldDiscard bool
	ep := filtercond.EventParameters{Event: &event}
	for num := range f.expressions {
		ret, err := f.expressions[num].Eval(&ep)
		if err == nil {
			if r, ok := ret.(bool); ok {
				shouldDiscard = r
			} else {
				goglog.Logger.Warn("filter cond condition returns not a boolean, ignored")
			}
		} else {
			goglog.Logger.Errorf("%s: %s", ModuleName, err.Error())
		}
		if shouldDiscard {
			break
		}
	}
	if f.NegateConditions {
		shouldDiscard = !shouldDiscard
	}
	if shouldDiscard {
		event.FilterPos = logevent.DiscardEvent
		return event, false
	}
	return event, true
}

// backgroundtask monitors backpressure and updates the internal status. Is only started if we need to monitor backpressure.
func (f *FilterConfig) backgroundtask() {
	for {
		select {
		case <-f.control.PauseSignal():
			atomic.AddUint64(&f.hasPressure, 1)
		case <-f.control.ResumeSignal():
			atomic.StoreUint64(&f.hasPressure, 0)
		case <-f.ctx.Done():
			return
		}
	}
}
