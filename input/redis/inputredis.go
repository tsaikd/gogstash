package inputredis

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/fzzy/radix/redis"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
)

// ModuleName is the name used in config file
const ModuleName = "redis"

const invalidJsonError = "Invalid JSON received from Redis input. Decoder error: %+v"

// InputConfig holds the configuration json fields and internal objects
type InputConfig struct {
	config.InputConfig
	Key               string `json:"key"`
	Host              string `json:"host"`
	DataType          string `json:"data_type,omitempty"` // one of ["list", "channel"] TODO!`
	Connections       int    `json:"connections"`
	ReconnectInterval int    `json:"reconnect_interval,omitempty"` // TODO!

	clients []*redis.Client // all configured clients
}

// DefaultInputConfig returns an InputConfig struct with default values
func DefaultInputConfig() InputConfig {
	return InputConfig{
		InputConfig: config.InputConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		Key:               "gogstash",
		DataType:          "list",
		Connections:       5,
		ReconnectInterval: 1,
	}
}

// InitHandler initialize the input plugin
func InitHandler(confraw *config.ConfigRaw) (config.TypeInputConfig, error) {
	conf := DefaultInputConfig()
	if err := config.ReflectConfig(confraw, &conf); err != nil {
		return nil, err
	}

	if err := conf.initRedisClients(); err != nil {
		return nil, err
	}

	return &conf, nil
}

func (i *InputConfig) closeRedisClients() (err error) {
	for _, client := range i.clients {
		client.Close()
	}
	i.clients = i.clients[:0]
	return
}

// Open connections to redis server
func (i *InputConfig) initRedisClients() error {
	i.closeRedisClients()

	for j := 0; j < i.Connections; j++ {
		if client, err := redis.Dial("tcp", i.Host); err == nil {
			i.clients = append(i.clients, client)
		} else {
			logrus.Fatalf("Redis connection to '%s' failed: %s", i.Host, err)
			return err
		}
	}

	return nil
}

// Start wraps the actual function starting the plugin
func (i *InputConfig) Start() {
	i.Invoke(i.start)
}

// spawn all input handlers
func (i *InputConfig) start(logger *logrus.Logger, inchan config.InChan) {
	for _, client := range i.clients {
		logger.Debug("Spawning redis input handler")
		go i.inputHandler(logger, inchan, client)
	}
}

// launch redis input handler to listen for incoming messages and publish them on the InChan
func (i *InputConfig) inputHandler(logger *logrus.Logger, inchan config.InChan, client *redis.Client) {
	logger.Debugf("Started input handler")
	for {
		// wait forever for a message from redis.
		reply := client.Cmd("blpop", i.Key, 0)
		if reply.Err != nil {
			// something bad happened with the blpop?
			logger.Errorf("Error with redis input client: %+v", reply.Err)
			break
		}

		var jsonMsg map[string]interface{}

		msg, err := reply.ListBytes()
		if err != nil {
			// I got no idea, but somehow ListBytes gave me an error.
			logger.Errorf("Error retrieving bytes from redis message: %+v", err)
			break
		}

		// we need to use msg[1] because BLPOP returns a tuple of (key,value) where key is
		// the redis key used to retrieve the message
		dec := json.NewDecoder(bytes.NewReader(msg[1]))

		// attempt to decode post body, if it fails, log it.
		if err := dec.Decode(&jsonMsg); err != nil {
			logger.Errorf(invalidJsonError, err)
			logger.Debugf("Invalid JSON: '%s'", msg[1])
			continue
		}

		event := logevent.LogEvent{Extra: jsonMsg}

		// try to fill basic log event by json message
		if value, ok := jsonMsg["message"]; ok {
			switch v := value.(type) {
			case string:
				event.Message = v
			}
		}
		if value, ok := jsonMsg["@timestamp"]; ok {
			switch v := value.(type) {
			case string:
				event.Timestamp, _ = time.Parse(time.RFC3339Nano, v)
			}
		}

		// send the event as it came to us
		inchan <- event
	}

	// Somehow I encountered an error and died. Don't stay dead!
	// TODO: count errors, and fatalf if error count exceeds retries
	logger.Errorf("Redis inputHandler thread died, respawning!")
	go i.inputHandler(logger, inchan, client)
}
