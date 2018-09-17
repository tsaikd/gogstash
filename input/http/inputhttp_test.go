package inputhttp

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
)

func init() {
	goglog.Logger.SetLevel(logrus.DebugLevel)
	config.RegistInputHandler(ModuleName, InitHandler)
}

func Test_input_http_module(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
input:
  - type: http
    method: GET
    url: "http://127.0.0.1/"
    interval: 3
    codec:
      - type: "default"
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	time.Sleep(500 * time.Millisecond)
	if event, err := conf.TestGetOutputEvent(100 * time.Millisecond); assert.NoError(err) {
		t.Log(event)
	}
}
