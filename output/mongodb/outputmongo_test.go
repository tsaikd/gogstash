package outputmongodb

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
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

func Test_output_stdout_module(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
output:
  - type: mongodb
    host:
      - 127.0.0.1:27017
    database: test
    collection: allLogs
    timeout: 10
    connections: 10
    username: username
    password: password
    mechanism: SCRAM-SHA-1
    retry_interval: 10
	`)))
	require.NoError(err)
	err = conf.Start(ctx)
	if err != nil {
		t.Logf("Skip test output mongodb module err='%s'", err.Error())
		require.True(ErrorConnectionMongoDBFailed1.In(err))
		return
	}

	conf.TestInputEvent(logevent.LogEvent{
		Timestamp: time.Now(),
		Message:   "output mongodb test message",
	})

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal("output mongodb test message", event.Message)
	}
}
