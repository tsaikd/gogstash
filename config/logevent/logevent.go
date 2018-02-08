package logevent

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/tsaikd/KDGoLib/jsonex"
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
	if len(t.Tags) > 0 {
		event["tags"] = t.Tags
	}
	for key, value := range t.Extra {
		event[key] = value
	}
	return event
}

func (t LogEvent) GetMap() map[string]interface{} {
	event := map[string]interface{}{
		"@timestamp": t.Timestamp,
	}
	if t.Message != "" {
		event["message"] = t.Message
	}
	if len(t.Tags) > 0 {
		event["tags"] = t.Tags
	}
	for key, value := range t.Extra {
		event[key] = value
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
		return getStringFromObject(t.Extra, field)
	}
}

func getStringFromObject(obj map[string]interface{}, field string) string {
	fieldSplits := strings.Split(field, ".")
	if len(fieldSplits) < 2 {
		if value, ok := obj[field]; ok {
			return fmt.Sprintf("%v", value)
		}
		return ""
	}

	switch child := obj[fieldSplits[0]].(type) {
	case map[string]interface{}:
		return getStringFromObject(child, strings.Join(fieldSplits[1:], "."))
	default:
		return fmt.Sprintf("%v", child)
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
