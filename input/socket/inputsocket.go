package inputsocket

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"os"
	"time"

	"github.com/Sirupsen/logrus"

	"github.com/tsaikd/KDGoLib/errutil"
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
	var l net.Listener

	switch i.Socket {
	case "unix", "unixpacket":
		// Remove existing unix socket
		os.Remove(i.Address)
		// Listen to socket
		address, err := net.ResolveUnixAddr(i.Socket, i.Address)
		if err != nil {
			logger.Fatal(err)
		}
		l, err = net.ListenUnix(i.Socket, address)
		if err != nil {
			logger.Fatal(err)
		}
		// Set socket permissions.
		if err = os.Chmod(i.Address, 0777); err != nil {
			logger.Fatal(err)
		}

	case "tcp":
		address, err := net.ResolveTCPAddr(i.Socket, i.Address)
		if err != nil {
			logger.Fatal(err)
		}
		l, err = net.ListenTCP(i.Socket, address)
		if err != nil {
			logger.Fatal(err)
		}

	default:
		logger.Fatal(errutil.NewFactory(i.Socket + " is not a valid socket type."))
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
	var streamCopy bytes.Buffer
	stream := io.TeeReader(conn, &streamCopy)

	dec := json.NewDecoder(stream)
	for {
		// Assume first the message is JSON and try to decode it
		var jsonMsg map[string]interface{}
		if err := dec.Decode(&jsonMsg); err == io.EOF {
			break
		} else if err != nil {
			// If decoding fail, split raw message by line
			// and send a log event per line
			for {
				line, err := streamCopy.ReadString('\n')
				evchan <- logevent.LogEvent{
					Timestamp: time.Now(),
					Message:   line,
				}
				if err != nil {
					break
				}
			}
			break
		}
		evchan <- logevent.LogEvent{Extra: jsonMsg}
	}
}
