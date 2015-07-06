package outputreport

import (
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/tsaikd/KDGoLib/logrusutil"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
)

func Test_main(t *testing.T) {
	assert := assert.New(t)

	logger := logrusutil.DefaultConsoleLogger
	logger.Level = logrus.DebugLevel
	config.RegistOutputHandler(ModuleName, InitHandler)

	conf, err := config.LoadFromString(`{
		"output": [{
			"type": "report",
			"interval": 1
		}]
	}`)
	assert.NoError(err)
	conf.Map(logger)

	evchan := make(chan logevent.LogEvent, 10)
	conf.Map(evchan)

	err = conf.RunOutputs(evchan, logger)
	assert.NoError(err)

	event := logevent.LogEvent{
		Timestamp: time.Now(),
		Message:   "outputreport test message",
	}

	evchan <- event
	evchan <- event
	time.Sleep(2 * time.Second)

	evchan <- event
	time.Sleep(2 * time.Second)

	evchan <- event
	evchan <- event
	evchan <- event
	evchan <- event
	time.Sleep(2 * time.Second)
}
