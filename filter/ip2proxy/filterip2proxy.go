package filterip2proxy

import (
	"context"
	"github.com/fsnotify/fsnotify"
	"github.com/ip2location/ip2proxy-go"
	"net"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
	ip2l "github.com/tsaikd/gogstash/filter/ip2location"
)

// ModuleName is the name used in config file
const ModuleName = "ip2proxy"

// ErrorTag tag added to event when process ip2location failed
const ErrorTag = "gogstash_filter_ip2location_error"

// FilterConfig holds the configuration json fields and internal objects
type FilterConfig struct {
	config.FilterConfig

	DBPath      string   `json:"db_path" yaml:"db_path"`           // ip2location .BIN file
	IPField     string   `json:"ip_field" yaml:"ip_field"`         // IP field to get geoip info
	Key         string   `json:"key"`                              // geoip destination field name, default: geoip
	QuietFail   bool     `json:"quiet" yaml:"quiet"`               // fail quietly
	SkipPrivate bool     `json:"skip_private" yaml:"skip_private"` // skip private IP addresses
	PrivateNet  []string `json:"private_net" yaml:"private_net"`   // list of own defined private IP addresses
	CacheSize   int      `json:"cache_size" yaml:"cache_size"`     // cache size

	db    *ip2proxy.DB
	dbMtx sync.RWMutex

	cache        *lru.Cache
	privateCIDRs []*net.IPNet

	watcher *fsnotify.Watcher
	ctx     context.Context
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

// InitHandler initialize the filter plugin
func InitHandler(ctx context.Context, raw *config.ConfigRaw) (config.TypeFilterConfig, error) {
	conf := DefaultFilterConfig()
	err := config.ReflectConfig(raw, &conf)
	if err != nil {
		return nil, err
	}

	conf.db, err = ip2proxy.OpenDB(conf.DBPath)
	if err != nil {
		return nil, err
	}

	conf.cache, err = lru.New(conf.CacheSize)
	if err != nil {
		return nil, err
	}

	conf.ctx = ctx

	var cidrs []string
	if len(conf.PrivateNet) > 0 {
		cidrs = conf.PrivateNet
	} else {
		cidrs = ip2l.DefaultCIDR
	}

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

	for _, cidr := range cidrs {
		_, privateCIDR, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, err
		}
		conf.privateCIDRs = append(conf.privateCIDRs, privateCIDR)
	}

	return &conf, nil
}

// not_supported is copied from ip2proxy and is the field entry for each field that is not supported by the current database
const not_supported string = "NOT SUPPORTED"

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
	var record map[string]string
	// single-thread here
	if c, ok := f.cache.Get(ipstr); ok {
		record = c.(map[string]string)
	} else {
		var err error
		f.dbMtx.RLock()
		record, err = f.db.GetAll(ipstr)
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
	m := make(map[string]string)
	for k, v := range record {
		if v != not_supported {
			m[k] = v
		}
	}
	event.SetValue(f.Key, m)
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
	newDb, err := ip2proxy.OpenDB(fc.DBPath)
	if err != nil {
		goglog.Logger.Errorf("%s failed to update %s: %s", ModuleName, fc.DBPath, err.Error())
		return
	}
	oldDb := fc.db
	fc.dbMtx.Lock()
	fc.db = newDb
	fc.dbMtx.Unlock()
	oldDb.Close()
	fc.cache.Purge()
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
			case <-fc.watcher.Events:
				timer.Reset(pauseDelay)
			case err := <-fc.watcher.Errors:
				goglog.Logger.Errorf("%s: %s", ModuleName, err.Error())
			case <-fc.ctx.Done():
				return
			}
		}
	}()
}
