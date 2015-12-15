package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_LoadConfig(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)

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
	assert.NoError(err)

	inputs, err := conf.getInputs(nil)
	assert.Error(err)
	assert.Len(inputs, 0)

	outputs, err := conf.getOutputs()
	assert.Error(err)
	assert.Len(outputs, 0)
}

func Test_FormatWithEnv(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)

	path := FormatWithEnv("%{PATH}")
	assert.NotEqual("%{PATH}", path)
}
