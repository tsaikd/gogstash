package filtertypeconv

import (
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

func Test_filter_typeconv_module_error(t *testing.T) {
	require := require.New(t)
	require.NotNil(require)

	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: typeconv
    conv_type: foobar
	`)))
	require.NoError(err)
	require.Error(conf.Start())
}

func Test_filter_typeconv_module_convert_string(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: typeconv
    conv_type: string
    fields: ["foo", "bar"]
	`)))
	require.NoError(err)
	require.NoError(conf.Start())

	expectedEvent := logevent.LogEvent{
		Extra: map[string]interface{}{
			"foo":   "123",
			"bar":   "3.14",
			"extra": "foo bar",
		},
	}

	conf.TestInputEvent(logevent.LogEvent{
		Extra: map[string]interface{}{
			"foo":   123,
			"bar":   3.14,
			"extra": "foo bar",
		},
	})

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal(expectedEvent, event)
	}
}

func Test_filter_typeconv_module_convert_int64(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: typeconv
    conv_type: int64
    fields: ["foo", "bar", "foostr", "barstr"]
	`)))
	require.NoError(err)
	require.NoError(conf.Start())

	expectedEvent := logevent.LogEvent{
		Extra: map[string]interface{}{
			"foo":    int64(123),
			"bar":    int64(3),
			"foostr": int64(123),
			"barstr": int64(3),
			"extra":  "foo bar",
		},
	}

	conf.TestInputEvent(logevent.LogEvent{
		Extra: map[string]interface{}{
			"foo":    123,
			"bar":    3.14,
			"foostr": "123",
			"barstr": "3.14",
			"extra":  "foo bar",
		},
	})

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal(expectedEvent, event)
	}
}

func Test_filter_typeconv_module_convert_float64(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: typeconv
    conv_type: float64
    fields: ["foo", "bar", "foostr", "barstr"]
	`)))
	require.NoError(err)
	require.NoError(conf.Start())

	expectedEvent := logevent.LogEvent{
		Extra: map[string]interface{}{
			"foo":    float64(123),
			"bar":    float64(3.14),
			"foostr": float64(123),
			"barstr": float64(3.14),
			"extra":  "foo bar",
		},
	}

	conf.TestInputEvent(logevent.LogEvent{
		Extra: map[string]interface{}{
			"foo":    123,
			"bar":    3.14,
			"foostr": "123",
			"barstr": "3.14",
			"extra":  "foo bar",
		},
	})

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal(expectedEvent, event)
	}
}
