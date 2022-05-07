package outputstatsd

import (
	"context"
	"io"
	"io/ioutil"
	"net"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
)

const (
	testAddr = ":0"
)

func init() {
	goglog.Logger.SetLevel(logrus.DebugLevel)
	config.RegistOutputHandler(ModuleName, InitHandler)
}

func Test_output_statsd_module(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	cCount := int32(0)
	cInc := int32(0)
	cDec := int32(0)
	cGauge := int32(0)
	cTiming := int32(0)
	received := make(chan bool)
	server := newServer(t, "udp", testAddr, func(p []byte) {
		s := strings.Split(string(p), "\n")
		l := len(s)
		if l > 0 && len(s[l-1]) == 0 {
			l--
		}
		for j := 0; j < l; j++ {
			if s[j] == "Log.staging.all.increment.200:1|c" {
				atomic.AddInt32(&cInc, 1) // increment
			} else if s[j] == "Log.staging.all.increment2.200:1|c" {
				atomic.AddInt32(&cInc, 1) // increment
			} else if s[j] == "Log.staging.all.decrement.200:-1|c" {
				atomic.AddInt32(&cDec, 1) // decrement
			} else if s[j] == "Log.staging.all.responce_time:0.12|ms" {
				atomic.AddInt32(&cTiming, 1)
			} else if s[j] == "Log.staging.all.responce_time2:0.12|ms" {
				atomic.AddInt32(&cTiming, 1)
			} else if s[j] == "Log.staging.all.count.200:4|c" {
				atomic.AddInt32(&cCount, 1)
			} else if s[j] == "Log.staging.all.gauge.200:4|g" {
				atomic.AddInt32(&cGauge, 1)
			} else {
				t.Errorf("invalid output [%d of %d]: %q", j, len(s), s[j])
			}
		}
		received <- true
	})
	defer server.Close()

	time.Sleep(10 * time.Millisecond)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
output:
  - type: statsd
    host: "` + server.addr + `"
    prefix: "Log.staging"
    increment:
      - "all.increment.%{logmsg.status}"
      - "all.increment2.%{logmsg.status}"
    decrement:
      - "all.decrement.%{logmsg.status}"
    count:
      - name: "all.count.%{logmsg.status}"
        value: "%{logmsg.count}"
    gauge:
      - name: "all.gauge.%{logmsg.status}"
        value: "%{logmsg.count}"
    timing:
      - name: "all.responce_time"
        value: "%{logmsg.time}"
      - name: "all.responce_time2"
        value: "%{logmsg.time}"
    `)))
	require.NoError(err)
	err = conf.Start(ctx)
	assert.NoErrorf(err, "statsd client error")

	conf.TestInputEvent(logevent.LogEvent{
		Timestamp: time.Now(),
		Message:   "outputstatsd test message",
		Extra: map[string]interface{}{
			"logmsg": map[string]interface{}{
				"status": int32(200),
				"time":   float64(0.12),
				"count":  int32(4),
			},
		},
	})

	select {
	case <-time.After(200 * time.Millisecond):
		t.Errorf("server received nothing after 100ms")
	case <-received:
	}

	assert.Equal(int32(1), cCount, "Count count")
	assert.Equal(int32(2), cInc, "Increment count")
	assert.Equal(int32(1), cDec, "Decrement count")
	assert.Equal(int32(1), cGauge, "Gauge count")
	assert.Equal(int32(2), cTiming, "Timing count")
}

type server struct {
	t      testing.TB
	addr   string
	closer io.Closer
	closed chan bool
}

func newServer(t testing.TB, network, addr string, f func([]byte)) *server {
	s := &server{t: t, closed: make(chan bool)}
	switch network {
	case "udp":
		laddr, err := net.ResolveUDPAddr("udp", addr)
		if err != nil {
			t.Fatal(err)
		}
		conn, err := net.ListenUDP("udp", laddr)
		if err != nil {
			t.Fatal(err)
		}
		s.closer = conn
		s.addr = conn.LocalAddr().String()
		go func() {
			buf := make([]byte, 1024)
			for {
				n, err := conn.Read(buf)
				if err != nil {
					s.closed <- true
					return
				}
				if n > 0 {
					f(buf[:n])
				}
			}
		}()
	case "tcp":
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			t.Fatal(err)
		}
		s.closer = ln
		s.addr = ln.Addr().String()
		go func() {
			for {
				conn, err := ln.Accept()
				if err != nil {
					s.closed <- true
					return
				}
				p, err := ioutil.ReadAll(conn)
				if err != nil {
					t.Fatal(err)
				}
				if err := conn.Close(); err != nil {
					t.Fatal(err)
				}
				f(p)
			}
		}()
	default:
		t.Fatalf("Invalid network: %q", network)
	}

	return s
}

func (s *server) Close() {
	if s.closer != nil {
		if err := s.closer.Close(); err != nil {
			s.t.Error(err)
		}
		<-s.closed
	}
}
