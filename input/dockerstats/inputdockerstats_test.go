package inputdockerstats

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
			"type": "dockerstats",
			"dockerurl": "unix:///var/run/docker.sock",
			"stat_interval": 3
		}]
	}`)
	assert.NoError(err)

	err = conf.RunInputs()
	assert.NoError(err)

	waitsec := 10
	logger.Debugf("Wait for %d seconds", waitsec)
	time.Sleep(time.Duration(waitsec) * time.Second)
}
