package outputelastic

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
	elastic "gopkg.in/olivere/elastic.v5"
)

func init() {
	goglog.Logger.SetLevel(logrus.DebugLevel)
	config.RegistOutputHandler(ModuleName, InitHandler)
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
  - type: elastic
    url: ["http://127.0.0.1:9200"]
    index: "gogstash-index-test"
    document_type: "testtype"
    document_id: "%{fieldstring}"
    bulk_actions: 0
	`)))
	require.NoError(err)
	err = conf.Start(ctx)
	if err != nil {
		require.True(ErrorCreateClientFailed1.In(err))
		t.Skip("skip test output elastic module")
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

	client, err := elastic.NewClient(
		elastic.SetURL("http://127.0.0.1:9200"),
		elastic.SetSniff(false),
	)
	require.NoError(err)
	require.NotNil(client)

	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Millisecond)
	defer cancel()
	result, err := client.Get().Index("gogstash-index-test").Id("ABC").Do(ctx)
	require.NoError(err)
	require.NotNil(result)
	require.NotNil(result.Source)
	require.Equal(`{"@timestamp":"2017-04-18T19:53:01.000000002Z","fieldnumber":123,"fieldstring":"ABC","message":"output elastic test message"}`, string(*result.Source))

	_, err = client.DeleteIndex("gogstash-index-test").Do(ctx)
	require.NoError(err)
}
