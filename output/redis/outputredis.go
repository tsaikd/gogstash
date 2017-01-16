package outputredis

import (
	"errors"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/fzzy/radix/redis"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
)

// ModuleName is the name used in config file
const ModuleName = "redis"

// OutputConfig holds the configuration json fields and internal objects
type OutputConfig struct {
	config.OutputConfig
	Key               string   `json:"key"`
	Host              []string `json:"host"`
	DataType          string   `json:"data_type,omitempty"` // one of ["list", "channel"]
	Timeout           int      `json:"timeout,omitempty"`
	ReconnectInterval int      `json:"reconnect_interval,omitempty"`

	clients []*redis.Client // all configured clients
	client  *redis.Client   // cache last success client
	evchan  chan logevent.LogEvent
}

// DefaultOutputConfig returns an OutputConfig struct with default values
func DefaultOutputConfig() OutputConfig {
	return OutputConfig{
		OutputConfig: config.OutputConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		Key:               "gogstash",
		DataType:          "list",
		Timeout:           5,
		ReconnectInterval: 1,

		evchan: make(chan logevent.LogEvent),
	}
}

// InitHandler initialize the output plugin
func InitHandler(confraw *config.ConfigRaw) (retconf config.TypeOutputConfig, err error) {
	conf := DefaultOutputConfig()
	if err = config.ReflectConfig(confraw, &conf); err != nil {
		return
	}

	go conf.loop()
	if err = conf.initRedisClient(); err != nil {
		return
	}

	retconf = &conf
	return
}

func (self *OutputConfig) Event(event logevent.LogEvent) (err error) {
	self.evchan <- event
	return
}

func (self *OutputConfig) loop() (err error) {
	for {
		event := <-self.evchan
		self.sendEvent(event)
	}
}

func (self *OutputConfig) initRedisClient() (err error) {
	var (
		client *redis.Client
	)

	self.closeRedisClient()

	for _, addr := range self.Host {
		if client, err = redis.DialTimeout("tcp", addr, time.Duration(self.Timeout)*time.Second); err == nil {
			self.clients = append(self.clients, client)
		} else {
			log.Warnf("Redis connection failed: %q\n%s", addr, err)
		}
	}

	if len(self.clients) > 0 {
		self.client = self.clients[0]
		err = nil
	} else {
		self.client = nil
		err = errors.New("no valid redis server connection")
	}

	return
}

func (self *OutputConfig) closeRedisClient() (err error) {
	var (
		client *redis.Client
	)

	for _, client = range self.clients {
		client.Close()
	}

	self.clients = self.clients[:0]

	return
}

func (self *OutputConfig) sendEvent(event logevent.LogEvent) (err error) {
	var (
		client *redis.Client
		raw    []byte
		key    string
	)

	if raw, err = event.MarshalJSON(); err != nil {
		log.Errorf("event Marshal failed: %v", event)
		return
	}
	key = event.Format(self.Key)

	// try previous client first
	if self.client != nil {
		if err = self.redisSend(self.client, key, raw); err == nil {
			return
		}
	}

	// try to log forever
	for {
		// reconfig all clients
		if err = self.initRedisClient(); err != nil {
			return
		}

		// find valid client
		for _, client = range self.clients {
			if err = self.redisSend(client, key, raw); err == nil {
				self.client = client
				return
			}
		}

		time.Sleep(time.Duration(self.ReconnectInterval) * time.Second)
	}
}

func (self *OutputConfig) redisSend(client *redis.Client, key string, raw []byte) (err error) {
	var (
		res *redis.Reply
	)

	switch self.DataType {
	case "list":
		res = client.Cmd("rpush", key, raw)
		err = res.Err
	case "channel":
		res = client.Cmd("publish", key, raw)
		err = res.Err
	default:
		err = errors.New("unknown DataType: " + self.DataType)
	}

	return
}
