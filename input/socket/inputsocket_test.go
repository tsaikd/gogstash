package inputsocket

import (
	"context"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	codecjson "github.com/tsaikd/gogstash/codec/json"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
)

func init() {
	goglog.Logger.SetLevel(logrus.DebugLevel)
	config.RegistInputHandler(ModuleName, InitHandler)
	config.RegistCodecHandler(codecjson.ModuleName, codecjson.InitHandler)
}

func Test_input_socket_module_unix(t *testing.T) {
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
input:
  - type: socket
    socket: unix
    address: "/tmp/gogstash-test-unix.sock"
  - type: socket
    socket: unixpacket
    address: "/tmp/gogstash-test-unixpacket.sock"
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	waitsec := 10
	goglog.Logger.Infof("Wait for %d seconds", waitsec)
	time.Sleep(time.Duration(waitsec) * time.Second)
	os.Remove("/tmp/gogstash-test-unix.sock")
	os.Remove("/tmp/gogstash-test-unixpacket.sock")
}

func Test_input_socket_module_tcp(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
input:
  - type: socket
    socket: tcp
    address: "127.0.0.1:9999"
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	time.Sleep(500 * time.Millisecond)

	conn, err := net.Dial("tcp", "127.0.0.1:9999")
	require.NoError(err)
	defer conn.Close()

	testWriteData(t, conf, conn)
}

func Test_input_socket_module_udp(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
input:
  - type: socket
    socket: udp
    address: "127.0.0.1:9998"
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	time.Sleep(500 * time.Millisecond)

	conn, err := net.Dial("udp", "127.0.0.1:9998")
	require.NoError(err)
	defer conn.Close()

	testWriteData(t, conf, conn)
}

func testWriteData(t *testing.T, conf config.Config, conn net.Conn) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	_, err := conn.Write([]byte("{\"foo\":\"bar\"}\n"))
	require.NoError(err)

	time.Sleep(200 * time.Millisecond)
	if event, err := conf.TestGetOutputEvent(100 * time.Millisecond); assert.NoError(err) {
		assert.Equal(map[string]any{"foo": "bar"}, event.Extra)
	}

	// malformed data
	_, err = conn.Write([]byte("114514\n"))
	require.NoError(err)

	time.Sleep(200 * time.Millisecond)
	if event, err := conf.TestGetOutputEvent(100 * time.Millisecond); assert.NoError(err) {
		assert.Equal("114514\n", event.Message)
	}

	_, err = conn.Write([]byte("{\"foo\":\"bar\"dr.who}\n"))
	require.NoError(err)

	time.Sleep(200 * time.Millisecond)
	if event, err := conf.TestGetOutputEvent(100 * time.Millisecond); assert.NoError(err) {
		assert.Equal("{\"foo\":\"bar\"dr.who}\n", event.Message)
	}

	// continued
	_, err = conn.Write([]byte("{\"bar\":\"foo\"}\n"))
	require.NoError(err)

	time.Sleep(200 * time.Millisecond)
	if event, err := conf.TestGetOutputEvent(100 * time.Millisecond); assert.NoError(err) {
		assert.Equal(map[string]any{"bar": "foo"}, event.Extra)
	}
}
