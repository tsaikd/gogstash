package outputkafka

import (
	"context"

	"github.com/Shopify/sarama"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
)

// ModuleName is the name used in config file
const ModuleName = "kafka"

// OutputConfig holds the configuration json fields and internal objects
type OutputConfig struct {
	config.OutputConfig
	Version          string   `json:"version"`                     // Kafka cluster version, eg: 0.10.2.0
	Brokers          []string `json:"brokers"`                     // Kafka bootstrap brokers to connect to, as a comma separated list
	Topics           []string `json:"topics"`                      // Kafka topics to be consumed, as a comma seperated list
	SecurityProtocol string   `json:"security_protocol,omitempty"` // use SASL authentication
	User             string   `json:"sasl_username,omitempty"`     // SASL authentication username
	Password         string   `json:"sasl_password,omitempty"`     // SASL authentication password

	client sarama.AsyncProducer
}

// DefaultOutputConfig returns an OutputConfig struct with default values
func DefaultOutputConfig() OutputConfig {
	return OutputConfig{
		OutputConfig: config.OutputConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		SecurityProtocol: "",
		User:             "",
		Password:         "",
	}
}

// InitHandler initialize the output plugin
func InitHandler(ctx context.Context, raw config.ConfigRaw) (config.TypeOutputConfig, error) {
	conf := DefaultOutputConfig()
	if err := config.ReflectConfig(raw, &conf); err != nil {
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

	if len(conf.Topics) < 0 {
		goglog.Logger.Error("topics should not be empty")
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

	conf.client, err = sarama.NewAsyncProducer(conf.Brokers, sarConfig)
	if err != nil {
		goglog.Logger.Errorf("Error creating producer client: %v", err)
		return nil, err
	}

	return &conf, nil
}

// Output event
func (t *OutputConfig) Output(ctx context.Context, event logevent.LogEvent) (err error) {

	raw, err := event.MarshalJSON()
	if err != nil {
		return err
	}

	ch := t.client.Input()
	for _, topic := range t.Topics {
		msg := &sarama.ProducerMessage{Topic: topic, Value: sarama.ByteEncoder(raw)}
		ch <- msg
	}

	return nil
}
