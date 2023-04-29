package outputgelf

import (
	"context"
	"regexp"
	"strings"

	"github.com/tsaikd/KDGoLib/errutil"

	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
	"github.com/tsaikd/gogstash/config/queue"
)

// ModuleName is the name used in config file
const ModuleName = "gelf"

// errors
var (
	ErrNoValidHosts          = errutil.NewFactory("no valid Hosts found")
	ErrInvalidHost           = errutil.NewFactory("Some host are invalid. Should respect the format: hostname:port")
	ErrorCreateClientFailed1 = errutil.NewFactory("create gelf client failed: %q")
)

// OutputConfig holds the configuration json fields and internal objects
type OutputConfig struct {
	config.OutputConfig
	ChunkSize        int      `json:"chunk_size" yaml:"chunk_size"`
	CompressionLevel int      `json:"compression_level" yaml:"compression_level"`
	CompressionType  int      `json:"compression_type" yaml:"compression_type"`
	Hosts            []string `json:"hosts" yaml:"hosts"` // List of Gelf Host servers, format: ip:port

	RetryInterval uint `json:"retry_interval" yaml:"retry_interval"` // seconds before a new retry in case on error
	MaxQueueSize  int  `json:"max_queue_size" yaml:"max_queue_size"` // max size of queue before deleting events (-1=no limit, 0=disable)

	gelfWriters []GELFWriter
	queue       queue.Queue // our queue
}

// DefaultOutputConfig returns an OutputConfig struct with default values
func DefaultOutputConfig() OutputConfig {
	return OutputConfig{
		OutputConfig: config.OutputConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		RetryInterval: 30,
	}
}

// InitHandler initialize the output plugin
func InitHandler(
	ctx context.Context,
	raw config.ConfigRaw,
	control config.Control,
) (config.TypeOutputConfig, error) {
	conf := DefaultOutputConfig()
	if err := config.ReflectConfig(raw, &conf); err != nil {
		return nil, err
	}

	if len(conf.Hosts) == 0 {
		return nil, ErrNoValidHosts
	}

	// host validation regex
	r := regexp.MustCompile(`[^\:]+:[0-9]{1,5}`)

	for _, host := range conf.Hosts {
		if !r.MatchString(host) {
			return nil, ErrInvalidHost
		}

		gelfWriter, err := NewWriter(GELFConfig{
			Host:             host,
			ChunkSize:        conf.ChunkSize,
			CompressionLevel: conf.CompressionLevel,
			CompressionType:  CompressType(conf.CompressionType),
		})
		if err != nil {
			return nil, ErrorCreateClientFailed1.New(err, host)
		}

		conf.gelfWriters = append(conf.gelfWriters, gelfWriter)
	}

	// create the queue
	conf.queue = queue.NewSimpleQueue(ctx, control, &conf, nil, conf.MaxQueueSize, conf.RetryInterval)

	return conf.queue, nil
}

// OutputEvent handle message from the queue
func (t *OutputConfig) OutputEvent(ctx context.Context, event logevent.LogEvent) (err error) {
	var host string
	var level int32
	for k, v := range event.Extra {
		lk := strings.ToLower(k)

		if lk == "host" || lk == "hostname" {
			if h, ok := v.(string); ok {
				host = h
			}
			delete(event.Extra, k)
		}

		if lk == "level" {
			if l, ok := v.(int32); ok {
				level = l
			}
			delete(event.Extra, k)
		}
	}

	if len(event.Tags) > 0 {
		event.Extra["tags"] = strings.Join(event.Tags, ", ")
	}

	for _, w := range t.gelfWriters {
		err := w.WriteMessage(
			&SimpleMessage{
				Extra:     event.Extra,
				Host:      host,
				Level:     level,
				Message:   event.Message,
				Timestamp: event.Timestamp,
			},
		)
		if err != nil {
			goglog.Logger.Errorf("outputgelf: %s", err.Error())
			err = t.queue.Queue(ctx, event)
			if err != nil {
				goglog.Logger.Errorf("outputgelf: %s", err.Error())
			}
		}
	}

	return t.queue.Resume(ctx)
}
