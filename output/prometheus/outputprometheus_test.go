package outputprometheus

import (
	"io/ioutil"
	"net/http"
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
	config.RegistOutputHandler(ModuleName, InitHandler)
}

func Test_main(t *testing.T) {
	require := require.New(t)
	require.NotNil(require)

	conf, err := config.LoadFromJSON([]byte(`{
		"output": [{
			"type": "prometheus"
		}]
	}`))
	require.NoError(err)

	err = conf.RunOutputs()
	time.Sleep(1 * time.Second)

	require.NoError(err)

	outchan := conf.Get(reflect.TypeOf(make(config.OutChan))).
		Interface().(config.OutChan)
	event := logevent.LogEvent{
		Timestamp: time.Now(),
		Message:   "outputreport test message",
	}

	// sending 1st event
	outchan <- event
	value, err := getMetric()
	if err != nil {
		t.Fail()
	}
	require.Equal("processed_messages_total 1", value)

	// sending second event
	outchan <- event
	value, err = getMetric()
	if err != nil {
		t.Fail()
	}
	require.Equal("processed_messages_total 2", value)
}

func getMetric() (string, error) {
	resp, err := http.Get("http://localhost:8080/metrics")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(body), "\n")
	return lines[len(lines)-2], nil
}
