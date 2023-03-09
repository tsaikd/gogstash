package filterdrop

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
	filtercond "github.com/tsaikd/gogstash/filter/cond"
)

func init() {
	goglog.Logger.SetLevel(logrus.DebugLevel)
	config.RegistFilterHandler(filtercond.ModuleName, filtercond.InitHandler)
	config.RegistFilterHandler(ModuleName, InitHandler)
}

func Test_filter_drop_module(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: drop
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	conf.TestInputEvent(logevent.LogEvent{
		Timestamp: time.Now(),
		Message:   "filter test message",
		Drop:      false,
		Extra: map[string]interface{}{
			"foo": "bar",
		},
	})
	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal(logevent.LogEvent{}, event)
		require.Equal(err, nil)
	}
}

func Test_filter_drop_module_cond(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: cond
    condition: "level == 'DEBUG'"
    filter:
      - type: drop
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	timestamp := time.Now()
	expectedEvent1 := logevent.LogEvent{
		Timestamp: timestamp,
		Message:   "filter test message 1",
		Drop:      false,
		Extra: map[string]interface{}{
			"level": "ERROR",
		},
	}

	conf.TestInputEvent(logevent.LogEvent{
		Timestamp: timestamp,
		Message:   "filter test message 1",
		Drop:      false,
		Extra: map[string]interface{}{
			"level": "ERROR",
		},
	})
	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal(expectedEvent1, event)
	}

	conf.TestInputEvent(logevent.LogEvent{
		Timestamp: timestamp,
		Message:   "filter test message 2",
		Drop:      false,
		Extra: map[string]interface{}{
			"level": "DEBUG",
		},
	})
	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal(logevent.LogEvent{}, event)
		require.Equal(err, nil)
	}
}
