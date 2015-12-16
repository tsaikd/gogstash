package outputreport

import (
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"

	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
)

const (
	ModuleName = "report"
)

type OutputConfig struct {
	config.OutputConfig
	Interval     int    `json:"interval,omitempty"`
	TimeFormat   string `json:"time_format,omitempty"`
	ReportPrefix string `json:"report_prefix,omitempty"`

	ProcessCount int `json:"-"`
}

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

func InitHandler(confraw *config.ConfigRaw, logger *logrus.Logger) (retconf config.TypeOutputConfig, err error) {
	conf := DefaultOutputConfig()
	if err = config.ReflectConfig(confraw, &conf); err != nil {
		return
	}

	go conf.ReportLoop(logger)

	retconf = &conf
	return
}

func (t *OutputConfig) Event(event logevent.LogEvent) (err error) {
	t.ProcessCount++
	return
}

func (t *OutputConfig) ReportLoop(logger *logrus.Logger) (err error) {
	for {
		if err = t.Report(logger); err != nil {
			logger.Errorln(fmt.Sprintf("ReportLoop failed: %v", err))
			return
		}
		time.Sleep(time.Duration(t.Interval) * time.Second)
	}
}

func (t *OutputConfig) Report(logger *logrus.Logger) (err error) {
	if t.ProcessCount > 0 {
		fmt.Printf(
			"%s %sProcess %d events\n",
			time.Now().Format(t.TimeFormat),
			t.ReportPrefix,
			t.ProcessCount,
		)
		t.ProcessCount = 0
	}
	return
}
