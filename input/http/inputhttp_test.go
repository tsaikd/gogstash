package inputhttp

import (
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/tsaikd/gogstash/config"
)

var (
	logger = config.Logger
)

func init() {
	logger.Level = logrus.DebugLevel
	config.RegistInputHandler(ModuleName, InitHandler)
}

func Test_main(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)

	conf, err := config.LoadFromString(`{
		"input": [{
			"type": "http",
			"method": "GET",
			"url": "http://127.0.0.1/",
			"interval": 3
		}]
	}`)
	assert.NoError(err)

	err = conf.RunInputs()
	assert.NoError(err)

	waitsec := 10
	logger.Infof("Wait for %d seconds", waitsec)
	time.Sleep(time.Duration(waitsec) * time.Second)
}
