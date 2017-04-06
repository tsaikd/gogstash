package filtergeoip2

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/tsaikd/KDGoLib/futil"
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

	if futil.IsNotExist("GeoLite2-City.mmdb") {
		t.Log("No geoip2 database found, skip test ...")
		return
	}

	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
filter:
  - type: geoip2
    ip_field: clientip
	`)))
	require.NoError(err)

	timestamp, err := time.Parse("2006-01-02T15:04:05Z", "2016-12-04T09:09:41.193Z")
	require.NoError(err)

	expectedEvent := logevent.LogEvent{
		Timestamp: timestamp,
		Message:   `223.137.229.27 - - [20/Mar/2017:00:42:51 +0000] "GET /explore HTTP/1.1" 200 1320 "-" "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/56.0.2924.87 Safari/537.36"`,
		Extra: map[string]interface{}{
			"clientip": "223.137.229.27",
			"geoip": map[string]interface{}{
				"city": map[string]interface{}{
					"name": "Taipei",
				},
				"continent": map[string]interface{}{
					"code": "AS",
					"name": "Asia",
				},
				"country": map[string]interface{}{
					"code": "TW",
					"name": "Taiwan",
				},
				"ip":        "223.137.229.27",
				"latitude":  float64(25.0418),
				"location":  []float64{float64(121.4966), float64(25.0418)},
				"longitude": float64(121.4966),
				"timezone":  "Asia/Taipei",
			},
		},
	}

	inchan := conf.Get(reflect.TypeOf(make(config.InChan))).
		Interface().(config.InChan)

	outchan := conf.Get(reflect.TypeOf(make(config.OutChan))).
		Interface().(config.OutChan)

	err = conf.RunFilters()
	require.NoError(err)

	inchan <- logevent.LogEvent{
		Timestamp: timestamp,
		Message:   `223.137.229.27 - - [20/Mar/2017:00:42:51 +0000] "GET /explore HTTP/1.1" 200 1320 "-" "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/56.0.2924.87 Safari/537.36"`,
		Extra: map[string]interface{}{
			"clientip": "223.137.229.27",
		},
	}

	event := <-outchan

	require.Equal(expectedEvent, event)
	require.NoError(err)
}
