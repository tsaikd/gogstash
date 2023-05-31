package filterconvert

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

func Test_filter_convert_module_error(t *testing.T) {
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: convert
	`)))
	require.NoError(err)
	require.Error(conf.Start(ctx))
}

func Test_filter_convert_module_to_int_1(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: convert
    to_int: ["key", "1"]
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	expectedEvent := logevent.LogEvent{
		Extra: map[string]any{
			"key": int64(99),
		},
	}

	conf.TestInputEvent(logevent.LogEvent{
		Extra: map[string]any{
			"key": "99",
		},
	})

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal(expectedEvent, event)
	}
}

func Test_filter_convert_module_to_int_100(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: convert
    to_int: ["key", "100"]
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	expectedEvent := logevent.LogEvent{
		Extra: map[string]any{
			"key": int64(9900),
		},
	}

	conf.TestInputEvent(logevent.LogEvent{
		Extra: map[string]any{
			"key": "99",
		},
	})

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal(expectedEvent, event)
	}
}

func Test_filter_convert_module_to_int_error_value(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: convert
    to_int: ["key", "1"]
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	conf.TestInputEvent(logevent.LogEvent{
		Extra: map[string]any{
			"key": "99ms",
		},
	})

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		assert.Equal(event.Extra["key"], "99ms")
	}
}

func Test_filter_convert_module_to_int_error_factor(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: convert
    to_int: ["key", "1ms"]
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	expectedEvent := logevent.LogEvent{
		Extra: map[string]any{
			"key": int64(99),
		},
	}
	conf.TestInputEvent(logevent.LogEvent{
		Extra: map[string]any{
			"key": "99",
		},
	})

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal(expectedEvent, event)
	}
}

func Test_filter_convert_module_to_float_1(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: convert
    to_float: ["key", "1"]
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	expectedEvent := logevent.LogEvent{
		Extra: map[string]any{
			"key": 99.9,
		},
	}

	conf.TestInputEvent(logevent.LogEvent{
		Extra: map[string]any{
			"key": "99.9",
		},
	})

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal(expectedEvent, event)
	}
}

func Test_filter_convert_module_to_float_1_0(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: convert
    to_float: ["key", "1.0"]
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	expectedEvent := logevent.LogEvent{
		Extra: map[string]any{
			"key": 99.9,
		},
	}

	conf.TestInputEvent(logevent.LogEvent{
		Extra: map[string]any{
			"key": "99.9",
		},
	})

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal(expectedEvent, event)
	}
}

func Test_filter_convert_module_to_float_10(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: convert
    to_float: ["key", "10.0"]
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	expectedEvent := logevent.LogEvent{
		Extra: map[string]any{
			"key": 999.0,
		},
	}

	conf.TestInputEvent(logevent.LogEvent{
		Extra: map[string]any{
			"key": "99.9",
		},
	})

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		assert.InDelta(expectedEvent.Extra["key"], event.Extra["key"], 0.00000000000001)
	}
}

func Test_filter_convert_module_to_float_0_1(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: convert
    to_float: ["key", "0.1"]
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	expectedEvent := logevent.LogEvent{
		Extra: map[string]any{
			"key": 9.99,
		},
	}

	conf.TestInputEvent(logevent.LogEvent{
		Extra: map[string]any{
			"key": "99.9",
		},
	})

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		assert.InDelta(expectedEvent.Extra["key"], event.Extra["key"], 0.00000000000001)
	}
}

func Test_filter_convert_module_to_float_error_value(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: convert
    to_float: ["key", "0.1"]
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	conf.TestInputEvent(logevent.LogEvent{
		Extra: map[string]any{
			"key": "99.9ms",
		},
	})

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		assert.Equal(event.Extra["key"], "99.9ms")
	}
}

func Test_filter_convert_module_to_float_error_factor(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: convert
    to_float: ["key", "0,1"]
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	expectedEvent := logevent.LogEvent{
		Extra: map[string]any{
			"key": 99.9,
		},
	}
	conf.TestInputEvent(logevent.LogEvent{
		Extra: map[string]any{
			"key": "99.9",
		},
	})

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		assert.InDelta(expectedEvent.Extra["key"], event.Extra["key"], 0.00000000000001)
	}
}
