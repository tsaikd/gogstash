package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadFromJSON(t *testing.T) {
	require := require.New(t)
	require.NotNil(require)

	conf, err := LoadFromJSON([]byte(`{
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
	}`))
	require.NoError(err)

	require.NotNil(conf.chInFilter)
	require.NotNil(conf.chFilterOut)
	require.Nil(conf.ctx)
	require.Nil(conf.eg)
	require.Len(conf.InputRaw, 2)
	require.Len(conf.FilterRaw, 0)
	require.Len(conf.OutputRaw, 1)

	inputs, err := conf.getInputs()
	require.Error(err)
	require.Len(inputs, 0)

	filters, err := conf.getFilters()
	require.NoError(err)
	require.Len(filters, 0)

	outputs, err := conf.getOutputs()
	require.Error(err)
	require.Len(outputs, 0)

	conf, err = LoadFromJSON([]byte(`{
		"input": [{
			"type": "exec",
			"command": "uptime",
			"args": [],
			"interval": 3,
			"message_prefix": "%{@timestamp} "
		}],
		"output": [{
			"type": "stdout"
		}],
	}`))
	require.Error(err)
}

func TestLoadFromYAML(t *testing.T) {
	require := require.New(t)
	require.NotNil(require)

	conf, err := LoadFromYAML([]byte(strings.TrimSpace(`
input:
  - type: exec
    command: uptime
    args: []
    interval: 3
    message_prefix: "%{@timestamp} "
  - type: exec
    command: whoami
    args: []
    interval: 4
    message_prefix: "%{@timestamp} "
output:
  - type: stdout
	`)))
	require.NoError(err)

	require.NotNil(conf.chInFilter)
	require.NotNil(conf.chFilterOut)
	require.Nil(conf.ctx)
	require.Nil(conf.eg)
	require.Len(conf.InputRaw, 2)
	require.Len(conf.FilterRaw, 0)
	require.Len(conf.OutputRaw, 1)

	inputs, err := conf.getInputs()
	require.Error(err)
	require.Len(inputs, 0)

	filters, err := conf.getFilters()
	require.NoError(err)
	require.Len(filters, 0)

	outputs, err := conf.getOutputs()
	require.Error(err)
	require.Len(outputs, 0)
}
