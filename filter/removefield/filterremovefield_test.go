package filterremovefield

import (
	"reflect"
	"strings"
	"testing"
	"time"

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

func Test_empty_fields(t *testing.T) {
	require := require.New(t)
	require.NotNil(require)

	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
filter:
  - type: remove_field
	`)))
	require.NoError(err)

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

	inchan := conf.Get(reflect.TypeOf(make(config.InChan))).
		Interface().(config.InChan)

	outchan := conf.Get(reflect.TypeOf(make(config.OutChan))).
		Interface().(config.OutChan)

	err = conf.RunFilters()
	require.NoError(err)

	inchan <- logevent.LogEvent{
		Timestamp: timestamp,
		Message:   "filter test message",
		Extra: map[string]interface{}{
			"fieldA": "foo",
			"fieldB": "bar",
		},
	}

	event := <-outchan

	require.Equal(expectedEvent, event)
	require.NoError(err)
}

func Test_remove_one_field(t *testing.T) {
	require := require.New(t)
	require.NotNil(require)

	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
filter:
  - type: remove_field
    fields:
      - fieldA
	`)))
	require.NoError(err)

	timestamp, err := time.Parse("2006-01-02T15:04:05Z", "2017-04-05T18:30:41.193Z")
	require.NoError(err)

	expectedEvent := logevent.LogEvent{
		Timestamp: timestamp,
		Message:   "filter test message",
		Extra: map[string]interface{}{
			"fieldB": "bar",
		},
	}

	inchan := conf.Get(reflect.TypeOf(make(config.InChan))).
		Interface().(config.InChan)

	outchan := conf.Get(reflect.TypeOf(make(config.OutChan))).
		Interface().(config.OutChan)

	err = conf.RunFilters()
	require.NoError(err)

	inchan <- logevent.LogEvent{
		Timestamp: timestamp,
		Message:   "filter test message",
		Extra: map[string]interface{}{
			"fieldA": "foo",
			"fieldB": "bar",
		},
	}

	event := <-outchan

	require.Equal(expectedEvent, event)
	require.NoError(err)
}

func Test_remove_two_field(t *testing.T) {
	require := require.New(t)
	require.NotNil(require)

	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
filter:
  - type: remove_field
    fields:
      - fieldA
      - fieldB
	`)))
	require.NoError(err)

	timestamp, err := time.Parse("2006-01-02T15:04:05Z", "2017-04-05T18:30:41.193Z")
	require.NoError(err)

	expectedEvent := logevent.LogEvent{
		Timestamp: timestamp,
		Message:   "filter test message",
		Extra:     map[string]interface{}{},
	}

	inchan := conf.Get(reflect.TypeOf(make(config.InChan))).
		Interface().(config.InChan)

	outchan := conf.Get(reflect.TypeOf(make(config.OutChan))).
		Interface().(config.OutChan)

	err = conf.RunFilters()
	require.NoError(err)

	inchan <- logevent.LogEvent{
		Timestamp: timestamp,
		Message:   "filter test message",
		Extra: map[string]interface{}{
			"fieldA": "foo",
			"fieldB": "bar",
		},
	}

	event := <-outchan

	require.Equal(expectedEvent, event)
	require.NoError(err)
}

func Test_remove_child_field(t *testing.T) {
	require := require.New(t)
	require.NotNil(require)

	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
filter:
  - type: remove_field
    fields:
      - fieldA.childA
	`)))
	require.NoError(err)

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

	inchan := conf.Get(reflect.TypeOf(make(config.InChan))).
		Interface().(config.InChan)

	outchan := conf.Get(reflect.TypeOf(make(config.OutChan))).
		Interface().(config.OutChan)

	err = conf.RunFilters()
	require.NoError(err)

	inchan <- logevent.LogEvent{
		Timestamp: timestamp,
		Message:   "filter test message",
		Extra: map[string]interface{}{
			"fieldA": map[string]interface{}{
				"childA": "child test A",
				"childB": "child test B",
			},
			"fieldB": "bar",
		},
	}

	event := <-outchan

	require.Equal(expectedEvent, event)
	require.NoError(err)
}
