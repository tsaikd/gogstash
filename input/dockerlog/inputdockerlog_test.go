package inputdockerlog

import (
	"strings"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
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

func Test_input_dockerlog_module(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
input:
  - type: dockerlog
    dockerurl: "unix:///var/run/docker.sock"
    sincepath: "sincedb-test"
	`)))
	require.NoError(err)
	err = conf.Start()
	if err != nil {
		t.Log("skip test input dockerlog module")
		require.True(ErrorPingFailed.In(err))
		return
	}

	time.Sleep(500 * time.Millisecond)
	if event, err := conf.TestGetOutputEvent(100 * time.Millisecond); assert.NoError(err) {
		t.Log(event)
	}
}
