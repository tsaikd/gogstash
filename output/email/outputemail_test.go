package outputemail

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

func Test_output_email_module(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	// Please fill the correct email info xxx is just a placeholder
	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
output:
  - type: email
    address: "xxx"
    from: "xxx"
    to: "xxx"
    cc: "xxx"
    use_tls: false
    port: 25
    username: "xxx"
    password: "xxx"
    subject: "outputemail test subject"
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	conf.TestInputEvent(logevent.LogEvent{
		Timestamp: time.Now(),
		Message:   "outputemail test message",
	})

	if event, err := conf.TestGetOutputEvent(500 * time.Millisecond); assert.NoError(err) {
		require.Equal("outputemail test message", event.Message)
	}
}
