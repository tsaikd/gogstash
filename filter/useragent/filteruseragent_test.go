package filteruseragent

import (
	"context"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
)

var (
	fileName = "regexes.yaml"
	fileData = []byte(strings.TrimSpace(`
user_agent_parsers:
  # Chrome/Chromium/major_version.minor_version
  - regex: '(Chromium|Chrome)/(\d+)\.(\d+)(?:\.(\d+)|)(?:\.(\d+)|)'

os_parsers:
  - regex: '(Windows NT 6\.1)'
    os_replacement: 'Windows'
    os_v1_replacement: '7'
	`))
)

func init() {
	goglog.Logger.SetLevel(logrus.DebugLevel)
	config.RegistFilterHandler(ModuleName, InitHandler)

	err := ioutil.WriteFile(fileName, fileData, 0644)
	if err != nil {
		panic(err)
	}
}

func Test_filter_useragent_module_default(t *testing.T) {
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: useragent
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))
}

func Test_filter_useragent_module_parse(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
filter:
  - type: useragent
    source: agent
    target: user_agent
    regexes: "./regexes.yaml"
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	uagent := "Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.98 Safari/537.36"
	expectedEvent := logevent.LogEvent{
		Extra: map[string]interface{}{
			"agent": uagent,
			"user_agent": map[string]interface{}{
				"device":   "Other",
				"major":    "71",
				"minor":    "0",
				"name":     "Chrome",
				"os":       "Windows",
				"os_name":  "Windows",
				"os_major": "7",
				"patch":    "3578",
			},
		},
	}

	conf.TestInputEvent(logevent.LogEvent{
		Extra: map[string]interface{}{
			"agent": uagent,
		},
	})

	if event, err := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err) {
		require.Equal(expectedEvent, event)
	}

	err = os.Remove(fileName)
	require.NoError(err)
}
