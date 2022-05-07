package filterip2location

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	lru "github.com/hashicorp/golang-lru"
	"github.com/ip2location/ip2location-go/v9"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
)

// ModuleName is the name used in config file
const ModuleName = "ip2location"

// ErrorTag tag added to event when process ip2location failed
const ErrorTag = "gogstash_filter_ip2location_error"

// FilterConfig holds the configuration json fields and internal objects
type FilterConfig struct {
	config.FilterConfig

	DBPath      string   `json:"db_path" yaml:"db_path"`           // ip2location .BIN file
	IPField     string   `json:"ip_field" yaml:"ip_field"`         // IP field to get geolocation for
	Key         string   `json:"key"`                              // destination field name
	QuietFail   bool     `json:"quiet" yaml:"quiet"`               // fail quietly
	SkipPrivate bool     `json:"skip_private" yaml:"skip_private"` // skip private IP addresses
	PrivateNet  []string `json:"private_net" yaml:"private_net"`   // list of own defined private IP addresses
	CacheSize   int      `json:"cache_size" yaml:"cache_size"`     // cache size

	dbMtx sync.RWMutex    // lock for db
	db    *ip2location.DB // database

	cache        *lru.Cache
	privateCIDRs []*net.IPNet

	watcher *fsnotify.Watcher
	ctx     context.Context
}

// DefaultCIDR is the default list of CIDRs to filter out
var DefaultCIDR = []string{
	"10.0.0.0/8",
	"100.64.0.0/10",
	"127.0.0.0/8",
	"172.16.0.0/12",
	"192.168.0.0/16",
	"fc00::/7",
	"fe80::/10",
	"169.254.0.0/16",
}

// DefaultFilterConfig returns an FilterConfig struct with default values
func DefaultFilterConfig() FilterConfig {
	return FilterConfig{
		FilterConfig: config.FilterConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		Key:         ModuleName,
		QuietFail:   false, // backwards compatible
		SkipPrivate: false,
		CacheSize:   100000,
	}
}

// reloadFile reloads the file from disk invalidates the cache
func (fc *FilterConfig) reloadFile() {
	newDb, err := ip2location.OpenDB(fc.DBPath)
	if err != nil {
		goglog.Logger.Errorf("ip2location failed to update %s: %s", fc.DBPath, err.Error())
		return
	}
	oldDb := fc.db
	fc.dbMtx.Lock()
	fc.db = newDb
	fc.dbMtx.Unlock()
	oldDb.Close()
	fc.cache.Purge()
	goglog.Logger.Infof("ip2location reloaded file %s", fc.DBPath)
}

// initFsnotifyEventHandler is called by InitHandler and sets up a background thread that watches for changes
func (fc *FilterConfig) initFsnotifyEventHandler() {
	const pauseDelay = 5 * time.Second // used to let all changes be done before reloading the file
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
			case <-fc.watcher.Events:
				timer.Reset(pauseDelay)
			case err := <-fc.watcher.Errors:
				goglog.Logger.Errorf("ip2location: %s", err.Error())
			case <-fc.ctx.Done():
				return
			}
		}
	}()
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

	conf.ctx = ctx

	// open database
	conf.db, err = ip2location.OpenDB(conf.DBPath)
	if err != nil {
		return nil, err
	}

	// init cache
	conf.cache, err = lru.New(conf.CacheSize)
	if err != nil {
		return nil, err
	}

	goglog.Logger.Info("ip2location fsnotify initialized for", conf.DBPath)

	// init fsnotify
	conf.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		goglog.Logger.Errorf("ip2location failed to init watcher, %s", err.Error())
	}
	err = conf.watcher.Add(conf.DBPath)
	if err != nil {
		goglog.Logger.Errorf("ip2location failed to add file: %s", err.Error())
	}
	conf.initFsnotifyEventHandler()

	var cidrs []string
	if len(conf.PrivateNet) > 0 {
		cidrs = conf.PrivateNet
	} else {
		cidrs = DefaultCIDR
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

// not_supported is copied from ip2location and is the field entry for each field that is not supported by the current database
const not_supported string = "This parameter is unavailable for selected data file. Please upgrade the data file."

// Event the main filter event
func (f *FilterConfig) Event(ctx context.Context, event logevent.LogEvent) (logevent.LogEvent, bool) {
	ipstr := event.GetString(f.IPField)
	if ipstr == "" {
		// Passthru if empty
		return event, false
	}
	ip := net.ParseIP(ipstr)
	if ip == nil {
		return event, false
	}
	if f.SkipPrivate && f.privateIP(ip) {
		// Passthru
		return event, false
	}
	var record *ip2location.IP2Locationrecord
	if c, ok := f.cache.Get(ipstr); ok {
		record = c.(*ip2location.IP2Locationrecord)
	} else {
		f.dbMtx.RLock()
		r2, err := f.db.Get_all(ipstr)
		f.dbMtx.RUnlock()
		record = &r2
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

	m := map[string]interface{}{
		"country_code": record.Country_short,
		"country_name": record.Country_long,
	}
	if record.City != not_supported {
		m["city_name"] = record.City
		m["region_name"] = record.Region
	}
	if record.Isp != not_supported {
		m["ISP"] = record.Isp
	}
	if record.Latitude != 0 || record.Longitude != 0 {
		location := make(map[string]interface{})
		location["lon"] = record.Longitude
		location["lat"] = record.Latitude
		m["location"] = location
	}

	event.SetValue(f.Key, m)
	return event, true
}

// privateIP returns true if ip is in list of private IP addresses
func (f *FilterConfig) privateIP(ip net.IP) bool {
	for _, cidr := range f.privateCIDRs {
		if cidr.Contains(ip) {
			return true
		}
	}
	return false
}
