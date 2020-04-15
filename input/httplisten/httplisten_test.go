package inputhttplisten

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

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

func Test_input_httplisten_module(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
input:
  - type: httplisten
    address: "127.0.0.1:8089"
    path: "/"
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	time.Sleep(500 * time.Millisecond)

	resp, err := http.Post("http://127.0.0.1:8089/", "application/json", bytes.NewReader([]byte("{\"foo\":\"bar\"}")))
	require.NoError(err)
	defer resp.Body.Close()

	assert.Equal(http.StatusOK, resp.StatusCode)
	data, err := ioutil.ReadAll(resp.Body)
	require.NoError(err)
	assert.Equal([]byte{}, data)

	if event, err := conf.TestGetOutputEvent(100 * time.Millisecond); assert.NoError(err) {
		assert.Equal(map[string]interface{}{"foo": "bar"}, event.Extra)
	}
}

func Test_input_httpslisten_module(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
input:
  - type: httplisten
    address: "127.0.0.1:8989"
    path: "/tls/"
    cert: "./server.pem"
    key:  "./server.key"
        `)))

	require.NoError(err)
	require.NoError(conf.Start(ctx))

	time.Sleep(500 * time.Millisecond)

	rootPEM, err := ioutil.ReadFile("./root.pem")
	require.NoError(err)
	roots := x509.NewCertPool()
	assert.NotNil(roots)
	ok := roots.AppendCertsFromPEM(rootPEM)
	assert.Equal(ok, true)

	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{RootCAs: roots}}}

	resp, err := client.Post("https://127.0.0.1:8989/tls/", "application/json", bytes.NewReader([]byte("{\"foo\":\"bar\"}")))
	require.NoError(err)
	defer resp.Body.Close()
	assert.Equal(http.StatusOK, resp.StatusCode)
	data, err := ioutil.ReadAll(resp.Body)
	require.NoError(err)
	assert.Equal([]byte{}, data)

	if event, err := conf.TestGetOutputEvent(100 * time.Millisecond); assert.NoError(err) {
		assert.Equal(map[string]interface{}{"foo": "bar"}, event.Extra)
	}
}

func Test_input_httpslisten2_module(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
input:
  - type: httplisten
    address: "127.0.0.1:8999"
    path: "/tls2/"
    cert: "./server.pem"
    key:  "./server.key"
    ca:   "./root.pem"
        `)))

	require.NoError(err)
	require.NoError(conf.Start(ctx))
	time.Sleep(500 * time.Millisecond)

	rootPEM, err := ioutil.ReadFile("./root.pem")
	require.NoError(err)
	roots := x509.NewCertPool()
	assert.NotNil(roots)
	ok := roots.AppendCertsFromPEM(rootPEM)
	assert.Equal(ok, true)

	clientCert, err := tls.LoadX509KeyPair("./client.pem", "./client.key")
	require.NoError(err)

	// case 1: without client cert
	tlsConfig := tls.Config{
		RootCAs: roots,
	}

	transport := http.Transport{
		TLSClientConfig: &tlsConfig,
	}

	client := &http.Client{Transport: &transport}

	_, err = client.Post("https://127.0.0.1:8999/tls2/", "application/json", bytes.NewReader([]byte("{\"foo2\":\"bar2\"}")))
	assert.NotNil(err)

	// case 2: with correct client cert
	tlsConfig.Certificates = []tls.Certificate{clientCert}
	resp, err := client.Post("https://127.0.0.1:8999/tls2/", "application/json", bytes.NewReader([]byte("{\"foo2\":\"bar2\"}")))
	require.NoError(err)
	defer resp.Body.Close()
	assert.Equal(http.StatusOK, resp.StatusCode)

	data, err := ioutil.ReadAll(resp.Body)
	require.NoError(err)
	assert.Equal([]byte{}, data)

	if event, err := conf.TestGetOutputEvent(100 * time.Millisecond); assert.NoError(err) {
		assert.Equal(map[string]interface{}{"foo2": "bar2"}, event.Extra)
	}
}
