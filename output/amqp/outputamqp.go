package outputamqp

import (
	"errors"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/bitly/go-hostpool"
	"github.com/streadway/amqp"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
)

// ModuleName is the name used in config file
const ModuleName = "amqp"

// OutputConfig holds the output configuration json fields and internal objects
type OutputConfig struct {
	config.OutputConfig
	URLs               []string `json:"urls"`                           // Array of AMQP connection strings formatted per the [RabbitMQ URI Spec](http://www.rabbitmq.com/uri-spec.html).
	RoutingKey         string   `json:"routing_key,omitempty"`          // The message routing key used to bind the queue to the exchange. Defaults to empty string.
	Exchange           string   `json:"exchange"`                       // AMQP exchange name
	ExchangeType       string   `json:"exchange_type"`                  // AMQP exchange type (fanout, direct, topic or headers).
	ExchangeDurable    bool     `json:"exchange_durable,omitempty"`     // Whether the exchange should be configured as a durable exchange. Defaults to false.
	ExchangeAutoDelete bool     `json:"exchange_auto_delete,omitempty"` // Whether the exchange is deleted when all queues have finished and there is no publishing. Defaults to true.
	Persistent         bool     `json:"persistent,omitempty"`           // Whether published messages should be marked as persistent or transient. Defaults to false.
	Retries            int      `json:"retries,omitempty"`              // Number of attempts to send a message. Defaults to 3.
	ReconnectDelay     int      `json:"reconnect_delay,omitempty"`      // Delay between each attempt to reconnect to AMQP server. Defaults to 30 seconds.
	hostPool           hostpool.HostPool
	amqpClients        map[string]*amqp.Channel
	evchan             chan logevent.LogEvent
}

type amqpConn struct {
	Channel    *amqp.Channel
	Connection *amqp.Connection
}

// DefaultOutputConfig returns an OutputConfig struct with default values
func DefaultOutputConfig() OutputConfig {
	return OutputConfig{
		OutputConfig: config.OutputConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		RoutingKey:         "",
		ExchangeDurable:    false,
		ExchangeAutoDelete: true,
		Persistent:         false,
		Retries:            3,
		ReconnectDelay:     30,
		amqpClients:        map[string]*amqp.Channel{},

		evchan: make(chan logevent.LogEvent),
	}
}

// InitHandler initialize the output plugin
func InitHandler(confraw *config.ConfigRaw) (retconf config.TypeOutputConfig, err error) {
	conf := DefaultOutputConfig()
	if err = config.ReflectConfig(confraw, &conf); err != nil {
		return
	}

	if err = conf.initAmqpClients(); err != nil {
		return
	}

	retconf = &conf
	return
}

func (o *OutputConfig) initAmqpClients() error {
	var hosts []string

	for _, url := range o.URLs {
		if conn, err := amqp.Dial(url); err == nil {
			if ch, err := conn.Channel(); err == nil {
				o.amqpClients[url] = ch
				err := ch.ExchangeDeclare(
					o.Exchange,
					o.ExchangeType,
					o.ExchangeDurable,
					o.ExchangeAutoDelete,
					false,
					false,
					nil,
				)
				if err != nil {
					return err
				}
				hosts = append(hosts, url)
			}
		}
	}

	if len(hosts) == 0 {
		return errors.New("no valid amqp server connection found")
	}

	o.hostPool = hostpool.New(hosts)
	return nil
}

// Event send the event through AMQP
func (o *OutputConfig) Event(event logevent.LogEvent) (err error) {
	raw, err := event.MarshalJSON()
	if err != nil {
		logrus.Errorf("event Marshal failed: %v", event)
		return
	}

	exchange := event.Format(o.Exchange)
	routingKey := event.Format(o.RoutingKey)

	for i := 0; i <= o.Retries; i++ {
		hp := o.hostPool.Get()
		err = o.amqpClients[hp.Host()].Publish(
			exchange,
			routingKey,
			false,
			false,
			amqp.Publishing{
				ContentType: "application/json",
				Body:        raw,
			},
		)
		if err != nil {
			hp.Mark(err)
			if len(o.amqpClients) > 1 {
				go o.reconnect(hp)
			} else {
				o.reconnect(hp)
			}
		} else {
			break
		}
	}

	return
}

func (o *OutputConfig) reconnect(hp hostpool.HostPoolResponse) {
	for {
		time.Sleep(time.Duration(o.ReconnectDelay) * time.Second)
		logrus.Info("Reconnecting to ", hp.Host())
		if conn, err := amqp.Dial(hp.Host()); err == nil {
			if ch, err := conn.Channel(); err == nil {
				err := ch.ExchangeDeclare(
					o.Exchange,
					o.ExchangeType,
					o.ExchangeDurable,
					o.ExchangeAutoDelete,
					false,
					false,
					nil,
				)
				if err == nil {
					hp.Mark(nil)
					logrus.Info("Reconnected to ", hp.Host())
					o.amqpClients[hp.Host()] = ch
					break
				}
			}
		}
	}
}
