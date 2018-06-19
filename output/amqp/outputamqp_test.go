package outputamqp

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsaikd/KDGoLib/errutil"
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

func Test_output_amqp_module(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
output:
  - type: amqp
    urls: ["amqp://guest:guest@localhost:5672/"]
    exchange: "amq.topic"
    exchange_type: "topic"
	`)))
	require.NoError(err)
	err = conf.Start(ctx)
	if err != nil {
		require.True(config.ErrorInitOutputFailed1.Match(err))
		require.True(ErrorNoValidConn.In(err))
		require.Implements((*errutil.ErrorObject)(nil), err)
		require.True(ErrorNoValidConn.Match(err.(errutil.ErrorObject).Parent()))
		t.Skip("skip test output amqp module")
	}

	conf.TestInputEvent(logevent.LogEvent{
		Timestamp: time.Now(),
		Message:   "outputstdout test message",
		Extra: map[string]interface{}{
			"fieldstring": "ABC",
			"fieldnumber": 123,
		},
	})

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		t.Log(event)
	}
}
