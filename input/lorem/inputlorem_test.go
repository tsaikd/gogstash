package inputlorem

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
)

func init() {
	goglog.Logger.SetLevel(logrus.DebugLevel)
	config.RegistInputHandler(ModuleName, InitHandler)
}

func Test_input_lorem_module(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
input:
  - type: lorem
    worker: 1
    duration: "1s"
	`)))
	require.NoError(err)
	start := time.Now()
	require.NoError(conf.Start(ctx))

	time.Sleep(500 * time.Millisecond)
	if event, err := conf.TestGetOutputEvent(100 * time.Millisecond); assert.NoError(err) {
		require.WithinDuration(start, event.Timestamp, 300*time.Millisecond)
	}
	if event, err := conf.TestGetOutputEvent(100 * time.Millisecond); assert.NoError(err) {
		require.WithinDuration(start, event.Timestamp, 300*time.Millisecond)
	}
}

func Test_input_lorem_module_format(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
input:
  - type: lorem
    worker: 1
    format: '{{.TimeFormat "20060102"}}|{{.Sentence 1 5}}'
    duration: "1s"
	`)))
	require.NoError(err)
	start := time.Now()
	require.NoError(conf.Start(ctx))

	time.Sleep(500 * time.Millisecond)
	if event, err := conf.TestGetOutputEvent(100 * time.Millisecond); assert.NoError(err) {
		require.WithinDuration(start, event.Timestamp, 300*time.Millisecond)
		require.Equal(start.Format("20060102"), strings.SplitN(event.Message, "|", 2)[0])
	}
}

func Test_input_lorem_module_fields(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
input:
  - type: lorem
    worker: 1
    duration: "1s"
    fields:
      host: host1
      prospector:
        type: log
	`)))
	require.NoError(err)
	start := time.Now()
	require.NoError(conf.Start(ctx))

	time.Sleep(500 * time.Millisecond)
	if event, err := conf.TestGetOutputEvent(100 * time.Millisecond); assert.NoError(err) {
		require.WithinDuration(start, event.Timestamp, 300*time.Millisecond)
		require.Equal("host1", event.GetString("host"))
		require.Equal("log", event.GetString("prospector.type"))
	}
}
