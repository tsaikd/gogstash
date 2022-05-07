package hash

import (
	"context"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"hash"
	"hash/adler32"
	"hash/fnv"
	"math/big"

	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
)

// ModuleName is the name used in the config file
const ModuleName = "hash"

// FilterConfig holds the configuration json fields and internal objects
type FilterConfig struct {
	config.FilterConfig
	Source       []string           `json:"source" yaml:"source"` // source message field name(s)
	Target       string             `json:"target" yaml:"target"` // target field where the hash should be stored
	Kind         string             `json:"kind" yaml:"kind"`     // kind of hash
	Format       string             `json:"format" yaml:"format"` // output format
	hash         hash.Hash          // the hasher we use with hash.Hash interface
	hash32       func() hash.Hash32 // init function for hash.Hash32 interface, used if above is nil
	outputFormat int                // output format
}

// DefaultFilterConfig returns an FilterConfig struct with default values
func DefaultFilterConfig() FilterConfig {
	return FilterConfig{
		FilterConfig: config.FilterConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		Target: "hash",
		Source: []string{"message"},
		Kind:   "sha1",
	}
}

// Enum of supported output types
const (
	outputBinary = iota
	outputBase64
	outputBigInt
	outputHex
)

// hashAlgo defines a hash implementing hash.Hash
type hashAlgo struct {
	Name string           // name of algo
	Init func() hash.Hash // init func for hash method
}

// hash32Algo defines a hash implementing hash.Hash32
type hash32Algo struct {
	Name string             // name of algo
	Init func() hash.Hash32 // init func for hash method
}

// hash32Algos is a list of supported hash.Hash32 algorithms
var hash32Algos = []hash32Algo{
	{"adler32", adler32.New},
	{"fnv32a", fnv.New32a},
}

// hashAlgos is a list of supported hash.Hash algorithms
var hashAlgos = []hashAlgo{
	{"sha1", sha1.New},
	{"sha256", sha256.New},
	{"md5", md5.New},
	{"fnv128a", fnv.New128a},
}

// InitHandler initialize the filter plugin
func InitHandler(
	ctx context.Context,
	raw config.ConfigRaw,
	control config.Control,
) (config.TypeFilterConfig, error) {
	conf := DefaultFilterConfig()
	if err := config.ReflectConfig(raw, &conf); err != nil {
		return nil, err
	}

	if len(conf.Source) == 0 {
		return nil, errors.New("hash: no source fields")
	}
	if len(conf.Target) == 0 {
		return nil, errors.New("hash: no destination field")
	}

	err := initHashConfig(&conf)

	if err == nil {
		goglog.Logger.Debug("hash filter initialized")
	}
	return &conf, err
}

// initHashConfig initializes hash configuration. Moved to its own function so it can be tested.
func initHashConfig(conf *FilterConfig) error {
	// we either prepare a hash or hash32 algo
	for _, v := range hashAlgos {
		if v.Name == conf.Kind {
			conf.hash = v.Init()
			break
		}
	}
	if conf.hash == nil {
		for _, v := range hash32Algos {
			if v.Name == conf.Kind {
				conf.hash32 = v.Init
				break
			}
		}
	}
	if conf.hash == nil && conf.hash32 == nil {
		return fmt.Errorf("filterhash unsupported hash algo: %s", conf.Kind)
	}

	switch conf.Format {
	case "binary":
		conf.outputFormat = outputBinary
	case "base64":
		conf.outputFormat = outputBase64
	case "int":
		conf.outputFormat = outputBigInt
	default:
		conf.outputFormat = outputHex
	}
	return nil
}

// Event is the main filter event
func (f *FilterConfig) Event(ctx context.Context, event logevent.LogEvent) (logevent.LogEvent, bool) {
	// concat all fields that should be hashed together, make hash and set value
	source := f.getUnhashedString(&event)
	hash := f.makeHash(source)
	event.SetValue(f.Target, hash)
	return event, true
}

// getUnhashedString returns all configured source fields concatenated together
func (f *FilterConfig) getUnhashedString(event *logevent.LogEvent) string {
	var sourceString string
	for _, v := range f.Source {
		field := event.Get(v)
		sourceString = fmt.Sprintf("%v%v", sourceString, field)
	}
	return sourceString
}

// i32tob convert uint32 to byte array
func i32tob(val uint32) []byte {
	r := make([]byte, 4)
	binary.LittleEndian.PutUint32(r, val)
	return r
}

// makeHash returns a hash of the source string using the configured type.
// In return a string in any any supported output format.
func (f *FilterConfig) makeHash(source string) (result string) {
	var bs []byte
	if f.hash != nil {
		bs = f.hash.Sum([]byte(source))
	} else {
		h := f.hash32()
		h.Write([]byte(source))
		sum := h.Sum32()
		bs = i32tob(sum)
	}
	switch f.outputFormat {
	case outputBinary:
		result = string(bs)
	case outputBase64:
		result = base64.StdEncoding.EncodeToString(bs)
	case outputBigInt:
		z := new(big.Int)
		z.SetBytes(bs)
		result = z.String()
	default:
		result = fmt.Sprintf("%x", bs)
	}
	return
}
