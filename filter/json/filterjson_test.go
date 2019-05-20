package filterjson

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

func Test_filter_json_module(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: json
    message: message
    timestamp: time
    timeformat: "2006-01-02T15:04:05Z"
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	timestamp, err := time.Parse("2006-01-02T15:04:05Z", "2016-12-04T09:09:41.193Z")
	require.NoError(err)

	expectedEvent := logevent.LogEvent{
		Timestamp: timestamp,
		Message:   "Test",
		Extra: map[string]interface{}{
			"host": "Hostname",
		},
		Tags: []string{"foo"},
	}

	conf.TestInputEvent(logevent.LogEvent{
		Timestamp: time.Now(),
		Message:   "{ \"message\": \"Test\", \"host\": \"Hostname\", \"time\":\"2016-12-04T09:09:41.193Z\", \"tags\": [ \"foo\" ] }",
	})

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal(expectedEvent, event)
	}
}

func Test_filter_json_module_source(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: json
    message: message
    timestamp: time
    source: myfield
    timeformat: "2006-01-02T15:04:05Z"
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	timestamp, err := time.Parse("2006-01-02T15:04:05Z", "2016-12-04T09:09:41.193Z")
	require.NoError(err)
	fieldvalue := `{ "message": "Test", "host":"Hostname", "time":"2016-12-04T09:09:41.193Z", "tags": [ "foo" ] }`

	expectedEvent := logevent.LogEvent{
		Timestamp: timestamp,
		Message:   "Test",
		Extra: map[string]interface{}{
			"host":    "Hostname",
			"myfield": fieldvalue,
		},
		Tags: []string{"foo"},
	}

	conf.TestInputEvent(logevent.LogEvent{
		Timestamp: time.Now(),
		Extra: map[string]interface{}{
			"myfield": fieldvalue,
		},
	})

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal(expectedEvent, event)
	}
}
