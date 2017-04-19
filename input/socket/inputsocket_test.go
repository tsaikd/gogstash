package inputsocket

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/tsaikd/gogstash/config"
)

var (
	logger = config.Logger
)

func init() {
	logger.Level = logrus.DebugLevel
	config.RegistInputHandler(ModuleName, InitHandler)
}

func Test_input_socket_module(t *testing.T) {
	require := require.New(t)
	require.NotNil(require)

	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
input:
  - type: socket
    socket: unix
    address: "/tmp/gogstash-test-unix.sock"
  - type: socket
    socket: unixpacket
    address: "/tmp/gogstash-test-unixpacket.sock"
  - type: socket
    socket: tcp
    address: ":9999"
	`)))
	require.NoError(err)
	require.NoError(conf.Start())

	waitsec := 10
	logger.Infof("Wait for %d seconds", waitsec)
	time.Sleep(time.Duration(waitsec) * time.Second)
	os.Remove("/tmp/gogstash-test-unix.sock")
	os.Remove("/tmp/gogstash-test-unixpacket.sock")
}
