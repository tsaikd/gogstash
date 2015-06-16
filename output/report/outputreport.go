package outputreport

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/tsaikd/gogstash/config"
)

const (
	ModuleName = "report"
)

type OutputConfig struct {
	config.CommonConfig
	Interval   int    `json:"interval,omitempty"`
	TimeFormat string `json:"time_format,omitempty"`

	ProcessCount int `json:"-"`
}

func DefaultOutputConfig() OutputConfig {
	return OutputConfig{
		CommonConfig: config.CommonConfig{
			Type: ModuleName,
		},
		Interval:   5,
		TimeFormat: "[2/Jan/2006:15:04:05 -0700]",
	}
}

func init() {
	config.RegistOutputHandler(ModuleName, func(mapraw map[string]interface{}) (retconf config.TypeOutputConfig, err error) {
		conf := DefaultOutputConfig()
		if err = config.ReflectConfig(mapraw, &conf); err != nil {
			return
		}

		go conf.ReportLoop()

		retconf = &conf
		return
	})
}

func (self *OutputConfig) Event(event config.LogEvent) (err error) {
	self.ProcessCount++
	return
}

func (self *OutputConfig) ReportLoop() (err error) {
	for {
		if err = self.Report(); err != nil {
			log.Errorf("ReportLoop failed: %v", err)
			return
		}
		time.Sleep(time.Duration(self.Interval) * time.Second)
	}
	return
}

func (self *OutputConfig) Report() (err error) {
	if self.ProcessCount > 0 {
		fmt.Printf("%s Process %d events\n", time.Now().Format(self.TimeFormat), self.ProcessCount)
		self.ProcessCount = 0
	}
	return
}
