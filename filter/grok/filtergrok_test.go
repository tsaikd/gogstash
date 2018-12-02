package filtergrok

import (
	"context"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
)

var (
	fileName = "patterns"
	fileData = []byte(`NGINXTEST %{IP:addr} - (?:%{USERNAME:auth}|-) \[%{HTTPDATE:time}\] "(?:%{WORD:method} %{URIPATHPARAM:request}(?: HTTP/%{NUMBER:httpversion})?|-)" %{NUMBER:status} (?:%{NUMBER:body_bytes}|-) "(?:%{URI:referrer}|-)" (?:%{QS:agent}|-) %{NUMBER:request_time} (?:%{HOSTPORT:upstream_addr}|-)` + "\n")
)

func init() {
	goglog.Logger.SetLevel(logrus.DebugLevel)
	config.RegistFilterHandler(ModuleName, InitHandler)

	err := ioutil.WriteFile(fileName, fileData, 0644)
	if err != nil {
		panic(err)
	}
}

func Test_filter_grok_module(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: grok
    source: message
    match: ["%{NGINXTEST}"]
    patterns_path: "patterns"
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	hostname, err := os.Hostname()
	require.NoError(err)
	timestamp, err := time.Parse("2006-01-02T15:04:05Z", "2016-12-04T09:09:41.193Z")
	require.NoError(err)

	expectedEvent := logevent.LogEvent{
		Timestamp: timestamp,
		Message:   `8.8.8.8 - - [18/Jul/2017:16:10:16 +0300] "GET /index.html HTTP/1.1" 200 756 "https://google.com/" "Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.115 Safari/537.36" 0.1 192.168.0.1:8080`,
		Extra: map[string]interface{}{
			"host":          hostname,
			"path":          "/test/file/path",
			"offset":        0,
			"addr":          "8.8.8.8",
			"auth":          "-",
			"time":          "18/Jul/2017:16:10:16 +0300",
			"referrer":      "https://google.com/",
			"request_time":  "0.1",
			"method":        "GET",
			"request":       "/index.html",
			"httpversion":   "1.1",
			"status":        "200",
			"body_bytes":    "756",
			"port":          "",
			"agent":         "\"Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.115 Safari/537.36\"",
			"upstream_addr": "192.168.0.1:8080",
		},
	}

	conf.TestInputEvent(logevent.LogEvent{
		Timestamp: timestamp,
		Message:   `8.8.8.8 - - [18/Jul/2017:16:10:16 +0300] "GET /index.html HTTP/1.1" 200 756 "https://google.com/" "Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.115 Safari/537.36" 0.1 192.168.0.1:8080`,
		Extra: map[string]interface{}{
			"host":   hostname,
			"path":   "/test/file/path",
			"offset": 0,
		},
	})

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal(expectedEvent, event)
	}

	conf.TestInputEvent(logevent.LogEvent{
		Timestamp: timestamp,
		Message:   `8.8.8.8 - - [18/Jul/2017:16:10:16 +0300] "GET /index.html HTTP/1.1" 200 756 "https://google.com/" "Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.115 Safari/537.36"`,
		Extra: map[string]interface{}{
			"host":   hostname,
			"path":   "/test/file/path",
			"offset": 0,
		},
	})

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Contains(event.Tags, ErrorTag)
	}

	err = os.Remove(fileName)
	require.NoError(err)
}
