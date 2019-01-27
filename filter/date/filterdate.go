package filterdate

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/tengattack/jodatime"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
)

// ModuleName is the name used in config file
const ModuleName = "date"

// ErrorTag tag added to event when process module failed
const ErrorTag = "gogstash_filter_date_error"

// FilterConfig holds the configuration json fields and internal objects
type FilterConfig struct {
	config.FilterConfig

	Format []string `json:"format"` // date parse format
	Source string   `json:"source"` // source message field name
	Joda   bool     `json:"joda"`   // whether using joda time format

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
		Format: []string{time.RFC3339Nano},
		Source: "message",
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
		if thisFormat == "UNIX" {
			var sec, nsec int64
			s := event.GetString(f.Source)
			dot := strings.Index(s, ".")
			if dot >= 0 {
				sec, err = strconv.ParseInt(s[:dot], 10, 64)
				if err != nil {
					continue
				}
				// fraction to nano seconds, avoid precision loss
				nsec, err = strconv.ParseInt(s[dot+1:], 10, 64)
				if err != nil {
					continue
				}
				nsec *= exponent(10, 9-(len(s)-dot-1))
			} else {
				sec, err = strconv.ParseInt(s, 10, 64)
				if err != nil {
					continue
				}
			}
			timestamp = time.Unix(sec, nsec)
		} else {
			timestamp, err = f.timeParser(thisFormat, event.GetString(f.Source))
		}
		if err == nil {
			break
		}
	}

	if err != nil {
		event.AddTag(ErrorTag)
		goglog.Logger.Error(err)
		return event
	}

	event.Timestamp = timestamp.UTC()
	return event
}

func exponent(a int64, n int) int64 {
	result := int64(1)
	for i := n; i > 0; i >>= 1 {
		if i&1 != 0 {
			result *= a
		}
		a *= a
	}
	return result
}
