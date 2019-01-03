package config

import (
	"context"
	"time"

	"github.com/icza/dyno"
	"github.com/tsaikd/KDGoLib/errutil"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
)

// errors
var (
	ErrorUnknownCodecType1 = errutil.NewFactory("unknown codec config type: %q")
	ErrorInitCodecFailed1  = errutil.NewFactory("initialize codec module failed: %v")
	ErrorNotImplement1     = errutil.NewFactory("%q is not implement")
)

// TypeCodecConfig is interface of codec module
type TypeCodecConfig interface {
	TypeCommonConfig
	// The codec’s decode method is where data coming in from an input is transformed into an event.
	// if event sent to msgChan, ok will be true
	Decode(ctx context.Context, data []byte, extra map[string]interface{}, msgChan chan<- logevent.LogEvent) (ok bool, err error)
	// The encode method takes an event and serializes it (encodes) into another format.
	Encode(ctx context.Context, event logevent.LogEvent, dataChan chan<- []byte) (ok bool, err error)
}

// CodecConfig is basic codec config struct
type CodecConfig struct {
	CommonConfig
}

// CodecHandler is a handler to regist codec module
type CodecHandler func(ctx context.Context, raw *ConfigRaw) (TypeCodecConfig, error)

var (
	mapCodecHandler = map[string]CodecHandler{}
)

// RegistCodecHandler regist a codec handler
func RegistCodecHandler(name string, handler CodecHandler) {
	mapCodecHandler[name] = handler
}

// GetCodec returns a codec based on the 'codec' configuration from provided 'ConfigRaw' input
func GetCodec(ctx context.Context, raw ConfigRaw) (TypeCodecConfig, error) {
	return GetCodecDefault(ctx, raw, DefaultCodecName)
}

// GetCodecDefault returns a codec based on the 'codec' configuration from provided 'ConfigRaw' input
// defaults to 'defaultType'
func GetCodecDefault(ctx context.Context, raw ConfigRaw, defaultType string) (TypeCodecConfig, error) {
	codecConfig, err := dyno.Get(map[string]interface{}(raw), "codec")
	if err != nil {
		// return default codec here
		return getCodec(ctx, ConfigRaw{"type": defaultType})
	}

	if codecConfig == nil {
		return nil, nil
	}

	switch codecConfig.(type) {
	case map[string]interface{}:
		return getCodec(ctx, ConfigRaw(codecConfig.(map[string]interface{})))
	case string:
		// shorthand codec config method:
		// codec: [codecTypeName]
		return getCodec(ctx, ConfigRaw{"type": codecConfig.(string)})
	default:
		return nil, ErrorUnknownCodecType1.New(nil, codecConfig)
	}
}

func getCodec(ctx context.Context, raw ConfigRaw) (codec TypeCodecConfig, err error) {
	handler, ok := mapCodecHandler[raw["type"].(string)]
	if !ok {
		return nil, ErrorUnknownCodecType1.New(nil, raw["type"])
	}

	if codec, err = handler(ctx, &raw); err != nil {
		return nil, ErrorInitCodecFailed1.New(err, raw)
	}

	return codec, nil
}

// DefaultCodecName default codec name
const DefaultCodecName = "default"

// DefaultCodec default struct for codec
type DefaultCodec struct {
	CodecConfig
}

// DefaultCodecInitHandler returns an TypeCodecConfig interface with default handler
func DefaultCodecInitHandler(context.Context, *ConfigRaw) (TypeCodecConfig, error) {
	return &DefaultCodec{
		CodecConfig: CodecConfig{
			CommonConfig: CommonConfig{
				Type: DefaultCodecName,
			},
		},
	}, nil
}

// Decode returns an event based on current timestamp and converting 'data' to 'string', adding provided 'eventExtra'
func (c *DefaultCodec) Decode(ctx context.Context, data []byte,
	eventExtra map[string]interface{},
	msgChan chan<- logevent.LogEvent) (ok bool, err error) {
	event := logevent.LogEvent{
		Timestamp: time.Now(),
		Message:   string(data),
		Extra:     eventExtra,
	}

	goglog.Logger.Debugf("%q %v", event.Message, event)
	msgChan <- event

	return true, nil
}

// Encode function not implement (TODO)
func (c *DefaultCodec) Encode(ctx context.Context, event logevent.LogEvent, dataChan chan<- []byte) (ok bool, err error) {
	return false, ErrorNotImplement1.New(nil)
}
