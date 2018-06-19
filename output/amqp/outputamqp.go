package outputamqp

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"strings"
	"time"

	"github.com/bitly/go-hostpool"
	"github.com/sirupsen/logrus"
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

// OutputConfig holds the configuration json fields and internal objects
type OutputConfig struct {
	config.OutputConfig
	URLs               []string `json:"urls"`                           // Array of AMQP connection strings formatted per the [RabbitMQ URI Spec](http://www.rabbitmq.com/uri-spec.html).
	TLSCACerts         []string `json:"tls_ca_certs,omitempty"`         // Array of CA Certificates to load for TLS connections
	TLSCerts           []string `json:"tls_certs,omitempty"`            // Array of Certificates to load for TLS connections
	TLSCertKeys        []string `json:"tls_cert_keys,omitempty"`        // Array of Certificate Keys to load for TLS connections (Must NOT be password protected)
	TLSSkipVerify      bool     `json:"tls_cert_skip_verify,omitempty"` // Skip verification of certifcates. Defaults to false.
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
		TLSSkipVerify:      false,
		RoutingKey:         "",
		ExchangeDurable:    false,
		ExchangeAutoDelete: true,
		Persistent:         false,
		Retries:            3,
		ReconnectDelay:     30,

		amqpClients: map[string]amqpClient{},
	}
}

type amqpClient struct {
	client    *amqp.Channel
	reconnect chan hostpool.HostPoolResponse
}

// InitHandler initialize the output plugin
func InitHandler(ctx context.Context, raw *config.ConfigRaw) (config.TypeOutputConfig, error) {
	conf := DefaultOutputConfig()
	if err := config.ReflectConfig(raw, &conf); err != nil {
		return nil, err
	}

	if err := conf.initAmqpClients(); err != nil {
		return nil, err
	}

	return &conf, nil
}

func (o *OutputConfig) initAmqpClients() error {
	var hosts []string

	for _, url := range o.URLs {
		if conn, err := o.getConnection(url); err == nil {
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

// Output send the event through AMQP
func (o *OutputConfig) Output(ctx context.Context, event logevent.LogEvent) (err error) {
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

				if conn, err := o.getConnection(poolResponse.Host()); err == nil {
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

func (o *OutputConfig) getConnection(url string) (c *amqp.Connection, e error) {
	if strings.HasPrefix(url, "amqps") {
		cfg := new(tls.Config)
		cfg.RootCAs = x509.NewCertPool()

		cfg.InsecureSkipVerify = false
		if o.TLSSkipVerify == true {
			cfg.InsecureSkipVerify = true
		}

		for _, ca := range o.TLSCACerts {
			if cert, err := ioutil.ReadFile(ca); err == nil {
				cfg.RootCAs.AppendCertsFromPEM(cert)
			}
		}

		for index, cert := range o.TLSCerts {
			if cert, err := tls.LoadX509KeyPair(cert, o.TLSCertKeys[index]); err == nil {
				cfg.Certificates = append(cfg.Certificates, cert)
			}
		}

		conn, err := amqp.DialTLS(url, cfg)
		return conn, err
	}
	conn, err := amqp.Dial(url)
	return conn, err
}
