package filtergeoip2

import (
	"context"
	"github.com/fsnotify/fsnotify"
	"net"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru"
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

	DBPath      string   `json:"db_path" yaml:"db_path"`           // geoip2 db file path, default: GeoLite2-City.mmdb
	IPField     string   `json:"ip_field" yaml:"ip_field"`         // IP field to get geoip info
	Key         string   `json:"key" yaml:"key"`                   // geoip destination field name, default: geoip
	QuietFail   bool     `json:"quiet" yaml:"quiet"`               // fail quietly
	SkipPrivate bool     `json:"skip_private" yaml:"skip_private"` // skip private IP addresses
	PrivateNet  []string `json:"private_net" yaml:"private_net"`   // list of own defined private IP addresses
	FlatFormat  bool     `json:"flat_format" yaml:"flat_format"`   // flat format
	CacheSize   int      `json:"cache_size" yaml:"cache_size"`     // cache size

	db      *geoip2.Reader
	dbMtx   sync.RWMutex
	watcher *fsnotify.Watcher
	ctx     context.Context

	cache        *lru.Cache
	privateCIDRs []*net.IPNet
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
		CacheSize:   100000,
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
	conf.cache, err = lru.New(conf.CacheSize)
	if err != nil {
		return nil, err
	}

	conf.ctx = ctx

	// init fsnotify
	goglog.Logger.Infof("%s fsnotify initialized for %s", ModuleName, conf.DBPath)
	conf.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		goglog.Logger.Errorf("%s failed to init watcher, %s", ModuleName, err.Error())
	}
	err = conf.watcher.Add(conf.DBPath)
	if err != nil {
		goglog.Logger.Errorf("%s failed to add file: %s", ModuleName, err.Error())
	}
	conf.initFsnotifyEventHandler()

	cidrs := []string{
		"10.0.0.0/8",
		"100.64.0.0/10",
		"127.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"fc00::/7",
		"fe80::/10",
		"169.254.0.0/16",
	}
	if len(conf.PrivateNet) > 0 {
		cidrs = conf.PrivateNet
	}
	for _, cidr := range cidrs {
		_, privateCIDR, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, err
		}
		conf.privateCIDRs = append(conf.privateCIDRs, privateCIDR)
	}

	return &conf, nil
}

// Event the main filter event
func (f *FilterConfig) Event(ctx context.Context, event logevent.LogEvent) (logevent.LogEvent, bool) {
	ipstr := event.GetString(f.IPField)
	if ipstr == "" {
		// Passthru if empty
		return event, false
	}
	ip := net.ParseIP(ipstr)
	if ip == nil || (f.SkipPrivate && f.privateIP(ip)) {
		// Passthru
		return event, false
	}
	var err error
	var record *geoip2.City
	// single-thread here
	if c, ok := f.cache.Get(ipstr); ok {
		record = c.(*geoip2.City)
	} else {
		f.dbMtx.RLock()
		record, err = f.db.City(ip)
		f.dbMtx.RUnlock()
		if err != nil {
			if !f.QuietFail {
				goglog.Logger.Error(err)
			}
			event.AddTag(ErrorTag)
			return event, false
		}
		f.cache.Add(ipstr, record)
	}
	if record == nil {
		event.AddTag(ErrorTag)
		return event, false
	}
	if record.Location.Latitude == 0 && record.Location.Longitude == 0 {
		event.AddTag(ErrorTag)
		return event, false
	}

	if f.FlatFormat {
		m := map[string]interface{}{
			"continent_code": record.Continent.Code,
			"country_code":   record.Country.IsoCode,
			"country_name":   record.Country.Names["en"],
			"ip":             ipstr,
			"latitude":       record.Location.Latitude,
			"location":       []float64{record.Location.Longitude, record.Location.Latitude},
			"longitude":      record.Location.Longitude,
		}
		if record.City.Names != nil {
			m["city_name"] = record.City.Names["en"]
		}
		if record.Postal.Code != "" {
			m["postal_code"] = record.Postal.Code
		}
		if len(record.Subdivisions) > 0 {
			m["region_code"] = record.Subdivisions[0].IsoCode
			m["region_name"] = record.Subdivisions[0].Names["en"]
		}
		if record.Location.TimeZone != "" {
			m["timezone"] = record.Location.TimeZone
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

	return event, true
}

func (f *FilterConfig) privateIP(ip net.IP) bool {
	for _, cidr := range f.privateCIDRs {
		if cidr.Contains(ip) {
			return true
		}
	}
	return false
}

// reloadFile reloads a new file from disk and invalidates the cache
func (fc *FilterConfig) reloadFile() {
	newDb, err := geoip2.Open(fc.DBPath)
	if err != nil {
		goglog.Logger.Errorf("%s failed to update %s: %s", ModuleName, fc.DBPath, err.Error())
		return
	}
	oldDb := fc.db
	fc.dbMtx.Lock()
	fc.db = newDb
	fc.dbMtx.Unlock()
	fc.cache.Purge()
	oldDb.Close()
	goglog.Logger.Infof("%s reloaded file %s", ModuleName, fc.DBPath)
}

// initFsnotifyEventHandler is called by InitHandler and sets up a background thread that watches for changes
func (fc *FilterConfig) initFsnotifyEventHandler() {
	const pauseDelay = 5 * time.Second // used to let all changes be completed before reloading the file
	go func() {
		timer := time.NewTimer(0)
		defer timer.Stop()
		firstTime := true
		for {
			select {
			case <-timer.C:
				if firstTime {
					firstTime = false
				} else {
					fc.reloadFile()
				}
			case e := <-fc.watcher.Events:
				// geoip2 issues a chmod on itself so we need to filter those out
				if e.Op.String() != "CHMOD" {
					timer.Reset(pauseDelay)
				}
			case err := <-fc.watcher.Errors:
				goglog.Logger.Errorf("%s: %s", ModuleName, err.Error())
			case <-fc.ctx.Done():
				return
			}
		}
	}()
}
