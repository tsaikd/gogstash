package outputredis

import (
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
)

var (
	logger = config.Logger
)

func init() {
	logger.Level = logrus.DebugLevel
	config.RegistOutputHandler(ModuleName, InitHandler)
}

func Test_main(t *testing.T) {
	require := require.New(t)
	require.NotNil(require)

	conf, err := config.LoadFromString(`{
		"output": [{
			"type": "redis",
			"host": ["127.0.0.1:6379"]
		}]
	}`)
	require.NoError(err)

	err = conf.RunOutputs()
	require.NoError(err)

	evchan := conf.Get(reflect.TypeOf(make(chan logevent.LogEvent))).
		Interface().(chan logevent.LogEvent)
	evchan <- logevent.LogEvent{
		Timestamp: time.Now(),
		Message:   "outputstdout test message",
	}

	// test random time event only
	//testRandomTimeEvent(t, evchan)
}

func testRandomTimeEvent(t *testing.T, evchan chan logevent.LogEvent) {
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
