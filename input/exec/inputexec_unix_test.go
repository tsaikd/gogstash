package inputexec

import (
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/tsaikd/gogstash/config"
)

var (
	logger = config.Logger
)

func init() {
	logger.Level = logrus.DebugLevel
	config.RegistInputHandler(ModuleName, InitHandler)
}

func Test_text(t *testing.T) {
	require := require.New(t)
	require.NotNil(require)

	conf, err := config.LoadFromJSON([]byte(`{
		"input": [{
			"type": "exec",
			"command": "uptime",
			"args": [],
			"interval": 1,
			"message_prefix": "%{@timestamp} "
		},{
			"type": "exec",
			"command": "whoami",
			"args": [],
			"interval": 3,
			"message_prefix": "%{@timestamp} "
		}]
	}`))
	require.NoError(err)

	err = conf.RunInputs()
	require.NoError(err)

	waitsec := 7
	logger.Infof("Wait for %d seconds", waitsec)
	time.Sleep(time.Duration(waitsec) * time.Second)
}

func Test_json(t *testing.T) {
	require := require.New(t)
	require.NotNil(require)

	conf, err := config.LoadFromJSON([]byte(`{
		"input": [{
			"type": "exec",
			"command": "./test_json.sh",
			"interval": 1,
			"message_type": "json"
		}]
	}`))
	require.NoError(err)

	err = conf.RunInputs()
	require.NoError(err)

	waitsec := 3
	logger.Infof("Wait for %d seconds", waitsec)
	time.Sleep(time.Duration(waitsec) * time.Second)
}
