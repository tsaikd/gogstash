package config

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"regexp"

	log "github.com/Sirupsen/logrus"
)

type Config struct {
	InputRaw  []map[string]interface{} `json:"input,omitempty"`
	FilterRaw []map[string]interface{} `json:"filter,omitempty"`
	OutputRaw []map[string]interface{} `json:"output,omitempty"`
}

type CommonConfig struct {
	Type string `json:"type"`
}

type TypeConfig interface {
	Type() string
}

func (config *Config) Filter() (filters []interface{}) {
	for _, mapraw := range config.FilterRaw {
		switch mapraw["type"] {
		default:
			log.Errorf("Unknown type: %q", mapraw["type"])
		}
	}
	return
}

func LoadConfig(path string) (config Config, err error) {
	var (
		buffer []byte
	)

	if buffer, err = ioutil.ReadFile(path); err != nil {
		log.Errorf("Failed to read config file %q\n%s", path, err)
		return
	}

	if buffer, err = StripComments(buffer); err != nil {
		log.Errorf("Failed to strip comments from json\n%s", err)
		return
	}

	if err = json.Unmarshal(buffer, &config); err != nil {
		log.Errorf("Failed unmarshalling json\n%s", err)
		return
	}

	return
}

func StripComments(data []byte) ([]byte, error) {
	data = bytes.Replace(data, []byte("\r"), []byte(""), 0) // Windows
	lines := bytes.Split(data, []byte("\n"))
	filtered := make([][]byte, 0)

	for _, line := range lines {
		match, err := regexp.Match(`^\s*#`, line)
		if err != nil {
			return nil, err
		}
		if !match {
			filtered = append(filtered, line)
		}
	}

	return bytes.Join(filtered, []byte("\n")), nil
}
