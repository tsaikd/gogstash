package gogstash

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
)

func Test_main(t *testing.T) {
	assert := assert.New(t)

	conf, err := config.LoadFromString(`{
		"input": [{
			"type": "exec",
			"command": "uptime",
			"args": [],
			"interval": 3,
			"message_prefix": "%{@timestamp} "
		},{
			"type": "exec",
			"command": "whoami",
			"args": [],
			"interval": 4,
			"message_prefix": "%{@timestamp} "
		}],
		"output": [{
			"type": "stdout"
		}]
	}`)
	assert.NoError(err)
	conf.Map(logger)

	evchan := make(chan logevent.LogEvent, 100)
	conf.Map(evchan)

	_, err = conf.Invoke(conf.RunInputs)
	assert.NoError(err)

	_, err = conf.Invoke(conf.RunOutputs)
	assert.NoError(err)

	waitsec := 10
	logger.Infof("Wait for %d seconds", waitsec)
	time.Sleep(time.Duration(waitsec) * time.Second)
}
