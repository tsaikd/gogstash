package config

import (
	"context"

	"github.com/tsaikd/KDGoLib/errutil"
	"github.com/tsaikd/gogstash/config/logevent"
)

// errors
var (
	ErrorUnknownFilterType1 = errutil.NewFactory("unknown filter config type: %q")
	ErrorInitFilterFailed1  = errutil.NewFactory("initialize filter module failed: %v")
)

// TypeFilterConfig is interface of filter module
type TypeFilterConfig interface {
	TypeCommonConfig
	Event(context.Context, logevent.LogEvent) logevent.LogEvent
}

// FilterConfig is basic filter config struct
type FilterConfig struct {
	CommonConfig
}

// FilterHandler is a handler to regist filter module
type FilterHandler func(ctx context.Context, raw *ConfigRaw) (TypeFilterConfig, error)

var (
	mapFilterHandler = map[string]FilterHandler{}
)

// RegistFilterHandler regist a filter handler
func RegistFilterHandler(name string, handler FilterHandler) {
	mapFilterHandler[name] = handler
}

func (t *Config) getFilters() (filters []TypeFilterConfig, err error) {
	var filter TypeFilterConfig
	for _, raw := range t.FilterRaw {
		handler, ok := mapFilterHandler[raw["type"].(string)]
		if !ok {
			return filters, ErrorUnknownFilterType1.New(nil, raw["type"])
		}

		if filter, err = handler(t.ctx, &raw); err != nil {
			return filters, ErrorInitFilterFailed1.New(err, raw)
		}

		filters = append(filters, filter)
	}
	return
}

func (t *Config) startFilters() (err error) {
	filters, err := t.getFilters()
	if err != nil {
		return
	}

	t.eg.Go(func() error {
		for {
			select {
			case <-t.ctx.Done():
				return nil
			case event := <-t.chInFilter:
				for _, filter := range filters {
					event = filter.Event(t.ctx, event)
				}
				t.chFilterOut <- event
			}
		}
	})

	return
}
