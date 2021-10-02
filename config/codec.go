package config

import (
	"context"
	"errors"
	"time"

	"github.com/icza/dyno"
	"github.com/tsaikd/KDGoLib/errutil"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
)

// errors
var (
	ErrorUnknownCodecType1      = errutil.NewFactory("unknown codec config type: %q")
	ErrorInitCodecFailed1       = errutil.NewFactory("initialize codec module failed: %v")
	ErrorNotImplement1          = errutil.NewFactory("%q is not implement")
	ErrorUnsupportedTargetEvent = errors.New("unsupported target event to decode")
)

// TypeCodecConfig is interface of codec module
type TypeCodecConfig interface {
	TypeCommonConfig
	// Decode - The codec’s decode method is where data coming in from an input is transformed into an event.
	//  'ok' returns a boolean indicating if an event was created and sent to a provided 'msgChan' channel
	//  'error' is returned in case of any failure handling input 'data', but 'ok' == false DOES NOT indicate an error
	Decode(ctx context.Context, data interface{}, extra map[string]interface{}, tags []string, msgChan chan<- logevent.LogEvent) (ok bool, err error)
	// DecodeEvent decodes 'data' to 'event' pointer, creating new current timestamp if IsZero
	//  'error' is returned in case of any failure handling input 'data'
	DecodeEvent(data []byte, event *logevent.LogEvent) error
	// Encode - The encode method takes an event and serializes it (encodes) into another format.
	//  'ok' returns a boolean indicating if an event was encoded and sent to a provided 'dataChan' channel
	//  'error' is returned in case of any failure encoding 'event', but 'ok' == false DOES NOT indicate an error
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

// GetCodecOrDefault returns a codec based on the `codec` configuration, if specified in `ConfigRaw` input,
//  else an instance of `DefaultCodec` is returned
func GetCodecOrDefault(ctx context.Context, raw ConfigRaw) (TypeCodecConfig, error) {
	c, err := GetCodec(ctx, raw)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return DefaultCodecInitHandler(ctx, nil)
	}
	return c, nil
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

	switch cfg := codecConfig.(type) {
	case map[string]interface{}:
		return getCodec(ctx, ConfigRaw(cfg))
	case string:
		// shorthand codec config method:
		// codec: [codecTypeName]
		return getCodec(ctx, ConfigRaw{"type": cfg})
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

// DefaultErrorTag tag added to event when process module failed
const DefaultErrorTag = "gogstash_codec_default_error"

// codec errors
var (
	ErrDecodeData      = errors.New("decode data error")
	ErrDecodeNilTarget = errors.New("decode event target is nil")
)

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
func (c *DefaultCodec) Decode(ctx context.Context, data interface{},
	eventExtra map[string]interface{},
	tags []string,
	msgChan chan<- logevent.LogEvent) (ok bool, err error) {

	event := logevent.LogEvent{
		Timestamp: time.Now(),
		Extra:     eventExtra,
	}
	event.AddTag(tags...)

	switch v := data.(type) {
	case string:
		event.Message = v
	case []byte:
		event.Message = string(v)
	default:
		err = ErrDecodeData
		event.AddTag(DefaultErrorTag)
	}

	goglog.Logger.Debugf("%q %v", event.Message, event)
	msgChan <- event
	ok = true

	return
}

// DecodeEvent decodes data to event pointer, creating new current timestamp if IsZero
func (c *DefaultCodec) DecodeEvent(data []byte, event *logevent.LogEvent) error {
	if event == nil {
		goglog.Logger.Errorf("Provided DecodeEvent target event pointer is nil")
		return ErrDecodeNilTarget
	}

	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	event.Message = string(data)

	return nil
}

// Encode sends the message field, ignoring any extra fields
func (c *DefaultCodec) Encode(ctx context.Context, event logevent.LogEvent, dataChan chan<- []byte) (ok bool, err error) {
	// return if there is no message field
	if len(event.Message) == 0 {
		return false, nil
	}
	// send message
	dataChan <- []byte(event.Message)

	return true, nil
}
