package outputhttp

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
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

func init() {
	goglog.Logger.SetLevel(logrus.DebugLevel)
	config.RegistOutputHandler(ModuleName, InitHandler)
}

func Test_output_http_module(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	h := httptest.NewRecorder()
	handler := func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		h.HeaderMap = r.Header
		io.Copy(h, r.Body)
	}

	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
output:
  - type: http
    urls: ["` + server.URL + `"]
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	conf.TestInputEvent(logevent.LogEvent{
		Timestamp: time.Now(),
		Message:   "outputhttp test message",
	})

	time.Sleep(100 * time.Millisecond)
	assert.Contains(h.Body.String(), "\"message\":\"outputhttp test message\"")
}
