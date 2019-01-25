package filterdate

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
