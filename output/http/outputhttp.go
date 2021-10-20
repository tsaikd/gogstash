package outputhttp

import (
	"bytes"
	"context"
	"crypto/tls"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
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
	ErrNoValidURLs   = errutil.NewFactory("no valid URLs found")
	ErrEndpointDown1 = errutil.NewFactory("%q endpoint down")
)

// OutputConfig holds the configuration json fields and internal objects
type OutputConfig struct {
	config.OutputConfig
	URLs               []string `json:"urls"` // Array of HTTP connection strings
	AcceptedHttpResult []int    `json:"http_status_codes" yaml:"http_status_codes"`
	IgnoreSSL          bool     `json:"ignore_ssl" yaml:"ignore_ssl"`

	httpClient *http.Client
}

// DefaultOutputConfig returns an OutputConfig struct with default values
func DefaultOutputConfig() OutputConfig {
	return OutputConfig{
		OutputConfig: config.OutputConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		AcceptedHttpResult: []int{200, 201, 202},
	}
}

// InitHandler initialize the output plugin
func InitHandler(ctx context.Context, raw config.ConfigRaw) (config.TypeOutputConfig, error) {
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

	return &conf, nil
}

// Output event
func (t *OutputConfig) Output(ctx context.Context, event logevent.LogEvent) (err error) {
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
		return err
	}
	defer resp.Body.Close()

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if !t.checkIntInList(resp.StatusCode) {
		err = ErrEndpointDown1.New(nil, url)
		goglog.Logger.Errorf("output http: %v, status=%v", err, resp.StatusCode)
		return err
	}

	return nil
}

// checkIntInList checks if code is in configured list of accepted status codes
func (t *OutputConfig) checkIntInList(code int) bool {
	for _, v := range t.AcceptedHttpResult {
		if v == code {
			return true
		}
	}
	return false
}
