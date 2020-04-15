package filterurlparam

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
)

// ModuleName is the name used in config file
const ModuleName = "url_param"

// FilterConfig holds the configuration json fields and internal objects
type FilterConfig struct {
	config.FilterConfig

	// The field containing the url params string.
	Source string `json:"source"`

	// Include param keys, "all_fields" or "*" include all fields
	IncludeKeys []string `json:"include_keys"`
	includeAll  bool

	// url_decode params, "all_fields" or "*" decode all params values
	UrlDecode []string `json:"url_decode"`
	decodeAll bool

	// prefix for param name, default: request_url_args_
	Prefix string `json:"prefix"`

	RemoveEmptyValues bool `json:"remove_empty_values"`
}

// DefaultFilterConfig returns an FilterConfig struct with default values
func DefaultFilterConfig() FilterConfig {
	return FilterConfig{
		FilterConfig: config.FilterConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		Source:            "request_url",
		IncludeKeys:       []string{"*"},
		UrlDecode:         []string{"*"},
		Prefix:            "request_url_args_",
		RemoveEmptyValues: true,
	}
}

// InitHandler initialize the filter plugin
func InitHandler(ctx context.Context, raw *config.ConfigRaw) (config.TypeFilterConfig, error) {
	conf := DefaultFilterConfig()
	err := config.ReflectConfig(raw, &conf)
	if err != nil {
		return nil, err
	}

	if len(conf.IncludeKeys) == 0 {
		goglog.Logger.Fatalf("include_keys is empty configure.")
	}

	if conf.IncludeKeys[0] == "*" {
		conf.includeAll = true
	} else {
		conf.includeAll = false
	}

	if len(conf.UrlDecode) > 0 && conf.UrlDecode[0] == "*" {
		conf.decodeAll = true
	} else {
		conf.decodeAll = false
	}

	if conf.Prefix != "" {
		if strings.Contains(conf.Prefix, ".") {
			return nil, fmt.Errorf("prefix can not includ dot(\".\")")
		}
	}

	return &conf, nil
}

// Event the main filter event
func (f *FilterConfig) Event(ctx context.Context, event logevent.LogEvent) (logevent.LogEvent, bool) {
	message := event.GetString(f.Source)
	var (
		u      *url.URL
		params url.Values
		err    error
	)
	if u, err = url.Parse(message); err == nil {
		// parse http://domain/path?params, /path?params, ?params
		params = u.Query()
	} else {
		// params string
		params, err = url.ParseQuery(message)
	}
	if err != nil {
		goglog.Logger.Errorf("parse param failed, %s, %v", message, err)
		return event, false
	}

	//url decode
	if f.decodeAll {
		for k, v := range params {
			if nv, err := url.PathUnescape(v[0]); err != nil {
				params.Set(k, nv)
			}
		}
	} else {
		for _, k := range f.UrlDecode {
			if v := params.Get(k); v != "" {
				if nv, err := url.PathUnescape(v); err != nil {
					params.Set(k, nv)
				}
			}
		}
	}

	// add key
	if f.includeAll {
		for k, v := range params {
			k := f.Prefix + k
			event.SetValue(k, v[0])
		}
	} else {
		for _, k := range f.IncludeKeys {
			if v := params.Get(k); !f.RemoveEmptyValues || v != "" {
				k := f.Prefix + k
				event.SetValue(k, v)
			}
		}
	}
	return event, true
}
