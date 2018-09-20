package inputlorem

import (
	"context"
	"time"

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
	Empty    bool `json:"empty,omitempty"` // send empty messages without any lorem text, default: false
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
	}
}

// InitHandler initialize the input plugin
func InitHandler(ctx context.Context, raw *config.ConfigRaw) (config.TypeInputConfig, error) {
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
				msgChan <- logevent.LogEvent{
					Timestamp: time.Now(),
					Message:   lorem.Sentence(1, 5),
					Extra: map[string]interface{}{
						"loremurl": lorem.Url(),
					},
				}
			}
		}
	}
}
