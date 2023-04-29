package outputsocket

import (
	"context"
	"fmt"
	"math/rand"
	"net"
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

	pc, err := net.ListenPacket("udp", ":9876")
	require.NoError(err)
	defer pc.Close()

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
		expected = append(expected, fmt.Sprintf(`"message":"%s"`, m))
	}

	buf := make([]byte, 1024)
	for i, m := range messages {
		conf.TestInputEvent(m)
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
		n, _, err := pc.ReadFrom(buf)
		require.NoError(err)
		require.Contains(string(buf[:n]), expected[i])
	}
}
