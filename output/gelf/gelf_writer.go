// Copyright 2012 SocialCode. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package outputgelf

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/tsaikd/gogstash/internal/httpctx"
)

// regex used later
var forbiddenExtraKeys = [7]string{
	"_version",
	"_host",
	"_short_message",
	"_full_message",
	"_timestamp",
	"_level",
}

type GELFConfig struct {
	Host             string
	ChunkSize        int          // Default Value: 1420
	CompressionLevel int          // Default: 1
	CompressionType  CompressType // Default: Gzip
}

type GELFWriter interface {
	WriteCustomMessage(ctx context.Context, m *Message) error
	WriteMessage(ctx context.Context, sm *SimpleMessage) error
}

// UDPWriter implements io.Writer and is used to send both discrete
// messages to a graylog2 server, or data from a stream-oriented
// interface (like the functions in log).
type UDPWriter struct {
	config *GELFConfig

	conn net.Conn
	mu   sync.Mutex

	zw                 writerCloserResetter
	zwCompressionLevel int
	zwCompressionType  CompressType
}

// What compression type the writer should use when sending messages
// to the graylog2 server
type CompressType int

const (
	CompressGzip CompressType = iota
	CompressZlib
	NoCompress
)

// Message represents the contents of the GELF message.  It is gzipped
// before sending. https://docs.graylog.org/docs/gelf
type Message struct {
	Extra     map[string]any `json:"-"`
	Full      string         `json:"full_message"`
	Host      string         `json:"host"`
	Level     int32          `json:"level"`
	Short     string         `json:"short_message"`
	Timestamp int64          `json:"timestamp"`
	Version   string         `json:"version"`
}

type SimpleMessage struct {
	Extra     map[string]any
	Host      string
	Level     int32
	Message   string
	Timestamp time.Time
}

type innerMessage Message // against circular (Un)MarshalJSON

// Used to control GELF chunking.  Should be less than (MTU - len(UDP
// header)).
//
// TODO: generate dynamically using Path MTU Discovery?
const (
	DefaultChunkSize = 1420
	chunkedHeaderLen = 12
)

var (
	magicChunked = []byte{0x1e, 0x0f}
	hostname     string
)

// numChunks returns the number of GELF chunks necessary to transmit
// the given compressed buffer.
func numChunks(b []byte, chunkSize int) int {
	lenB := len(b)
	if lenB <= chunkSize {
		return 1
	}
	return len(b)/(chunkSize-chunkedHeaderLen) + 1
}

// NewWriter returns a new GELFWriter. This writer can be used to send the
// output of the standard Go log functions to a central GELF server by
// passing it to log.SetOutput()
func NewWriter(config GELFConfig) (GELFWriter, error) {
	// handle config
	if config.Host == "" {
		return nil, fmt.Errorf("missing host")
	}

	if config.ChunkSize == 0 {
		config.ChunkSize = DefaultChunkSize
	}

	if config.CompressionLevel == 0 {
		config.CompressionLevel = flate.BestSpeed
	}

	var err error
	if hostname, err = os.Hostname(); err != nil {
		panic(err)
	}

	// handle http / udp writing method
	if strings.HasPrefix(config.Host, "http") {
		return newHTTPWriter(&config)
	}

	return newUDPWriter(&config)
}

func newHTTPWriter(config *GELFConfig) (GELFWriter, error) {
	httpClient := &http.Client{
		Transport: &http.Transport{},
		Timeout:   10 * time.Second,
	}

	return HTTPWriter{
		config:     config,
		httpClient: httpClient,
	}, nil
}

func newUDPWriter(config *GELFConfig) (GELFWriter, error) {
	var err error

	w := &UDPWriter{
		config: config,
	}

	if w.conn, err = net.Dial("udp", config.Host); err != nil {
		return nil, err
	}

	return w, nil
}

