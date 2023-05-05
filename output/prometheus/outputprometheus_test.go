package outputprometheus

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/tsaikd/gogstash/internal/httpctx"

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

func Test_output_prometheus_module(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
output:
  - type: prometheus
    address: "127.0.0.1:8080"
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	time.Sleep(1000 * time.Millisecond)

	// sending 1st event
	conf.TestInputEvent(logevent.LogEvent{
		Timestamp: time.Now(),
		Message:   "output prometheus test message",
	})
	value, err := getMetric()
	require.NoError(err)
	require.Equal("processed_messages_total 1", value)

	// sending second event
	conf.TestInputEvent(logevent.LogEvent{
		Timestamp: time.Now(),
		Message:   "output prometheus test message",
	})
	time.Sleep(500 * time.Millisecond)
	value, err = getMetric()
	require.NoError(err)
	require.Equal("processed_messages_total 2", value)
}

func getMetric() (string, error) {
	resp, err := httpctx.Get(context.Background(), "http://127.0.0.1:8080/metrics")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(body), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "processed_messages_total") {
			return line, nil
		}
	}
	return "", nil
}
