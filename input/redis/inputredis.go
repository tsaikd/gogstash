package inputredis

import (
	"context"
	"strings"
	"time"

	"github.com/tsaikd/KDGoLib/errutil"
	codecjson "github.com/tsaikd/gogstash/codec/json"
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
	DB          int    `json:"db"`          // redis db, default: 0
	Password    string `json:"password"`    // redis password, default: ""
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
		DB:              0,
		Password:		 "",
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
func InitHandler(
	ctx context.Context,
	raw config.ConfigRaw,
	control config.Control,
) (config.TypeInputConfig, error) {
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
		DB:       conf.DB,
		Password: conf.Password,
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

	conf.Codec, err = config.GetCodec(ctx, raw["codec"], codecjson.ModuleName)
	if err != nil {
		return nil, err
	}

	return &conf, nil
}

func (i *InputConfig) queueMessage(
	ctx context.Context,
	message string,
	msgChan chan<- logevent.LogEvent,
) (err error) {
	_, err = i.Codec.Decode(ctx, []byte(message), nil, []string{}, msgChan)

	return
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
	return i.queueMessage(ctx, result[1], msgChan)
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
			if err := i.queueMessage(ctx, result.(string), msgChan); err != nil {
				return err
			}
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
