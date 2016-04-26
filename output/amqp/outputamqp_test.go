package outputamqp

import (
	"reflect"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/tsaikd/KDGoLib/errutil"
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

func Test_WithoutAMQPServer(t *testing.T) {
	require := require.New(t)
	require.NotNil(require)

	conf, err := config.LoadFromString(`{
		"output": [{
			"type": "amqp",
			"urls": ["amqp://guest:guest@localhost:5566/"],
			"exchange": "amq.topic",
			"exchange_type": "topic"
		}]
	}`)
	require.NoError(err)

	err = conf.RunOutputs()
	require.Error(err)
	require.True(config.ErrorRunOutput1.Match(err))
	require.True(ErrorNoValidConn.In(err))
	require.Implements((*errutil.ErrorObject)(nil), err)
	require.True(ErrorNoValidConn.Match(err.(errutil.ErrorObject).Parent()))
}

func Test_main(t *testing.T) {
	require := require.New(t)
	require.NotNil(require)

	conf, err := config.LoadFromString(`{
		"output": [{
			"type": "amqp",
			"urls": ["amqp://guest:guest@localhost:5672/"],
			"exchange": "amq.topic",
			"exchange_type": "topic"
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
