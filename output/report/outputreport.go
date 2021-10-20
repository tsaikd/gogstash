package outputreport

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
)

// ModuleName is the name used in config file
const ModuleName = "report"

// OutputConfig holds the configuration json fields and internal objects
type OutputConfig struct {
	config.OutputConfig
	Interval     int    `json:"interval,omitempty"`
	TimeFormat   string `json:"time_format,omitempty"`
	ReportPrefix string `json:"report_prefix,omitempty"`

	ProcessCount int64 `json:"-"`
}

// DefaultOutputConfig returns an OutputConfig struct with default values
func DefaultOutputConfig() OutputConfig {
	return OutputConfig{
		OutputConfig: config.OutputConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		Interval:   5,
		TimeFormat: "[2/Jan/2006:15:04:05 -0700]",
	}
}

// InitHandler initialize the output plugin
func InitHandler(ctx context.Context, raw config.ConfigRaw) (config.TypeOutputConfig, error) {
	conf := DefaultOutputConfig()
	if err := config.ReflectConfig(raw, &conf); err != nil {
		return nil, err
	}

	go conf.reportLoop(ctx)

	return &conf, nil
}

// Output event
func (t *OutputConfig) Output(ctx context.Context, event logevent.LogEvent) (err error) {
	atomic.AddInt64(&t.ProcessCount, 1)
	return
}

func (t *OutputConfig) reportLoop(ctx context.Context) {
	startChan := make(chan bool, 1) // startup tick
	ticker := time.NewTicker(time.Duration(t.Interval) * time.Second)
	defer ticker.Stop()

	startChan <- true

	for {
		select {
		case <-ctx.Done():
			return
		case <-startChan:
			t.report()
		case <-ticker.C:
			t.report()
		}
	}
}

func (t *OutputConfig) report() {
	count := atomic.LoadInt64(&t.ProcessCount)
	if count < 1 {
		return
	}
	fmt.Printf(
		"%s %sProcess %d events\n",
		time.Now().Format(t.TimeFormat),
		t.ReportPrefix,
		count,
	)
	atomic.StoreInt64(&t.ProcessCount, 0)
}
