package outputsentry

import (
	"context"
	"log"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
)

const ModuleName = "sentry"

// OutputConfig holds the configuration json fields and internal objects
type OutputConfig struct {
	config.OutputConfig
	DSN string `json:"dsn"`
}

// DefaultOutputConfig returns an OutputConfig struct with default values
func DefaultOutputConfig() OutputConfig {
	return OutputConfig{
		OutputConfig: config.OutputConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		DSN: "",
	}
}

func InitHandler(
	ctx context.Context,
	raw config.ConfigRaw,
	control config.Control,
) (config.TypeOutputConfig, error) {
	conf := DefaultOutputConfig()
	if err := config.ReflectConfig(raw, &conf); err != nil {
		return nil, err
	}
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              conf.DSN,
		TracesSampleRate: 1.0,
	})
	if err != nil {
		log.Fatalf("Failed to initialize Sentry: %v", err)
		return nil, err
	}

	return &conf, nil
}

func (o *OutputConfig) Output(ctx context.Context, event logevent.LogEvent) error {
	sentry.WithScope(func(scope *sentry.Scope) {
		for key, value := range event.Extra {
			scope.SetExtra(key, value)
		}
		sentry.CaptureMessage(event.Message)
	})
	sentry.Flush(2 * time.Second)
	return nil
}
