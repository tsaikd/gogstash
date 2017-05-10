package inputfile

import (
	"context"
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

func Test_input_file_module(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
input:
  - type: file
    path: "./README.md"
    sincedb_path: ""
    start_position: beginning
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	time.Sleep(500 * time.Millisecond)
	if event, err := conf.TestGetOutputEvent(100 * time.Millisecond); assert.NoError(err) {
		require.Equal("gogstash input file", event.Message)
	}
}
