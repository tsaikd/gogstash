package outputelastic

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/tsaikd/gogstash/config"
)

func Test_main(t *testing.T) {
	assert := assert.New(t)

	conftest, err := config.LoadConfig("config_test.json")
	assert.NoError(err)

	outputs := conftest.Output()
	assert.Len(outputs, 1)
	if len(outputs) > 0 {
		output := outputs[0].(*OutputConfig)
		assert.IsType(&OutputConfig{}, output)
		assert.Equal("elastic", output.GetType())

		output.Event(config.LogEvent{
			Timestamp: time.Now(),
			Message:   "outputstdout test message",
			Extra: map[string]interface{}{
				"fieldstring": "ABC",
				"fieldnumber": 123,
			},
		})
	}
}
