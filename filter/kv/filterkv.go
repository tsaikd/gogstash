package kv

import (
	"context"
	"strconv"
	"strings"

	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
)

// ModuleName is the name used in the config file
const ModuleName = "kv"

// FilterConfig holds the configuration json fields and internal objects
type FilterConfig struct {
	config.FilterConfig
	Source  string   `json:"source" yaml:"source"`   // source message field name
	Target  string   `json:"target" yaml:"target"`   // target field where date should be stored
	Strings []string `json:"strings" yaml:"strings"` // values to keep as strings even if they are numbers
}

// DefaultFilterConfig returns an FilterConfig struct with default values
func DefaultFilterConfig() FilterConfig {
	return FilterConfig{
		FilterConfig: config.FilterConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
	}
}

// InitHandler initialize the filter plugin
func InitHandler(ctx context.Context, raw config.ConfigRaw) (config.TypeFilterConfig, error) {
	conf := DefaultFilterConfig()
	if err := config.ReflectConfig(raw, &conf); err != nil {
		return nil, err
	}

	return &conf, nil
}

// Event the main filter event
func (f *FilterConfig) Event(ctx context.Context, event logevent.LogEvent) (logevent.LogEvent, bool) {
	if value, ok := event.Get(f.Source).(string); ok {
		if len(value) > 0 {
			kvpairs := splitQuotedStringsBySpace(value)
			if len(kvpairs) > 0 {
				kv := splitIntoKV(kvpairs, f.Strings)
				if f.Target == "" {
					for k, v := range kv {
						event.SetValue(k, v)
					}
				} else {
					event.SetValue(f.Target, kv)
				}
				return event, true
			}
		}
	}
	return event, false
}

// splitQuotedStringsBySpace returns an array of key/values
func splitQuotedStringsBySpace(input string) (result []string) {
	var head, quote int
	inBetween := true
	for x := 0; x < len(input); x++ {
		switch r := input[x]; r {
		case ' ':
			if inBetween == false && x > head && quote == 0 {
				s := input[head:x]
				if strings.IndexRune(s, '=') > 0 {
					result = append(result, s)
				}
				head = x + 1
			} else if quote == 0 {
				inBetween = true
				head = x + 1
			}
		case '"':
			if quote == 0 {
				quote = 1
			} else {
				quote = 0
			}
		default:
			inBetween = false
		}
	}
	// get last element if there is one
	if head != len(input) {
		s := input[head:len(input)]
		if strings.IndexRune(s, '=') > 0 {
			result = append(result, s)
		}
	}
	return
}

// contains checks if a key is in the list of elements
func contains(key string, elements *[]string) bool {
	for _, v := range *elements {
		if v == key {
			return true
		}
	}
	return false
}

// splitIntoKV splits an array of input strings into a map with the keys/values where value is either string or integer
func splitIntoKV(input []string, keepAsString []string) map[string]interface{} {
	result := make(map[string]interface{})
	for _, v := range input {
		separator := strings.IndexRune(v, '=')
		if separator > 0 && separator < len(v)-1 {
			key := v[:separator]
			var val string
			if v[separator+1] == '"' {
				val = v[separator+2 : len(v)-1]
			} else {
				val = v[separator+1:]
			}
			// check if val is an integer
			number, err := strconv.Atoi(val)
			if err == nil && contains(key, &keepAsString) == false {
				result[key] = number
			} else {
				result[key] = val
			}
		}
	}
	return result
}
