package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_LoadConfig(t *testing.T) {
	var (
		assert = assert.New(t)
		err    error
		config Config
	)

	config, err = LoadConfig("config_test.json")
	assert.NoError(err)

	inputs := config.Input()
	assert.Len(inputs, 0)

	outputs := config.Output()
	assert.Len(outputs, 0)

	t.Log("Previous error log is correct, because of no import necessary module in this testing")
}
