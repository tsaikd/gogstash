package config

import (
	"github.com/codegangsta/inject"
	"github.com/tsaikd/KDGoLib/errutil"
	"github.com/tsaikd/KDGoLib/injectutil"
	"github.com/tsaikd/gogstash/config/logevent"
)

// errors
var (
	ErrorUnknownFilterType = errutil.NewFactory("unknown filter config type: %q")
	ErrorInitFilter        = errutil.NewFactory("filter initialization failed: %q")
)

type TypeFilterConfig interface {
	TypeConfig
	Event(logevent.LogEvent) logevent.LogEvent
}

type FilterConfig struct {
	CommonConfig
}

type FilterHandler interface{}

var (
	mapFilterHandler = map[string]FilterHandler{}
)

func RegistFilterHandler(name string, handler FilterHandler) {
	mapFilterHandler[name] = handler
}

func (t *Config) RunFilters() (err error) {
	return t.InvokeSimple(t.runFilters)
}

func (c *Config) runFilters(inchan InChan, outchan OutChan) (err error) {
	filters, err := c.getFilters()
	if err != nil {
		return
	}

	go func() {
		for {
			select {
			case event := <-inchan:
				for _, filter := range filters {
					event = filter.Event(event)
				}
				outchan <- event
			}
		}
	}()
	return
}

func (c *Config) getFilters() (filters []TypeFilterConfig, err error) {
	for _, confraw := range c.FilterRaw {
		handler, ok := mapFilterHandler[confraw["type"].(string)]
		if !ok {
			err = ErrorUnknownFilterType.New(nil, confraw["type"])
			return
		}

		inj := inject.New()
		inj.SetParent(c)
		inj.Map(&confraw)
		refvs, err := injectutil.Invoke(inj, handler)
		if err != nil {
			err = ErrorInitFilter.New(err, confraw)
			return []TypeFilterConfig{}, err
		}

		for _, refv := range refvs {
			if !refv.CanInterface() {
				continue
			}
			if conf, ok := refv.Interface().(TypeFilterConfig); ok {
				conf.SetInjector(inj)
				filters = append(filters, conf)
			}
		}
	}
	return
}
