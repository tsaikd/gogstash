package outputredis

import (
	"context"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
)

func init() {
	goglog.Logger.SetLevel(logrus.DebugLevel)
	config.RegistOutputHandler(ModuleName, InitHandler)
}

func Test_output_redis_module(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
output:
  - type: redis
    host:
      - localhost:6379
    key: gogstash-test
    data_type: list
	`)))
	require.NoError(err)
	err = conf.Start(ctx)
	if err != nil {
		t.Log("skip test output redis module")
		require.True(ErrorPingFailed.In(err))
		return
	}

	conf.TestInputEvent(logevent.LogEvent{
		Timestamp: time.Now(),
		Message:   "outputredis test message",
	})

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal("outputredis test message", event.Message)
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
