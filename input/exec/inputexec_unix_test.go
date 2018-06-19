package inputexec

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsaikd/gogstash/config"
)

var (
	logger = config.Logger
)

func init() {
	logger.Level = logrus.DebugLevel
	config.RegistInputHandler(ModuleName, InitHandler)
}

func Test_input_exec_module(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
input:
  - type: exec
    command: uptime
    interval: 1
    message_prefix: "%{@timestamp} test_uptime "
  - type: exec
    command: whoami
    interval: 3
    message_prefix: "%{@timestamp} test_whoami "
	`)))
	require.NoError(err)
	start := time.Now()
	require.NoError(conf.Start(ctx))

	time.Sleep(500 * time.Millisecond)
	if event, err := conf.TestGetOutputEvent(100 * time.Millisecond); assert.NoError(err) {
		require.WithinDuration(start, event.Timestamp, 300*time.Millisecond)
	}
	if event, err := conf.TestGetOutputEvent(100 * time.Millisecond); assert.NoError(err) {
		require.WithinDuration(start, event.Timestamp, 300*time.Millisecond)
	}

	time.Sleep(1000 * time.Millisecond)
	if event, err := conf.TestGetOutputEvent(100 * time.Millisecond); assert.NoError(err) {
		require.Contains(event.Message, "test_uptime")
		require.WithinDuration(start.Add(1*time.Second), event.Timestamp, 300*time.Millisecond)
	}

	time.Sleep(1000 * time.Millisecond)
	if event, err := conf.TestGetOutputEvent(100 * time.Millisecond); assert.NoError(err) {
		require.Contains(event.Message, "test_uptime")
		require.WithinDuration(start.Add(2*time.Second), event.Timestamp, 300*time.Millisecond)
	}

	time.Sleep(1000 * time.Millisecond)
	if event, err := conf.TestGetOutputEvent(100 * time.Millisecond); assert.NoError(err) {
		require.WithinDuration(start.Add(3*time.Second), event.Timestamp, 300*time.Millisecond)
	}
	if event, err := conf.TestGetOutputEvent(100 * time.Millisecond); assert.NoError(err) {
		require.WithinDuration(start.Add(3*time.Second), event.Timestamp, 300*time.Millisecond)
	}
}

func Test_input_exec_module_json(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
input:
  - type: exec
    command: "./test_json.sh"
    interval: 1
    message_type: json
	`)))
	require.NoError(err)
	start := time.Now()
	require.NoError(conf.Start(ctx))

	time.Sleep(500 * time.Millisecond)
	if event, err := conf.TestGetOutputEvent(100 * time.Millisecond); assert.NoError(err) {
		require.WithinDuration(start, event.Timestamp, 300*time.Millisecond)
		require.EqualValues(123, event.Extra["num"])
		require.Equal("this is a test text", event.Extra["text"])
		require.Equal(map[string]interface{}{"data": "text in child"}, event.Extra["child"])
	}
}
