package inputnats

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
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

func TestInputNats(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	s, err := server.NewServer(&server.Options{
		Host:                  "127.0.0.1",
		Trace:                 true,
		Debug:                 true,
		DisableShortFirstPing: true,
		NoLog:                 true,
		NoSigs:                true,
	})
	require.NoError(err)
	go s.Start()
	defer s.Shutdown()
	time.Sleep(500 * time.Millisecond)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
input:
  - type:  "nats"
    host:  "` + s.ClientURL() + `"
    topic: "test.*"
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	// start a publisher client
	opts := []nats.Option{nats.Name("test")}

	nc, err := nats.Connect(nats.DefaultURL, opts...)
	require.NoError(err)

	err = nc.Publish("test.1", []byte(`{"foo":"bar"}`))
	require.NoError(err)

	err = nc.Flush()
	require.NoError(err)

	// check event
	if event, err := conf.TestGetOutputEvent(100 * time.Millisecond); assert.NoError(err) {
		require.EqualValues(map[string]any{"foo": "bar"}, event.Extra)
	}
}
