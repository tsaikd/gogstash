package outputreport

import (
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/tsaikd/gogstash/config"
)

func Test_main(t *testing.T) {
	var (
		assert   = assert.New(t)
		err      error
		conftest config.Config
	)

	log.SetLevel(log.DebugLevel)

	conftest, err = config.LoadConfig("config_test.json")
	assert.NoError(err)

	outputs := conftest.Output()
	assert.Len(outputs, 1)
	if len(outputs) > 0 {
		output := outputs[0].(*OutputConfig)
		assert.IsType(&OutputConfig{}, output)
		assert.Equal("report", output.GetType())
		assert.Equal(1, output.Interval)

		event := config.LogEvent{
			Timestamp: time.Now(),
			Message:   "outputreport test message",
		}

		output.Event(event)
		output.Event(event)
		time.Sleep(2 * time.Second)

		output.Event(event)
		time.Sleep(2 * time.Second)

		output.Event(event)
		output.Event(event)
		output.Event(event)
		output.Event(event)
		time.Sleep(2 * time.Second)
	}
}
