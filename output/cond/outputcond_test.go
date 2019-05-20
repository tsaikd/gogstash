package outputcond

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
	"github.com/tsaikd/gogstash/output/stdout"
)

func init() {
	goglog.Logger.SetLevel(logrus.DebugLevel)
	config.RegistOutputHandler(outputstdout.ModuleName, outputstdout.InitHandler)
	config.RegistOutputHandler(ModuleName, InitHandler)
}

func Test_filter_cond_module_invalid(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
output:
  - type: cond
    condition: "!'level' == 'ERROR'"
    output:
      - type: add_field
        key: foo
        value: bar    
    `)))
	require.Nil(err)
	_, err = InitHandler(context.TODO(), &conf.OutputRaw[0])
	require.NotNil(err)
}

func Test_output_cond_module(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
output:
  - type: cond
    condition: "level == 'ERROR'"
    output:
      - type: stdout
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	conf.TestInputEvent(logevent.LogEvent{
		Timestamp: time.Now(),
		Message:   "outputstdout test message",
		Extra: map[string]interface{}{
			"level": "ERROR",
			"foo":   "bar",
		},
	})

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal("outputstdout test message", event.Message)
	}

	conf.TestInputEvent(logevent.LogEvent{
		Timestamp: time.Now(),
		Message:   "outputstdout test message",
		Extra: map[string]interface{}{
			"level": "WARN",
		},
	})

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal("outputstdout test message", event.Message)
	}
}
