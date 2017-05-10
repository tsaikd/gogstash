package filteraddfield

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
	config.RegistFilterHandler(ModuleName, InitHandler)
}

func Test_filter_add_field_module(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: add_field
    key: foo
    value: bar
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	timestamp := time.Now()
	expectedEvent := logevent.LogEvent{
		Timestamp: timestamp,
		Message:   "filter test message",
		Extra: map[string]interface{}{
			"foo": "bar",
		},
	}

	conf.TestInputEvent(logevent.LogEvent{
		Timestamp: timestamp,
		Message:   "filter test message",
	})

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal(expectedEvent, event)
	}
}
