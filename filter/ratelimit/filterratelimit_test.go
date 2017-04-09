package filterratelimit

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

func Test_ratelimit(t *testing.T) {
	require := require.New(t)
	require.NotNil(require)

	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
filter:
  - type: rate_limit
    rate: 10
    burst: 1
	`)))
	require.NoError(err)

	inchan := conf.Get(reflect.TypeOf(make(config.InChan))).
		Interface().(config.InChan)

	outchan := conf.Get(reflect.TypeOf(make(config.OutChan))).
		Interface().(config.OutChan)

	err = conf.RunFilters()
	require.NoError(err)

	start := time.Now()

	inchan <- logevent.LogEvent{}
	inchan <- logevent.LogEvent{}
	inchan <- logevent.LogEvent{}
	inchan <- logevent.LogEvent{}
	inchan <- logevent.LogEvent{}

	<-outchan
	<-outchan
	<-outchan
	<-outchan
	<-outchan

	require.True(time.Now().Sub(start) > (450 * time.Millisecond))
}

func Test_ratelimit_burst(t *testing.T) {
	require := require.New(t)
	require.NotNil(require)

	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
filter:
  - type: rate_limit
    rate: 10
    burst: 3
	`)))
	require.NoError(err)

	inchan := conf.Get(reflect.TypeOf(make(config.InChan))).
		Interface().(config.InChan)

	outchan := conf.Get(reflect.TypeOf(make(config.OutChan))).
		Interface().(config.OutChan)

	err = conf.RunFilters()
	require.NoError(err)

	time.Sleep(500 * time.Millisecond)

	start := time.Now()

	inchan <- logevent.LogEvent{}
	inchan <- logevent.LogEvent{}
	inchan <- logevent.LogEvent{}
	inchan <- logevent.LogEvent{}
	inchan <- logevent.LogEvent{}

	<-outchan
	<-outchan
	<-outchan
	<-outchan
	<-outchan

	require.True(time.Now().Sub(start) < (450 * time.Millisecond))
}

func Test_ratelimit_delay(t *testing.T) {
	require := require.New(t)
	require.NotNil(require)

	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
filter:
  - type: rate_limit
    rate: 10
    burst: 1
	`)))
	require.NoError(err)

	inchan := conf.Get(reflect.TypeOf(make(config.InChan))).
		Interface().(config.InChan)

	outchan := conf.Get(reflect.TypeOf(make(config.OutChan))).
		Interface().(config.OutChan)

	err = conf.RunFilters()
	require.NoError(err)

	time.Sleep(500 * time.Millisecond)

	start := time.Now()

	inchan <- logevent.LogEvent{}
	inchan <- logevent.LogEvent{}
	inchan <- logevent.LogEvent{}
	inchan <- logevent.LogEvent{}
	inchan <- logevent.LogEvent{}
	inchan <- logevent.LogEvent{}

	<-outchan
	<-outchan
	<-outchan
	<-outchan
	<-outchan
	<-outchan

	require.WithinDuration(start.Add(450*time.Millisecond), time.Now(), 80*time.Millisecond)
}
