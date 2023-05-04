package nsq

import (
	"context"
	"errors"

	"github.com/tsaikd/KDGoLib/version"

	"github.com/nsqio/go-nsq"

	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
)

// ModuleName is the name used in config file
const ModuleName = "nsq"

// InputConfig holds the configuration json fields and internal objects
type InputConfig struct {
	config.InputConfig
	NSQ      string         `json:"nsq" yaml:"nsq"`                   // NSQd to connect to
	Lookupd  string         `json:"lookupd" yaml:"lookupd"`           // lookupd to connect to, can be a NSQd or lookupd
	Topic    string         `json:"topic" yaml:"topic"`               // topic to listen from
	Channel  string         `json:"channel" yaml:"channel"`           // channel to subscribe to
	InFlight uint           `json:"max_inflight" yaml:"max_inflight"` // max number of messages inflight
	control  config.Control // backpressure control
}

// DefaultInputConfig returns an InputConfig struct with default values
func DefaultInputConfig() InputConfig {
	return InputConfig{
		InputConfig: config.InputConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		InFlight: 75,
	}
}

// InitHandler initialize the input plugin
func InitHandler(
	ctx context.Context,
	raw config.ConfigRaw,
	control config.Control,
) (config.TypeInputConfig, error) {
	conf := DefaultInputConfig()
	conf.control = control
	err := config.ReflectConfig(raw, &conf)
	if err != nil {
		return nil, err
	}
	if conf.Lookupd == "" && conf.NSQ == "" {
		return nil, errors.New("nsq: you need to specify nsq or lookupd")
	}
	if conf.Topic == "" {
		return nil, errors.New("nsq: missing topic")
	}
	if conf.Channel == "" {
		return nil, errors.New("nsq: missing channel")
	}
	conf.Codec, err = config.GetCodecOrDefault(ctx, raw["codec"])
	if err != nil {
		return nil, err
	}

	return &conf, err
}

// Start wraps the actual function starting the plugin
func (t *InputConfig) Start(ctx context.Context, msgChan chan<- logevent.LogEvent) error {
	// setup the handler
	handler := nsqhandler{
		msgChan: msgChan,
		ctx:     ctx,
		i:       t,
	}
	// setup the consumer
	conf := nsq.NewConfig()
	conf.MaxInFlight = int(t.InFlight)
	conf.UserAgent = "gogstash/" + version.VERSION
	consumer, err := nsq.NewConsumer(t.Topic, t.Channel, conf)
	if err != nil {
		goglog.Logger.Errorf("nsq: %s", err.Error())
		return err
	}
	consumer.AddHandler(&handler)
	if len(t.NSQ) > 0 {
		if err := consumer.ConnectToNSQD(t.NSQ); err != nil {
			return err
		}
	}
	if len(t.Lookupd) > 0 {
		if err := consumer.ConnectToNSQLookupd(t.Lookupd); err != nil {
			return err
		}
	}
	// wait for stop signal and exit
outer_loop:
	for {
		select {
		case <-ctx.Done():
			break outer_loop
		case <-t.control.PauseSignal():
			goglog.Logger.Info("nsq: received pause")
			consumer.ChangeMaxInFlight(0)
		case <-t.control.ResumeSignal():
			goglog.Logger.Info("nsq: received resume")
			consumer.ChangeMaxInFlight(int(t.InFlight))
		}
	}
	consumer.Stop()
	<-consumer.StopChan
	goglog.Logger.Info("nsq stopped")
	return err
}

// nsqhandler implements a handler to receive messages
type nsqhandler struct {
	msgChan chan<- logevent.LogEvent
	ctx     context.Context
	i       *InputConfig
}

// HandleMessage receives a message from NSQ
func (h *nsqhandler) HandleMessage(m *nsq.Message) error {
	ok, err := h.i.Codec.Decode(h.ctx, m.Body, nil, []string{}, h.msgChan)
	if !ok {
		goglog.Logger.Errorf("nsq: nok ok, error %s", err.Error())
	} else if err != nil {
		goglog.Logger.Errorf("nsq: %s", err.Error())
	}
	return err
}
