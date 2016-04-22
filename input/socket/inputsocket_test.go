package inputsocket

import (
	"os"
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

func TestSocketInput(t *testing.T) {
	require := require.New(t)
	require.NotNil(require)

	conf, err := config.LoadFromString(`{
		"input": [{
        "type": "socket",
        "socket": "unix",
        "address": "/tmp/unix.sock"
    },
    {
        "type": "socket",
        "socket": "unixpacket",
        "address": "/tmp/unixpacket.sock"
    },
    {
        "type": "socket",
        "socket": "tcp",
        "address": ":9999"
    }]
	}`)
	require.NoError(err)

	err = conf.RunInputs()
	require.NoError(err)

	waitsec := 10
	logger.Infof("Wait for %d seconds", waitsec)
	time.Sleep(time.Duration(waitsec) * time.Second)
	os.Remove("/tmp/unix.sock")
	os.Remove("/tmp/unixpacket.sock")
}
