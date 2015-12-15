package outputreport

import (
	"reflect"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
)

var (
	logger = config.Logger
)

func init() {
	logger.Level = logrus.DebugLevel
	config.RegistOutputHandler(ModuleName, InitHandler)
}

func Test_main(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)

	conf, err := config.LoadFromString(`{
		"output": [{
			"type": "report",
			"interval": 1
		}]
	}`)
	assert.NoError(err)

	err = conf.RunOutputs()
	assert.NoError(err)

	evchan := conf.Get(reflect.TypeOf(make(chan logevent.LogEvent))).
		Interface().(chan logevent.LogEvent)
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
