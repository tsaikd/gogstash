package nsq

import (
	"context"
	"errors"

	"github.com/nsqio/go-nsq"
	"github.com/tsaikd/KDGoLib/version"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
)

// ModuleName is the name used in config file
const ModuleName = "nsq"

// InputConfig holds the configuration json fields and internal objects
type InputConfig struct {
	config.InputConfig
	NSQ      string `json:"nsq" yaml:"nsq"`                   // NSQd to connect to
	Lookupd  string `json:"lookupd" yaml:"lookupd"`           // lookupd to connect to, can be a NSQd or lookupd
	Topic    string `json:"topic" yaml:"topic"`               // topic to listen from
	Channel  string `json:"channel" yaml:"channel"`           // channel to subscribe to
	InFlight uint   `json:"max_inflight" yaml:"max_inflight"` // max number of messages inflight
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
func InitHandler(ctx context.Context, raw config.ConfigRaw) (config.TypeInputConfig, error) {
	conf := DefaultInputConfig()
	err := config.ReflectConfig(raw, &conf)
	if err != nil {
		return nil, err
	}
	if len(conf.Lookupd) == 0 && len(conf.NSQ) == 0 {
		return nil, errors.New("nsq: you need to specify nsq or lookupd")
	}
	if len(conf.Topic) == 0 {
		return nil, errors.New("nsq: missing topic")
	}
	if len(conf.Channel) == 0 {
		return nil, errors.New("nsq: missing channel")
	}
	conf.Codec, err = config.GetCodecOrDefault(ctx, raw["codec"])
	if err != nil {
		return nil, err
	}

	return &conf, err
}

// Start wraps the actual function starting the plugin
func (t *InputConfig) Start(ctx context.Context, msgChan chan<- logevent.LogEvent) (err error) {
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
		return
	}
	consumer.AddHandler(&handler)
	consumer.ConnectToNSQD(t.NSQ)
	consumer.ConnectToNSQLookupd(t.Lookupd)
	// wait for stop signal and exit
	<-ctx.Done()
	consumer.Stop()
	<-consumer.StopChan
	return
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
