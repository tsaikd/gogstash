package inputsocket

import (
	"bufio"
	"context"
	"io"
	"net"
	"os"
	"time"

	jsoniter "github.com/json-iterator/go"
	reuse "github.com/libp2p/go-reuseport"
	"github.com/tsaikd/KDGoLib/errutil"
	codecjson "github.com/tsaikd/gogstash/codec/json"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
	"golang.org/x/sync/errgroup"
)

// ModuleName is the name used in config file
const ModuleName = "socket"

// ErrorTag tag added to event when process module failed
const ErrorTag = "gogstash_input_socket_error"

// InputConfig holds the configuration json fields and internal objects
type InputConfig struct {
	config.InputConfig
	Socket string `json:"socket"` // Type of socket, must be one of ["tcp", "udp", "unix", "unixpacket"].
	// For TCP or UDP, address must have the form `host:port`.
	// For Unix networks, the address must be a file system path.
	Address   string `json:"address"`
	ReusePort bool   `json:"reuseport"`
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
func (i *InputConfig) Start(ctx context.Context, msgChan chan<- logevent.LogEvent) error {
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
		if i.ReusePort {
			l, err = reuse.Listen(i.Socket, address.String())
		} else {
			l, err = net.ListenTCP(i.Socket, address)
		}
		if err != nil {
			return err
		}
		defer l.Close()
	case "udp":
		address, err := net.ResolveUDPAddr(i.Socket, i.Address)
		logger.Debugf("listen %q on %q", i.Socket, address.String())
		var conn net.PacketConn
		if i.ReusePort {
			conn, err = reuse.ListenPacket(i.Socket, i.Address)
		} else {
			conn, err = net.ListenPacket(i.Socket, i.Address)
		}
		if err != nil {
			return err
		}
		return handleUDP(ctx, conn, msgChan)
	default:
		return ErrorUnknownSocketType1.New(nil, i.Socket)
	}

	eg, ctx := errgroup.WithContext(ctx)

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
					defer conn.Close()
					parse(ctx, conn, msgChan)
					return nil
				})
			}(conn)
		}
	})

	return eg.Wait()
}

func handleUDP(ctx context.Context, conn net.PacketConn, msgChan chan<- logevent.LogEvent) error {
	eg, ctx := errgroup.WithContext(ctx)
	b := make([]byte, 1500) // read buf
	pr, pw := io.Pipe()
	defer pw.Close()

	eg.Go(func() error {
		select {
		case <-ctx.Done():
			pr.Close()
			conn.Close()
			return nil
		}
	})

	eg.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return nil
			default:
			}
			n, _, err := conn.ReadFrom(b)
			if err == io.EOF {
				break
			} else if err != nil {
				return err
			}
			pw.Write(b[:n])
		}
		return nil
	})

	eg.Go(func() error {
		parse(ctx, pr, msgChan)
		return nil
	})

	return eg.Wait()
}

func parse(ctx context.Context, r io.Reader, msgChan chan<- logevent.LogEvent) {
	b := bufio.NewReader(r)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		line, err := b.ReadBytes('\n')
		if err != nil {
			// EOF
			return
		}

		event := logevent.LogEvent{
			Timestamp: time.Now(),
			Message:   string(line),
			Extra:     map[string]interface{}{},
		}

		if err := jsoniter.Unmarshal([]byte(event.Message), &event.Extra); err != nil {
			event.AddTag(ErrorTag)
			goglog.Logger.Error(err)
		}

		// try to fill basic log event by json message
		if value, ok := event.Extra["message"]; ok {
			switch v := value.(type) {
			case string:
				event.Message = v
			}
		}
		if value, ok := event.Extra["@timestamp"]; ok {
			switch v := value.(type) {
			case string:
				if timestamp, err := time.Parse(time.RFC3339Nano, v); err == nil {
					event.Timestamp = timestamp
				}
			}
		}

		msgChan <- event
	}
}
