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

	serverRecvMsg := make(chan []byte, 1)
	handler := func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		data, err := io.ReadAll(r.Body)
		require.NoError(err)
		serverRecvMsg <- data
	}

	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
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

	select {
	case <-ctx.Done():
		t.Fatal("timeout")
	case msg := <-serverRecvMsg:
		require.Contains(string(msg), `"message":"outputhttp test message"`)
	}
}

func TestMapFromInts(t *testing.T) {
	input := []int{100, 200, 300}
	myMap := MapFromInts(input)
	// check length
	if len(myMap) != len(input) {
		t.Error("Length incorrect")
	}
	// check if first elem is in list
	if _, ok := myMap[input[0]]; !ok {
		t.Error("First element not in list")
	}
	// check element not in list
	if _, ok := myMap[-1]; ok {
		t.Error("Found element not in list")
	}
}
