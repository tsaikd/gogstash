package outputemail

import (
	"reflect"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
)

var (
	logger = config.Logger
)

func init() {
	logger.Level = logrus.DebugLevel
	config.RegistOutputHandler(ModuleName, InitHandler)
}

func Test_main(t *testing.T) {
	require := require.New(t)
	require.NotNil(require)

	// Please fill the correct email info xxx is just a placeholder
	conf, err := config.LoadFromString(`{
		"output": [{
			"type": "email",
			"address": "xxx",
			"from": "xxx",
			"to": "xxx",
            "cc": "xxx",
			"use_tls": false,
            "port": 25,
            "username": "xxx",
            "password": "xxx",
            "subject": "outputemail test subject"
		}]
	}`)
	require.NoError(err)

	err = conf.RunOutputs()
	require.NoError(err)

	outchan := conf.Get(reflect.TypeOf(make(config.OutChan))).
		Interface().(config.OutChan)
	outchan <- logevent.LogEvent{
		Timestamp: time.Now(),
		Message:   "outputemail test message",
	}

	waitsec := 1
	logger.Infof("Wait for %d seconds", waitsec)
	time.Sleep(time.Duration(waitsec) * time.Second)
}
