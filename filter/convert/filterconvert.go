package filterconvert

import (
	"context"
	"errors"
	"strconv"

	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
)

const (
	// ModuleName is the name used in config file
	ModuleName = "convert"
	// ErrorTag tag added to event when process module failed
	ErrorTag = "gogstash_filter_convert_error"
)

// errors
var (
	ErrNotConfigured = errors.New("filter convert not configured")
)

// FilterConfig holds the configuration json fields and internal objects
type FilterConfig struct {
	config.FilterConfig

	To_int   [2]string `yaml:"to_int"`
	To_float [2]string `yaml:"to_float"`
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
	err := config.ReflectConfig(raw, &conf)
	if err != nil {
		return nil, err
	}

	if conf.To_int[0] == "" && conf.To_float[0] == "" && !conf.IsConfigured() {
		return nil, ErrNotConfigured
	}

	return &conf, nil
}

// Event the main filter event
func (f *FilterConfig) Event(ctx context.Context, event logevent.LogEvent) (logevent.LogEvent, bool) {
	if f.To_int[0] != "" {
		if intVar, err := strconv.ParseInt(event.GetString(f.To_int[0]), 10, 64); err == nil {
			if factorVar, errf := strconv.ParseInt(f.To_int[1], 10, 64); errf == nil {
				event.SetValue(f.To_int[0], intVar*factorVar)
			} else {
				event.SetValue(f.To_int[0], intVar)
			}
		}
	}
	if f.To_float[0] != "" {
		if floatVar, err := strconv.ParseFloat(event.GetString(f.To_float[0]), 64); err == nil {
			if factorVar, errf := strconv.ParseFloat(f.To_float[1], 64); errf == nil {
				event.SetValue(f.To_float[0], floatVar*factorVar)
			} else {
				event.SetValue(f.To_float[0], floatVar)
			}
		}
	}

	// always return true here for configured filter
	return event, true
}

func mergeField(event logevent.LogEvent, destinationName, source string) logevent.LogEvent {
	destinationValue := event.Get(destinationName)
	value := event.Format(source)
	if destinationValue == nil {
		destinationValue = []string{value}
		event.SetValue(destinationName, destinationValue)
		return event
	}
	switch currentDestination := destinationValue.(type) {
	case string:
		var newDestination []string
		if currentDestination != "" {
			newDestination = append(newDestination, currentDestination)
		}
		newDestination = append(newDestination, value)
		event.SetValue(destinationName, newDestination)
	case []string:
		currentDestination = append(currentDestination, value)
		event.SetValue(destinationName, currentDestination)
	default:
		goglog.Logger.Warnf("convert: destination field %s is not string nor []string", destinationName)
		event.AddTag(ErrorTag)
	}
	return event
}
