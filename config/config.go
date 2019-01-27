package config

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/tsaikd/KDGoLib/errutil"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
	"golang.org/x/sync/errgroup"
	yaml "gopkg.in/yaml.v2"
)

// errors
var (
	ErrorReadConfigFile1     = errutil.NewFactory("Failed to read config file: %q")
	ErrorUnmarshalJSONConfig = errutil.NewFactory("Failed unmarshalling config in JSON format")
	ErrorUnmarshalYAMLConfig = errutil.NewFactory("Failed unmarshalling config in YAML format")
	ErrorTimeout1            = errutil.NewFactory("timeout: %v")
)

// Config contains all config
type Config struct {
	InputRaw  []ConfigRaw `json:"input,omitempty" yaml:"input"`
	FilterRaw []ConfigRaw `json:"filter,omitempty" yaml:"filter"`
	OutputRaw []ConfigRaw `json:"output,omitempty" yaml:"output"`

	Event *logevent.Config `json:"event,omitempty" yaml:"event"`

	// channel size: chInFilter, chFilterOut, chOutDebug
	ChannelSize int `json:"chsize,omitempty" yaml:"chsize"`

	// worker number, defaults to 1
	Worker int `json:"worker,omitempty" yaml:"worker"`

	// enable debug channel, used for testing
	DebugChannel bool `json:"debugch,omitempty" yaml:"debugch"`

	chInFilter  MsgChan // channel from input to filter
	chFilterOut MsgChan // channel from filter to output
	chOutDebug  MsgChan // channel from output to debug
	ctx         context.Context
	eg          *errgroup.Group
}

var defaultConfig = Config{
	ChannelSize: 100,
	Worker:      1,
}

// MsgChan message channel type
type MsgChan chan logevent.LogEvent

// LoadFromFile load config from filepath
func LoadFromFile(path string) (config Config, err error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return config, ErrorReadConfigFile1.New(err, path)
	}

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".yml", ".yaml":
		return LoadFromYAML(data)
	default:
		return LoadFromJSON(data)
	}
}

// LoadFromJSON load config from []byte in JSON format
func LoadFromJSON(data []byte) (config Config, err error) {
	if data, err = cleanComments(data); err != nil {
		return
	}

	if err = json.Unmarshal(data, &config); err != nil {
		return config, ErrorUnmarshalJSONConfig.New(err)
	}

	initConfig(&config)
	return
}

// LoadFromYAML load config from []byte in YAML format
func LoadFromYAML(data []byte) (config Config, err error) {
	if err = yaml.Unmarshal(data, &config); err != nil {
		return config, ErrorUnmarshalYAMLConfig.New(err)
	}
	initConfig(&config)
	return
}

func initConfig(config *Config) {
	rv := reflect.ValueOf(&config)
	formatReflect(rv)

	if config.ChannelSize < 1 {
		config.ChannelSize = defaultConfig.ChannelSize
	}
	if config.Worker < 1 {
		config.Worker = defaultConfig.Worker
	}
	if config.Event != nil {
		logevent.SetConfig(config.Event)
	}

	config.chInFilter = make(MsgChan, config.ChannelSize)
	config.chFilterOut = make(MsgChan, config.ChannelSize)
	if config.DebugChannel {
		config.chOutDebug = make(MsgChan, config.ChannelSize)
	}
}

// Start config in goroutines
func (t *Config) Start(ctx context.Context) (err error) {
	ctx = contextWithOSSignal(ctx, goglog.Logger, os.Interrupt, os.Kill)
	t.eg, t.ctx = errgroup.WithContext(ctx)

	if err = t.startInputs(); err != nil {
		return
	}
	if err = t.startFilters(); err != nil {
		return
	}
	if err = t.startOutputs(); err != nil {
		return
	}
	return
}

// Wait blocks until all filters returned, then
// returns the first non-nil error (if any) from them.
func (t *Config) Wait() (err error) {
	return t.eg.Wait()
}

// TestInputEvent send an event to chInFilter, used for testing
func (t *Config) TestInputEvent(event logevent.LogEvent) {
	t.chInFilter <- event
}

// TestGetOutputEvent get an event from chOutDebug, used for testing
func (t *Config) TestGetOutputEvent(timeout time.Duration) (event logevent.LogEvent, err error) {
	ctx, cancel := context.WithTimeout(t.ctx, timeout)
	defer cancel()
	select {
	case <-ctx.Done():
		return
	case ev := <-t.chOutDebug:
		return ev, nil
	case <-time.After(timeout + 10*time.Millisecond):
		return event, ErrorTimeout1.New(nil, timeout)
	}
}
