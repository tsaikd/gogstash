package outputreport

import (
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
	Interval   int    `json:"interval,omitempty"`
	TimeFormat string `json:"time_format,omitempty"`

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

func (self *OutputConfig) Event(event logevent.LogEvent) (err error) {
	self.ProcessCount++
	return
}

func (self *OutputConfig) ReportLoop(logger *logrus.Logger) (err error) {
	for {
		if err = self.Report(logger); err != nil {
			logger.Errorf("ReportLoop failed: %v", err)
			return
		}
		time.Sleep(time.Duration(self.Interval) * time.Second)
	}
	return
}

func (self *OutputConfig) Report(logger *logrus.Logger) (err error) {
	if self.ProcessCount > 0 {
		logger.Infof("%s Process %d events\n", time.Now().Format(self.TimeFormat), self.ProcessCount)
		self.ProcessCount = 0
	}
	return
}
