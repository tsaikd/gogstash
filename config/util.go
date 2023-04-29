package config

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"reflect"
	"regexp"
	"strings"

	"github.com/icza/dyno"
	"github.com/tsaikd/KDGoLib/logutil"

	"github.com/tsaikd/gogstash/config/logevent"
)

// ReflectConfig set conf from confraw
func ReflectConfig(confraw ConfigRaw, conf any) (err error) {
	data, err := json.Marshal(dyno.ConvertMapI2MapS(map[string]any(confraw)))
	if err != nil {
		return
	}

	if err = json.Unmarshal(data, conf); err != nil {
		return
	}

	rv := reflect.ValueOf(conf).Elem()
	formatReflect(rv)

	return
}

// GetFromObject obtaining value from specified field recursively
func GetFromObject(obj map[string]any, field string) any {
	fieldSplits := strings.Split(field, ".")
	for i, key := range fieldSplits {
		if i >= len(fieldSplits)-1 {
			if v, ok := obj[key]; ok {
				return v
			}
			return nil
		} else if node, ok := obj[key]; ok {
			switch v := node.(type) {
			case map[string]any:
				obj = v
			default:
				return nil
			}
		} else {
			break
		}
	}
	return nil
}

func formatReflect(rv reflect.Value) {
	if !rv.IsValid() {
		return
	}

	switch rv.Kind() {
	case reflect.Ptr:
		if !rv.IsNil() {
			formatReflect(rv.Elem())
		}
	case reflect.Struct:
		for i := 0; i < rv.NumField(); i++ {
			field := rv.Field(i)
			formatReflect(field)
		}
	case reflect.String:
		if !rv.CanSet() {
			return
		}
		value := rv.Interface().(string)
		value = logevent.FormatWithEnv(value)
		rv.SetString(value)
	}
}

// cleanComments used for remove non-standard json comments.
// Supported comment formats
// format 1: ^\s*#
// format 2: ^\s*//
func cleanComments(data []byte) ([]byte, error) {
	reForm1 := regexp.MustCompile(`^\s*#`)
	reForm2 := regexp.MustCompile(`^\s*//`)
	data = bytes.ReplaceAll(data, []byte("\r"), []byte("")) // Windows
	lines := bytes.Split(data, []byte("\n"))
	var filtered [][]byte

	for _, line := range lines {
		if reForm1.Match(line) {
			continue
		}
		if reForm2.Match(line) {
			continue
		}
		filtered = append(filtered, line)
	}

	return bytes.Join(filtered, []byte("\n")), nil
}

func contextWithOSSignal(parent context.Context, logger logutil.LevelLogger, sig ...os.Signal) context.Context {
	osSignalChan := make(chan os.Signal, 1)
	signal.Notify(osSignalChan, sig...)

	ctx, cancel := context.WithCancel(parent)

	go func(cancel context.CancelFunc) {
		sig := <-osSignalChan
		logger.Info(sig)
		cancel()
	}(cancel)

	return ctx
}
