package outputreport

import (
	"reflect"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/require"
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
	require := require.New(t)
	require.NotNil(require)

	conf, err := config.LoadFromJSON([]byte(`{
		"output": [{
			"type": "report",
			"interval": 1
		}]
	}`))
	require.NoError(err)

	err = conf.RunOutputs()
	require.NoError(err)

	outchan := conf.Get(reflect.TypeOf(make(config.OutChan))).
		Interface().(config.OutChan)
	event := logevent.LogEvent{
		Timestamp: time.Now(),
		Message:   "outputreport test message",
	}

	outchan <- event
	outchan <- event
	time.Sleep(2 * time.Second)

	outchan <- event
	time.Sleep(2 * time.Second)

	outchan <- event
	outchan <- event
	outchan <- event
	outchan <- event
	time.Sleep(2 * time.Second)
}
