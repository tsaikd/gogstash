package inputkafka

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Shopify/sarama"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
)

// ModuleName is the name used in config file
const ModuleName = "kafka"

// InputConfig holds the configuration json fields and internal objects
type InputConfig struct {
	config.InputConfig
	Version          string   `json:"version"`                     // Kafka cluster version, eg: 0.10.2.0
	Brokers          []string `json:"brokers"`                     // Kafka bootstrap brokers to connect to, as a comma separated list
	Topics           []string `json:"topics"`                      // Kafka topics to be consumed, as a comma seperated list
	Group            string   `json:"group"`                       // Kafka consumer group definition
	OffsetOldest     bool     `json:"offset_oldest"`               // Kafka consumer consume initial offset from oldest
	Assignor         string   `json:"assignor"`                    // Consumer group partition assignment strategy (range, roundrobin)
	SecurityProtocol string   `json:"security_protocol,omitempty"` // use SASL authentication
	User             string   `json:"sasl_username,omitempty"`     // SASL authentication username
	Password         string   `json:"sasl_password,omitempty"`     // SASL authentication password

	saConf *sarama.Config
}

// DefaultInputConfig returns an InputConfig struct with default values
func DefaultInputConfig() InputConfig {
	return InputConfig{
		InputConfig: config.InputConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		SecurityProtocol: "",
		User:             "",
		Password:         "",
	}
}

// InitHandler initialize the input plugin
func InitHandler(ctx context.Context, raw config.ConfigRaw) (config.TypeInputConfig, error) {
	conf := DefaultInputConfig()
	err := config.ReflectConfig(raw, &conf)
	if err != nil {
		return nil, err
	}

	sarama.Logger = goglog.Logger

	version, err := sarama.ParseKafkaVersion(conf.Version)
	if err != nil {
		goglog.Logger.Errorf("Error parsing Kafka version: %v", err)
		return nil, err
	}

	/**
	 * Construct a new Sarama configuration.
	 * The Kafka cluster version has to be defined before the consumer/producer is initialized.
	 */
	sarConfig := sarama.NewConfig()
	sarConfig.Version = version

	switch conf.Assignor {
	case "roundrobin":
		sarConfig.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	case "range":
		sarConfig.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange
	default:
		goglog.Logger.Errorf("Unrecognized consumer group partition assignor: %s", conf.Assignor)
	}

	if conf.OffsetOldest {
		sarConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	}

	if len(conf.Topics) < 0 {
		goglog.Logger.Error("topics should not be empty")
		return nil, err
	}

	if conf.Group == "" {
		goglog.Logger.Error("group should not be empty")
		return nil, err
	}

	if len(conf.Brokers) == 0 {
		goglog.Logger.Error("topics should not be empty")
		return nil, err
	}

	if conf.SecurityProtocol == "SASL" {
		sarConfig.Net.SASL.Enable = true
		sarConfig.Net.SASL.User = conf.User
		sarConfig.Net.SASL.Password = conf.Password
	}

	conf.saConf = sarConfig

	conf.Codec, err = config.GetCodecOrDefault(ctx, raw["codec"])
	if err != nil {
		return nil, err
	}

	return &conf, nil
}

// Start wraps the actual function starting the plugin
func (t *InputConfig) Start(ctx context.Context, msgChan chan<- logevent.LogEvent) (err error) {
	/**
	 * Setup a new Sarama consumer group
	 */
	cum := consumerHandle{
		i:     t,
		ch:    msgChan,
		ready: make(chan bool),
		ctx:   ctx,
	}

	ct, cancel := context.WithCancel(ctx)
	defer cancel()
	client, err := sarama.NewConsumerGroup(t.Brokers, t.Group, t.saConf)
	if err != nil {
		goglog.Logger.Errorf("Error creating consumer group client: %v", err)
		return err
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if err := client.Consume(ct, t.Topics, &cum); err != nil {
				goglog.Logger.Errorf("Error from consumer: %v", err)
			}
			// check if context was cancelled, signaling that the consumer should stop
			if ct.Err() != nil {
				return
			}

			cum.ready = make(chan bool)
		}
	}()

	<-cum.ready // Await till the consumer has been set up
	goglog.Logger.Println("Sarama consumer up and running!...")

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-ct.Done():
		goglog.Logger.Println("terminating: context cancelled")
	case <-sigterm:
		goglog.Logger.Println("terminating: via signal")
	}
	cancel()
	wg.Wait()
	if err = client.Close(); err != nil {
		goglog.Logger.Errorf("Error closing client: %v", err)
		return err
	}
	return nil
}

// consumerHandle represents a Sarama consumer group consumer
type consumerHandle struct {
	i     *InputConfig
	ch    chan<- logevent.LogEvent
	ready chan bool
	ctx   context.Context
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (c *consumerHandle) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(c.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (c *consumerHandle) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (c *consumerHandle) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {

	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/Shopify/sarama/blob/master/consumer_group.go#L27-L29
	for message := range claim.Messages() {
		//goglog.Logger.Printf("Message claimed: value = %s, timestamp = %v, topic = %s", string(message.Value), message.Timestamp, message.Topic)
		var extra = map[string]interface{}{
			"topic":     message.Topic,
			"timestamp": message.Timestamp,
		}
		ok, err := c.i.Codec.Decode(c.ctx, string(message.Value), extra, []string{}, c.ch)
		if !ok {
			goglog.Logger.Errorf("decode message to msg chan error : %v", err)
		}
		session.MarkMessage(message, "")
	}

	return nil
}
