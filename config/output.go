package config

import (
	log "github.com/Sirupsen/logrus"
)

type TypeOutputConfig interface {
	TypeConfig
	Event(event LogEvent) (err error)
}

type OutputHandler func(mapraw map[string]interface{}) (conf TypeOutputConfig, err error)

var (
	mapOutputHandler = map[string]OutputHandler{}
)

func RegistOutputHandler(name string, handler OutputHandler) {
	mapOutputHandler[name] = handler
}

func (config *Config) Output() (outputs []TypeOutputConfig) {
	var (
		conf    TypeOutputConfig
		err     error
		handler OutputHandler
		ok      bool
	)
	for _, mapraw := range config.OutputRaw {
		if handler, ok = mapOutputHandler[mapraw["type"].(string)]; !ok {
			log.Errorf("unknown output config type %q", mapraw["type"])
			continue
		}
		if conf, err = handler(mapraw); err != nil {
			log.Errorf("handle output config failed: %q\n%s", mapraw, err)
			continue
		}
		outputs = append(outputs, conf)
	}
	return
}
