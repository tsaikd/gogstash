package inputsocket

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"os"
	"time"

	"github.com/tsaikd/KDGoLib/errutil"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
	"golang.org/x/sync/errgroup"
)

// ModuleName is the name used in config file
const ModuleName = "socket"

// InputConfig holds the configuration json fields and internal objects
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

// errors
var (
	ErrorUnknownSocketType1 = errutil.NewFactory("%q is not a valid socket type")
	ErrorSocketAccept       = errutil.NewFactory("socket accept error")
)

// InitHandler initialize the input plugin
func InitHandler(ctx context.Context, raw *config.ConfigRaw) (config.TypeInputConfig, error) {
	conf := DefaultInputConfig()
	if err := config.ReflectConfig(raw, &conf); err != nil {
		return nil, err
	}
	return &conf, nil
}

// Start wraps the actual function starting the plugin
func (i *InputConfig) Start(ctx context.Context, msgChan chan<- logevent.LogEvent) error {
	eg, ctx := errgroup.WithContext(ctx)
	logger := goglog.Logger
	var l net.Listener

	switch i.Socket {
	case "unix", "unixpacket":
		// Remove existing unix socket
		os.Remove(i.Address)
		// Listen to socket
		address, err := net.ResolveUnixAddr(i.Socket, i.Address)
		if err != nil {
			return err
		}
		logger.Debugf("listen %q on %q", i.Socket, i.Address)
		l, err = net.ListenUnix(i.Socket, address)
		if err != nil {
			return err
		}
		defer l.Close()
		// Set socket permissions.
		if err = os.Chmod(i.Address, 0777); err != nil {
			return err
		}
	case "tcp":
		address, err := net.ResolveTCPAddr(i.Socket, i.Address)
		if err != nil {
			return err
		}
		logger.Debugf("listen %q on %q", i.Socket, address.String())
		l, err = net.ListenTCP(i.Socket, address)
		if err != nil {
			return err
		}
		defer l.Close()
	default:
		return ErrorUnknownSocketType1.New(nil, i.Socket)
	}

	eg.Go(func() error {
		select {
		case <-ctx.Done():
			return l.Close()
		}
	})

	eg.Go(func() error {
		for {
			conn, err := l.Accept()
			if err != nil {
				return ErrorSocketAccept.New(err)
			}
			func(conn net.Conn) {
				eg.Go(func() error {
					parse(ctx, conn, msgChan)
					return nil
				})
			}(conn)
		}
	})

	return eg.Wait()
}

func parse(ctx context.Context, conn net.Conn, msgChan chan<- logevent.LogEvent) {
	defer conn.Close()

	// Duplicate buffer to be able to read it even after failed json decoding
	var streamCopy bytes.Buffer
	stream := io.TeeReader(conn, &streamCopy)

	dec := json.NewDecoder(stream)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Assume first the message is JSON and try to decode it
		var jsonMsg map[string]interface{}
		if err := dec.Decode(&jsonMsg); err == io.EOF {
			break
		} else if err != nil {
			// If decoding fail, split raw message by line
			// and send a log event per line
			for {
				line, err := streamCopy.ReadString('\n')
				msgChan <- logevent.LogEvent{
					Timestamp: time.Now(),
					Message:   line,
				}
				if err != nil {
					break
				}
			}
			break
		}
		msgChan <- logevent.LogEvent{
			Timestamp: time.Now(),
			Extra:     jsonMsg,
		}
	}
}
