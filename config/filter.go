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
	Event(context.Context, logevent.LogEvent) (logevent.LogEvent, bool)
	CommonFilter(context.Context, logevent.LogEvent) logevent.LogEvent
}

// IsConfigured returns whether common configuration has been setup
func (f *FilterConfig) IsConfigured() bool {
	return len(f.AddTags) != 0 || len(f.AddFields) != 0 || len(f.RemoveTags) != 0 || len(f.RemoveFields) != 0
}

func (f *FilterConfig) CommonFilter(ctx context.Context, event logevent.LogEvent) logevent.LogEvent {

	event.AddTag(f.AddTags...)
	event.RemoveTag(f.RemoveTags...)
	for _, field := range f.RemoveFields {
		event.Remove(field)
	}
	for _, f := range f.AddFields {
		event.SetValue(f.Key, event.Format(f.Value))
	}
	return event
}

// FilterConfig is basic filter config struct
type FilterConfig struct {
	CommonConfig
	AddTags      []string      `yaml:"add_tag" json:"add_tag"`
	RemoveTags   []string      `yaml:"remove_tag" json:"remove_tag"`
	AddFields    []FieldConfig `yaml:"add_field" json:"add_field"`
	RemoveFields []string      `yaml:"remove_field" json:"remove_field"`
}

// FieldConfig is a name/value field config
type FieldConfig struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
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

// GetFilters get filters from config
func GetFilters(ctx context.Context, filterRaw []ConfigRaw) (filters []TypeFilterConfig, err error) {
	var filter TypeFilterConfig
	for _, raw := range filterRaw {
		handler, ok := mapFilterHandler[raw["type"].(string)]
		if !ok {
			return filters, ErrorUnknownFilterType1.New(nil, raw["type"])
		}

		if filter, err = handler(ctx, &raw); err != nil {
			return filters, ErrorInitFilterFailed1.New(err, raw)
		}

		filters = append(filters, filter)
	}
	return
}

func (t *Config) getFilters() (filters []TypeFilterConfig, err error) {
	return GetFilters(t.ctx, t.FilterRaw)
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
				if len(t.chInFilter) < 1 {
					return nil
				}
			case event := <-t.chInFilter:
				var ok bool
				for _, filter := range filters {
					event, ok = filter.Event(t.ctx, event)
					if ok {
						event = filter.CommonFilter(t.ctx, event)
					}
				}
				t.chFilterOut <- event
			}
		}
	})

	return
}
