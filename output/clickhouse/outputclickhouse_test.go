package outputclickhouse

import (
	"context"
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
	// register this output handler for tests
	config.RegistOutputHandler(ModuleName, InitHandler)
}

func Test_output_clickhouse_module(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
output:
  - type: clickhouse
    urls: ["http://127.0.0.1:8123"]
    table: "logs.test"
    batch_size: 1000
    flush_interval: 10s
`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	conf.TestInputEvent(logevent.LogEvent{
		Timestamp: time.Now(),
		Message:   "outputclickhouse test message",
		Extra:     map[string]any{"App": "app1", "int": 12, "float": 0.3},
	})

	// wait a bit to allow the event to go through the debug channel
	time.Sleep(1000 * time.Millisecond)

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal("outputclickhouse test message", event.Message)
	}
}
