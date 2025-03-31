package codecprometheus

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
)

// ModuleName is the name used in config file
const ModuleName = "prometheus"

// ErrorTag tag added to event when process module failed
const ErrorTag = "gogstash_codec_prometheus_error"

// Codec default struct for codec
type Codec struct {
	config.CodecConfig
}

var (
	fullStringMatch      = `(\S+)\{([^\}]+)\}\s+(.+)`
	fullStringMatchRegex *regexp.Regexp
)

// InitHandler initialize the codec plugin
func InitHandler(context.Context, config.ConfigRaw) (config.TypeCodecConfig, error) {
	var err error
	fullStringMatchRegex, err = regexp.Compile(fullStringMatch)
	return &Codec{
		CodecConfig: config.CodecConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
	}, err
}

// Decode returns an event from 'data' as JSON format, adding provided 'eventExtra'
func (c *Codec) Decode(ctx context.Context, data any, eventExtra map[string]any, tags []string, msgChan chan<- logevent.LogEvent) (ok bool, err error) {
	event := logevent.LogEvent{
		Timestamp: time.Now(),
		Extra:     eventExtra,
	}
	if eventExtra == nil {
		event.Extra = map[string]any{}
	}
	event.AddTag(tags...)

	datastr := ""
	switch data := data.(type) {
	case string:
		datastr = data
	case []byte:
		datastr = string(data)
	default:
		err = config.ErrDecodeData
		event.AddTag(ErrorTag)
		goglog.Logger.Error(err)
		return false, err
	}

	// Split into individual metrics

	// Convert to string, make sure we have a \n suffix
	if !strings.HasSuffix(datastr, "\n") {
		datastr += "\n"
	}

	lines := strings.Split(datastr, "\n")
	savetype := map[string]string{}
	for _, line := range lines {
		if len(strings.TrimSpace(line)) == 0 {
			continue
		}

		// fmt.Printf("Decode line = '%s'\n", line)
		// Preserve formatted type definitions
		if strings.HasPrefix(line, "# TYPE") {
			p := strings.Split(line, " ")
			if len(p) < 4 {
				goglog.Logger.Error(fmt.Errorf("prometheus: TYPE line: %s", line))
				continue
			}
			savetype[p[2]] = p[3]
		}
		// Skip line if comment, we don't process them further
		if strings.HasPrefix(line, "#") {
			continue
		}

		// Decode lines, every metric is an possible individual event
		{
			e := event
			err = c.decodePrometheusEvent(line, savetype, &e)
			if err == nil {
				msgChan <- e
				ok = true
			}
		}
	}
	return ok, err
}

func (c *Codec) DecodeEvent(msg []byte, event *logevent.LogEvent) (err error) {
	return fmt.Errorf("unimplemented")
}

// decodePrometheusEvent decodes 'data' as prometheus format to event
func (c *Codec) decodePrometheusEvent(line string, typeref map[string]string, event *logevent.LogEvent) (err error) {
	// fmt.Printf("decodePrometheusEvent line = '%s'\n", line)
	// If the pointer is empty, raise error
	if event == nil {
		goglog.Logger.Errorf("Provided DecodeEvent target event pointer is nil")
		return config.ErrDecodeNilTarget
	}

	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	event.Message = line

	p := strings.Split(line, " ")
	if len(p) < 2 {
		return err
	}
	name := p[0]
	value := strings.TrimSpace(strings.Join(p[1:], " "))

	// Detect bracket format
	if !fullStringMatchRegex.MatchString(line) {
		// fmt.Printf("Does not match regex!!!\n")
		// Deal with straight name/value pair
		event.Extra["name"] = name
		event.Extra["value"], err = strconv.ParseFloat(value, 64)
		if err != nil {
			event.Extra["value"] = value
		}
		t, ok := typeref[name]
		if ok {
			event.Extra["type"] = t
		}
		c.populateEventExtras(event)
		return err
	}

	m := fullStringMatchRegex.FindStringSubmatch(line)
	if len(m) < 4 {
		goglog.Logger.Errorf("incorrect bracket format")
		return config.ErrorInitCodecFailed1
	}

	name = strings.Split(line, "{")[0]
	event.Extra["name"] = name
	event.Extra["value"], err = strconv.ParseFloat(strings.Split(line, "}")[1], 64)
	if err != nil {
		event.Extra["value"] = strings.Split(line, "}")[1]
	}
	t, ok := typeref[name]
	if ok {
		event.Extra["type"] = t
	}

	dimensions := map[string]string{}
	for _, d := range strings.Split(m[2], ",") {
		kv := strings.Split(d, "=")
		if len(kv) < 2 {
			continue
		}
		dimensions[strings.ToLower(kv[0])] = strings.ReplaceAll(kv[1], `"`, "")
	}
	event.Extra["dimensions"] = dimensions

	return err
}

// Encode encodes the event to a JSON encoded message
func (c *Codec) Encode(ctx context.Context, event logevent.LogEvent, dataChan chan<- []byte) (ok bool, err error) {
	output, err := event.MarshalJSON()
	if err != nil {
		return false, err
	}
	select {
	case <-ctx.Done():
		return false, nil
	case dataChan <- output:
	}
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
