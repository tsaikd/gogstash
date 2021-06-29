package outputstatsd

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	stringutils "github.com/msaf1980/go-stringutils"
	statsd "github.com/msaf1980/statsd"
	"github.com/tsaikd/KDGoLib/errutil"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
)

// ModuleName is the name used in config file
const ModuleName = "statsd"

// ErrorTag tag added to event when process module failed
const ErrorTag = "gogstash_output_statsd_error"

var (
	ErrorPingFailed = errutil.NewFactory("ping statsd server failed")
)

type NameValue struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type templatePair struct {
	name  stringutils.Template
	value stringutils.Template
}

type statsdPool struct {
	last    int64
	mu      sync.Mutex
	used    int32
	clients map[string]*statsd.Client
}

func (sp *statsdPool) Client(
	host, proto string, prefix string,
	timeout time.Duration, flushInterval time.Duration,
) (*statsd.Client, error) {

	var (
		s   *statsd.Client
		err error
		ok  bool
	)
	key := proto + ":" + host
	sp.mu.Lock()
	defer sp.mu.Unlock()
	if s, ok = sp.clients[key]; !ok {
		if s, err = statsd.New(
			statsd.Network(proto),
			statsd.Address(host),
			statsd.Timeout(timeout),
			statsd.FlushPeriod(flushInterval),
			statsd.ErrorHandler(errorHandler),
		); err != nil {
			if _, ok := err.(net.Error); !ok {
				return nil, err
			} else {
				// network error can be non-fatal
				err = ErrorPingFailed.New(err)
			}
		}
		sp.clients[key] = s
	}
	atomic.AddInt32(&sp.used, 1)
	newClient := s.Clone(statsd.Prefix(prefix))
	return newClient, err
}

func (sp *statsdPool) Close() {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	if sp.clients != nil {
		if atomic.AddInt32(&sp.used, -1) == 0 {
			goglog.Logger.Infof("%s: closing statsd clients", ModuleName)
			for _, s := range sp.clients {
				s.Close()
			}
			clientPool.clients = nil
		}
	}
}

func InitStatsdPool() {
	clientPool.mu.Lock()
	defer clientPool.mu.Unlock()

	if clientPool.clients == nil {
		clientPool.clients = make(map[string]*statsd.Client)
		clientPool.last = time.Now().Truncate(time.Minute).Unix()
		clientPool.used = int32(0)
	}
}

var (
	clientPool statsdPool
)

// OutputConfig holds the configuration json fields and internal objects
type OutputConfig struct {
	config.OutputConfig
	Host   string `json:"host"`
	Proto  string `json:"protocol"`
	Prefix string `json:"prefix"`

	Timeout       time.Duration `json:"timeout,omitempty"`
	FlushInterval time.Duration `json:"flush_interval,omitempty"`

	Increment    []string               `json:"increment"`
	IncrementTpl []stringutils.Template `json:"-"`

	Decrement    []string               `json:"decrement"`
	DecrementTpl []stringutils.Template `json:"-"`

	Count    []NameValue    `json:"count"`
	CountTpl []templatePair `json:"-"`

	Gauge    []NameValue    `json:"gauge"`
	GaugeTpl []templatePair `json:"-"`

	Timing    []NameValue    `json:"timing"`
	TimingTpl []templatePair `json:"-"`

	client *statsd.Client `json:"-"`
}

// DefaultOutputConfig returns an OutputConfig struct with default values
func DefaultOutputConfig() OutputConfig {
	return OutputConfig{
		OutputConfig: config.OutputConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		Host:          "localhost:8125",
		Proto:         "udp",
		Timeout:       5 * time.Second,
		FlushInterval: 100 * time.Millisecond,
	}
}

// errors
var (
	ErrorFlushFailed          = errutil.NewFactory("flush to statsd server failed")
	ErrorEventMarshalFailed1  = errutil.NewFactory("event Marshal failed: %v")
	ErrorUnsupportedDataType1 = errutil.NewFactory("unsupported data type: %q")
)

func errorHandler(err error) {
	goglog.Logger.Errorf("%s: %s", ModuleName, err.Error())
}

