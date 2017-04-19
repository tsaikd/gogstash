package inputredis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/tsaikd/KDGoLib/errutil"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
	"gopkg.in/redis.v5"
)

// ModuleName is the name used in config file
const ModuleName = "redis"

// ErrorTag tag added to event when process module failed
const ErrorTag = "gogstash_input_redis_error"

// InputConfig holds the configuration json fields and internal objects
type InputConfig struct {
	config.InputConfig
	Host        string `json:"host"`        // redis server host:port, default: "localhost:6379"
	Key         string `json:"key"`         // where to get data, default: "gogstash"
	Connections int    `json:"connections"` // maximum number of socket connections, default: 10

	client *redis.Client
}

// DefaultInputConfig returns an InputConfig struct with default values
func DefaultInputConfig() InputConfig {
	return InputConfig{
		InputConfig: config.InputConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		Host:        "localhost:6379",
		Key:         "gogstash",
		Connections: 10,
	}
}

// errors
var (
	ErrorPingFailed = errutil.NewFactory("ping redis server failed")
)

// InitHandler initialize the input plugin
func InitHandler(ctx context.Context, raw *config.ConfigRaw) (config.TypeInputConfig, error) {
	conf := DefaultInputConfig()
	err := config.ReflectConfig(raw, &conf)
	if err != nil {
		return nil, err
	}

	conf.client = redis.NewClient(&redis.Options{
		Addr:     conf.Host,
		PoolSize: conf.Connections,
	})
	conf.client = conf.client.WithContext(ctx)

	if _, err := conf.client.Ping().Result(); err != nil {
		return nil, ErrorPingFailed.New(err)
	}

	return &conf, nil
}

// Start wraps the actual function starting the plugin
func (i *InputConfig) Start(ctx context.Context, msgChan chan<- logevent.LogEvent) error {
	logger := config.Logger

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		result, err := i.client.BLPop(600*time.Second, i.Key).Result()
		if err != nil {
			switch err {
			case redis.Nil: // BLPOP timeout
				continue
			default:
				return err
			}
		}

		event := logevent.LogEvent{
			Timestamp: time.Now(),
			// we need to use msg[1] because BLPOP returns a tuple of (key,value) where key is
			// the redis key used to retrieve the message
			Message: result[1],
			Extra:   map[string]interface{}{},
		}

		if err = json.Unmarshal([]byte(event.Message), &event.Extra); err != nil {
			event.AddTag(ErrorTag)
			logger.Error(err)
		}

		// try to fill basic log event by json message
		if value, ok := event.Extra["message"]; ok {
			switch v := value.(type) {
			case string:
				event.Message = v
			}
		}
		if value, ok := event.Extra["@timestamp"]; ok {
			switch v := value.(type) {
			case string:
				if timestamp, err := time.Parse(time.RFC3339Nano, v); err == nil {
					event.Timestamp = timestamp
				}
			}
		}

		msgChan <- event
	}
}
