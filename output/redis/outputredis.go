package outputredis

import (
	"encoding/json"
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/fzzy/radix/redis"

	"github.com/tsaikd/gogstash/config"
)

type OutputConfig struct {
	config.CommonConfig
	Key      string   `json:"key"`
	Host     []string `json:"host"`
	DataType string   `json:"data_type,omitempty"` // one of ["list", "channel"], TODO: implement channel mode
	Timeout  int      `json:"timeout,omitempty"`
	client   *redis.Client
}

func DefaultOutputConfig() OutputConfig {
	return OutputConfig{
		CommonConfig: config.CommonConfig{
			Type: "redis",
		},
		Key:      "logstash-test",
		DataType: "list",
		Timeout:  5,
	}
}

func init() {
	config.RegistOutputHandler("redis", func(mapraw map[string]interface{}) (conf config.TypeOutputConfig, err error) {
		var (
			raw []byte
		)
		if raw, err = json.Marshal(mapraw); err != nil {
			log.Error(err)
			return
		}
		defconf := DefaultOutputConfig()
		conf = &defconf
		if err = json.Unmarshal(raw, &conf); err != nil {
			log.Error(err)
			return
		}
		return
	})
}

func (self *OutputConfig) Type() string {
	return self.CommonConfig.Type
}

func (self *OutputConfig) Event(event config.LogEvent) (err error) {
	var (
		raw []byte
		res *redis.Reply
		key string
	)

	if self.client == nil {
		for _, addr := range self.Host {
			if self.client, err = redis.DialTimeout("tcp", addr, time.Duration(self.Timeout)*time.Second); err != nil {
				log.Errorf("Redis connection failed: %q\n%s", addr, err)
				return
			}
		}
	}

	if self.client == nil {
		err = fmt.Errorf("no valid redis server connection")
		return
	}

	if raw, err = event.Marshal(); err != nil {
		log.Errorf("event Marshal failed: %v", event)
		return
	}
	key = event.Format(self.Key)
	res = self.client.Cmd("rpush", key, raw)
	log.Debug("redisoutput", res, event)
	return
}
