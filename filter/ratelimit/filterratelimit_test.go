package filterratelimit

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

func Test_filter_ratelimit_module(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: rate_limit
    rate: 10
    burst: 1
	`)))
	require.NoError(err)
	require.NoError(conf.Start())

	start := time.Now()

	conf.TestInputEvent(logevent.LogEvent{})
	conf.TestInputEvent(logevent.LogEvent{})
	conf.TestInputEvent(logevent.LogEvent{})
	conf.TestInputEvent(logevent.LogEvent{})
	conf.TestInputEvent(logevent.LogEvent{})

	_, err = conf.TestGetOutputEvent(100 * time.Millisecond)
	require.NoError(err)
	_, err = conf.TestGetOutputEvent(100 * time.Millisecond)
	require.NoError(err)
	_, err = conf.TestGetOutputEvent(100 * time.Millisecond)
	require.NoError(err)
	_, err = conf.TestGetOutputEvent(100 * time.Millisecond)
	require.NoError(err)
	_, err = conf.TestGetOutputEvent(100 * time.Millisecond)
	require.NoError(err)

	require.WithinDuration(start.Add(400*time.Millisecond), time.Now(), 150*time.Millisecond)
}

func Test_filter_ratelimit_module_burst(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: rate_limit
    rate: 10
    burst: 4
	`)))
	require.NoError(err)
	require.NoError(conf.Start())

	time.Sleep(600 * time.Millisecond)

	start := time.Now()

	conf.TestInputEvent(logevent.LogEvent{})
	conf.TestInputEvent(logevent.LogEvent{})
	conf.TestInputEvent(logevent.LogEvent{})
	conf.TestInputEvent(logevent.LogEvent{})
	conf.TestInputEvent(logevent.LogEvent{})
	conf.TestInputEvent(logevent.LogEvent{})

	_, err = conf.TestGetOutputEvent(100 * time.Millisecond)
	require.NoError(err)
	_, err = conf.TestGetOutputEvent(100 * time.Millisecond)
	require.NoError(err)
	_, err = conf.TestGetOutputEvent(100 * time.Millisecond)
	require.NoError(err)
	_, err = conf.TestGetOutputEvent(100 * time.Millisecond)
	require.NoError(err)
	_, err = conf.TestGetOutputEvent(100 * time.Millisecond)
	require.NoError(err)
	_, err = conf.TestGetOutputEvent(100 * time.Millisecond)
	require.NoError(err)

	require.WithinDuration(start.Add(150*time.Millisecond), time.Now(), 100*time.Millisecond)
}

func Test_filter_ratelimit_module_delay(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: rate_limit
    rate: 10
    burst: 1
	`)))
	require.NoError(err)
	require.NoError(conf.Start())

	time.Sleep(500 * time.Millisecond)

	start := time.Now()

	conf.TestInputEvent(logevent.LogEvent{})
	conf.TestInputEvent(logevent.LogEvent{})
	conf.TestInputEvent(logevent.LogEvent{})
	conf.TestInputEvent(logevent.LogEvent{})
	conf.TestInputEvent(logevent.LogEvent{})

	_, err = conf.TestGetOutputEvent(100 * time.Millisecond)
	require.NoError(err)
	_, err = conf.TestGetOutputEvent(100 * time.Millisecond)
	require.NoError(err)
	_, err = conf.TestGetOutputEvent(100 * time.Millisecond)
	require.NoError(err)
	_, err = conf.TestGetOutputEvent(100 * time.Millisecond)
	require.NoError(err)
	_, err = conf.TestGetOutputEvent(100 * time.Millisecond)
	require.NoError(err)

	require.WithinDuration(start.Add(400*time.Millisecond), time.Now(), 150*time.Millisecond)
}
