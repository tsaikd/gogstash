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

func TestOutputConfig_checkIntInList(t1 *testing.T) {
	myList := []int{100, 200, 300}
	type fields struct {
		AcceptedHttpResult []int
	}
	type args struct {
		code int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{"100", fields{myList}, args{code: 100}, true},
		{"101", fields{myList}, args{code: 101}, false},
		{"empty list", fields{[]int{}}, args{code: 101}, false},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &OutputConfig{
				AcceptedHttpResult: tt.fields.AcceptedHttpResult,
			}
			if got := t.checkIntInList(tt.args.code); got != tt.want {
				t1.Errorf("checkIntInList() = %v, want %v", got, tt.want)
			}
		})
	}
}
