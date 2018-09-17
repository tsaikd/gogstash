package logevent

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/tsaikd/KDGoLib/jsonex"
	"github.com/tsaikd/gogstash/config/goglog"
)

type LogEvent struct {
	Timestamp time.Time              `json:"timestamp"`
	Message   string                 `json:"message"`
	Tags      []string               `json:"tags,omitempty"`
	Extra     map[string]interface{} `json:"-"`
}

const timeFormat = `2006-01-02T15:04:05.999999999Z`

func appendIfMissing(slice []string, s string) []string {
	for _, ele := range slice {
		if ele == s {
			return slice
		}
	}
	return append(slice, s)
}

func (t *LogEvent) AddTag(tags ...string) {
	for _, tag := range tags {
		ftag := t.Format(tag)
		t.Tags = appendIfMissing(t.Tags, ftag)
	}
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
		if _, ok := event["tags"]; ok {
			// extra contains tags field
			switch tags := event["tags"].(type) {
			case []interface{}:
				ok := true
				stringTags := make([]string, 0, len(tags))
			tags_loop:
				for _, v := range tags {
					switch tag := v.(type) {
					case string:
						stringTags = append(stringTags, tag)
					default:
						ok = false
						break tags_loop
					}
				}
				if ok {
					event["tags"] = append(stringTags, t.Tags...)
				} else {
					goglog.Logger.Warnf("event %v contains malformed tags", t)
				}
			case []string:
				event["tags"] = append(tags, t.Tags...)
			case nil:
				event["tags"] = t.Tags
			default:
				goglog.Logger.Warnf("event %v contains malformed tags", t)
			}
		} else {
			event["tags"] = t.Tags
		}
	}
	return event
}

func (t LogEvent) MarshalJSON() (data []byte, err error) {
	event := t.getJSONMap()
	return jsonex.Marshal(event)
}

func (t LogEvent) MarshalIndent() (data []byte, err error) {
	event := t.getJSONMap()
	return jsonex.MarshalIndent(event, "", "\t")
}

func (t LogEvent) Get(field string) (v interface{}) {
	switch field {
	case "@timestamp":
		v = t.Timestamp
	case "message":
		v = t.Message
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
		v, ok := getValueFromObject(t.Extra, field)
		if ok {
			return fmt.Sprintf("%v", v)
		}
		return ""
	}
}

func (t LogEvent) GetValue(field string) (interface{}, bool) {
	return getValueFromObject(t.Extra, field)
}

func (t *LogEvent) SetValue(field string, v interface{}) bool {
	if t.Extra == nil {
		t.Extra = map[string]interface{}{}
	}
	return setValueToObject(t.Extra, field, v)
}

func getValueFromObject(obj map[string]interface{}, field string) (interface{}, bool) {
	fieldSplits := strings.Split(field, ".")
	if len(fieldSplits) < 2 {
		val, ok := obj[field]
		return val, ok
	}

	switch child := obj[fieldSplits[0]].(type) {
	case map[string]interface{}:
		return getValueFromObject(child, strings.Join(fieldSplits[1:], "."))
	default:
		return nil, false
	}
}

func setValueToObject(obj map[string]interface{}, field string, v interface{}) bool {
	fieldSplits := strings.Split(field, ".")
	if len(fieldSplits) < 2 {
		obj[field] = v
		return true
	}

	switch child := obj[fieldSplits[0]].(type) {
	case nil:
		obj[fieldSplits[0]] = map[string]interface{}{}
		return setValueToObject(obj[fieldSplits[0]].(map[string]interface{}),
			strings.Join(fieldSplits[1:], "."), v)
	case map[string]interface{}:
		return setValueToObject(child, strings.Join(fieldSplits[1:], "."), v)
	default:
		return false
	}
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
