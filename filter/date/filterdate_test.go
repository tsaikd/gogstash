package filterdate

import (
	"reflect"
	"strings"
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

	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
filter:
  - type: date
    format: "02/Jan/2006:15:04:05 -0700"
    source: time_local
	`)))
	require.NoError(err)

	timestamp, err := time.Parse("2006-01-02T15:04:05Z", "2017-03-20T00:42:51Z")
	require.NoError(err)

	expectedEvent := logevent.LogEvent{
		Timestamp: timestamp,
		Extra: map[string]interface{}{
			"time_local": "20/Mar/2017:00:42:51 +0000",
		},
	}

	inchan := conf.Get(reflect.TypeOf(make(config.InChan))).
		Interface().(config.InChan)

	outchan := conf.Get(reflect.TypeOf(make(config.OutChan))).
		Interface().(config.OutChan)

	err = conf.RunFilters()
	require.NoError(err)

	inchan <- logevent.LogEvent{
		Extra: map[string]interface{}{
			"time_local": "20/Mar/2017:00:42:51 +0000",
		},
	}

	event := <-outchan

	require.Equal(expectedEvent, event)
	require.NoError(err)
}
