package filtercond

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
	filteraddfield "github.com/tsaikd/gogstash/filter/addfield"
	filtermutate "github.com/tsaikd/gogstash/filter/mutate"
)

func init() {
	goglog.Logger.SetLevel(logrus.DebugLevel)
	config.RegistFilterHandler(filteraddfield.ModuleName, filteraddfield.InitHandler)
	config.RegistFilterHandler(ModuleName, InitHandler)
}

func Test_filter_cond_module_invalid(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: cond
    condition: "!'level' == 'ERROR'"
    filter:
      - type: add_field
        key: foo
        value: bar
    `)))
	require.Nil(err)
	_, err = InitHandler(context.TODO(), conf.FilterRaw[0])
	require.NotNil(err)
}

func Test_filter_cond_module(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: cond
    condition: "level == 'ERROR'"
    filter:
      - type: add_field
        key: foo
        value: bar
    else_filter:
      - type: add_field
        key: foo
        value: bar2
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	timestamp := time.Now()
	expectedEvent1 := logevent.LogEvent{
		Timestamp: timestamp,
		Message:   "filter test message 1",
		Extra: map[string]interface{}{
			"level": "ERROR",
			"foo":   "bar",
		},
	}
	expectedEvent2 := logevent.LogEvent{
		Timestamp: timestamp,
		Message:   "filter test message 2",
		Extra: map[string]interface{}{
			"level": "WARN",
			"foo":   "bar2",
		},
	}

	conf.TestInputEvent(logevent.LogEvent{
		Timestamp: timestamp,
		Message:   "filter test message 1",
		Extra: map[string]interface{}{
			"level": "ERROR",
		},
	})
	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal(expectedEvent1, event)
	}

	conf.TestInputEvent(logevent.LogEvent{
		Timestamp: timestamp,
		Message:   "filter test message 2",
		Extra: map[string]interface{}{
			"level": "WARN",
		},
	})
	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal(expectedEvent2, event)
	}
}

func Test_filter_cond_module_common(t *testing.T) {
	config.RegistFilterHandler(filtermutate.ModuleName, filtermutate.InitHandler)
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)
	// if branch
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: cond
    condition: "true"
    filter:
      - type: mutate
        add_tag: ["added"]
    `)))
	require.Nil(err)
	require.NoError(conf.Start(context.Background()))
	conf.TestInputEvent(logevent.LogEvent{})
	if output, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal([]string{"added"}, output.Tags)
	}

	// else branch
	conf, err = config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: cond
    condition: "false"
    filter:
      - type: mutate
        add_tag: ["error"]
    else_filter:
      - type: mutate
        add_tag: ["added"]
    `)))
	require.Nil(err)
	require.NoError(conf.Start(context.Background()))
	conf.TestInputEvent(logevent.LogEvent{})
	if output, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal([]string{"added"}, output.Tags)
	}
}
