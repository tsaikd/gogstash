package filterratelimit

import (
	"context"
	"time"

	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
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
func InitHandler(ctx context.Context, raw *config.ConfigRaw) (config.TypeFilterConfig, error) {
	conf := DefaultFilterConfig()
	if err := config.ReflectConfig(raw, &conf); err != nil {
		return nil, err
	}

	if conf.Rate <= 0 {
		goglog.Logger.Warn("filter ratelimit config rate should > 0, ignored")
		return &conf, nil
	}

	conf.throttle = make(chan time.Time, conf.Burst)
	tick := time.NewTicker(time.Second / time.Duration(conf.Rate))

	go func() {
		defer tick.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case t := <-tick.C:
				select {
				case <-ctx.Done():
					return
				case conf.throttle <- t:
				default:
				}
			}
		}
	}()

	return &conf, nil
}

// Event the main filter event
func (f *FilterConfig) Event(ctx context.Context, event logevent.LogEvent) (logevent.LogEvent, bool) {
	if event.Extra == nil {
		event.Extra = map[string]interface{}{}
	}

	if f.throttle == nil {
		return event, false
	}

	<-f.throttle
	return event, true
}
