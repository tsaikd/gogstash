package filtertypeconv

import (
	"context"
	"fmt"
	"strconv"

	"github.com/tsaikd/KDGoLib/errutil"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
)

// ModuleName is the name used in config file
const ModuleName = "typeconv"

// ErrorTag tag added to event when process typeconv failed
const ErrorTag = "gogstash_filter_typeconv_error"

// Errors
var (
	ErrorInvalidConvType1 = errutil.NewFactory(`%q is not one of ["string", "int64", "float64"]`)
)

// FilterConfig holds the configuration json fields and internal objects
type FilterConfig struct {
	config.FilterConfig

	ConvType string   `json:"conv_type"` // one of ["string", "int64", "float64"]
	Fields   []string `json:"fields"`    // fields to convert type
}

const convTypeString = "string"
const convTypeInt64 = "int64"
const convTypeFloat64 = "float64"

// DefaultFilterConfig returns an FilterConfig struct with default values
func DefaultFilterConfig() FilterConfig {
	return FilterConfig{
		FilterConfig: config.FilterConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		ConvType: "string",
	}
}

// InitHandler initialize the filter plugin
func InitHandler(ctx context.Context, raw config.ConfigRaw) (config.TypeFilterConfig, error) {
	conf := DefaultFilterConfig()
	if err := config.ReflectConfig(raw, &conf); err != nil {
		return nil, err
	}

	switch conf.ConvType {
	case convTypeString, convTypeInt64, convTypeFloat64:
	default:
		return nil, ErrorInvalidConvType1.New(nil, conf.ConvType)
	}

	return &conf, nil
}

// Event the main filter event
func (f *FilterConfig) Event(ctx context.Context, event logevent.LogEvent) (logevent.LogEvent, bool) {
	for _, field := range f.Fields {
		if value, ok := event.GetValue(field); ok {
			switch f.ConvType {
			case convTypeString:
				switch v := value.(type) {
				case string:
				default:
					event.SetValue(field, fmt.Sprintf("%v", v))
				}
			case convTypeInt64:
				switch v := value.(type) {
				case string:
					if vparse, err := strconv.ParseInt(v, 0, 64); err == nil {
						event.SetValue(field, vparse)
					} else if vparse, err := strconv.ParseFloat(fmt.Sprintf("%v", v), 64); err == nil {
						event.SetValue(field, int64(vparse))
					} else {
						goglog.Logger.Error(err)
						event.AddTag(ErrorTag)
					}
				case int:
					event.SetValue(field, int64(v))
				case int8:
					event.SetValue(field, int64(v))
				case int16:
					event.SetValue(field, int64(v))
				case int32:
					event.SetValue(field, int64(v))
				case int64:
				case float32:
					event.SetValue(field, int64(v))
				case float64:
					event.SetValue(field, int64(v))
				default:
					if vparse, err := strconv.ParseInt(fmt.Sprintf("%v", v), 0, 64); err == nil {
						event.SetValue(field, vparse)
					} else if vparse, err := strconv.ParseFloat(fmt.Sprintf("%v", v), 64); err == nil {
						event.SetValue(field, int64(vparse))
					} else {
						goglog.Logger.Error(err)
						event.AddTag(ErrorTag)
					}
				}
			case convTypeFloat64:
				switch v := value.(type) {
				case string:
					if vparse, err := strconv.ParseFloat(v, 64); err == nil {
						event.SetValue(field, vparse)
					} else {
						goglog.Logger.Error(err)
						event.AddTag(ErrorTag)
					}
				case int:
					event.SetValue(field, float64(v))
				case int8:
					event.SetValue(field, float64(v))
				case int16:
					event.SetValue(field, float64(v))
				case int32:
					event.SetValue(field, float64(v))
				case int64:
					event.SetValue(field, float64(v))
				case float32:
					event.SetValue(field, float64(v))
				case float64:
				default:
					if vparse, err := strconv.ParseFloat(fmt.Sprintf("%v", v), 64); err == nil {
						event.SetValue(field, vparse)
					} else {
						goglog.Logger.Error(err)
						event.AddTag(ErrorTag)
					}
				}
			}
		}
	}

	// TODO: no converts return false
	return event, true
}
