package outputhttp

import (
	"bytes"
	"container/list"
	"context"
	"crypto/tls"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/tsaikd/KDGoLib/errutil"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
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

const (
	statusDelivering = iota // if we are in running mode - delivering messages
	statusPaused            // if we have paused the inputs
)

// OutputConfig holds the configuration json fields and internal objects
type OutputConfig struct {
	config.OutputConfig
	URLs                []string `json:"urls"` // Array of HTTP connection strings
	AcceptedHttpResult  []int    `json:"http_status_codes" yaml:"http_status_codes"`
	PermanentHttpErrors []int    `json:"http_error_codes" yaml:"http_error_codes"` // HTTP codes that will not retry an event
	RetryInterval       uint     `json:"retry_interval" yaml:"retry_interval"`     // seconds before a new retry in case on error
	IgnoreSSL           bool     `json:"ignore_ssl" yaml:"ignore_ssl"`
	MaxQueueSize        int      `json:"max_queue_size" yaml:"max_queue_size"` // max size of queue before deleting events (-1=no limit, 0=disable)

	httpClient *http.Client
	control    config.Control
	isInPause  uint32 // set to either statusDelivering or statusPaused

	queue      chan logevent.LogEvent // channel to send events to internal queue
	retryqueue list.List              // list of queued messages; not multithread safe, only accessed from backgroundtask()
}

// DefaultOutputConfig returns an OutputConfig struct with default values
func DefaultOutputConfig() OutputConfig {
	return OutputConfig{
		OutputConfig: config.OutputConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		AcceptedHttpResult:  []int{http.StatusOK, http.StatusCreated, http.StatusAccepted},
		PermanentHttpErrors: []int{http.StatusNotImplemented, http.StatusMethodNotAllowed, http.StatusNotFound, http.StatusAlreadyReported, http.StatusHTTPVersionNotSupported},
		RetryInterval:       30,
		isInPause:           statusDelivering,
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

	if len(conf.URLs) <= 0 {
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
	conf.control = control
	conf.queue = make(chan logevent.LogEvent, 5)
	if conf.MaxQueueSize == 0 || conf.MaxQueueSize < -1 {
		conf.MaxQueueSize = 1
	} // we need at least one event in the queue to allow for resume
	go conf.backgroundtask()

	return &conf, nil
}

// Output is receiving the event from gogstash handler
func (t *OutputConfig) Output(ctx context.Context, event logevent.LogEvent) (err error) {
	// see if output has requested pause and if so just queue the event instead of trying
	if atomic.LoadUint32(&t.isInPause) == statusPaused {
		t.queue <- event
		return nil
	}
	// If we are not in pause mode then call the sender method
	return t.OutputEvent(ctx, event)
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
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "gogstash/output"+ModuleName)

	resp, err := t.httpClient.Do(req)
	if err != nil {
		t.failedDelivery(ctx, event)
		return err
	}
	defer resp.Body.Close()

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.failedDelivery(ctx, event)
		return err
	}
	if t.checkIntInDiscardedList(resp.StatusCode) {
		return ErrPermanentError.New(nil, url, resp.StatusCode)
	}
	if !t.checkIntInAcceptedList(resp.StatusCode) {
		t.failedDelivery(ctx, event)
		return ErrSoftError.New(nil, url, resp.StatusCode)
	}
	// the event was sent correctly, we now have to resume inputs if we have requested a pause.
	if atomic.CompareAndSwapUint32(&t.isInPause, statusPaused, statusDelivering) {
		goglog.Logger.Debug("outputhttp requesting resume")
		err := t.control.RequestResume(ctx)
		if err != nil {
			goglog.Logger.Error("outputhttp", err.Error())
		}
	}
	return nil
}

// checkIntInAcceptedList checks if code is in configured list of accepted status codes
func (t *OutputConfig) checkIntInAcceptedList(code int) bool {
	for _, v := range t.AcceptedHttpResult {
		if v == code {
			return true
		}
	}
	return false
}

// checkIntInAcceptedList checks if code is in configured list of status codes where we should discard the message
func (t *OutputConfig) checkIntInDiscardedList(code int) bool {
	for _, v := range t.PermanentHttpErrors {
		if v == code {
			return true
		}
	}
	return false
}

// failedDelivery receives an event that failed delivery and triggers a pause event if we have not done so already
// and places the message in the retry-queue.
func (t *OutputConfig) failedDelivery(ctx context.Context, event logevent.LogEvent) {
	// see if we need to send a pause signal
	if atomic.CompareAndSwapUint32(&t.isInPause, statusDelivering, statusPaused) {
		goglog.Logger.Debug("outputhttp requesting pause")
		err := t.control.RequestPause(ctx)
		if err != nil {
			goglog.Logger.Error("outputhttp ", err.Error())
		}
	}
	// queue the event for later retry
	t.queue <- event
}

// backgroundtask is running in the background and adds new events to the queue, tries to send them out on the RetryInterval.
func (t *OutputConfig) backgroundtask() {
	dur := time.Duration(t.RetryInterval) * time.Second
	ticker := time.NewTicker(dur)
	defer ticker.Stop()
	for {
		select {
		case event := <-t.queue:
			if (t.MaxQueueSize > 0 && t.retryqueue.Len() < t.MaxQueueSize) || t.MaxQueueSize == -1 {
				t.retryqueue.PushBack(event)
			}
		case <-ticker.C:
			// We have reached a RetryInterval. If there are any events in the queue, lets send one back.
			// If we are still in pause mode we will send one, if we are in normal mode we will send all back.
			if e := t.retryqueue.Front(); e != nil {
				if atomic.LoadUint32(&t.isInPause) == statusPaused {
					event := e.Value.(logevent.LogEvent)
					t.retryqueue.Remove(e)
					go func() {
						ctx, cancel := context.WithTimeout(context.Background(), dur)
						err := t.OutputEvent(ctx, event)
						if err != nil {
							goglog.Logger.Error("outputhttp background sendone ", err.Error())
						}
						cancel()
					}()
				} else {
					// we are not in pause mode and will queue all the events in the queue for sending.
					// First we need to empty the queue and get all events to send.
					myList := []*logevent.LogEvent{}
					for {
						e := t.retryqueue.Front()
						if e == nil {
							break
						}
						event := e.Value.(logevent.LogEvent)
						myList = append(myList, &event)
						t.retryqueue.Remove(e)
					}
					// Now we have to send them all out
					go func() {
						for x := range myList {
							ctx, cancel := context.WithTimeout(context.Background(), dur)
							err := t.Output(ctx, *myList[x])
							if err != nil {
								goglog.Logger.Error("outputhttp background sendall", err.Error())
							}
							cancel()
						}
					}()
				}
			}
		}
	}
}
