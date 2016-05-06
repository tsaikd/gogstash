package filteraddfield

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
	config.RegistFilterHandler(ModuleName, InitHandler)
}

func Test_main(t *testing.T) {
	require := require.New(t)
	require.NotNil(require)

	conf, err := config.LoadFromString(`{
		"filter": [{
			"type": "add_field",
			"key": "foo",
			"value": "bar"
		}]
	}`)
	require.NoError(err)

	timestamp := time.Now()
	expectedEvent := logevent.LogEvent{
		Timestamp: timestamp,
		Message:   "filter test message",
		Extra: map[string]interface{}{
			"foo": "bar",
		},
	}

	inchan := conf.Get(reflect.TypeOf(make(config.InChan))).
		Interface().(config.InChan)

	outchan := conf.Get(reflect.TypeOf(make(config.OutChan))).
		Interface().(config.OutChan)

	err = conf.RunFilters()

	inchan <- logevent.LogEvent{
		Timestamp: timestamp,
		Message:   "filter test message",
	}

	event := <-outchan

	require.Equal(expectedEvent, event)
	require.NoError(err)
}
