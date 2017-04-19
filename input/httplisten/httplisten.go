package inputhttplisten

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
)

// ModuleName is the name used in config file
const ModuleName = "httplisten"

const invalidMethodError = "Method not allowed: '%v'"
const invalidJSONError = "Invalid JSON received on HTTP listener. Decoder error: %+v"
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
	return &conf, nil
}

// Start wraps the actual function starting the plugin
func (i *InputConfig) Start(ctx context.Context, msgChan chan<- logevent.LogEvent) (err error) {
	logger := config.Logger
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
		i.postHandler(logger, msgChan, rw, req)
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
func (i *InputConfig) postHandler(logger *logrus.Logger, msgChan chan<- logevent.LogEvent, rw http.ResponseWriter, req *http.Request) {
	logger.Debugf("Received request")

	var jsonMsg map[string]interface{}
	dec := json.NewDecoder(req.Body)

	// attempt to decode post body, if it fails, log it.
	if err := dec.Decode(&jsonMsg); err != nil {
		logger.Warnf(invalidJSONError, err)
		logger.Debugf("Invalid JSON: '%s'", req.Body)
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte(fmt.Sprintf(invalidJSONError, err)))
		return
	}

	// send the event as it came to us
	msgChan <- logevent.LogEvent{Extra: jsonMsg}
}
