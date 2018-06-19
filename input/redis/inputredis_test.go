package inputredis

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
	"github.com/tsaikd/gogstash/output/redis"
)

var (
	logger = config.Logger
)

func init() {
	logger.Level = logrus.DebugLevel
	config.RegistInputHandler(ModuleName, InitHandler)
	config.RegistOutputHandler(outputredis.ModuleName, outputredis.InitHandler)
}

func Test_input_redis_module(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	// write test event to redis
	ctx := context.Background()
	confWrite, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
output:
  - type: redis
    host:
      - localhost:6379
    key: gogstash-test
    data_type: list
	`)))
	require.NoError(err)
	err = confWrite.Start(ctx)
	if err != nil {
		require.True(outputredis.ErrorPingFailed.In(err))
		t.Skip("skip test input redis module")
	}

	timestamp := time.Now()
	confWrite.TestInputEvent(logevent.LogEvent{
		Timestamp: timestamp,
		Message:   "inputredis test message",
	})

	if event, err2 := confWrite.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err2) {
		require.Equal(timestamp.UnixNano(), event.Timestamp.UnixNano())
		require.Equal("inputredis test message", event.Message)
	}

	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
input:
  - type: redis
    host: localhost:6379
    key: gogstash-test
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	time.Sleep(500 * time.Millisecond)
	if event, err := conf.TestGetOutputEvent(100 * time.Millisecond); assert.NoError(err) {
		require.Equal(timestamp.UnixNano(), event.Timestamp.UnixNano())
		require.Equal("inputredis test message", event.Message)
	}
}
