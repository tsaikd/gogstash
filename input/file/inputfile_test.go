package inputfile

import (
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

func Test_main(t *testing.T) {
	require := require.New(t)
	require.NotNil(require)

	conf, err := config.LoadFromJSON([]byte(`{
		"input": [{
			"type": "file",
			"path": "/tmp/log/syslog",
			"sincedb_path": "",
			"start_position": "beginning"
		}]
	}`))
	require.NoError(err)

	err = conf.RunInputs()
	require.NoError(err)

	waitsec := 10
	logger.Infof("Wait for %d seconds", waitsec)
	time.Sleep(time.Duration(waitsec) * time.Second)
}
