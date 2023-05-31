package inputbeats

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	client "github.com/elastic/go-lumber/client/v2"
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

func Test_input_beats_module(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
input:
  - type: beats
    host: "127.0.0.1"
    port: 5044
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	time.Sleep(1000 * time.Millisecond)

	c, err := client.Dial("127.0.0.1:5044")
	require.NoError(err)
	defer c.Close()

	ts := time.Date(2019, time.January, 4, 0, 55, 36, 0, time.UTC)
	eventData := map[string]any{
		"@timestamp": ts.Format(time.RFC3339),
		"message":    "test message",
	}
	data := []any{eventData}
	err = c.Send(data)
	require.NoError(err)

	if event, err := conf.TestGetOutputEvent(500 * time.Millisecond); assert.NoError(err) {
		assert.Equal(ts, event.Timestamp)
		assert.Equal(eventData["message"], event.Message)
	}
}

func Test_input_beats_module_tls(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	// generate key pair
	// https://golang.org/src/crypto/tls/generate_cert.go
	var priv any
	var err error
	rsaBits := 2048
	priv, err = rsa.GenerateKey(rand.Reader, rsaBits)
	require.NoError(err)

	notBefore := time.Now()
	notAfter := notBefore.Add(time.Hour * 24 * 30) // valid for 30 days

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	require.NoError(err)

	ip := net.ParseIP("127.0.0.1")
	require.NotNil(ip)

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"gogstash"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,

		// IP addresses
		IPAddresses: []net.IP{ip},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey(priv), priv)
	require.NoError(err)

	certOut, err := os.Create("cert.pem")
	require.NoError(err)
	err = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	require.NoError(err)
	err = certOut.Close()
	require.NoError(err)
	defer os.Remove("cert.pem")

	keyOut, err := os.OpenFile("key.pem", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	require.NoError(err)
	err = pem.Encode(keyOut, pemBlockForKey(priv))
	require.NoError(err)
	err = keyOut.Close()
	require.NoError(err)
	defer os.Remove("key.pem")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
input:
  - type: beats
    host: "127.0.0.1"
    port: 5044
    ssl: true
    ssl_certificate: cert.pem
    ssl_key: key.pem
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	time.Sleep(1000 * time.Millisecond)

	conn, err := tls.Dial("tcp", "127.0.0.1:5044", &tls.Config{InsecureSkipVerify: true})
	require.NoError(err)
	c, err := client.NewWithConn(conn)
	require.NoError(err)
	defer c.Close()
	t.Log("client: connected to:", conn.RemoteAddr())

	state := conn.ConnectionState()
	for _, v := range state.PeerCertificates {
		pubkey, _ := x509.MarshalPKIXPublicKey(v.PublicKey)
		t.Log("client: subject:", v.Subject, "pubkey:", pubkey)
	}
	t.Log("client: handshake:", state.HandshakeComplete)

	ts := time.Date(2019, time.January, 4, 0, 55, 36, 0, time.UTC)
	eventData := map[string]any{
		"@timestamp": ts.Format(time.RFC3339),
		"message":    "test message",
	}
	data := []any{eventData}
	err = c.Send(data)
	require.NoError(err)

	if event, err := conf.TestGetOutputEvent(500 * time.Millisecond); assert.NoError(err) {
		assert.Equal(ts, event.Timestamp)
		assert.Equal(eventData["message"], event.Message)
	}
}

func publicKey(priv any) any {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
	}
}

func pemBlockForKey(priv any) *pem.Block {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)}
	case *ecdsa.PrivateKey:
		b, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to marshal ECDSA private key: %v", err)
			os.Exit(2)
		}
		return &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}
	default:
		return nil
	}
}
