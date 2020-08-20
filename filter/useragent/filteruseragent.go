package filteruseragent

import (
	"context"

	lru "github.com/hashicorp/golang-lru"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
	"github.com/ua-parser/uap-go/uaparser"
)

// ModuleName is the name used in config file
const ModuleName = "useragent"

type uaFields struct {
	Name    string
	OS      string
	OSName  string
	OSMajor string
	OSMinor string
	Device  string
	Major   string
	Minor   string
	Patch   string
	Build   string
}

// Init user agent fields
func (f *uaFields) Init(target string) {
	if target != "" {
		target = target + "."
	}
	f.Name = target + "name"
	f.OS = target + "os"
	f.OSName = target + "os_name"
	f.OSMajor = target + "os_major"
	f.OSMinor = target + "os_minor"
	f.Device = target + "device"
	f.Major = target + "major"
	f.Minor = target + "minor"
	f.Patch = target + "patch"
	f.Build = target + "build"
}

// FilterConfig holds the configuration json fields and internal objects
type FilterConfig struct {
	config.FilterConfig

	// The field containing the user agent string.
	Source string `json:"source"`

	// The name of the field to assign user agent data into.
	// If not specified user agent data will be stored in the root of the event.
	Target string `json:"target"`

	// `regexes.yaml` file to use
	//
	// You can find the latest version of this here:
	// <https://github.com/ua-parser/uap-core/blob/master/regexes.yaml>
	Regexes string `json:"regexes"`

	// UA parsing is surprisingly expensive.
	// We can optimize it by adding a cache
	CacheSize int `json:"cache_size"`

	fields uaFields
	parser *uaparser.Parser
	cache  *lru.Cache
}

// DefaultFilterConfig returns an FilterConfig struct with default values
func DefaultFilterConfig() FilterConfig {
	return FilterConfig{
		FilterConfig: config.FilterConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		Target:    "",
		CacheSize: 100000,
	}
}

// InitHandler initialize the filter plugin
func InitHandler(ctx context.Context, raw *config.ConfigRaw) (config.TypeFilterConfig, error) {
	conf := DefaultFilterConfig()
	err := config.ReflectConfig(raw, &conf)
	if err != nil {
		return nil, err
	}

	if conf.Regexes == "" {
		conf.parser = uaparser.NewFromSaved()
	} else {
		conf.parser, err = uaparser.New(conf.Regexes)
		if err != nil {
			return nil, err
		}
	}

	conf.fields.Init(conf.Target)
	conf.cache, err = lru.New(conf.CacheSize)
	if err != nil {
		return nil, err
	}

	return &conf, nil
}

// Event the main filter event
func (f *FilterConfig) Event(ctx context.Context, event logevent.LogEvent) (logevent.LogEvent, bool) {
	ua := event.GetString(f.Source)
	if ua == "" {
		return event, false
	}
	var client *uaparser.Client
	// single-thread here
	if c, ok := f.cache.Get(ua); ok {
		client = c.(*uaparser.Client)
	} else {
		client = f.parser.Parse(ua)
		f.cache.Add(ua, client)
	}
	if client.Os != nil {
		event.SetValue(f.fields.OS, client.Os.Family)
		if client.Os.Family != "" {
			event.SetValue(f.fields.OSName, client.Os.Family)
		}
		if client.Os.Major != "" {
			event.SetValue(f.fields.OSMajor, client.Os.Major)
		}
		if client.Os.Minor != "" {
			event.SetValue(f.fields.OSMinor, client.Os.Minor)
		}
	}
	if client.Device != nil {
		event.SetValue(f.fields.Device, client.Device.Family)
	}
	if client.UserAgent != nil {
		event.SetValue(f.fields.Name, client.UserAgent.Family)
		if client.UserAgent.Major != "" {
			event.SetValue(f.fields.Major, client.UserAgent.Major)
		}
		if client.UserAgent.Minor != "" {
			event.SetValue(f.fields.Minor, client.UserAgent.Minor)
		}
		if client.UserAgent.Patch != "" {
			event.SetValue(f.fields.Patch, client.UserAgent.Patch)
		}
	}
	return event, true
}
