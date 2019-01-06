package inputredis

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	codecjson "github.com/tsaikd/gogstash/codec/json"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
)

var s *miniredis.Miniredis
var timeNow time.Time

func init() {
	goglog.Logger.SetLevel(logrus.DebugLevel)
	config.RegistInputHandler(ModuleName, InitHandler)
	config.RegistCodecHandler(codecjson.ModuleName, codecjson.InitHandler)
}

func TestMain(m *testing.M) {
	// initialize redis server
	s = miniredis.NewMiniRedis()
	err := s.StartAddr("localhost:6380") // change the port
	if err != nil {
		panic(err)
	}
	defer s.Close()

	timeNow = time.Now().UTC()
	ret := m.Run()

	os.Exit(ret)
}

func Test_input_redis_module_batch(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	for i := 0; i < 10; i++ {
		s.Lpush("gogstash-test", fmt.Sprintf("{\"@timestamp\":\"%s\",\"message\":\"inputredis test message\"}", timeNow.Format(time.RFC3339Nano)))
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
input:
  - type: redis
    host: localhost:6380
    key: gogstash-test
    connections: 1
    batch_count: 10
    codec: json
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	time.Sleep(500 * time.Millisecond)
	for i := 0; i < 10; i++ {
		if event, err := conf.TestGetOutputEvent(100 * time.Millisecond); assert.NoError(err) {
			require.Equal(timeNow.UnixNano(), event.Timestamp.UnixNano())
			require.Equal("inputredis test message", event.Message)
		}
	}
}

func Test_input_redis_module_single(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	s.Lpush("gogstash-test", fmt.Sprintf("{\"@timestamp\":\"%s\",\"message\":\"inputredis test message\"}", timeNow.Format(time.RFC3339Nano)))

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
input:
  - type: redis
    host: localhost:6380
    key: gogstash-test
    batch_count: 1
    blocking_timeout: 5s
    codec: json
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	time.Sleep(500 * time.Millisecond)
	if event, err := conf.TestGetOutputEvent(100 * time.Millisecond); assert.NoError(err) {
		require.Equal(timeNow.UnixNano(), event.Timestamp.UnixNano())
		require.Equal("inputredis test message", event.Message)
	}
}
