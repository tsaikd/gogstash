package inputhttplisten

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	codecjson "github.com/tsaikd/gogstash/codec/json"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
)

// ModuleName is the name used in config file
const ModuleName = "httplisten"

const invalidMethodError = "Method not allowed: '%v'"
const invalidRequestError = "Invalid request received on HTTP listener. Decoder error: %+v"
const invalidAccessToken = "Invalid access token. Access denied."

// InputConfig holds the configuration json fields and internal objects
type InputConfig struct {
	config.InputConfig
	Address       string   `json:"address"`        // host:port to listen on
	Path          string   `json:"path"`           // The path to accept json HTTP POST requests on
	RequireHeader []string `json:"require_header"` // Require this header to be present to accept the POST ("X-Access-Token: Potato")
}

// DefaultInputConfig returns an InputConfig struct with default values
func DefaultInputConfig() InputConfig {
	return InputConfig{
		InputConfig: config.InputConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		Address:       "0.0.0.0:8080",
		Path:          "/",
		RequireHeader: []string{},
	}
}

// InitHandler initialize the input plugin
func InitHandler(ctx context.Context, raw *config.ConfigRaw) (config.TypeInputConfig, error) {
	conf := DefaultInputConfig()
	err := config.ReflectConfig(raw, &conf)
	if err != nil {
		return nil, err
	}

	conf.Codec, err = config.GetCodecDefault(ctx, *raw, codecjson.ModuleName)
	if err != nil {
		return nil, err
	}

	return &conf, nil
}

// Start wraps the actual function starting the plugin
func (i *InputConfig) Start(ctx context.Context, msgChan chan<- logevent.LogEvent) (err error) {
	logger := goglog.Logger
	http.HandleFunc(i.Path, func(rw http.ResponseWriter, req *http.Request) {
		// Only allow POST requests (for now).
		if req.Method != http.MethodPost {
			logger.Warnf(invalidMethodError, req.Method)
			rw.WriteHeader(http.StatusMethodNotAllowed)
			rw.Write([]byte(fmt.Sprintf(invalidMethodError, req.Method)))
			return
		}
		// Check for header
		if len(i.RequireHeader) == 2 {
			// get returns empty string if header not found
			if req.Header.Get(i.RequireHeader[0]) != i.RequireHeader[1] {
				logger.Warn(invalidAccessToken)
				rw.WriteHeader(http.StatusForbidden)
				rw.Write([]byte(invalidAccessToken))
				return
			}
		}
		i.postHandler(msgChan, rw, req)
	})
	go func() {
		logger.Infof("accepting POST requests to %s%s", i.Address, i.Path)
		if err = http.ListenAndServe(i.Address, nil); err != nil {
			logger.Fatal(err)
		}
	}()
	return nil
}

// Handle HTTP POST requests
func (i *InputConfig) postHandler(msgChan chan<- logevent.LogEvent, rw http.ResponseWriter, req *http.Request) {
	logger := goglog.Logger
	logger.Debugf("Received request")

	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		logger.Errorf("read request body error: %v", err)
		return
	}

	ok, err := i.Codec.Decode(context.TODO(), data, nil, msgChan)
	if err != nil {
		logger.Errorf("decode request body error: %v", err)
	}
	if !ok {
		// event not sent to msgChan
		rw.WriteHeader(http.StatusInternalServerError)
		if err != nil {
			rw.Write([]byte(err.Error()))
		}
	} else if err != nil {
		// event sent to msgChan
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte(fmt.Sprintf(invalidRequestError, err)))
	}
}
