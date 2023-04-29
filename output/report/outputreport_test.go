package outputreport

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
	config.RegistOutputHandler(ModuleName, InitHandler)
}

func Test_output_report_module(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
output:
  - type: report
    interval: 1
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	conf.TestInputEvent(logevent.LogEvent{})
	conf.TestInputEvent(logevent.LogEvent{})
	time.Sleep(2 * time.Second)

	conf.TestInputEvent(logevent.LogEvent{})
	time.Sleep(2 * time.Second)

	conf.TestInputEvent(logevent.LogEvent{})
	conf.TestInputEvent(logevent.LogEvent{})
	conf.TestInputEvent(logevent.LogEvent{})
	conf.TestInputEvent(logevent.LogEvent{})
	time.Sleep(2 * time.Second)
}
