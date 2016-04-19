package outputelastic

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

	conf, err := config.LoadFromString(`{
		"output": [{
			"type": "elastic",
			"url": "http://127.0.0.1:9200",
			"index": "testindex",
			"document_type": "testtype",
			"document_id": "%{fieldstring}"
		}]
	}`)
	require.NoError(err)

	err = conf.RunOutputs()
	require.NoError(err)

	evchan := conf.Get(reflect.TypeOf(make(chan logevent.LogEvent))).
		Interface().(chan logevent.LogEvent)
	evchan <- logevent.LogEvent{
		Timestamp: time.Now(),
		Message:   "outputstdout test message",
		Extra: map[string]interface{}{
			"fieldstring": "ABC",
			"fieldnumber": 123,
		},
	}

	waitsec := 1
	logger.Infof("Wait for %d seconds", waitsec)
	time.Sleep(time.Duration(waitsec) * time.Second)
}