func constructMessage(sm *SimpleMessage) *Message {
	message := Message{
		Extra:     sm.Extra,
		Level:     6, // Default: Info
		Timestamp: int64(time.Now().UnixNano() / int64(time.Millisecond) / 1000),
		Version:   "1.1",
	}

	message.Host = hostname

	// Hostname
	if sm.Host != "" {
		message.Host = sm.Host
	}

	if sm.Level != 0 {
		message.Level = sm.Level
	}

	if !sm.Timestamp.IsZero() {
		message.Timestamp = int64(sm.Timestamp.UnixNano() / int64(time.Millisecond) / 1000)
	}

	if sm.Message != "" {
		message.Short = sm.Message

		if lines := strings.Split(sm.Message, "\n"); len(lines) > 0 {
			message.Short = lines[0]
			message.Full = sm.Message
		}
	}

	return &message
}

func prepareExtra(e map[string]any) (map[string]any, error) {
	cleanExtra := make(map[string]any)

	for k, v := range e {
		newKey := k

		if !strings.HasPrefix(newKey, "_") {
			newKey = "_" + newKey
		}

		for _, forbiddenKey := range forbiddenExtraKeys {
			if newKey == forbiddenKey {
				return nil, fmt.Errorf("key %s is not allowed", k)
			}
		}

		cleanExtra[newKey] = v
	}

	return cleanExtra, nil
}

/**
* UDP IMPLEMENTATION
**/

// writes the gzip compressed byte array to the connection as a series
// of GELF chunked messages.  The header format is documented at
// https://github.com/Graylog2/graylog2-docs/wiki/GELF as:
//
//	2-byte magic (0x1e 0x0f), 8 byte id, 1 byte sequence id, 1 byte
//	total, chunk-data
func (w *UDPWriter) writeChunked(zBytes []byte) (err error) {
	b := make([]byte, 0, w.config.ChunkSize)
	buf := bytes.NewBuffer(b)
	nChunksI := numChunks(zBytes, w.config.ChunkSize)
	if nChunksI > 255 {
		return fmt.Errorf("msg too large, would need %d chunks", nChunksI)
	}
	nChunks := uint8(nChunksI)
	// use urandom to get a unique message id
	msgID := make([]byte, 8)
	n, err := io.ReadFull(rand.Reader, msgID)
	if err != nil || n != 8 {
		return fmt.Errorf("rand.Reader: %d/%s", n, err)
	}

	bytesLeft := len(zBytes)
	for i := uint8(0); i < nChunks; i++ {
		buf.Reset()
		// manually write header.  Don't care about
		// host/network byte order, because the spec only
		// deals in individual bytes.
		buf.Write(magicChunked) // magic
		buf.Write(msgID)
		buf.WriteByte(i)
		buf.WriteByte(nChunks)
		// slice out our chunk from zBytes
		chunkLen := (w.config.ChunkSize - chunkedHeaderLen)
		if chunkLen > bytesLeft {
			chunkLen = bytesLeft
		}
		off := int(i) * (w.config.ChunkSize - chunkedHeaderLen)
		chunk := zBytes[off : off+chunkLen]
		buf.Write(chunk)

		// write this chunk, and make sure the write was good
		n, err := w.conn.Write(buf.Bytes())
		if err != nil {
			return fmt.Errorf("Write (chunk %d/%d): %s", i,
				nChunks, err)
		}
		if n != len(buf.Bytes()) {
			return fmt.Errorf("Write len: (chunk %d/%d) (%d/%d)",
				i, nChunks, n, len(buf.Bytes()))
		}

		bytesLeft -= chunkLen
	}

	if bytesLeft != 0 {
		return fmt.Errorf("error: %d bytes left after sending", bytesLeft)
	}
	return nil
}

type bufferedWriter struct {
	buffer io.Writer
}

func (bw bufferedWriter) Write(p []byte) (n int, err error) {
	return bw.buffer.Write(p)
}

func (bw bufferedWriter) Close() error {
	return nil
}

func (bw *bufferedWriter) Reset(w io.Writer) {
	bw.buffer = w
}

