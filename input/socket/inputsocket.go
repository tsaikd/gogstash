package inputsocket

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"os"

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
	Address    string `json:"address"`
	ReusePort  bool   `json:"reuseport"`
	BufferSize int    `json:"buffer_size" yaml:"buffer_size"`
	// packetmode is only valid for UDP sessions and handles each packet as a message on its own
	PacketMode bool `json:"packetmode"`
}

// DefaultInputConfig returns an InputConfig struct with default values
func DefaultInputConfig() InputConfig {
	return InputConfig{
		InputConfig: config.InputConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		BufferSize: 4096,
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
		if err != nil {
			return err
		}
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
		if i.PacketMode {
			return i.handleUDPpacketMode(ctx, conn, msgChan)
		} else {
			return i.handleUDP(ctx, conn, msgChan)
		}
	default:
		return ErrorUnknownSocketType1.New(nil, i.Socket)
	}

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		<-ctx.Done()
		return l.Close()
	})

	eg.Go(func() error {
		for {
			conn, err := l.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					return nil
				}
				return ErrorSocketAccept.New(err)
			}
			doneCh := make(chan struct{})
			eg.Go(func() error {
				select {
				case <-doneCh:
				case <-ctx.Done():
					conn.Close()
				}
				return nil
			})
			func(conn net.Conn) {
				eg.Go(func() error {
					defer conn.Close()
					i.parse(ctx, conn, msgChan)
					close(doneCh)
					return nil
				})
			}(conn)
		}
	})

	return eg.Wait()
}

// handleUDPpacketMode receives UDP packets and sends it out as messages one packet at a time
func (i *InputConfig) handleUDPpacketMode(ctx context.Context, conn net.PacketConn, msgChan chan<- logevent.LogEvent) error {
	eg, ctx := errgroup.WithContext(ctx)
	b := make([]byte, i.BufferSize) // make read buffer
	logger := goglog.Logger

	// code to handle cancellation as ReadFrom is blocking and could take some time to time out if there is no traffic
	eg.Go(func() error {
		<-ctx.Done()
		conn.Close()
		return nil
	})

	// code to process input packets
	eg.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return nil
			default:
			}
			n, addr, err := conn.ReadFrom(b)
			// handle data
			if n > 0 {
				if n >= i.BufferSize {
					const bufIncSize = 600
					logger.Errorf("%v UDP receive buffers (%v bytes) too small, should be increased!", i.Address, i.BufferSize)
					i.BufferSize += bufIncSize
					b = make([]byte, i.BufferSize)
				} else {
					extras := map[string]interface{}{
						"host_ip": addr.String(),
					}
					_, codecErr := i.Codec.Decode(ctx, b[:n], extras, []string{}, msgChan)
					if codecErr != nil {
						logger.Errorf("Input socket %v: %v", i.Address, codecErr)
					}
				}
			}
			// handle error
			switch err {
			case nil:
			// continue processing
			case io.EOF:
				return nil
			default:
				if errors.Is(err, net.ErrClosed) {
					return nil // do not return error on closed socket
				}
				return err
			}
		}
	})

	return eg.Wait()
}

func (i *InputConfig) handleUDP(ctx context.Context, conn net.PacketConn, msgChan chan<- logevent.LogEvent) error {
	eg, ctx := errgroup.WithContext(ctx)
	b := make([]byte, i.BufferSize) // read buf
	pr, pw := io.Pipe()
	defer pw.Close()

	eg.Go(func() error {
		<-ctx.Done()
		pr.Close()
		conn.Close()
		return nil
	})

	eg.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return nil
			default:
			}
			n, _, err := conn.ReadFrom(b)
			if n > 0 {
				pw.Write(b[:n])
			}
			if errors.Is(err, net.ErrClosed) || err == io.EOF {
				break
			} else if err != nil {
				return err
			}
		}
		return nil
	})

	eg.Go(func() error {
		i.parse(ctx, pr, msgChan)
		return nil
	})

	return eg.Wait()
}

func (i *InputConfig) parse(ctx context.Context, r io.Reader, msgChan chan<- logevent.LogEvent) {
	b := bufio.NewReader(r)
	logger := goglog.Logger
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

		_, codecErr := i.Codec.Decode(ctx, line, nil, []string{}, msgChan)
		if codecErr != nil {
			logger.Errorf("Input socket %v: %v", i.Address, codecErr)
		}
	}
}
