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
	Timestamp time.Time              `json:"timestamp"`
	Message   string                 `json:"message"`
	Tags      []string               `json:"tags,omitempty"`
	Extra     map[string]interface{} `json:"-"`
}

type Config struct {
	SortMapKeys bool     `yaml:"sort_map_keys"`
	RemoveField []string `yaml:"remove_field"`

	jsonMarshal       func(v interface{}) ([]byte, error)
	jsonMarshalIndent func(v interface{}, prefix, indent string) ([]byte, error)
}

// TagsField is the event tags field name
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
func (t *LogEvent) ParseTags(tags interface{}) bool {
	switch v := tags.(type) {
	case []interface{}:
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

func (t LogEvent) getJSONMap() map[string]interface{} {
	event := map[string]interface{}{
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

func (t LogEvent) Get(field string) (v interface{}) {
	switch field {
	case "@timestamp":
		v = t.Timestamp
	case "message":
		v = t.Message
	case TagsField:
		v = t.Tags
	default:
		v = t.Extra[field]
	}
	return
}

func (t LogEvent) GetString(field string) string {
	switch field {
	case "@timestamp":
		return t.Timestamp.UTC().Format(timeFormat)
	case "message":
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

func (t LogEvent) GetValue(field string) (interface{}, bool) {
	return getPathValue(t.Extra, field)
}

func (t *LogEvent) SetValue(field string, v interface{}) bool {
	if field == "message" {
		if value, ok := v.(string); ok {
			t.Message = value
			return false
		}
	}
	if t.Extra == nil {
		t.Extra = map[string]interface{}{}
	}
	return setPathValue(t.Extra, field, v)
}

func (t *LogEvent) Remove(field string) bool {
	return removePathValue(t.Extra, field)
}

var (
	reCurrentTime = regexp.MustCompile(`%{\+([^}]+)}`)
	reEventTime   = regexp.MustCompile(`%{\+@([^}]+)}`)
	revar         = regexp.MustCompile(`%{([\w@\.]+)}`)
)

// FormatWithEnv format string with environment value, ex: %{HOSTNAME}
func FormatWithEnv(text string) (result string) {
	result = text

	matches := revar.FindAllStringSubmatch(result, -1)
	for _, submatches := range matches {
		field := submatches[1]
		value := os.Getenv(field)
		if value != "" {
			result = strings.Replace(result, submatches[0], value, -1)
		} else if field == "HOSTNAME" {
			if value, _ := os.Hostname(); value != "" {
				result = strings.Replace(result, submatches[0], value, -1)
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
		result = strings.Replace(result, submatches[0], value, -1)
	}

	return
}

// FormatWithEventTime format string with event time, ex: %{+@2006-01-02}
func FormatWithEventTime(text string, evevtTime time.Time) (result string) {
	result = text

	matches := reEventTime.FindAllStringSubmatch(result, -1)
	for _, submatches := range matches {
		value := evevtTime.Format(submatches[1])
		result = strings.Replace(result, submatches[0], value, -1)
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
			out = strings.Replace(out, submatches[0], value, -1)
		}
	}

	out = FormatWithEnv(out)

	return
}
