package outputredis

import (
	"errors"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/fzzy/radix/redis"

	"github.com/tsaikd/gogstash/config"
)

const (
	ModuleName = "redis"
)

type OutputConfig struct {
	config.CommonConfig
	Key               string   `json:"key"`
	Host              []string `json:"host"`
	DataType          string   `json:"data_type,omitempty"` // one of ["list", "channel"]
	Timeout           int      `json:"timeout,omitempty"`
	ReconnectInterval int      `json:"reconnect_interval,omitempty"`

	clients []*redis.Client // all configured clients
	client  *redis.Client   // cache last success client
	chEvent chan config.LogEvent
}

func DefaultOutputConfig() OutputConfig {
	return OutputConfig{
		CommonConfig: config.CommonConfig{
			Type: ModuleName,
		},
		Key:               "gogstash",
		DataType:          "list",
		Timeout:           5,
		ReconnectInterval: 1,

		chEvent: make(chan config.LogEvent),
	}
}

func init() {
	config.RegistOutputHandler(ModuleName, func(mapraw map[string]interface{}) (retconf config.TypeOutputConfig, err error) {
		conf := DefaultOutputConfig()
		if err = config.ReflectConfig(mapraw, &conf); err != nil {
			return
		}

		go conf.loop()
		if err = conf.initRedisClient(); err != nil {
			return
		}

		retconf = &conf
		return
	})
}

func (self *OutputConfig) Event(event config.LogEvent) (err error) {
	self.chEvent <- event
	return
}

func (self *OutputConfig) loop() (err error) {
	var (
		event config.LogEvent
	)

	for {
		event = <-self.chEvent
		self.sendEvent(event)
	}

	return
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

func (self *OutputConfig) sendEvent(event config.LogEvent) (err error) {
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

	return
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
