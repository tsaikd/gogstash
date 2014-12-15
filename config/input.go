package config

import (
	log "github.com/Sirupsen/logrus"
)

type TypeInputConfig interface {
	TypeConfig
	Event(evchan chan LogEvent) (err error)
}

type InputHandler func(mapraw map[string]interface{}) (conf TypeInputConfig, err error)

var (
	mapInputHandler = map[string]InputHandler{}
)

func RegistInputHandler(name string, handler InputHandler) {
	mapInputHandler[name] = handler
}

func (config *Config) Input() (inputs []TypeInputConfig) {
	var (
		conf    TypeInputConfig
		err     error
		handler InputHandler
		ok      bool
	)
	for _, mapraw := range config.InputRaw {
		if handler, ok = mapInputHandler[mapraw["type"].(string)]; !ok {
			log.Errorf("unknown input config type %q", mapraw["type"])
			continue
		}
		if conf, err = handler(mapraw); err != nil {
			log.Errorf("handle input config failed: %q\n%s", mapraw, err)
			continue
		}
		inputs = append(inputs, conf)
	}
	return
}
