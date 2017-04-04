package cmd

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tsaikd/gogstash/config"
)

func Test_gogstash(t *testing.T) {
	require := require.New(t)
	require.NotNil(require)

	logger := config.Logger

	conf, err := config.LoadFromJSON([]byte(`{
		"input": [{
			"type": "exec",
			"command": "uptime",
			"args": [],
			"interval": 2,
			"message_prefix": "%{@timestamp} "
		},{
			"type": "exec",
			"command": "whoami",
			"args": [],
			"interval": 3,
			"message_prefix": "%{@timestamp} "
		}],
		"output": [{
			"type": "stdout"
		}]
	}`))
	require.NoError(err)

	_, err = conf.Invoke(conf.RunInputs)
	require.NoError(err)

	_, err = conf.Invoke(conf.RunOutputs)
	require.NoError(err)

	waitsec := 5
	logger.Infof("Wait for %d seconds", waitsec)
	time.Sleep(time.Duration(waitsec) * time.Second)
}
