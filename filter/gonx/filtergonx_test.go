package filtergonx

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
)

var (
	logger = config.Logger
)

func init() {
	logger.Level = logrus.DebugLevel
	config.RegistFilterHandler(ModuleName, InitHandler)
}

func Test_filter_gonx_module(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: gonx
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	hostname, err := os.Hostname()
	require.NoError(err)
	timestamp, err := time.Parse("2006-01-02T15:04:05Z", "2016-12-04T09:09:41.193Z")
	require.NoError(err)

	expectedEvent := logevent.LogEvent{
		Timestamp: timestamp,
		Message:   `223.137.229.27 - - [20/Mar/2017:00:42:51 +0000] "GET /explore HTTP/1.1" 200 1320 "-" "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/56.0.2924.87 Safari/537.36"`,
		Extra: map[string]interface{}{
			"body_bytes_sent": "1320",
			"host":            hostname,
			"http_referer":    "-",
			"http_user_agent": "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/56.0.2924.87 Safari/537.36",
			"offset":          0,
			"path":            "/test/file/path",
			"remote_addr":     "223.137.229.27",
			"remote_user":     "-",
			"request":         "GET /explore HTTP/1.1",
			"status":          "200",
			"time_local":      "20/Mar/2017:00:42:51 +0000",
		},
	}

	conf.TestInputEvent(logevent.LogEvent{
		Timestamp: timestamp,
		Message:   `223.137.229.27 - - [20/Mar/2017:00:42:51 +0000] "GET /explore HTTP/1.1" 200 1320 "-" "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/56.0.2924.87 Safari/537.36"`,
		Extra: map[string]interface{}{
			"host":   hostname,
			"path":   "/test/file/path",
			"offset": 0,
		},
	})

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal(expectedEvent, event)
	}
}
