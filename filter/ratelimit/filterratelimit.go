package filterratelimit

import (
	"time"

	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
)

// ModuleName is the name used in config file
const ModuleName = "rate_limit"

// FilterConfig holds the configuration json fields and internal objects
type FilterConfig struct {
	config.FilterConfig

	Rate  int64 `json:"rate"`  // event number per second
	Burst int64 `json:"burst"` // burst limit

	throttle chan time.Time
}

// DefaultFilterConfig returns an FilterConfig struct with default values
func DefaultFilterConfig() FilterConfig {
	return FilterConfig{
		FilterConfig: config.FilterConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		Burst: 100,
	}
}

// InitHandler initialize the filter plugin
func InitHandler(confraw *config.ConfigRaw) (retconf config.TypeFilterConfig, err error) {
	conf := DefaultFilterConfig()
	if err = config.ReflectConfig(confraw, &conf); err != nil {
		return
	}

	if conf.Rate <= 0 {
		config.Logger.Warn("filter ratelimit config rate should > 0, ignored")
	} else {
		conf.throttle = make(chan time.Time, conf.Burst)
		tick := time.NewTicker(time.Second / time.Duration(conf.Rate))
		// no tick.Stop() will not be called, this will leak, but gogstash will not reuse InitHandler

		go func() {
			for t := range tick.C {
				select {
				case conf.throttle <- t:
				default:
				}
			} // exits after tick.Stop()
		}()
	}

	retconf = &conf
	return
}

// Event the main filter event
func (f *FilterConfig) Event(event logevent.LogEvent) logevent.LogEvent {
	if event.Extra == nil {
		event.Extra = map[string]interface{}{}
	}

	if f.throttle == nil {
		return event
	}

	<-f.throttle
	return event
}
