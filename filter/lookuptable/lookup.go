package lookuptable

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	lru "github.com/hashicorp/golang-lru"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
)

// ModuleName is the name used in config file
const ModuleName = "lookuptable"

const separator = ':'

// FilterConfig holds the configuration json fields and internal objects
type FilterConfig struct {
	config.FilterConfig

	// The field containing the user agent string.
	Source string `json:"source"`

	// The name of the field to assign user agent data into.
	// If not specified user agent data will be stored in the root of the event.
	Target string `json:"target"`

	// LookupFile contains one key-value mapping per line
	LookupFile string `json:"lookup_file"`

	// UA parsing is surprisingly expensive.
	// We can optimize it by adding a cache
	CacheSize int `json:"cache_size"`

	cache *lru.Cache

	file *os.File
}

// DefaultFilterConfig returns an FilterConfig struct with default values
func DefaultFilterConfig() FilterConfig {
	return FilterConfig{
		FilterConfig: config.FilterConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		Target:    "",
		CacheSize: 1000,
	}
}

// InitHandler initialize the filter plugin
func InitHandler(
	ctx context.Context,
	raw config.ConfigRaw,
	control config.Control,
) (config.TypeFilterConfig, error) {
	conf := DefaultFilterConfig()
	err := config.ReflectConfig(raw, &conf)
	if err != nil {
		return nil, err
	}

	if conf.LookupFile == "" {
		return &conf, errors.New("no lookup file defined")
	}

	conf.file, err = os.OpenFile(conf.LookupFile, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return &conf, fmt.Errorf("read lookup file: %v", err)
	}

	conf.cache, err = lru.New(conf.CacheSize)
	if err != nil {
		return nil, err
	}
	return &conf, nil
}

// Event the main filter event
func (f *FilterConfig) Event(ctx context.Context, event logevent.LogEvent) (logevent.LogEvent, bool) {
	source := event.GetString(f.Source)
	if source == "" {
		return event, false
	}
	val, err := f.findFromFile(source)
	if err != nil {
		return event, false
	}

	if val != nil {
		event.SetValue(f.Target, val)
	}
	return event, true
}

// findFromFile translates key to value.
func (f *FilterConfig) findFromFile(key string) (interface{}, error) {
	cached, ok := f.cache.Get(key)
	if ok {
		return cached, nil
	}
	_, err := f.file.Seek(0, io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("seek cleanup file: %v", err)
	}
	var value interface{}
	scanner := bufio.NewScanner(f.file)

	line := 0
	for {
		ok := scanner.Scan()
		if !ok {
			break
		}

		text := scanner.Text()
		line += 1
		tokens, err := tokenizeLine(text)
		if err != nil {
			goglog.Logger.Warningf("invalid lookup value (line %d): must have format <str>:<str>", line)
			continue
		}
		if key == tokens[0] {
			value = tokens[1]
			break
		}
	}

	if value != nil {
		f.cache.Add(key, value)
	}
	return value, nil
}

// tokenize single line in lookuptable file returning key-value and possible error
func tokenizeLine(line string) ([2]string, error) {
	var runes []rune
	escaped := false
	tokens := [2]string{}
	var err error

	inKey := true
	for _, r := range line {
		switch {
		case escaped:
			escaped = false
			fallthrough
		default:
			runes = append(runes, r)
		case r == '\\':
			escaped = true
			continue
		case r == separator:
			if inKey {
				tokens[0] = string(runes)
				inKey = false
				runes = runes[:0]
			} else {
				err = errors.New("too many tokens")
				break
			}
		}
	}
	tokens[1] = string(runes)
	if escaped {
		err = errors.New("invalid escaped ")
	}

	if err != nil {
		return [2]string{}, err
	}

	for i, v := range tokens {
		tokens[i] = strings.Trim(v, " \n")
	}

	return tokens, nil
}
