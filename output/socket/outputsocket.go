package outputsocket

import (
	"context"
	"net"

	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
)

// ModuleName is the name used in config file
const ModuleName = "socket"

// OutputConfig holds the configuration json fields and internal objects
type OutputConfig struct {
	config.OutputConfig
	Socket       string `json:"socket"`  // Type of socket, must be one of ["tcp", "unix", "unixpacket"].
	Address      string `json:"address"` // For TCP, address must have the form `host:port`. For Unix networks, the address must be a file system path.
	outputSocket *net.Conn
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

	// init Socket
	conn, err := net.Dial(conf.Socket, conf.Address)
	if err != nil {
		return nil, err
	}
	go func() {
		select {
		case <-ctx.Done():
			conn.Close()
		}
	}()

	conf.outputSocket = &conn

	return &conf, nil
}

// Output event
func (t *OutputConfig) Output(ctx context.Context, event logevent.LogEvent) error {
	b, err := event.MarshalJSON()
	if err != nil {
		return err
	}
	b = append(b, '\n')
	if _, err := (*t.outputSocket).Write(b); err != nil {
		return err
	}
	return nil
}
