package inputnats

import (
	"context"
	"os"
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

func TestMain(m *testing.M) {
	opts := server.Options{
		Host:     "127.0.0.1",
		Port:     4222,
		HTTPPort: -1,
		Cluster:  server.ClusterOpts{Port: -1},
		NoLog:    true,
		NoSigs:   true,
		Debug:    true,
		Trace:    true,
	}

	s, err := server.NewServer(&opts)
	if err != nil {
		panic(err)
	}

	go s.Start()
	defer s.Shutdown()

	ret := m.Run()

	os.Exit(ret)
}

func TestInputNats(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
input:
  - type:  "nats"
    host:  "127.0.0.1:4222"
    topic: "test.*"
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	time.Sleep(500 * time.Millisecond)

	// start a publisher client
	opts := []nats.Option{nats.Name("test")}

	nc, err := nats.Connect(nats.DefaultURL, opts...)
	require.NoError(err)

	err = nc.Publish("test.1", []byte("{\"foo\":\"bar\"}"))
	require.NoError(err)

	err = nc.Flush()
	require.NoError(err)

	// check event
	if event, err := conf.TestGetOutputEvent(100 * time.Millisecond); assert.NoError(err) {
		assert.Equal(map[string]interface{}{"foo": "bar"}, event.Extra)
	}
}
