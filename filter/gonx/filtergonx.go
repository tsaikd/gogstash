package filtergonx

import (
	"regexp"
	"strings"

	"github.com/satyrius/gonx"
	"github.com/tsaikd/KDGoLib/errutil"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
)

// ModuleName is the name used in config file
const ModuleName = "nginx"

// Errors
var (
	ErrorNoFieldInNginxFormat1 = errutil.NewFactory("no field found in nginx format: %q")
)

// FilterConfig holds the configuration json fields and internal objects
type FilterConfig struct {
	config.FilterConfig

	Format string `json:"format"` // nginx log format
	Source string `json:"source"` // source message field name

	fields []string
}

// DefaultFilterConfig returns an FilterConfig struct with default values
func DefaultFilterConfig() FilterConfig {
	return FilterConfig{
		FilterConfig: config.FilterConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		Format: `$remote_addr - $remote_user [$time_local] "$request" $status $body_bytes_sent "$http_referer" "$http_user_agent"`,
		Source: "message",
	}
}

// InitHandler initialize the filter plugin
func InitHandler(confraw *config.ConfigRaw) (retconf config.TypeFilterConfig, err error) {
	conf := DefaultFilterConfig()
	if err = config.ReflectConfig(confraw, &conf); err != nil {
		return
	}

	reFields := regexp.MustCompile(`\$([A-Za-z0-9_]+)`)
	fieldMatches := reFields.FindAllStringSubmatch(conf.Format, -1)
	if len(fieldMatches) < 1 {
		return retconf, ErrorNoFieldInNginxFormat1.New(nil, conf.Format)
	}

	conf.fields = make([]string, len(fieldMatches))
	for i, fieldInfo := range fieldMatches {
		conf.fields[i] = fieldInfo[1]
	}

	retconf = &conf
	return
}

// Event the main filter event
func (f *FilterConfig) Event(event logevent.LogEvent) logevent.LogEvent {
	if event.Extra == nil {
		event.Extra = map[string]interface{}{}
	}

	message := event.GetString(f.Source)
	reader := gonx.NewReader(strings.NewReader(message), f.Format)
	entry, err := reader.Read()
	if err != nil {
		event.AddTag("filter_nginx_invalid_message_format")
		config.Logger.Errorf("%s: %q", err, message)
		return event
	}

	for _, field := range f.fields {
		event.Extra[field], _ = entry.Field(field)
	}

	return event
}
