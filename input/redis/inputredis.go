package inputredis

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/tsaikd/KDGoLib/errutil"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
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
	BatchCount  int    `json:"batch_count"` // The number of events to return from Redis using EVAL, default: 125

	// BlockingTimeout used for set the blocking timeout interval in redis BLPOP command
	// Defaults to 600s
	BlockingTimeout string `json:"blocking_timeout,omitempty"` // automatically
	blockingTimeout time.Duration

	client         *redis.Client
	batchScriptSha string
}

// DefaultInputConfig returns an InputConfig struct with default values
func DefaultInputConfig() InputConfig {
	return InputConfig{
		InputConfig: config.InputConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		Host:            "localhost:6379",
		Key:             "gogstash",
		Connections:     10,
		BatchCount:      125,
		BlockingTimeout: "600s",
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

	conf.blockingTimeout, err = time.ParseDuration(conf.BlockingTimeout)
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

	if conf.BatchCount > 1 {
		err = conf.loadBatchScript()
		if err != nil {
			return nil, err
		}
	}

	return &conf, nil
}

func queueMessage(message string, msgChan chan<- logevent.LogEvent) {
	event := logevent.LogEvent{
		Timestamp: time.Now(),
		Message:   message,
		Extra:     map[string]interface{}{},
	}

	if err := json.Unmarshal([]byte(event.Message), &event.Extra); err != nil {
		event.AddTag(ErrorTag)
		goglog.Logger.Error(err)
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

func (i *InputConfig) listSingle(ctx context.Context, msgChan chan<- logevent.LogEvent) error {
	result, err := i.client.BLPop(i.blockingTimeout, i.Key).Result()
	if err != nil {
		switch err {
		case redis.Nil: // BLPOP timeout
			return nil
		default:
			return err
		}
	}

	// we need to use msg[1] because BLPOP returns a tuple of (key, value) where key is
	// the redis key used to retrieve the message
	queueMessage(result[1], msgChan)

	return nil
}

const batchEmptySleep = time.Duration(250000000) // 250ms

func (i *InputConfig) listBatch(ctx context.Context, msgChan chan<- logevent.LogEvent) error {
retry:
	r, err := i.client.EvalSha(i.batchScriptSha, []string{i.Key}, i.BatchCount-1).Result()
	if err != nil {
		if strings.Contains(err.Error(), "NOSCRIPT") {
			// redis server may have been restarted, try reloading batch EVAL script
			goglog.Logger.Warnf("%s: %v", ModuleName, err)
			err = i.loadBatchScript()
			if err != nil {
				return err
			}
			goto retry
		}
		return err
	}

	switch results := r.(type) {
	case []interface{}:
		for _, result := range results {
			queueMessage(result.(string), msgChan)
		}
		if len(results) <= 0 {
			time.Sleep(time.Duration(batchEmptySleep))
		}
	}
	return nil
}

func (i *InputConfig) loadBatchScript() (err error) {
	i.batchScriptSha, err = i.client.ScriptLoad(`
		local batchsize = tonumber(ARGV[1])
    local result = redis.call('lrange', KEYS[1], 0, batchsize)
    redis.call('ltrim', KEYS[1], batchsize + 1, -1)
    return result
	`).Result()
	return
}

// Start wraps the actual function starting the plugin
func (i *InputConfig) Start(ctx context.Context, msgChan chan<- logevent.LogEvent) error {
	var err error

	for {
		select {
		case <-ctx.Done():
			goglog.Logger.Info("input redis stopped")
			return nil
		default:
		}

		if i.BatchCount > 1 {
			err = i.listBatch(ctx, msgChan)
		} else {
			err = i.listSingle(ctx, msgChan)
		}
		if err != nil {
			return err
		}
	}
}
