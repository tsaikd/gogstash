package config

// TypeCommonConfig is interface of basic config
type TypeCommonConfig interface {
	GetType() string
}

// CommonConfig is basic config struct
type CommonConfig struct {
	Type     string `json:"type"`
	Disabled bool   `json:"disabled" yaml:"disabled"` // if set the input/output/filter will be disabled
}

// GetType return module type of config
func (t CommonConfig) GetType() string {
	return t.Type
}

// ConfigRaw is general config struct
type ConfigRaw map[string]any
