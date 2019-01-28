package codecjson

import (
	"context"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
)

// ModuleName is the name used in config file
const ModuleName = "json"

// ErrorTag tag added to event when process module failed
const ErrorTag = "gogstash_codec_json_error"

// Codec default struct for codec
type Codec struct {
	config.CodecConfig
}

// InitHandler initialize the codec plugin
func InitHandler(context.Context, *config.ConfigRaw) (config.TypeCodecConfig, error) {
	return &Codec{
		CodecConfig: config.CodecConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
	}, nil
}

// Decode returns an event from 'data' as JSON format, adding provided 'eventExtra'
func (c *Codec) Decode(ctx context.Context, data interface{},
	eventExtra map[string]interface{},
	msgChan chan<- logevent.LogEvent) (ok bool, err error) {

	event := logevent.LogEvent{
		Timestamp: time.Now(),
		Extra:     eventExtra,
	}

	switch v := data.(type) {
	case string:
		if err = jsoniter.Unmarshal([]byte(v), &event.Extra); err != nil {
			event.Message = v
		}
	case []byte:
		if err = jsoniter.Unmarshal(v, &event.Extra); err != nil {
			event.Message = string(v)
		}
	case map[string]interface{}:
		if event.Extra != nil {
			for k, val := range v {
				event.Extra[k] = val
			}
		} else {
			event.Extra = v
		}
	default:
		err = config.ErrDecodeData
	}
	if err != nil {
		event.AddTag(ErrorTag)
		goglog.Logger.Error(err)
	}

	if event.Extra != nil {
		// try to fill basic log event by json message
		if value, ok := event.Extra["message"]; ok {
			switch v := value.(type) {
			case string:
				event.Message = v
				delete(event.Extra, "message")
			}
		}
		if value, ok := event.Extra["@timestamp"]; ok {
			switch v := value.(type) {
			case string:
				if timestamp, err2 := time.Parse(time.RFC3339Nano, v); err2 == nil {
					event.Timestamp = timestamp
					delete(event.Extra, "@timestamp")
				}
			}
		}
		if value, ok := event.Extra[logevent.TagsField]; ok {
			if event.ParseTags(value) {
				delete(event.Extra, logevent.TagsField)
			} else {
				goglog.Logger.Warnf("malformed tags: %v", value)
			}
		}
	}

	msgChan <- event
	ok = true

	return
}

// DecodeEvent decodes 'data' as JSON format to event
func (c *Codec) DecodeEvent(data []byte, v interface{}) error {
	event := logevent.LogEvent{
		Timestamp: time.Now(),
	}

	if err := jsoniter.Unmarshal(data, &event.Extra); err != nil {
		event.Message = string(data)
		event.AddTag(ErrorTag)
		goglog.Logger.Error(err)
	}

	if event.Extra != nil {
		// try to fill basic log event by json message
		if value, ok := event.Extra["message"]; ok {
			switch v := value.(type) {
			case string:
				event.Message = v
				delete(event.Extra, "message")
			}
		}
		if value, ok := event.Extra["@timestamp"]; ok {
			switch v := value.(type) {
			case string:
				if timestamp, err2 := time.Parse(time.RFC3339Nano, v); err2 == nil {
					event.Timestamp = timestamp
					delete(event.Extra, "@timestamp")
				}
			}
		}
		if value, ok := event.Extra[logevent.TagsField]; ok {
			if event.ParseTags(value) {
				delete(event.Extra, logevent.TagsField)
			} else {
				goglog.Logger.Warnf("malformed tags: %v", value)
			}
		}
	}

	switch e := v.(type) {
	case *interface{}:
		*e = event
	case *logevent.LogEvent:
		*e = event
	default:
		return config.ErrorUnsupportedTargetEvent
	}
	return nil
}

// Encode function not implement (TODO)
func (c *Codec) Encode(ctx context.Context, event logevent.LogEvent, dataChan chan<- []byte) (ok bool, err error) {
	return false, config.ErrorNotImplement1.New(nil)
}
