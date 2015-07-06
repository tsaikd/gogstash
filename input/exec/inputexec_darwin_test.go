package inputexec

import (
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/tsaikd/KDGoLib/logrusutil"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
)

func Test_main(t *testing.T) {
	assert := assert.New(t)

	logger := logrusutil.DefaultConsoleLogger
	logger.Level = logrus.DebugLevel
	config.RegistInputHandler(ModuleName, InitHandler)

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
		}]
	}`)
	assert.NoError(err)
	conf.Map(logger)

	eventChan := make(chan logevent.LogEvent, 10)
	inputs, err := conf.Input(eventChan)
	assert.NoError(err)
	assert.Len(inputs, 2)

	if len(inputs) > 0 {
		for _, inputtype := range inputs {
			input := inputtype.(*InputConfig)
			assert.Equal("exec", input.GetType())

			go input.Start()
		}

		go func() {
			for {
				<-eventChan
			}
		}()

		waitsec := 10
		logger.Infof("Wait for %d seconds", waitsec)
		time.Sleep(time.Duration(waitsec) * time.Second)
	}
}
