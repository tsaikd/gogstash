package filtergeoip2

import (
	"context"
	"net"

	geoip2 "github.com/oschwald/geoip2-golang"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
)

// ModuleName is the name used in config file
const ModuleName = "geoip2"

// ErrorTag tag added to event when process geoip2 failed
const ErrorTag = "gogstash_filter_geoip2_error"

// FilterConfig holds the configuration json fields and internal objects
type FilterConfig struct {
	config.FilterConfig

	DBPath      string `json:"db_path"`      // geoip2 db file path, default: GeoLite2-City.mmdb
	IPField     string `json:"ip_field"`     // IP field to get geoip info
	Key         string `json:"key"`          // geoip destination field name, default: geoip
	QuietFail   bool   `json:"quiet"`        // fail quietly
	SkipPrivate bool   `json:"skip_private"` // skip private IP addresses
	FlatFormat  bool   `json:"flat_format"`  // flat format

	db *geoip2.Reader
}

// DefaultFilterConfig returns an FilterConfig struct with default values
func DefaultFilterConfig() FilterConfig {
	return FilterConfig{
		FilterConfig: config.FilterConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		DBPath:      "GeoLite2-City.mmdb",
		Key:         "geoip",
		QuietFail:   false, // backwards compatible
		SkipPrivate: false,
		FlatFormat:  false,
	}
}

// InitHandler initialize the filter plugin
func InitHandler(ctx context.Context, raw *config.ConfigRaw) (config.TypeFilterConfig, error) {
	conf := DefaultFilterConfig()
	err := config.ReflectConfig(raw, &conf)
	if err != nil {
		return nil, err
	}

	conf.db, err = geoip2.Open(conf.DBPath)
	if err != nil {
		return nil, err
	}

	return &conf, nil
}

// Event the main filter event
func (f *FilterConfig) Event(ctx context.Context, event logevent.LogEvent) logevent.LogEvent {
	ipstr := event.GetString(f.IPField)
	if ipstr == "" {
		// Passthru if empty
		return event
	}
	ip := net.ParseIP(ipstr)
	if f.SkipPrivate && f.privateIP(ip) {
		// Passthru
		return event
	}
	record, err := f.db.City(ip)
	if err != nil {
		if f.QuietFail {
			goglog.Logger.Error(err)
		}
		event.AddTag(ErrorTag)
		return event
	}
	if record == nil {
		event.AddTag(ErrorTag)
		return event
	}
	if record.Location.Latitude == 0 && record.Location.Longitude == 0 {
		event.AddTag(ErrorTag)
		return event
	}

	if f.FlatFormat {
		m := map[string]interface{}{
			"city_name":      record.City.Names["en"],
			"continent_code": record.Continent.Code,
			"country_code":   record.Country.IsoCode,
			"country_name":   record.Country.Names["en"],
			"ip":             ipstr,
			"latitude":       record.Location.Latitude,
			"location":       []float64{record.Location.Longitude, record.Location.Latitude},
			"longitude":      record.Location.Longitude,
			"postal_code":    record.Postal.Code,
			"timezone":       record.Location.TimeZone,
		}
		if len(record.Subdivisions) > 0 {
			m["region_code"] = record.Subdivisions[0].IsoCode
			m["region_name"] = record.Subdivisions[0].Names["en"]
		}
		event.SetValue(f.Key, m)
	} else {
		m := map[string]interface{}{
			"city": map[string]interface{}{
				"name": record.City.Names["en"],
			},
			"continent": map[string]interface{}{
				"code": record.Continent.Code,
				"name": record.Continent.Names["en"],
			},
			"country": map[string]interface{}{
				"code": record.Country.IsoCode,
				"name": record.Country.Names["en"],
			},
			"ip":        ipstr,
			"latitude":  record.Location.Latitude,
			"location":  []float64{record.Location.Longitude, record.Location.Latitude},
			"longitude": record.Location.Longitude,
			"timezone":  record.Location.TimeZone,
		}
		if len(record.Subdivisions) > 0 {
			m["region"] = map[string]interface{}{
				"code": record.Subdivisions[0].IsoCode,
				"name": record.Subdivisions[0].Names["en"],
			}
		}
		event.SetValue(f.Key, m)
	}

	return event
}

func (f *FilterConfig) privateIP(ip net.IP) bool {
	_, private24BitBlock, _ := net.ParseCIDR("10.0.0.0/8")
	_, private20BitBlock, _ := net.ParseCIDR("172.16.0.0/12")
	_, private16BitBlock, _ := net.ParseCIDR("192.168.0.0/16")
	return private24BitBlock.Contains(ip) || private20BitBlock.Contains(ip) || private16BitBlock.Contains(ip)
}
