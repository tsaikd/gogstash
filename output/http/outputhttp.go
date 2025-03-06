package outputhttp

import (
	"bytes"
	"context"
	"crypto/tls"
	"io"
	"math/rand"
	"net"
	"net/http"
	"time"

	"github.com/tsaikd/KDGoLib/errutil"

	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
	"github.com/tsaikd/gogstash/config/queue"
)

// ModuleName is the name used in config file
const ModuleName = "http"

// errors
var (
	ErrNoValidURLs    = errutil.NewFactory("no valid URLs found")
	ErrEndpointDown1  = errutil.NewFactory("%q endpoint down")
	ErrPermanentError = errutil.NewFactory("%q permanent error %v (discarding event)")
	ErrSoftError      = errutil.NewFactory("%q retryable error %v")
)

// OutputConfig holds the configuration json fields and internal objects
type OutputConfig struct {
	config.OutputConfig
	URLs                []string          `json:"urls" yaml:"urls"`                           // Array of HTTP connection strings
	AcceptedHttpResult  []int             `json:"http_status_codes" yaml:"http_status_codes"` // HTTP codes that indicate success
	PermanentHttpErrors []int             `json:"http_error_codes" yaml:"http_error_codes"`   // HTTP codes that will not retry an event
	RetryInterval       uint              `json:"retry_interval" yaml:"retry_interval"`       // seconds before a new retry in case on error
	IgnoreSSL           bool              `json:"ignore_ssl" yaml:"ignore_ssl"`               //
	ContentType         string            `json:"content_type" yaml:"content_type"`           // HTTP content type
	Format              string            `json:"format" yaml:"format"`                       // Format of data (defaults to json)
	Headers             map[string]string `json:"headers" yaml:"headers"`                     // Map of additional headers
	MaxQueueSize        int               `json:"max_queue_size" yaml:"max_queue_size"`       // max size of queue before deleting events (-1=no limit, 0=disable)

	acceptedHttpResult  map[int]struct{} // a map containing the accepted result codes
	permanentHttpErrors map[int]struct{} // a map containing the permanent error codes

	httpClient *http.Client
	queue      queue.Queue // our queue
}

// DefaultOutputConfig returns an OutputConfig struct with default values
func DefaultOutputConfig() OutputConfig {
	return OutputConfig{
		OutputConfig: config.OutputConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		RetryInterval:       30,
		permanentHttpErrors: MapFromInts([]int{http.StatusNotImplemented, http.StatusMethodNotAllowed, http.StatusNotFound, http.StatusAlreadyReported, http.StatusHTTPVersionNotSupported}),
		acceptedHttpResult:  MapFromInts([]int{http.StatusOK, http.StatusCreated, http.StatusAccepted}),
	}
}

// InitHandler initialize the output plugin
func InitHandler(
	ctx context.Context,
	raw config.ConfigRaw,
	control config.Control,
) (config.TypeOutputConfig, error) {
	conf := DefaultOutputConfig()
	if err := config.ReflectConfig(raw, &conf); err != nil {
		return nil, err
	}

	if len(conf.AcceptedHttpResult) > 0 {
		conf.acceptedHttpResult = MapFromInts(conf.AcceptedHttpResult)
	}
	if len(conf.PermanentHttpErrors) > 0 {
		conf.permanentHttpErrors = MapFromInts(conf.PermanentHttpErrors)
	}

	if len(conf.URLs) == 0 {
		return nil, ErrNoValidURLs
	}
	conf.httpClient = &http.Client{Transport: &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: conf.IgnoreSSL},
	}}
	// we need at least one event in the queue to allow for resume
	if conf.MaxQueueSize == 0 || conf.MaxQueueSize < -1 {
		conf.MaxQueueSize = 1
	}
	// create the queue
	conf.queue = queue.NewSimpleQueue(ctx, control, &conf, nil, conf.MaxQueueSize, conf.RetryInterval)

	return conf.queue, nil
}

// OutputEvent tries to send a message, requeueing if is has a temporary error
func (t *OutputConfig) OutputEvent(ctx context.Context, event logevent.LogEvent) (err error) {
	i := rand.Intn(len(t.URLs))

	raw, err := event.MarshalJSON()
	if err != nil {
		return err
	}

	url := t.URLs[i]
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	if t.ContentType != "" {
		req.Header.Set("Content-Type", t.ContentType)
	} else {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("User-Agent", "gogstash/output"+ModuleName)

	// Add custom headers, if they exist
	for k, v := range t.Headers {
		req.Header.Set(k, v)
	}

	resp, err := t.httpClient.Do(req)
	if err != nil {
		t.failedDelivery(ctx, event)
		return err
	}
	defer resp.Body.Close()

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		t.failedDelivery(ctx, event)
		return err
	}
	if _, isinlist := t.permanentHttpErrors[resp.StatusCode]; isinlist {
		return ErrPermanentError.New(nil, url, resp.StatusCode)
	}
	if _, ok := t.acceptedHttpResult[resp.StatusCode]; !ok {
		t.failedDelivery(ctx, event)
		return ErrSoftError.New(nil, url, resp.StatusCode)
	}
	// the event was sent correctly, we now have to resume inputs if we earlier has requested a pause.
	return t.queue.Resume(ctx)
}

// failedDelivery receives an event that failed delivery and triggers a pause event if we have not done so already
// and places the message in the retry-queue.
func (t *OutputConfig) failedDelivery(ctx context.Context, event logevent.LogEvent) {
	err := t.queue.Queue(ctx, event)
	if err != nil {
		goglog.Logger.Error("outputhttp ", err.Error())
	}
}

// MapFromInts returns a map containing all the ints in the array
func MapFromInts(nums []int) map[int]struct{} {
	result := make(map[int]struct{}, len(nums))
	for _, x := range nums {
		result[x] = struct{}{}
	}
	return result
}
