package inputlorem

import (
	"bytes"
	"context"
	"errors"
	"text/template"
	"time"

	"github.com/tsaikd/gogstash/config/goglog"

	lorem "github.com/drhodes/golorem"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
	"golang.org/x/sync/errgroup"
)

// ModuleName is the name used in config file
const ModuleName = "lorem"

// InputConfig holds the configuration json fields and internal objects
type InputConfig struct {
	config.InputConfig
	Worker   int    `json:"worker,omitempty"`   // worker count to generate lorem, default: 1
	Duration string `json:"duration,omitempty"` // duration to generate lorem, set 0 to generate forever, default: 30s
	duration time.Duration

	// format event message using go text/template, defualt: {{.Sentence 1 5}}
	// support functions:
	//   `TimeFormat(layout string) string`
	//   `Word(min, max int) string`
	//   `Sentence(min, max int) string`
	//   `Paragraph(min, max int) string`
	//   `Email() string`
	//   `Host() string`
	//   `Url() string`
	Format string                 `json:"format,omitempty"`
	Fields map[string]interface{} `json:"fields,omitempty"` // event extra fields
	Empty  bool                   `json:"empty,omitempty"`  // send empty messages without any lorem text, default: false

	template *template.Template
}

type loremTemplate struct {
	tpl       *template.Template
	timestamp time.Time
}

func (t *loremTemplate) TimeFormat(layout string) string {
	return t.timestamp.Format(layout)
}

func (t *loremTemplate) Word(min, max int) string {
	return lorem.Word(min, max)
}

func (t *loremTemplate) Sentence(min, max int) string {
	return lorem.Sentence(min, max)
}

func (t *loremTemplate) Paragraph(min, max int) string {
	return lorem.Paragraph(min, max)
}

func (t *loremTemplate) Email() string {
	return lorem.Email()
}

func (t *loremTemplate) Host() string {
	return lorem.Host()
}

func (t *loremTemplate) Url() string {
	return lorem.Url()
}

// DefaultInputConfig returns an InputConfig struct with default values
func DefaultInputConfig() InputConfig {
	return InputConfig{
		InputConfig: config.InputConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		Worker:   1,
		Duration: "30s",
		duration: 30 * time.Second,
		Format:   "{{.Sentence 1 5}}",
		template: nil,
	}
}

// errors
var (
	ErrNoMessageFormat = errors.New("no message format for lorem input")
)

// InitHandler initialize the input plugin
func InitHandler(ctx context.Context, raw config.ConfigRaw) (config.TypeInputConfig, error) {
	conf := DefaultInputConfig()
	err := config.ReflectConfig(raw, &conf)
	if err != nil {
		return nil, err
	}

	if conf.Worker < 1 {
		conf.Worker = 1
	}

	if conf.duration, err = time.ParseDuration(conf.Duration); err != nil {
		return nil, err
	}

	if conf.Format != "" {
		conf.template, err = template.New("gogstash-input-lorem").Parse(conf.Format)
		if err != nil {
			return nil, err
		}
	} else if !conf.Empty {
		return nil, ErrNoMessageFormat
	}

	return &conf, nil
}

// Start wraps the actual function starting the plugin
func (t *InputConfig) Start(ctx context.Context, msgChan chan<- logevent.LogEvent) (err error) {
	eg, ctx := errgroup.WithContext(ctx)
	for i := 0; i < t.Worker; i++ {
		eg.Go(func() error {
			t.exec(ctx, msgChan)
			return nil
		})
	}

	return eg.Wait()
}

func (t *InputConfig) exec(ctx context.Context, msgChan chan<- logevent.LogEvent) {
	var stopTimer <-chan time.Time
	if t.duration > 0 {
		stopTimer = time.After(t.duration)
	} else {
		stopTimer = make(chan time.Time, 1)
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-stopTimer:
			return
		default:
			if t.Empty {
				msgChan <- logevent.LogEvent{
					Timestamp: time.Now(),
				}
			} else {
				timestamp := time.Now()
				tpl := &loremTemplate{tpl: t.template, timestamp: timestamp}
				message := bytes.Buffer{}
				err := tpl.tpl.Execute(&message, tpl)
				if err != nil {
					goglog.Logger.Errorf("input lorem template error: %v", err)
				}
				event := logevent.LogEvent{
					Timestamp: time.Now(),
					Message:   message.String(),
				}
				if t.Fields != nil {
					// copy map values
					event.Extra = make(map[string]interface{})
					for k, v := range t.Fields {
						event.Extra[k] = v
						if err != nil {
							goglog.Logger.Errorf("input lorem copy fields error: %v", err)
						}
					}
				}
				msgChan <- event
			}
		}
	}
}
