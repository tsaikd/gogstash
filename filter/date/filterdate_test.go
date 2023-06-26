package filterdate

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	filterjson "github.com/tsaikd/gogstash/filter/json"

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
		Extra: map[string]any{
			"time_local": "20/Mar/2017:00:42:51 +0000",
		},
	}

	conf.TestInputEvent(logevent.LogEvent{
		Extra: map[string]any{
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
		Extra: map[string]any{
			"time_local": "1548237471.471",
		},
	}

	conf.TestInputEvent(logevent.LogEvent{
		Extra: map[string]any{
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
		Extra: map[string]any{
			"time_local": "1548237471.471",
		},
	}

	expectedEvent := logevent.LogEvent{
		Timestamp: eventIn.Timestamp,
		Extra: map[string]any{
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
		Extra: map[string]any{
			"time_local": "2018-09-19 19:50:26,208",
		},
	}

	conf.TestInputEvent(logevent.LogEvent{
		Extra: map[string]any{
			"time_local": "2018-09-19 19:50:26,208",
		},
	})

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal(expectedEvent, event)
	}
}

func Test_filter_date_module_computeTrue(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: date
    format: ["MMdd HH:mm:ss.SSSSSS"]
    source: time_local
    joda: true
    compute_year_if_missing: true
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	now := time.Now()

	timestamp1 := time.Date(now.Year()-1, now.Month(), now.Day()+1, 0, 0, 0, 123456000, time.UTC)
	timestampStr1 := fmt.Sprintf("%02d%02d %02d:%02d:%02d.%06d", timestamp1.Month(), timestamp1.Day(), timestamp1.Hour(), timestamp1.Minute(), timestamp1.Second(), timestamp1.Nanosecond())
	fmt.Println(timestampStr1)
	expectedEvent := logevent.LogEvent{
		Timestamp: timestamp1,
		Extra: map[string]any{
			"time_local": timestampStr1,
		},
	}

	conf.TestInputEvent(logevent.LogEvent{
		Extra: map[string]any{
			"time_local": timestampStr1,
		},
	})

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal(expectedEvent, event)
	}

	timestamp2 := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 123456000, time.UTC)
	timestampStr2 := fmt.Sprintf("%02d%02d %02d:%02d:%02d.%06d", timestamp2.Month(), timestamp2.Day(), timestamp2.Hour(), timestamp2.Minute(), timestamp2.Second(), timestamp2.Nanosecond())
	fmt.Println(timestampStr2)

	expectedEvent = logevent.LogEvent{
		Timestamp: timestamp2,
		Extra: map[string]any{
			"time_local": timestampStr2,
		},
	}

	conf.TestInputEvent(logevent.LogEvent{
		Extra: map[string]any{
			"time_local": timestampStr2,
		},
	})

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal(expectedEvent, event)
	}
}

func Test_filter_date_module_computeFalse(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: date
    format: ["MMdd HH:mm:ss.SSSSSS"]
    source: time_local
    joda: true
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	now := time.Now()

	timestamp := time.Date(now.Year()-1, now.Month(), now.Day()+2, 0, 0, 0, 123456000, time.UTC)
	timestampStr := fmt.Sprintf("%02d%02d %02d:%02d:%02d.%06d", timestamp.Month(), timestamp.Day(), timestamp.Hour(), timestamp.Minute(), timestamp.Second(), timestamp.Nanosecond())

	expectedEvent := logevent.LogEvent{
		Timestamp: timestamp,
		Extra: map[string]any{
			"time_local": timestampStr,
		},
	}

	conf.TestInputEvent(logevent.LogEvent{
		Extra: map[string]any{
			"time_local": timestampStr,
		},
	})

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.NotEqual(expectedEvent, event)
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
		Extra: map[string]any{
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
		input any
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
