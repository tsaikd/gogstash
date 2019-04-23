package filterurlparam

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
	"strings"
	"testing"
	"time"
)

func init() {
	goglog.Logger.SetLevel(logrus.DebugLevel)
	config.RegistFilterHandler(ModuleName, InitHandler)
}

func Test_filter_param_module(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: url_param 
    source: request_url
    prefix: request_url_args_
    include_keys: ["foo", "bar", "date_time", "empty_key"]
	`)))

	require.NoError(err)
	require.NoError(conf.Start(ctx))

	// test "http://domain/path?params"
	expectedEvent := logevent.LogEvent{
		Extra: map[string]interface{}{
			"request_url":                "http://www.example.com/path?foo=my_foo&bar=my_bar&date_time=2019-04-17%2010:12&empty_key=&discard_key=discard_value",
			"request_url_args_foo":       "my_foo",
			"request_url_args_bar":       "my_bar",
			"request_url_args_date_time": "2019-04-17 10:12",
		},
	}

	conf.TestInputEvent(logevent.LogEvent{
		Extra: map[string]interface{}{
			"request_url": "http://www.example.com/path?foo=my_foo&bar=my_bar&date_time=2019-04-17%2010:12&empty_key=&discard_key=discard_value",
		},
	})

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal(expectedEvent, event)
	}

	// test "/path?params"
	expectedEvent = logevent.LogEvent{
		Extra: map[string]interface{}{
			"request_url":                "/path?foo=my_foo&bar=my_bar&date_time=2019-04-17%2010:12&empty_key=&discard_key=discard_value",
			"request_url_args_foo":       "my_foo",
			"request_url_args_bar":       "my_bar",
			"request_url_args_date_time": "2019-04-17 10:12",
		},
	}

	conf.TestInputEvent(logevent.LogEvent{
		Extra: map[string]interface{}{
			"request_url": "/path?foo=my_foo&bar=my_bar&date_time=2019-04-17%2010:12&empty_key=&discard_key=discard_value",
		},
	})

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal(expectedEvent, event)
	}

	// test "?params"
	expectedEvent = logevent.LogEvent{
		Extra: map[string]interface{}{
			"request_url":                "?foo=my_foo&bar=my_bar&date_time=2019-04-17%2010:12&empty_key=&discard_key=discard_value",
			"request_url_args_foo":       "my_foo",
			"request_url_args_bar":       "my_bar",
			"request_url_args_date_time": "2019-04-17 10:12",
		},
	}

	conf.TestInputEvent(logevent.LogEvent{
		Extra: map[string]interface{}{
			"request_url": "?foo=my_foo&bar=my_bar&date_time=2019-04-17%2010:12&empty_key=&discard_key=discard_value",
		},
	})

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal(expectedEvent, event)
	}

	// test "params"
	expectedEvent = logevent.LogEvent{
		Extra: map[string]interface{}{
			"request_url":                "foo=my_foo&bar=my_bar&date_time=2019-04-17%2010:12&empty_key=&discard_key=discard_value",
			"request_url_args_foo":       "my_foo",
			"request_url_args_bar":       "my_bar",
			"request_url_args_date_time": "2019-04-17 10:12",
		},
	}

	conf.TestInputEvent(logevent.LogEvent{
		Extra: map[string]interface{}{
			"request_url": "foo=my_foo&bar=my_bar&date_time=2019-04-17%2010:12&empty_key=&discard_key=discard_value",
		},
	})

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal(expectedEvent, event)
	}

	// test "http://domain/path"
	expectedEvent = logevent.LogEvent{
		Extra: map[string]interface{}{
			"request_url": "http://example.com/path",
		},
	}

	conf.TestInputEvent(logevent.LogEvent{
		Extra: map[string]interface{}{
			"request_url": "http://example.com/path",
		},
	})

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal(expectedEvent, event)
	}

	// test "string"
	expectedEvent = logevent.LogEvent{
		Extra: map[string]interface{}{
			"request_url": "nothing",
		},
	}

	conf.TestInputEvent(logevent.LogEvent{
		Extra: map[string]interface{}{
			"request_url": "nothing",
		},
	})

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal(expectedEvent, event)
	}
}
