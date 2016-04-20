package inputsocket

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net"
	"os"
	"time"

	"github.com/Sirupsen/logrus"

	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
)

// ModuleName is the name used in config file
const ModuleName = "socket"

// InputConfig holds the output configuration json fields
type InputConfig struct {
	config.InputConfig
	Socket  string `json:"socket"`  // Type of socket, must be one of ["tcp", "unix", "unixpacket"].
	Address string `json:"address"` // For TCP, address must have the form `host:port`. For Unix networks, the address must be a file system path.
}

// DefaultInputConfig returns an InputConfig struct with default values
func DefaultInputConfig() InputConfig {
	return InputConfig{
		InputConfig: config.InputConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
	}
}

// InitHandler initialize the input plugin
func InitHandler(confraw *config.ConfigRaw) (config.TypeInputConfig, error) {
	conf := DefaultInputConfig()
	if err := config.ReflectConfig(confraw, &conf); err != nil {
		return nil, err
	}
	return &conf, nil
}

// Start wraps the actual function starting the plugin
func (i *InputConfig) Start() {
	i.Invoke(i.start)
}

func (i *InputConfig) start(logger *logrus.Logger, evchan chan logevent.LogEvent) {

	switch i.Socket {
	case "unix", "unixpacket":
		// Remove existing unix socket
		os.Remove(i.Address)
	}

	l, err := net.Listen(i.Socket, i.Address)
	if err != nil {
		logger.Error(ModuleName, ": Unable to listen to socket.", err)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			logger.Error(ModuleName, ": socket accept error.", err)
		}
		go parse(conn, logger, evchan)
	}
}

func parse(conn net.Conn, logger *logrus.Logger, evchan chan logevent.LogEvent) {
	defer conn.Close()

	// Duplicate buffer to be able to read it even after failed json decoding
	buf, _ := ioutil.ReadAll(conn)
	js := bytes.NewBuffer(buf)
	raw := bytes.NewBuffer(buf)

	dec := json.NewDecoder(js)
	for {
		// Assume first the message is JSON and try to decode it
		var jsonMsg map[string]interface{}
		if err := dec.Decode(&jsonMsg); err == io.EOF {
			break
		} else if err != nil {
			// if decoding fail, use raw message in LogEvent "Message" field
			evchan <- logevent.LogEvent{
				Timestamp: time.Now(),
				Message:   raw.String(),
			}
			break
		}
		evchan <- logevent.LogEvent{Extra: jsonMsg}
	}
}