// InitHandler initialize the output plugin
func InitHandler(ctx context.Context, raw *config.ConfigRaw) (config.TypeOutputConfig, error) {
	conf := DefaultOutputConfig()
	err := config.ReflectConfig(raw, &conf)
	if err != nil {
		return nil, err
	}

	// Init templates for StatsD metric types
	if len(conf.Increment) > 0 {
		conf.IncrementTpl = make([]stringutils.Template, len(conf.Increment))
		for i, v := range conf.Increment {
			if conf.IncrementTpl[i], err = stringutils.InitTemplate(v); err != nil {
				return nil, fmt.Errorf("%s: Init Increment template '%s': %s", ModuleName, v, err.Error())
			}
		}
	}
	if len(conf.Decrement) > 0 {
		conf.DecrementTpl = make([]stringutils.Template, len(conf.Decrement))
		for i, v := range conf.Decrement {
			if conf.DecrementTpl[i], err = stringutils.InitTemplate(v); err != nil {
				return nil, fmt.Errorf("%s: Init Decrement template '%s': %s", ModuleName, v, err.Error())
			}
		}
	}
	if len(conf.Count) > 0 {
		conf.CountTpl = make([]templatePair, len(conf.Count))
		for i, v := range conf.Count {
			if conf.CountTpl[i].name, err = stringutils.InitTemplate(v.Name); err != nil {
				return nil, fmt.Errorf("%s: Init Count Name template '%s': %s", ModuleName, v.Name, err.Error())
			}
			if conf.CountTpl[i].value, err = stringutils.InitTemplate(v.Value); err != nil {
				return nil, fmt.Errorf("%s: Init Count Value template '%s': %s", ModuleName, v.Name, err.Error())
			}
		}
	}
	if len(conf.Gauge) > 0 {
		conf.GaugeTpl = make([]templatePair, len(conf.Gauge))
		for i, v := range conf.Gauge {
			if conf.GaugeTpl[i].name, err = stringutils.InitTemplate(v.Name); err != nil {
				return nil, fmt.Errorf("%s: Init Gauge Name template '%s': %s", ModuleName, v.Name, err.Error())
			}
			if conf.GaugeTpl[i].value, err = stringutils.InitTemplate(v.Value); err != nil {
				return nil, fmt.Errorf("%s: Init Gauge Value template '%s': %s", ModuleName, v.Name, err.Error())
			}
		}
	}
	if len(conf.Timing) > 0 {
		for _, v := range conf.Timing {
			var (
				nameTpl  stringutils.Template
				valueTpl stringutils.Template
				err      error
			)
			if nameTpl, err = stringutils.InitTemplate(v.Name); err != nil {
				return nil, fmt.Errorf("%s: Init Timing Name template '%s': %s", ModuleName, v.Name, err.Error())
			}
			if valueTpl, err = stringutils.InitTemplate(v.Value); err != nil {
				return nil, fmt.Errorf("%s: Init Timing Value template '%s': %s", ModuleName, v.Value, err.Error())
			}
			conf.TimingTpl = append(conf.TimingTpl, templatePair{name: nameTpl, value: valueTpl})
		}
	}

	InitStatsdPool()

	if conf.client, err = clientPool.Client(
		conf.Host, conf.Proto,
		conf.Prefix,
		conf.Timeout,
		conf.FlushInterval,
	); err != nil {
		if !ErrorPingFailed.In(err) {
			return nil, err
		}
	}

	atomic.StoreInt64(&clientPool.last, time.Now().Truncate(time.Minute).Unix())

	go func() {
		ticker := time.NewTicker(time.Second * 10)
		defer ticker.Stop()
	LOOP:
		for {
			select {
			case <-ticker.C:
				atomic.StoreInt64(&clientPool.last, time.Now().Truncate(time.Minute).Unix())
			case <-ctx.Done():
				break LOOP
			}
		}
		clientPool.Close()
	}()

	return &conf, nil
}

// Output event
func (o *OutputConfig) Output(ctx context.Context, event logevent.LogEvent) error {
	t := atomic.LoadInt64(&clientPool.last)
	if event.Timestamp.Unix() < t {
		// old event, skip
		return nil
	}
	for _, tpl := range o.IncrementTpl {
		if name, err := tpl.Execute(event.Extra); err == nil && len(name) > 0 {
			o.client.Increment(name)
		}
	}

	for _, tpl := range o.DecrementTpl {
		if name, err := tpl.Execute(event.Extra); err == nil && len(name) > 0 {
			o.client.Decrement(name)
		}
	}

	for _, tpl := range o.CountTpl {
		var (
			name  string
			value string
			err   error
		)
		if name, err = tpl.name.Execute(event.Extra); err == nil && len(name) > 0 {
			if value, err = tpl.value.Execute(event.Extra); err == nil && len(value) > 0 {
				if n, err := strconv.ParseInt(value, 10, 64); err == nil {
					o.client.Count(name, n)
				}
			}
		}
	}

	for _, tpl := range o.GaugeTpl {
		var (
			name  string
			value string
			err   error
		)
		if name, err = tpl.name.Execute(event.Extra); err == nil && len(name) > 0 {
			if value, err = tpl.value.Execute(event.Extra); err == nil && len(value) > 0 {
				if f, err := strconv.ParseFloat(value, 64); err == nil {
					o.client.Gauge(name, f)
				}
			}
		}
	}
	for _, tpl := range o.TimingTpl {
		var (
			name  string
			value string
			err   error
		)
		if name, err = tpl.name.Execute(event.Extra); err != nil || len(name) == 0 {
			continue
		}
		if value, err = tpl.value.Execute(event.Extra); err != nil || len(value) == 0 {
			continue
		}
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			o.client.Timing(name, f)
		}
	}
	return nil
}
