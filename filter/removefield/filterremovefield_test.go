package filterremovefield

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
	config.RegistFilterHandler(ModuleName, InitHandler)
}

func Test_filter_remove_field_module_empty_fields(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: remove_field
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	timestamp, err := time.Parse("2006-01-02T15:04:05Z", "2017-04-05T18:30:41.193Z")
	require.NoError(err)

	expectedEvent := logevent.LogEvent{
		Timestamp: timestamp,
		Message:   "filter test message",
		Extra: map[string]interface{}{
			"fieldA": "foo",
			"fieldB": "bar",
		},
	}

	conf.TestInputEvent(logevent.LogEvent{
		Timestamp: timestamp,
		Message:   "filter test message",
		Extra: map[string]interface{}{
			"fieldA": "foo",
			"fieldB": "bar",
		},
	})

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal(expectedEvent, event)
	}
}

func Test_filter_remove_field_module_remove_one_field(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: remove_field
    fields:
      - fieldA
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	timestamp, err := time.Parse("2006-01-02T15:04:05Z", "2017-04-05T18:30:41.193Z")
	require.NoError(err)

	expectedEvent := logevent.LogEvent{
		Timestamp: timestamp,
		Message:   "filter test message",
		Extra: map[string]interface{}{
			"fieldB": "bar",
		},
	}

	conf.TestInputEvent(logevent.LogEvent{
		Timestamp: timestamp,
		Message:   "filter test message",
		Extra: map[string]interface{}{
			"fieldA": "foo",
			"fieldB": "bar",
		},
	})

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal(expectedEvent, event)
	}
}

func Test_filter_remove_field_module_remove_two_fields(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: remove_field
    fields:
      - fieldA
      - fieldB
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	timestamp, err := time.Parse("2006-01-02T15:04:05Z", "2017-04-05T18:30:41.193Z")
	require.NoError(err)

	expectedEvent := logevent.LogEvent{
		Timestamp: timestamp,
		Message:   "filter test message",
		Extra:     map[string]interface{}{},
	}

	conf.TestInputEvent(logevent.LogEvent{
		Timestamp: timestamp,
		Message:   "filter test message",
		Extra: map[string]interface{}{
			"fieldA": "foo",
			"fieldB": "bar",
		},
	})

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal(expectedEvent, event)
	}
}

func Test_filter_remove_field_module_remove_child_field(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: remove_field
    fields:
      - fieldA.childA
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	timestamp, err := time.Parse("2006-01-02T15:04:05Z", "2017-04-05T18:30:41.193Z")
	require.NoError(err)

	expectedEvent := logevent.LogEvent{
		Timestamp: timestamp,
		Message:   "filter test message",
		Extra: map[string]interface{}{
			"fieldA": map[string]interface{}{
				"childB": "child test B",
			},
			"fieldB": "bar",
		},
	}

	conf.TestInputEvent(logevent.LogEvent{
		Timestamp: timestamp,
		Message:   "filter test message",
		Extra: map[string]interface{}{
			"fieldA": map[string]interface{}{
				"childA": "child test A",
				"childB": "child test B",
			},
			"fieldB": "bar",
		},
	})

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal(expectedEvent, event)
	}
}

func Test_filter_remove_field_module_remove_message(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: remove_field
    remove_message: true
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	timestamp, err := time.Parse("2006-01-02T15:04:05Z", "2017-04-05T18:30:41.193Z")
	require.NoError(err)

	expectedEvent := logevent.LogEvent{
		Timestamp: timestamp,
		Message:   "",
		Extra: map[string]interface{}{
			"fieldA": "foo",
			"fieldB": "bar",
		},
	}

	conf.TestInputEvent(logevent.LogEvent{
		Timestamp: timestamp,
		Message:   "filter test message",
		Extra: map[string]interface{}{
			"fieldA": "foo",
			"fieldB": "bar",
		},
	})

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal(expectedEvent, event)
	}
}
