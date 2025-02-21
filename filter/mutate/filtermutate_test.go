package filtermutate

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

func Test_filter_mutate_module_error(t *testing.T) {
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: mutate
	`)))
	require.NoError(err)
	require.Error(conf.Start(ctx))
}

func Test_filter_mutate_module_rename(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: mutate
    rename: ["key", "key2"]
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	expectedEvent := logevent.LogEvent{
		Extra: map[string]any{
			"key2": "foo,bar",
		},
	}

	conf.TestInputEvent(logevent.LogEvent{
		Extra: map[string]any{
			"key": "foo,bar",
		},
	})

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal(expectedEvent, event)
	}
}
func Test_filter_mutate_module_configured(t *testing.T) {
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: mutate
    add_tag: ["testing"]
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))
}

func Test_filter_mutate_module_split(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: mutate
    split: ["key", ","]
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	expectedEvent := logevent.LogEvent{
		Extra: map[string]any{
			"key": []string{"foo", "bar"},
		},
	}

	conf.TestInputEvent(logevent.LogEvent{
		Extra: map[string]any{
			"key": "foo,bar",
		},
	})

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal(expectedEvent, event)
	}
}

func Test_filter_mutate_module_replace(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: mutate
    replace: ["key", ",", "|"]
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	expectedEvent := logevent.LogEvent{
		Extra: map[string]any{
			"key": "foo|bar",
		},
	}

	conf.TestInputEvent(logevent.LogEvent{
		Extra: map[string]any{
			"key": "foo,bar",
		},
	})

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal(expectedEvent, event)
	}
}

func Test_filter_mutate_module_replace_with_extrapolation(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: mutate
    replace: ["key", ",", "%{spacer}"]
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	expectedEvent := logevent.LogEvent{
		Extra: map[string]any{
			"key":    "foo~~bar",
			"spacer": "~~",
		},
	}

	conf.TestInputEvent(logevent.LogEvent{
		Extra: map[string]any{
			"key":    "foo,bar",
			"spacer": "~~",
		},
	})

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal(expectedEvent, event)
	}
}
func Test_filter_mutate_module_merge(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: mutate
    merge: ["key", "value"]
  - type: mutate
    merge: ["key", "%{field}"]
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	expectedEvent := logevent.LogEvent{
		Extra: map[string]any{
			"key":   []string{"value", "fieldvalue"},
			"field": "fieldvalue",
		},
	}

	conf.TestInputEvent(logevent.LogEvent{
		Extra: map[string]any{
			"field": "fieldvalue",
		},
	})

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal(expectedEvent, event)
	}
}

func Test_filter_mutate_module_merge_error(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: mutate
    merge: ["key", "value"]
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	expectedEvent := logevent.LogEvent{
		Extra: map[string]any{
			"key": 1,
		},
		Tags: []string{ErrorTag},
	}

	conf.TestInputEvent(logevent.LogEvent{
		Extra: map[string]any{
			"key": 1,
		},
	})

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal(expectedEvent, event)
	}
}
