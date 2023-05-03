package codecazurejson

import (
	"context"
	"time"

	jsoniter "github.com/json-iterator/go"

	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
)

// ModuleName is the name used in config file
const ModuleName = "azureeventhubjson"

// ErrorTag tag added to event when process module failed
const ErrorTag = "gogstash_codec_json_error"

// Codec default struct for codec
type Codec struct {
	config.CodecConfig
}

// InitHandler initialize the codec plugin
func InitHandler(context.Context, config.ConfigRaw) (config.TypeCodecConfig, error) {
	return &Codec{
		CodecConfig: config.CodecConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
	}, nil
}

// Decode returns an event from 'data' as JSON format, adding provided 'eventExtra'
func (c *Codec) Decode(ctx context.Context, data any,
	eventExtra map[string]any, tags []string,
	msgChan chan<- logevent.LogEvent) (ok bool, err error) {
	clonedEventExtras := make(map[string]any, len(eventExtra))
	for k, v := range eventExtra {
		clonedEventExtras[k] = v
	}

	event := logevent.LogEvent{
		Timestamp: time.Now(),
		Extra:     clonedEventExtras,
	}
	event.AddTag(tags...)

	switch v := data.(type) {
	case string:
		err = c.DecodeEvent([]byte(v), &event)
	case []byte:
		err = c.DecodeEvent(v, &event)
	case map[string]any:
		if event.Extra != nil {
			for k, val := range v {
				event.Extra[k] = val
			}
		} else {
			event.Extra = v
		}
		c.populateEventExtras(&event)
	default:
		err = config.ErrDecodeData
	}
	if err != nil {
		event.AddTag(ErrorTag)
		goglog.Logger.Error(err)
	}

	if records, rok := event.Extra["records"]; rok && len(event.Message) == 0 {
		for _, record := range records.([]any) {
			if _, err = c.Decode(ctx, record, eventExtra, tags, msgChan); err != nil {
				event.AddTag(ErrorTag)
				goglog.Logger.Error(err)
			}
		}
	} else {
		if ctx.Err() == context.Canceled {
			ok = false
			return
		}
		msgChan <- event
	}

	ok = true
	return ok, err
}

// DecodeEvent decodes 'data' as JSON format to event
func (c *Codec) DecodeEvent(data []byte, event *logevent.LogEvent) (err error) {
	// If the pointer is empty, raise error
	if event == nil {
		goglog.Logger.Errorf("Provided DecodeEvent target event pointer is nil")
		return config.ErrDecodeNilTarget
	}

	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	if err = jsoniter.Unmarshal(data, &event.Extra); err != nil {
		event.Message = string(data)
		event.AddTag(ErrorTag)
		goglog.Logger.Error(err)
	}

	c.populateEventExtras(event)

	return
}

// Encode encodes the event to a JSON encoded message
func (c *Codec) Encode(_ context.Context, event logevent.LogEvent, dataChan chan<- []byte) (ok bool, err error) {
	output, err := event.MarshalJSON()
	if err != nil {
		return false, err
	}

	dataChan <- output
	return true, nil
}

func (c *Codec) populateEventExtras(event *logevent.LogEvent) {
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
}