type writerCloserResetter interface {
	io.WriteCloser
	Reset(w io.Writer)
}

// WriteCustomMessage sends the specified message to the GELF server
// specified in the call to NewWriter(). It assumes all the fields are
// filled out appropriately. In general, clients will want to use
// Write, rather than WriteMessage.
func (w *UDPWriter) WriteCustomMessage(ctx context.Context, m *Message) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	mBytes, err := json.Marshal(m)
	if err != nil {
		return err
	}

	var zBuf bytes.Buffer

	// . If compression settings have changed, a new writer is required.
	if w.zwCompressionType != w.config.CompressionType || w.zwCompressionLevel != w.config.CompressionLevel {
		w.zw = nil
	}

	switch w.config.CompressionType {
	case CompressGzip:
		if w.zw == nil {
			w.zw, err = gzip.NewWriterLevel(&zBuf, w.config.CompressionLevel)
		}
	case CompressZlib:
		if w.zw == nil {
			w.zw, err = zlib.NewWriterLevel(&zBuf, w.config.CompressionLevel)
		}
	case NoCompress:
		w.zw = &bufferedWriter{}
	default:
		panic(fmt.Sprintf("unknown compression type %d",
			w.config.CompressionType))
	}

	if err != nil {
		return err
	}

	w.zw.Reset(&zBuf)

	if _, err = w.zw.Write(mBytes); err != nil {
		return err
	}
	w.zw.Close()

	zBytes := zBuf.Bytes()
	if numChunks(zBytes, w.config.ChunkSize) > 1 {
		return w.writeChunked(zBytes)
	}

	n, err := w.conn.Write(zBytes)
	if err != nil {
		return err
	}
	if n != len(zBytes) {
		return fmt.Errorf("bad write (%d/%d)", n, len(zBytes))
	}

	return nil
}

// WriteMessage allow to send messsage to gelf Server
// It only request basic fields and will handle conversion & co
func (w *UDPWriter) WriteMessage(ctx context.Context, sm *SimpleMessage) error {
	cleanExtra, err := prepareExtra(sm.Extra)
	if err != nil {
		return err
	}

	sm.Extra = cleanExtra

	return w.WriteCustomMessage(ctx, constructMessage(sm))
}

func (m *Message) MarshalJSON() ([]byte, error) {
	var err error
	var b, eb []byte

	extra := m.Extra
	b, err = json.Marshal((*innerMessage)(m))
	m.Extra = extra
	if err != nil {
		return nil, err
	}

	if len(extra) == 0 {
		return b, nil
	}

	if eb, err = json.Marshal(extra); err != nil {
		return nil, err
	}

	// merge serialized message + serialized extra map
	b[len(b)-1] = ','
	return append(b, eb[1:]...), nil
}

/**
* HTTP IMPLEMENTATION
**/

// HTTPWriter implements the GELFWriter interface, and cannot be used
// as an io.Writer
type HTTPWriter struct {
	config     *GELFConfig
	httpClient *http.Client
}

func (h HTTPWriter) WriteCustomMessage(ctx context.Context, m *Message) error {
	mBytes, err := json.Marshal(m)
	if err != nil {
		return err
	}

	resp, err := httpctx.ClientPost(ctx, h.httpClient, h.config.Host, "application/json", bytes.NewBuffer(mBytes))
	if err != nil {
		return err
	}
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	if resp.StatusCode != 204 {
		return fmt.Errorf("got code %s, expected 204", resp.Status)
	}

	return nil
}

// WriteMessage allow to send messsage to gelf Server
// It only request basic fields and will handle conversion & co
func (h HTTPWriter) WriteMessage(ctx context.Context, sm *SimpleMessage) error {
	cleanExtra, err := prepareExtra(sm.Extra)
	if err != nil {
		return err
	}

	sm.Extra = cleanExtra

	return h.WriteCustomMessage(ctx, constructMessage(sm))
}
