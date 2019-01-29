package outputhttp

import (
	"bytes"
	"context"
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
	URLs []string `json:"urls"` // Array of HTTP connection strings

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
	}
}

// InitHandler initialize the output plugin
func InitHandler(ctx context.Context, raw *config.ConfigRaw) (config.TypeOutputConfig, error) {
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
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
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
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(raw))
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
	if resp.StatusCode != http.StatusOK {
		err = ErrEndpointDown1.New(nil, url)
		goglog.Logger.Errorf("output http: %v", err)
		return err
	}

	return nil
}
