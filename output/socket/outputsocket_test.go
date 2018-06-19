package socket

import (
	"context"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stvp/go-udp-testing"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
)

var (
	logger = config.Logger
)

func init() {
	logger.Level = logrus.DebugLevel
	config.RegistOutputHandler(ModuleName, InitHandler)
}

func Test_output_report_module(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	udp.SetAddr(":9876")
	// This is needed to make sure that UDP Listener listens for data a bit longer, otherwise it will quit after a millisecond
	udp.Timeout = 5 * time.Second

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
output:
  - type: socket
    socket: udp
    address: localhost:9876
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	var messages []logevent.LogEvent
	var expected []string
	for _, m := range []string{"one", "two", "three", "four", "five", "six", "seven"} {
		msg := logevent.LogEvent{Message: m}
		messages = append(messages, msg)
		json, _ := msg.MarshalJSON()
		expected = append(expected, string(json)+"\n")
	}

	udp.ShouldReceiveAll(t, expected, func() {
		for _, m := range messages {
			conf.TestInputEvent(m)
			time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
		}
	})
}
