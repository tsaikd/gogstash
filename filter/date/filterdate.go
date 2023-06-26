package filterdate

import (
	"context"
	"math"
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

// DefaultTarget default event field to store date as
const DefaultTarget = "@timestamp"

// FilterConfig holds the configuration json fields and internal objects
type FilterConfig struct {
	config.FilterConfig

	Format               []string `json:"format"`                  // date parse format
	Source               string   `json:"source"`                  // source message field name
	Joda                 bool     `json:"joda"`                    // whether using joda time format
	Target               string   `json:"target"`                  // target field where date should be stored
	ComputeYearIfMissing bool     `json:"compute_year_if_missing"` // try to find missing year; dont support logs older than 1 year

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
		Format:               []string{time.RFC3339Nano},
		Source:               "message",
		Target:               DefaultTarget,
		ComputeYearIfMissing: false,
	}
}

// InitHandler initialize the filter plugin
func InitHandler(
	ctx context.Context,
	raw config.ConfigRaw,
	control config.Control,
) (config.TypeFilterConfig, error) {
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
func (f *FilterConfig) Event(ctx context.Context, event logevent.LogEvent) (logevent.LogEvent, bool) {
	var (
		timestamp time.Time
		err       error
	)
	for _, thisFormat := range f.Format {
		switch thisFormat {
		case "UNIX":
			var sec, nsec int64
			value := event.Get(f.Source)
			switch value := value.(type) {
			case float64:
				sec, nsec = convertFloat(value)
			default:
				sec, nsec, err = convert(event.GetString(f.Source))
			}
			if err != nil {
				continue
			}
			timestamp = time.Unix(sec, nsec)
		case "UNIXNANO":
			var nsec int64
			value := event.Get(f.Source)
			switch v := value.(type) {
			case int64:
				nsec = v
			case int:
				nsec = int64(v)
			case string:
				nsec, err = strconv.ParseInt(v, 10, 64)
			default:
				continue
			}
			if err != nil {
				continue
			}
			timestamp = time.Unix(0, nsec)
		default:
			timestamp, err = f.timeParser(thisFormat, event.GetString(f.Source))
			if timestamp.Year() == 0 && f.ComputeYearIfMissing {
				now := time.Now()

				timestamptest := time.Date(now.Year(), timestamp.Month(), timestamp.Day(), timestamp.Hour(), timestamp.Minute(), timestamp.Second(), timestamp.Nanosecond(), timestamp.Location())
				if timestamptest.After(now) {
					timestamp = time.Date(now.Year()-1, timestamp.Month(), timestamp.Day(), timestamp.Hour(), timestamp.Minute(), timestamp.Second(), timestamp.Nanosecond(), timestamp.Location())
				} else {
					timestamp = timestamptest
				}

				err = nil
			}
		}
		if err == nil {
			break
		}
	}

	if err != nil {
		event.AddTag(ErrorTag)
		goglog.Logger.Error(err)
		return event, false
	}
	if f.Target == DefaultTarget {
		event.Timestamp = timestamp.UTC()
	} else {
		event.SetValue(f.Target, timestamp.UTC())
	}

	return event, true
}

func convertFloat(value float64) (sec int64, nsec int64) {
	sec = int64(value)
	rounded := value - float64(sec)
	nsec = int64(rounded * 1000000000)
	return sec, nsec
}

func convert(s string) (sec int64, nsec int64, err error) {
	dot := strings.Index(s, ".")

	if indexOfe := strings.Index(s, "e"); dot == 1 && indexOfe != -1 {
		// looks like exponential notation
		result, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0, 0, err
		}
		sec = int64(result)
		rounded := math.Round((result - float64(sec)) * 1000)
		nsec = int64(rounded)
		return sec, nsec, nil
	}
	if dot >= 0 {
		sec, err = strconv.ParseInt(s[:dot], 10, 64)
		if err != nil {
			return 0, 0, err
		}
		// fraction to nano seconds, avoid precision loss
		nsec, err = strconv.ParseInt(s[dot+1:], 10, 64)
		if err != nil {
			return 0, 0, err
		}
		nsec *= exponent(10, 9-(len(s)-dot-1))
	} else {
		sec, err = strconv.ParseInt(s, 10, 64)
		if err != nil {
			return 0, 0, err
		}
	}
	return sec, nsec, nil
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
