package config

import (
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/tsaikd/KDGoLib/logrusutil"
)

func Test_LoadConfig(t *testing.T) {
	assert := assert.New(t)

	logger := logrusutil.DefaultConsoleLogger
	logger.Level = logrus.DebugLevel

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
	conf.Map(logger)

	inputs, err := conf.getInputs(nil)
	assert.Error(err)
	assert.Len(inputs, 0)

	outputs, err := conf.getOutputs()
	assert.Error(err)
	assert.Len(outputs, 0)
}
