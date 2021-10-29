package outputelasticv5

import (
	"context"
	"net/http"
	"net/http/httptest"
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
	"gopkg.in/olivere/elastic.v5"
)

const testIndexName = "gogstash-index-test"

func init() {
	goglog.Logger.SetLevel(logrus.DebugLevel)
	config.RegistOutputHandler(ModuleName, InitHandler)
}

func TestSSLCertValidation(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer ts.Close()

	ctx := context.Background()
	// check default ssl_certificate_validation is 'true'
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
output:
  - type: ` + ModuleName + `
    url: ["` + ts.URL + `"]
    index: "` + testIndexName + `"
    document_type: "testtype"
    document_id: "%{fieldstring}"
    bulk_actions: 0
	`)))
	require.NoError(err)
	require.NotNil(conf)
	_, err = InitHandler(ctx, conf.OutputRaw[0], nil)
	// expect error as certificate is not trusted by default
	require.Error(err)
	require.True(ErrorCreateClientFailed1.In(err), "%+v", err)
	require.Contains(err.Error(), "certificate signed by unknown authority", "%+v", err)

	conf, err = config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
output:
  - type: ` + ModuleName + `
    url: ["` + ts.URL + `"]
    index: "` + testIndexName + `"
    document_type: "testtype"
    document_id: "%{fieldstring}"
    bulk_actions: 0
    ssl_certificate_validation: true
	`)))
	require.NoError(err)
	require.NotNil(conf)
	_, err = InitHandler(ctx, conf.OutputRaw[0], nil)
	// again expect error as certificate is not trusted and we requested ssl_certificate_validation
	require.Error(err)
	require.True(ErrorCreateClientFailed1.In(err), "%+v", err)
	require.Contains(err.Error(), "certificate signed by unknown authority", "%+v", err)

	conf, err = config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
output:
  - type: ` + ModuleName + `
    url: ["` + ts.URL + `"]
    index: "` + testIndexName + `"
    document_type: "testtype"
    document_id: "%{fieldstring}"
    bulk_actions: 0
    ssl_certificate_validation: false
	`)))
	require.NoError(err)
	require.NotNil(conf)
	_, err = InitHandler(ctx, conf.OutputRaw[0], nil)
	// expect no error this time as ssl_certificate_validation is false
	require.NoError(err)
}

func TestResolveVars(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer ts.Close()

	err := os.Setenv("MYVAR", ts.URL)
	require.NoError(err)
	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
output:
  - type: ` + ModuleName + `
    url: ["%{MYVAR}"]
    index: "` + testIndexName + `"
    document_type: "testtype"
    document_id: "%{fieldstring}"
    bulk_actions: 0
	`)))
	require.NoError(err)
	require.NotNil(conf)
	resolvedConf, err := InitHandler(ctx, conf.OutputRaw[0], nil)
	require.NoError(err)
	outputConf := resolvedConf.(*OutputConfig)
	require.Equal(ts.URL, outputConf.resolvedURLs[0])
}

func Test_output_elastic_module(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
output:
  - type: ` + ModuleName + `
    url: ["http://127.0.0.1:9200"]
    index: "` + testIndexName + `"
    document_type: "testtype"
    document_id: "%{fieldstring}"
    bulk_actions: 0
	`)))
	require.NoError(err)
	err = conf.Start(ctx)
	if err != nil {
		require.True(ErrorCreateClientFailed1.In(err), "%+v", err)
		t.Skipf("skip test output %s module: %+v", ModuleName, err)
	}

	conf.TestInputEvent(logevent.LogEvent{
		Timestamp: time.Date(2017, 4, 18, 19, 53, 1, 2, time.UTC),
		Message:   "output elastic test message",
		Extra: map[string]interface{}{
			"fieldstring": "ABC",
			"fieldnumber": 123,
		},
	})

	if event, err2 := conf.TestGetOutputEvent(300 * time.Millisecond); assert.NoError(err2) {
		require.Equal("output elastic test message", event.Message)
	}
	time.Sleep(time.Second)

	client, err := elastic.NewClient(
		elastic.SetURL("http://127.0.0.1:9200"),
		elastic.SetSniff(false),
	)
	require.NoError(err)
	require.NotNil(client)

	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Millisecond)
	defer cancel()
	result, err := client.Get().Index(testIndexName).Id("ABC").Do(ctx)
	require.NoError(err)
	require.NotNil(result)
	require.NotNil(result.Source)
	require.JSONEq(`{"@timestamp":"2017-04-18T19:53:01.000000002Z","fieldnumber":123,"fieldstring":"ABC","message":"output elastic test message"}`, string(*result.Source))

	_, err = client.DeleteIndex(testIndexName).Do(ctx)
	require.NoError(err)
}
