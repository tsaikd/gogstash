package outputamqp

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/bitly/go-hostpool"
	"github.com/streadway/amqp"
	"github.com/tsaikd/KDGoLib/errutil"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
)

// ModuleName is the name used in config file
const ModuleName = "amqp"

// errors
var (
	ErrorNoValidConn = errutil.NewFactory("no valid amqp server connection found")
)

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
	amqpClients        map[string]amqpClient
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

		amqpClients: map[string]amqpClient{},
		evchan:      make(chan logevent.LogEvent),
	}
}

type amqpClient struct {
	client    *amqp.Channel
	reconnect chan hostpool.HostPoolResponse
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
				o.amqpClients[url] = amqpClient{
					client:    ch,
					reconnect: make(chan hostpool.HostPoolResponse, 1),
				}
				go o.reconnect(url)
				hosts = append(hosts, url)
			}
		}
	}

	if len(hosts) == 0 {
		return ErrorNoValidConn.New(nil)
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
		if err := o.amqpClients[hp.Host()].client.Publish(
			exchange,
			routingKey,
			false,
			false,
			amqp.Publishing{
				ContentType: "application/json",
				Body:        raw,
			},
		); err != nil {
			hp.Mark(err)
			o.amqpClients[hp.Host()].reconnect <- hp
		} else {
			break
		}
	}

	return
}

func (o *OutputConfig) reconnect(url string) {
	for {
		select {
		case poolResponse := <-o.amqpClients[url].reconnect:
			// When a reconnect event is received
			// start reconnect loop until reconnected
			for {
				time.Sleep(time.Duration(o.ReconnectDelay) * time.Second)

				logrus.Info("Reconnecting to ", poolResponse.Host())

				if conn, err := amqp.Dial(poolResponse.Host()); err == nil {
					if ch, err := conn.Channel(); err == nil {
						if err := ch.ExchangeDeclare(
							o.Exchange,
							o.ExchangeType,
							o.ExchangeDurable,
							o.ExchangeAutoDelete,
							false,
							false,
							nil,
						); err == nil {
							logrus.Info("Reconnected to ", poolResponse.Host())
							o.amqpClients[poolResponse.Host()] = amqpClient{
								client:    ch,
								reconnect: make(chan hostpool.HostPoolResponse, 1),
							}
							poolResponse.Mark(nil)
							break
						}
					}
				}

				logrus.Info("Failed to reconnect to ", url, ". Waiting ", o.ReconnectDelay, " seconds...")
			}
		}
	}
}
