package outputclickhouse

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/tsaikd/KDGoLib/errutil"

	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
)

// ModuleName is the name used in config file
const ModuleName = "clickhouse"

// errors
var (
	ErrNoValidURLs = errutil.NewFactory("clickhouse output: no valid URLs found")
	ErrNoTable     = errutil.NewFactory("clickhouse output: table is required")
)

// OutputConfig holds the configuration json fields and internal objects
type OutputConfig struct {
	config.OutputConfig

	// URLs points to one or more ClickHouse HTTP/HTTPS endpoints, e.g. "http://clickhouse:8123" or "https://clickhouse:8443"
	URLs []string `json:"urls"`

	// Auth is used for HTTP Basic Auth in the form "user:password"
	Auth string `json:"auth"`

	// Table is the full ClickHouse table name, e.g. "logs.ids" or "logs.ngfw"
	Table string `json:"table"`

	// BatchSize defines how many events are buffered before a flush is triggered
	BatchSize int `json:"batch_size,omitempty"`

	// FlushInterval is a string duration (e.g. "2s", "1m") read from config.
	// It is parsed into flushInterval at InitHandler.
	FlushInterval string `json:"flush_interval,omitempty"`

	// TsField, if set, defines which field from event.Extra
	// should be copied into the "ts" column sent to ClickHouse.
	// The value is converted to a ClickHouse-compatible string when possible.
	TsField string `json:"ts_field,omitempty"`

	// FailOnError controls whether the plugin should return an error
	// when ClickHouse responds with a non-2xx HTTP status.
	// If false, the plugin only logs the error and continues.
	FailOnError bool `json:"fail_on_error,omitempty"`

	// SSLSkipVerify controls TLS certificate verification.
	// If true, the HTTP client will accept self-signed / invalid certificates.
	// WARNING: this reduces security and should only be used in trusted environments.
	SSLSkipVerify bool `json:"ssl_insecure_skip_verify,omitempty"`

	// internal HTTP client
	httpClient *http.Client

	// internal parsed flush interval
	flushInterval time.Duration

	// internal buffer and synchronization
	mu        sync.Mutex
	buffer    []map[string]any
	lastFlush time.Time
}

// DefaultOutputConfig returns an OutputConfig struct with default values
func DefaultOutputConfig() OutputConfig {
	return OutputConfig{
		OutputConfig: config.OutputConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		BatchSize:     1000,
		FlushInterval: "2s",
	}
}

// InitHandler initializes the output plugin
func InitHandler(
	ctx context.Context,
	raw config.ConfigRaw,
	control config.Control,
) (config.TypeOutputConfig, error) {
	conf := DefaultOutputConfig()
	if err := config.ReflectConfig(raw, &conf); err != nil {
		return nil, err
	}

	if len(conf.URLs) == 0 {
		return nil, ErrNoValidURLs.New(nil)
	}
	if conf.Table == "" {
		return nil, ErrNoTable.New(nil)
	}
	if conf.BatchSize <= 0 {
		conf.BatchSize = 1000
	}
	if conf.FlushInterval == "" {
		conf.FlushInterval = "2s"
	}

	// Parse flush interval string (e.g. "2s", "1m")
	d, err := time.ParseDuration(conf.FlushInterval)
	if err != nil {
		return nil, fmt.Errorf("clickhouse output: invalid flush_interval %q: %w", conf.FlushInterval, err)
	}
	conf.flushInterval = d

	// HTTP client with optional TLS config
	tr := &http.Transport{
		DisableCompression: false,
	}
	if conf.SSLSkipVerify {
		tr.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true, //nolint:gosec // explicitly requested via config
		}
	}
	conf.httpClient = &http.Client{Transport: tr}

	conf.buffer = make([]map[string]any, 0, conf.BatchSize)
	conf.lastFlush = time.Now()

	// Periodic flush goroutine
	go conf.flushLoop(ctx)

	return &conf, nil
}

