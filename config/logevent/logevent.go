package logevent

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/tsaikd/KDGoLib/jsonex"
)

type LogEvent struct {
	Timestamp time.Time      `json:"timestamp"`
	Message   string         `json:"message"`
	Tags      []string       `json:"tags,omitempty"`
	Extra     map[string]any `json:"-"`
	Drop      bool
}

type Config struct {
	SortMapKeys bool     `yaml:"sort_map_keys"`
	RemoveField []string `yaml:"remove_field"`

	jsonMarshal       func(v any) ([]byte, error)
	jsonMarshalIndent func(v any, prefix, indent string) ([]byte, error)
}

// TagsField is the event tags field name
const TimestampField = "@timestamp"
const MessageField = "message"
const TagsField = "tags"

const timeFormat = `2006-01-02T15:04:05.999999999Z`

var config *Config

// SetConfig for LogEvent
func SetConfig(c *Config) {
	config = c
	json := jsoniter.Config{
		SortMapKeys:            c.SortMapKeys,
		ValidateJsonRawMessage: false,
		EscapeHTML:             false,
	}.Froze()
	config.jsonMarshal = json.Marshal
	config.jsonMarshalIndent = jsonex.MarshalIndent
}

func init() {
	SetConfig(&Config{SortMapKeys: false})
}

func appendIfMissing(slice []string, s string) []string {
	for _, ele := range slice {
		if ele == s {
			return slice
		}
	}
	return append(slice, s)
}

// AddTag add tags into event.Tags
func (t *LogEvent) AddTag(tags ...string) {
	for _, tag := range tags {
		ftag := t.Format(tag)
		t.Tags = appendIfMissing(t.Tags, ftag)
	}
}

// RemoveTag removes tags from event.Tags
func (t *LogEvent) RemoveTag(tags ...string) {
	for _, tag := range tags {
		ftag := t.Format(tag)
		var newTags []string
		for _, existingTag := range t.Tags {
			if existingTag != ftag {
				newTags = append(newTags, existingTag)
			}
		}
		t.Tags = newTags
	}
}

// ParseTags parse tags into event.Tags
func (t *LogEvent) ParseTags(tags any) bool {
	switch v := tags.(type) {
	case map[string]any:
		stringTags := make([]string, 0, len(v))
		for k := range v {
			stringTags = append(stringTags, k)
		}
		t.Tags = stringTags
		return true
	case []any:
		ok := true
		stringTags := make([]string, 0, len(v))
	tagsLoop:
		for _, t := range v {
			switch tag := t.(type) {
			case string:
				stringTags = append(stringTags, tag)
			default:
				ok = false
				break tagsLoop
			}
		}
		if ok {
			t.Tags = stringTags
			return true
		}
	case []string:
		t.Tags = v
		return true
	}
	return false
}

func (t LogEvent) getJSONMap() map[string]any {
	event := map[string]any{
		"@timestamp": t.Timestamp.UTC().Format(timeFormat),
	}
	if t.Message != "" {
		event["message"] = t.Message
	}
	for key, value := range t.Extra {
		event[key] = value
	}
	if len(t.Tags) > 0 {
		event[TagsField] = t.Tags
	}
	for _, field := range config.RemoveField {
		removePathValue(event, field)
	}
	return event
}

func (t LogEvent) MarshalJSON() (data []byte, err error) {
	event := t.getJSONMap()
	return config.jsonMarshal(event)
}

func (t LogEvent) MarshalIndent() (data []byte, err error) {
	event := t.getJSONMap()
	return config.jsonMarshalIndent(event, "", "\t")
}

func (t LogEvent) Get(field string) (v any) {
	switch field {
	case TimestampField:
		v = t.Timestamp
	case MessageField:
		v = t.Message
	case TagsField:
		v = t.Tags
	default:
		v, _ = getPathValue(t.Extra, field)
	}
	return
}

func (t LogEvent) GetString(field string) string {
	switch field {
	case TimestampField:
		return t.Timestamp.UTC().Format(timeFormat)
	case MessageField:
		return t.Message
	default:
		v, ok := getPathValue(t.Extra, field)
		if ok {
			if s, ok := v.(string); ok {
				return s
			}
			return fmt.Sprintf("%v", v)
		}
		return ""
	}
}

func (t LogEvent) GetValue(field string) (any, bool) {
	return getPathValue(t.Extra, field)
}

func (t *LogEvent) SetValue(field string, v any) bool {
	if field == "message" {
		if value, ok := v.(string); ok {
			t.Message = value
			return false
		}
	}
	if t.Extra == nil {
		t.Extra = map[string]any{}
	}
	return setPathValue(t.Extra, field, v)
}

func (t *LogEvent) Remove(field string) bool {
	return removePathValue(t.Extra, field)
}

var (
	reCurrentTime = regexp.MustCompile(`%{\+([^}]+)}`)
	reEventTime   = regexp.MustCompile(`%{\+@([^}]+)}`)
	revar         = regexp.MustCompile(`%{([\w@.]+)}`)
)

// FormatWithEnv format string with environment value, ex: %{HOSTNAME}
func FormatWithEnv(text string) (result string) {
	result = text

	matches := revar.FindAllStringSubmatch(result, -1)
	for _, submatches := range matches {
		field := submatches[1]
		value := os.Getenv(field)
		if value != "" {
			result = strings.ReplaceAll(result, submatches[0], value)
		} else if field == "HOSTNAME" {
			if value, _ := os.Hostname(); value != "" {
				result = strings.ReplaceAll(result, submatches[0], value)
			}
		}
	}

	return
}

// FormatWithCurrentTime format string with current time, ex: %{+2006-01-02}
func FormatWithCurrentTime(text string) (result string) {
	result = text

	matches := reCurrentTime.FindAllStringSubmatch(result, -1)
	for _, submatches := range matches {
		value := time.Now().Format(submatches[1])
		result = strings.ReplaceAll(result, submatches[0], value)
	}

	return
}

// FormatWithEventTime format string with event time, ex: %{+@2006-01-02}
func FormatWithEventTime(text string, evevtTime time.Time) (result string) {
	result = text

	matches := reEventTime.FindAllStringSubmatch(result, -1)
	for _, submatches := range matches {
		value := evevtTime.Format(submatches[1])
		result = strings.ReplaceAll(result, submatches[0], value)
	}

	return
}

// Format return string with current time / LogEvent field / ENV, ex: %{hostname}
func (t LogEvent) Format(format string) (out string) {
	out = format

	out = FormatWithEventTime(out, t.Timestamp)
	out = FormatWithCurrentTime(out)

	matches := revar.FindAllStringSubmatch(out, -1)
	for _, submatches := range matches {
		field := submatches[1]
		value := t.GetString(field)
		if value != "" {
			out = strings.ReplaceAll(out, submatches[0], value)
		}
	}

	out = FormatWithEnv(out)

	return
}
