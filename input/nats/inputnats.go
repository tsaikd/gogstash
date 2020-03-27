package inputnats

import (
	"context"
	"strings"

	"github.com/nats-io/nats.go"
	codecjson "github.com/tsaikd/gogstash/codec/json"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
)

// ModuleName is the name used in config file
const ModuleName = "nats"

// ErrorTag tag added to event when process module failed
const ErrorTag = "gogstash_input_nats_error"

// InputConfig holds the configuration json fields and internal objects
type InputConfig struct {
	config.InputConfig
	Host  string `json:"host"`  // redis server host:port, default: "localhost:6379"
	Creds string `json:"creds"` // where to get data, default: "gogstash"
	Topic string `json:"topic"` // topics to receive

	client *nats.Conn
}

// DefaultInputConfig returns an InputConfig struct with default values
func DefaultInputConfig() InputConfig {
	return InputConfig{
		InputConfig: config.InputConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		Host:  "localhost:4222",
		Creds: "",
	}
}

// errors
var (
	subList    map[string]*nats.Subscription
	msgChannel chan<- logevent.LogEvent
	cont       context.Context
	ic         *InputConfig
)

// InitHandler initialize the input plugin
func InitHandler(ctx context.Context, raw *config.ConfigRaw) (config.TypeInputConfig, error) {
	conf := DefaultInputConfig()
	err := config.ReflectConfig(raw, &conf)
	if err != nil {
		return nil, err
	}

	opts := []nats.Option{nats.Name("gostash")}
	if conf.Creds != "" {
		opts = append(opts, nats.UserCredentials(conf.Creds))
	}

	conf.client, err = nats.Connect(conf.Host, opts...)
	if err != nil {
		return nil, err
	}

	conf.Codec, err = config.GetCodecDefault(ctx, *raw, codecjson.ModuleName)
	if err != nil {
		return nil, err
	}

	subList = make(map[string]*nats.Subscription)

	return &conf, nil
}

func msgHandler(msg *nats.Msg) {
	goglog.Logger.Infof("Rx Msg Topic: %s", msg.Subject)
	_, err := ic.Codec.Decode(cont, msg.Data, nil, []string{}, msgChannel)
	if err != nil {
		goglog.Logger.Warnf("Decode failed: %v", err)
	}
}

// Start wraps the actual function starting the plugin
func (i *InputConfig) Start(ctx context.Context, msgChan chan<- logevent.LogEvent) error {
	ic = i
	cont = ctx
	msgChannel = msgChan
	topics := strings.Split(i.Topic, ",")
	for _, topic := range topics {
		topic = strings.TrimSpace(topic)
		sub, err := i.client.Subscribe(topic, msgHandler)
		if err != nil {
			goglog.Logger.Warnf("subscribe topic %s failed: %v", topic, err)
			// continue
		} else {
			goglog.Logger.Infof("Subscribed to topic %s", topic)
			subList[topic] = sub
		}
	}

	<-ctx.Done()
	goglog.Logger.Info("input nats stopped")
	for _, sub := range subList {
		if err := sub.Drain(); err != nil {
			goglog.Logger.Error(err)
		}
	}
	return nil
}
