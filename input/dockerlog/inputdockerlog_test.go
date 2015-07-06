package inputdockerlog

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
			"type": "dockerlog",
			"dockerurl": "unix:///var/run/docker.sock"
		}]
	}`)
	assert.NoError(err)
	conf.Map(logger)

	evchan := make(chan logevent.LogEvent, 10)
	conf.Map(evchan)

	err = conf.RunInputs(evchan)
	assert.NoError(err)

	go func() {
		for {
			event := <-evchan
			logger.Debugln(event)
		}
	}()

	waitsec := 10
	logger.Debugf("Wait for %d seconds", waitsec)
	time.Sleep(time.Duration(waitsec) * time.Second)
}