// Output is called for each log event
func (t *OutputConfig) Output(ctx context.Context, event logevent.LogEvent) (err error) {
	row := make(map[string]any)

	// Copy all Extra fields into the row
	for k, v := range event.Extra {
		row[k] = v
	}

	// Add message field
	if event.Message != "" {
		row["message"] = event.Message
	}

	// Add tags if present
	if len(event.Tags) > 0 {
		row["tags"] = event.Tags
	}

	// Optionally set explicit "ts" column from a given field
	if t.TsField != "" {
		if v, ok := event.Extra[t.TsField]; ok {
			if formatted, ok2 := formatTimestampValue(v); ok2 {
				row["ts"] = formatted
			} else {
				// Fallback: store raw value if conversion failed
				row["ts"] = v
			}
		}
	}

	t.mu.Lock()
	t.buffer = append(t.buffer, row)
	needFlush := len(t.buffer) >= t.BatchSize
	t.mu.Unlock()

	// Flush by size
	if needFlush {
		return t.flush(ctx)
	}

	return nil
}

// flushLoop performs time-based flushing
func (t *OutputConfig) flushLoop(ctx context.Context) {
	ticker := time.NewTicker(t.flushInterval / 2)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Final flush on shutdown
			_ = t.flush(ctx)
			return
		case <-ticker.C:
			t.mu.Lock()
			hasData := len(t.buffer) > 0
			shouldFlush := hasData && time.Since(t.lastFlush) >= t.flushInterval
			t.mu.Unlock()

			if shouldFlush {
				_ = t.flush(ctx)
			}
		}
	}
}

// flush sends the current batch to ClickHouse using JSONEachRow
func (t *OutputConfig) flush(ctx context.Context) error {
	t.mu.Lock()
	if len(t.buffer) == 0 {
		t.mu.Unlock()
		return nil
	}

	// Swap buffer to avoid blocking Output for long
	batch := t.buffer
	t.buffer = make([]map[string]any, 0, t.BatchSize)
	t.lastFlush = time.Now()
	t.mu.Unlock()

	var buf bytes.Buffer
	enc := jsoniter.NewEncoder(&buf)
	for _, row := range batch {
		if err := enc.Encode(row); err != nil {
			goglog.Logger.Errorf("output clickhouse: json encode error: %v", err)
		}
	}

	data := buf.Bytes()
	url := t.pickURL()

	// Build ClickHouse query: INSERT INTO <table> FORMAT JSONEachRow
	query := "INSERT INTO " + t.Table + " FORMAT JSONEachRow"

	// Send query via "query" parameter and JSONEachRow in the request body
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		goglog.Logger.Errorf("output clickhouse: %v", err)
		return err
	}

	q := req.URL.Query()
	q.Set("query", query)
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "gogstash/output"+ModuleName)

	if t.Auth != "" {
		req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(t.Auth)))
	}

	resp, err := t.httpClient.Do(req)
	if err != nil {
		goglog.Logger.Errorf("output clickhouse: %v", err)
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		goglog.Logger.Errorf("output clickhouse statusCode: %v error: %s", resp.StatusCode, body)
		if t.FailOnError {
			return fmt.Errorf("clickhouse output: non-2xx status %d: %s", resp.StatusCode, string(body))
		}
		// If FailOnError is false, we only log and continue
	}

	return nil
}

// pickURL selects one URL randomly from the configured list
func (t *OutputConfig) pickURL() string {
	if len(t.URLs) == 1 {
		return t.URLs[0]
	}
	i := rand.Intn(len(t.URLs))
	return t.URLs[i]
}

// formatTimestampValue tries to convert a value into a ClickHouse-compatible
// DateTime string. It returns (formatted, true) if conversion was successful.
func formatTimestampValue(v any) (string, bool) {
	switch ts := v.(type) {
	case time.Time:
		// RFC3339 is accepted by ClickHouse for DateTime in JSONEachRow
		return ts.UTC().Format(time.RFC3339), true
	case string:
		// Assume it's already in a ClickHouse-compatible format
		return ts, true
	case int64:
		// Treat as Unix seconds
		t := time.Unix(ts, 0).UTC()
		return t.Format(time.RFC3339), true
	case int:
		t := time.Unix(int64(ts), 0).UTC()
		return t.Format(time.RFC3339), true
	case float64:
		// Treat as Unix seconds, drop fractional part
		t := time.Unix(int64(ts), 0).UTC()
		return t.Format(time.RFC3339), true
	default:
		return "", false
	}
}
