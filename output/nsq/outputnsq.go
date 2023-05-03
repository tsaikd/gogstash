package nsq

import (
	"context"
	"errors"

	"github.com/nsqio/go-nsq"
	"github.com/tsaikd/KDGoLib/version"

	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
	"github.com/tsaikd/gogstash/config/queue"
)

// ModuleName is the name used in config file
const ModuleName = "nsq"

// OutputConfig holds the configuration json fields and internal objects
type OutputConfig struct {
	config.OutputConfig
	NSQ      string `json:"nsq" yaml:"nsq"`                   // NSQd to connect to
	Topic    string `json:"topic" yaml:"topic"`               // topic to publish to
	InFlight uint   `json:"max_inflight" yaml:"max_inflight"` // max number of messages inflight

	producer *nsq.Producer
	ctx      context.Context

	msg   chan []byte // channel to push message from codec to
	queue queue.Queue // our queue for backpressure handling
}

// DefaultOutputConfig returns an OutputConfig struct with default values
func DefaultOutputConfig() OutputConfig {
	return OutputConfig{
		OutputConfig: config.OutputConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		InFlight: 150,
		msg:      make(chan []byte),
	}
}

// InitHandler initialize the output plugin
func InitHandler(
	ctx context.Context,
	raw config.ConfigRaw,
	control config.Control,
) (config.TypeOutputConfig, error) {
	conf := DefaultOutputConfig()
	err := config.ReflectConfig(raw, &conf)
	if err != nil {
		return nil, err
	}

	if conf.NSQ == "" {
		return nil, errors.New("Missing NSQ server")
	}
	if conf.Topic == "" {
		return nil, errors.New("Missing topic")
	}

	conf.Codec, err = config.GetCodecOrDefault(ctx, raw["codec"])
	if err != nil {
		return nil, err
	}

	conf.ctx = ctx
	cfg := nsq.NewConfig()
	cfg.MaxInFlight = int(conf.InFlight)
	cfg.UserAgent = "gogstash/" + version.VERSION
	conf.producer, err = nsq.NewProducer(conf.NSQ, cfg)
	if err != nil {
		return nil, err
	}
	conf.queue = queue.NewSimpleQueue(ctx, control, &conf, conf.msg, 10, 30) // last values are queue size and retry interval in seconds
	go conf.nsqbackgroundtask()
	return conf.queue, nil
}

// nsqbackgroundtask runs in the background and handles messages and termination
func (t *OutputConfig) nsqbackgroundtask() {
	for {
		select {
		case <-t.ctx.Done():
			goglog.Logger.Debug("outputnsq: stopping")
			t.producer.Stop()
			goglog.Logger.Debug("outputnsq: stopped")
			close(t.msg)
			return
		case msg := <-t.msg:
			err := t.producer.Publish(t.Topic, msg)
			if err != nil {
				err = t.queue.Queue(t.ctx, msg)
				if err != nil {
					goglog.Logger.Errorf("outputnsq: %s", err.Error())
				}
			} else {
				_ = t.queue.Resume(t.ctx)
			}
		}
	}
}

// Output event
func (t *OutputConfig) OutputEvent(ctx context.Context, event logevent.LogEvent) (err error) {
	_, err = t.Codec.Encode(ctx, event, t.msg)
	return
}
