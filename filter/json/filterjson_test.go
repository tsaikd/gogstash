package filterjson

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

	conf, err := config.LoadFromJSON([]byte(`{
		"filter": [{
			"type": "json",
			"message": "message",
			"timestamp": "time",
			"timeformat": "2006-01-02T15:04:05Z"
		}]
	}`))
	require.NoError(err)

	timestamp, _ := time.Parse("2006-01-02T15:04:05Z", "2016-12-04T09:09:41.193Z")

	expectedEvent := logevent.LogEvent{
		Timestamp: timestamp,
		Message:   "Test",
		Extra: map[string]interface{}{
			"host": "Hostname",
		},
	}

	inchan := conf.Get(reflect.TypeOf(make(config.InChan))).
		Interface().(config.InChan)

	outchan := conf.Get(reflect.TypeOf(make(config.OutChan))).
		Interface().(config.OutChan)

	err = conf.RunFilters()

	inchan <- logevent.LogEvent{
		Timestamp: time.Now(),
		Message:   "{ \"message\": \"Test\", \"host\": \"Hostname\", \"time\":\"2016-12-04T09:09:41.193Z\" }",
	}

	event := <-outchan

	require.Equal(expectedEvent, event)
	require.NoError(err)
}
