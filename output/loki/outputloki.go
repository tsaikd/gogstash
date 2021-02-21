package outputloki

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/tsaikd/KDGoLib/errutil"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
)

// ModuleName is the name used in config file
const ModuleName = "loki"

// errors
var (
	ErrNoValidURLs   = errutil.NewFactory("no valid URLs found")
	ErrEndpointDown1 = errutil.NewFactory("%q endpoint down")
)

// OutputConfig holds the configuration json fields and internal objects
type OutputConfig struct {
	config.OutputConfig
	URLs []string `json:"urls"` // Array of HTTP connection strings
	Auth string   `json:"auth"`

	httpClient *http.Client
}

// func buildTags(stream []interface{}, labels map[string]interface{}) []interface{}{
// 	stream.append(labels)
// }

type LokiValue [2]string

type LokiStream struct {
	Stream map[string]interface{} `json:"stream"`
	Values []LokiValue            `json:"values"`
}

type LokiRequest struct {
	Streams []LokiStream `json:"streams"`
}

// ToStringE casts an empty interface to a string.
func ToStringE(i interface{}) (string, error) {
	switch s := i.(type) {
	case string:
		return s, nil
	case bool:
		return strconv.FormatBool(s), nil
	case float64:
		return strconv.FormatFloat(i.(float64), 'f', -1, 64), nil
	case int64:
		return strconv.FormatInt(i.(int64), 10), nil
	case int:
		return strconv.FormatInt(int64(i.(int)), 10), nil
	case []byte:
		return string(s), nil
	case nil:
		return "null", nil
	case error:
		return s.Error(), nil
	default:
		return "", errors.New("Unable to Cast to string")
	}
}

func buildLokiRequest(event logevent.LogEvent) ([]byte, error) {
	message := ""
	if event.Message != "" {
		message = event.Message
	}

	_time := time.Now().UnixNano()
	ts := strconv.FormatInt(_time, 10)
	value := LokiValue{ts, message}
	stream := LokiStream{}
	stream.Values = []LokiValue{value}

	_stream := make(map[string]interface{})
	for key, value := range event.Extra {
		v, err := ToStringE(value)
		if err != nil {
			goglog.Logger.Warn("key: %s error:%v", key, err)
		}
		_stream[key] = v
	}

	if len(event.Tags) > 0 {
		_stream["tag"] = event.Tags
	}
	stream.Stream = _stream

	request := LokiRequest{[]LokiStream{stream}}
	return jsoniter.Marshal(request)
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

	raw, err := buildLokiRequest(event)

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

	if t.Auth != "" {
		req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(t.Auth)))
	}

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if !(resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent) {
		err = ErrEndpointDown1.New(nil, url)
		goglog.Logger.Errorf("output http: %v", err)
		return err
	}

	return nil
}
