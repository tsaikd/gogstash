package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_load_config_without_regist_module(t *testing.T) {
	require := require.New(t)
	require.NotNil(require)

	conf, err := LoadFromString(`{
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
	require.NoError(err)

	inputs, err := conf.getInputs(nil)
	require.Error(err)
	require.Len(inputs, 0)

	outputs, err := conf.getOutputs()
	require.Error(err)
	require.Len(outputs, 0)
}
