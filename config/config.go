package config

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/codegangsta/inject"
	"github.com/tsaikd/KDGoLib/errutil"
	"github.com/tsaikd/KDGoLib/injectutil"
	"github.com/tsaikd/gogstash/config/logevent"
	"gopkg.in/yaml.v2"
)

// errors
var (
	ErrorReadConfigFile1     = errutil.NewFactory("Failed to read config file: %q")
	ErrorUnmarshalJSONConfig = errutil.NewFactory("Failed unmarshalling config in JSON format")
	ErrorUnmarshalYAMLConfig = errutil.NewFactory("Failed unmarshalling config in YAML format")
)

// Config contains all config
type Config struct {
	inject.Injector `json:"-"`
	InputRaw        []ConfigRaw `json:"input,omitempty" yaml:"input"`
	FilterRaw       []ConfigRaw `json:"filter,omitempty" yaml:"filter"`
	OutputRaw       []ConfigRaw `json:"output,omitempty" yaml:"output"`
}

// InChan input channel
type InChan chan logevent.LogEvent

// OutChan output channel
type OutChan chan logevent.LogEvent

// LoadFromFile load config from filepath
func LoadFromFile(path string) (config Config, err error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return config, ErrorReadConfigFile1.New(err, path)
	}

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".yml", ".yaml":
		return LoadFromYAML(data)
	default:
		return LoadFromJSON(data)
	}
}

// LoadFromJSON load config from []byte in JSON format
func LoadFromJSON(data []byte) (config Config, err error) {
	if data, err = CleanComments(data); err != nil {
		return
	}

	if err = json.Unmarshal(data, &config); err != nil {
		return config, ErrorUnmarshalJSONConfig.New(err)
	}

	config.Injector = inject.New()
	config.Map(Logger)

	inchan := make(InChan, 100)
	outchan := make(OutChan, 100)
	config.Map(inchan)
	config.Map(outchan)

	rv := reflect.ValueOf(&config)
	formatReflect(rv)

	return
}

// LoadFromYAML load config from []byte in YAML format
func LoadFromYAML(data []byte) (config Config, err error) {
	if err = yaml.Unmarshal(data, &config); err != nil {
		return config, ErrorUnmarshalYAMLConfig.New(err)
	}

	config.Injector = inject.New()
	config.Map(Logger)

	inchan := make(InChan, 100)
	outchan := make(OutChan, 100)
	config.Map(inchan)
	config.Map(outchan)

	rv := reflect.ValueOf(&config)
	formatReflect(rv)

	return
}

// ReflectConfig set conf from confraw
func ReflectConfig(confraw *ConfigRaw, conf interface{}) (err error) {
	data, err := json.Marshal(confraw)
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

// CleanComments used for remove non-standard json comments.
// Supported comment formats
// format 1: ^\s*#
// format 2: ^\s*//
func CleanComments(data []byte) (out []byte, err error) {
	reForm1 := regexp.MustCompile(`^\s*#`)
	reForm2 := regexp.MustCompile(`^\s*//`)
	data = bytes.Replace(data, []byte("\r"), []byte(""), 0) // Windows
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

	out = bytes.Join(filtered, []byte("\n"))
	return
}

// InvokeSimple invoke and handle return value for error type
func (t *Config) InvokeSimple(arg interface{}) (err error) {
	_, err = injectutil.Invoke(t.Injector, arg)
	return
}
