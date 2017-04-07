package filtertypeconv

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Sirupsen/logrus"
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

func Test_config_error(t *testing.T) {
	require := require.New(t)
	require.NotNil(require)

	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
filter:
  - type: typeconv
    conv_type: foobar
	`)))
	require.NoError(err)

	err = conf.RunFilters()
	require.Error(err)
}

func Test_convert_string(t *testing.T) {
	require := require.New(t)
	require.NotNil(require)

	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
filter:
  - type: typeconv
    conv_type: string
    fields: ["foo", "bar"]
	`)))
	require.NoError(err)

	expectedEvent := logevent.LogEvent{
		Extra: map[string]interface{}{
			"foo":   "123",
			"bar":   "3.14",
			"extra": "foo bar",
		},
	}

	inchan := conf.Get(reflect.TypeOf(make(config.InChan))).
		Interface().(config.InChan)

	outchan := conf.Get(reflect.TypeOf(make(config.OutChan))).
		Interface().(config.OutChan)

	err = conf.RunFilters()
	require.NoError(err)

	inchan <- logevent.LogEvent{
		Extra: map[string]interface{}{
			"foo":   123,
			"bar":   3.14,
			"extra": "foo bar",
		},
	}

	event := <-outchan

	require.Equal(expectedEvent, event)
	require.NoError(err)
}

func Test_convert_int64(t *testing.T) {
	require := require.New(t)
	require.NotNil(require)

	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
filter:
  - type: typeconv
    conv_type: int64
    fields: ["foo", "bar", "foostr", "barstr"]
	`)))
	require.NoError(err)

	expectedEvent := logevent.LogEvent{
		Extra: map[string]interface{}{
			"foo":    int64(123),
			"bar":    int64(3),
			"foostr": int64(123),
			"barstr": int64(3),
			"extra":  "foo bar",
		},
	}

	inchan := conf.Get(reflect.TypeOf(make(config.InChan))).
		Interface().(config.InChan)

	outchan := conf.Get(reflect.TypeOf(make(config.OutChan))).
		Interface().(config.OutChan)

	err = conf.RunFilters()
	require.NoError(err)

	inchan <- logevent.LogEvent{
		Extra: map[string]interface{}{
			"foo":    123,
			"bar":    3.14,
			"foostr": "123",
			"barstr": "3.14",
			"extra":  "foo bar",
		},
	}

	event := <-outchan

	require.Equal(expectedEvent, event)
	require.NoError(err)
}

func Test_convert_float64(t *testing.T) {
	require := require.New(t)
	require.NotNil(require)

	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
filter:
  - type: typeconv
    conv_type: float64
    fields: ["foo", "bar", "foostr", "barstr"]
	`)))
	require.NoError(err)

	expectedEvent := logevent.LogEvent{
		Extra: map[string]interface{}{
			"foo":    float64(123),
			"bar":    float64(3.14),
			"foostr": float64(123),
			"barstr": float64(3.14),
			"extra":  "foo bar",
		},
	}

	inchan := conf.Get(reflect.TypeOf(make(config.InChan))).
		Interface().(config.InChan)

	outchan := conf.Get(reflect.TypeOf(make(config.OutChan))).
		Interface().(config.OutChan)

	err = conf.RunFilters()
	require.NoError(err)

	inchan <- logevent.LogEvent{
		Extra: map[string]interface{}{
			"foo":    123,
			"bar":    3.14,
			"foostr": "123",
			"barstr": "3.14",
			"extra":  "foo bar",
		},
	}

	event := <-outchan

	require.Equal(expectedEvent, event)
	require.NoError(err)
}
