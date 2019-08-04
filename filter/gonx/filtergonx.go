package filtergonx

import (
	"context"
	"regexp"
	"strings"

	"github.com/satyrius/gonx"
	"github.com/tsaikd/KDGoLib/errutil"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
)

// ModuleName is the name used in config file
const ModuleName = "gonx"

// ErrorTag tag added to event when process module failed
const ErrorTag = "gogstash_filter_gonx_error"

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
	parser *gonx.Parser
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
func InitHandler(ctx context.Context, raw *config.ConfigRaw) (config.TypeFilterConfig, error) {
	conf := DefaultFilterConfig()
	err := config.ReflectConfig(raw, &conf)
	if err != nil {
		return nil, err
	}

	reFields := regexp.MustCompile(`\$([A-Za-z0-9_]+)`)
	fieldMatches := reFields.FindAllStringSubmatch(conf.Format, -1)
	if len(fieldMatches) < 1 {
		return nil, ErrorNoFieldInNginxFormat1.New(nil, conf.Format)
	}

	conf.fields = make([]string, len(fieldMatches))
	for i, fieldInfo := range fieldMatches {
		conf.fields[i] = fieldInfo[1]
	}

	conf.parser = gonx.NewParser(conf.Format)

	return &conf, nil
}

// Event the main filter event
func (f *FilterConfig) Event(ctx context.Context, event logevent.LogEvent) (logevent.LogEvent, bool) {
	message := event.GetString(f.Source)
	reader := gonx.NewParserReader(strings.NewReader(message), f.parser)
	entry, err := reader.Read()
	if err != nil {
		event.AddTag(ErrorTag)
		goglog.Logger.Errorf("%s: %q", err, message)
		return event, false
	}

	for _, field := range f.fields {
		s, _ := entry.Field(field)
		event.SetValue(field, s)
	}

	return event, true
}
