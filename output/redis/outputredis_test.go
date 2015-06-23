package outputredis

import (
	"math/rand"
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
		assert.Equal("redis", output.GetType())

		output.Event(config.LogEvent{
			Timestamp: time.Now(),
			Message:   "outputredis test message",
		})

		// test random time event only
		//test_random_time_event(t, output)
	}
}

func test_random_time_event(t *testing.T, output *OutputConfig) {
	var (
		assert = assert.New(t)
		ch     = make(chan int, 5)
	)

	rand.Seed(time.Now().UnixNano())
	for j := 0; j < 5; j++ {
		go func() {
			for i := 1; i < 120; i++ {
				assert.NoError(output.Event(config.LogEvent{
					Timestamp: time.Now(),
					Message:   "outputredis test message",
				}))

				time.Sleep(time.Duration(rand.Intn(2000)) * time.Millisecond)
			}
			ch <- j
		}()
	}
	for j := 0; j < 5; j++ {
		<-ch
	}

}
