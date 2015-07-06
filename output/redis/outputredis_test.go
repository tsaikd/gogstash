package outputredis

import (
	"math/rand"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/tsaikd/KDGoLib/logrusutil"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
)

func Test_main(t *testing.T) {
	assert := assert.New(t)

	logger := logrusutil.DefaultConsoleLogger
	logger.Level = logrus.DebugLevel
	config.RegistOutputHandler(ModuleName, InitHandler)

	conf, err := config.LoadFromString(`{
		"output": [{
			"type": "redis",
			"host": ["127.0.0.1:6379"]
		}]
	}`)
	assert.NoError(err)
	conf.Map(logger)

	evchan := make(chan logevent.LogEvent, 10)
	conf.Map(evchan)

	err = conf.RunOutputs(evchan, logger)
	assert.NoError(err)

	evchan <- logevent.LogEvent{
		Timestamp: time.Now(),
		Message:   "outputstdout test message",
	}

	// test random time event only
	//test_random_time_event(t, output)
}

func test_random_time_event(t *testing.T, evchan chan logevent.LogEvent) {
	ch := make(chan int, 5)

	rand.Seed(time.Now().UnixNano())
	for j := 0; j < 5; j++ {
		go func() {
			for i := 1; i < 120; i++ {
				evchan <- logevent.LogEvent{
					Timestamp: time.Now(),
					Message:   "outputredis test message",
				}

				time.Sleep(time.Duration(rand.Intn(2000)) * time.Millisecond)
			}
			ch <- j
		}()
	}
	for j := 0; j < 5; j++ {
		<-ch
	}

}
