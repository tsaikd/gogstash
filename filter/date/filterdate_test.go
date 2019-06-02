package filterdate

import (
	"context"
	"fmt"
	filterjson "github.com/tsaikd/gogstash/filter/json"
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
	config.RegistFilterHandler(filterjson.ModuleName, filterjson.InitHandler)
}

func Test_filter_date_module(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: date
    format: ["02/Jan/2006:15:04:05 -0700"]
    source: time_local
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	timestamp, err := time.Parse("2006-01-02T15:04:05Z", "2017-03-20T00:42:51Z")
	require.NoError(err)

	expectedEvent := logevent.LogEvent{
		Timestamp: timestamp,
		Extra: map[string]interface{}{
			"time_local": "20/Mar/2017:00:42:51 +0000",
		},
	}

	conf.TestInputEvent(logevent.LogEvent{
		Extra: map[string]interface{}{
			"time_local": "20/Mar/2017:00:42:51 +0000",
		},
	})

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal(expectedEvent, event)
	}
}

func Test_filter_date_module_UNIX(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: date
    format: ["UNIX"]
    source: time_local
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	timestamp, err := time.Parse("2006-01-02T15:04:05Z", "2019-01-23T09:57:51.471Z")
	require.NoError(err)

	expectedEvent := logevent.LogEvent{
		Timestamp: timestamp,
		Extra: map[string]interface{}{
			"time_local": "1548237471.471",
		},
	}

	conf.TestInputEvent(logevent.LogEvent{
		Extra: map[string]interface{}{
			"time_local": "1548237471.471",
		},
	})

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal(expectedEvent, event)
	}
}

func Test_filter_date_module_UNIX_target(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: date
    format: ["UNIX"]
    source: time_local
    target: mytimestamp
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	timestamp, err := time.Parse("2006-01-02T15:04:05Z", "2019-01-23T09:57:51.471Z")
	require.NoError(err)
	eventIn := logevent.LogEvent{
		Extra: map[string]interface{}{
			"time_local": "1548237471.471",
		},
	}

	expectedEvent := logevent.LogEvent{
		Timestamp: eventIn.Timestamp,
		Extra: map[string]interface{}{
			"time_local":  "1548237471.471",
			"mytimestamp": timestamp.UTC(),
		},
	}

	conf.TestInputEvent(eventIn)

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal(expectedEvent, event)
	}
}

func Test_filter_date_module_joda(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: date
    format: ["YYYY-MM-dd HH:mm:ss,SSS"]
    source: time_local
    joda: true
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	timestamp, err := time.Parse("2006-01-02T15:04:05.000Z", "2018-09-19T19:50:26.208Z")
	require.NoError(err)

	expectedEvent := logevent.LogEvent{
		Timestamp: timestamp,
		Extra: map[string]interface{}{
			"time_local": "2018-09-19 19:50:26,208",
		},
	}

	conf.TestInputEvent(logevent.LogEvent{
		Extra: map[string]interface{}{
			"time_local": "2018-09-19 19:50:26,208",
		},
	})

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal(expectedEvent, event)
	}
}

func Test_filter_date_module_float(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: json
    source: json
  - type: date
    format: ["UNIX"]
    source: time_local
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	conf.TestInputEvent(logevent.LogEvent{
		Extra: map[string]interface{}{
			"json": `{"time_local": 1558348989869}`,
		},
	})

	// this should NOT result in gogstash_filter_date_error tag
	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal(0, len(event.Tags))
	}
}

func Test_filter_date_convert(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)

	type inputOutput struct {
		name  string
		input interface{}
		sec   int64
		nsec  int64
	}
	tests := []inputOutput{
		{
			name:  "whole float",
			input: float64(1558348989869),
			sec:   int64(1558348989869),
			nsec:  int64(0),
		},
		{
			name:  "decimal float",
			input: float64(15583.48989869),
			sec:   int64(15583),
			nsec:  int64(489898690),
		},
		{
			name:  "whole string",
			input: "1558348989869",
			sec:   int64(1558348989869),
			nsec:  int64(0),
		},
		{
			name:  "fraction string",
			input: "1548237471.471",
			sec:   int64(1548237471),
			nsec:  int64(471000000),
		},
		{
			name:  "exp string",
			input: "1.558349e+12",
			sec:   int64(1558349000000),
			nsec:  int64(0),
		},
		{
			name:  "exp string 2",
			input: "1.558348989869e+12",
			sec:   int64(1558348989869),
			nsec:  int64(0),
		},
		{
			name:  "int",
			input: int64(1558348989869),
			sec:   int64(1558348989869),
			nsec:  int64(0),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := require.New(t)
			r.NotNil(r)
			var sec, nsec int64
			var err error
			switch input := test.input.(type) {
			case float64:
				sec, nsec = convertFloat(input)
			default:
				sec, nsec, err = convert(fmt.Sprintf("%v", input))
			}
			r.NoError(err)
			r.Equal(test.sec, sec)
			r.Equal(test.nsec, nsec)
		})
	}
}
