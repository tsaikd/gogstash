package outputkafka

import (
	"context"
	"github.com/tsaikd/gogstash/config/logevent"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
)

func init() {
	goglog.Logger.SetLevel(logrus.DebugLevel)
	config.RegistOutputHandler(ModuleName, InitHandler)
}

func Test_output_kafka_module(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
output:
  - type: kafka
    version: 0.10.2.0
    brokers:
      - 127.0.0.1:9092
    topics:
      - testTopic
	`)))
	require.NoError(err)
	err = conf.Start(ctx)
	if err != nil {
		t.Log("skip test output kafka module")
		require.NoError(err)
	}

	conf.TestInputEvent(logevent.LogEvent{
		Timestamp: time.Now(),
		Message:   "outputkafka test message",
	})

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal("outputkafka test message", event.Message)
	}

}
