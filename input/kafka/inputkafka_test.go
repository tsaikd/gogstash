package inputkafka

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
)

func init() {
	goglog.Logger.SetLevel(logrus.DebugLevel)
	config.RegistInputHandler(ModuleName, InitHandler)
}

func initClient() (sarama.SyncProducer, error) {
	// initialize kafka client
	saconf := sarama.NewConfig()
	saconf.Version = sarama.V0_10_2_0
	saconf.Producer.RequiredAcks = sarama.WaitForAll          // wait for both leader and follower checked
	saconf.Producer.Partitioner = sarama.NewRandomPartitioner // select one partition
	saconf.Producer.Return.Successes = true
	return sarama.NewSyncProducer([]string{"127.0.0.1:9092"}, saconf)
}

func Test_input_kafka_module_batch(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	client, err := initClient()
	if err != nil {
		t.Skipf("skip test output %s module: %+v", ModuleName, err)
	}
	require.NotNil(client)

	for i := 0; i < 10; i++ {
		msg := &sarama.ProducerMessage{
			Topic: "testTopic",
			Value: sarama.StringEncoder(fmt.Sprintf("this is a test log (%d)", i)),
		}
		partition, offset, err := client.SendMessage(msg)
		goglog.Logger.Infof("partition : %v, offset : %v, err : %v", partition, offset, err)
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	conf, err := config.LoadFromYAML([]byte(strings.TrimSpace(`
debugch: true
input:
  - type: kafka
    version: 0.10.2.0
    brokers:
      - 127.0.0.1:9092
    topics:
      - testTopic
    group: log_center
    offset_oldest: true
    assignor: roundrobin
	`)))
	require.NoError(err)
	require.NoError(conf.Start(ctx))

	time.Sleep(100 * time.Millisecond)
	for i := 0; i < 10; i++ {
		if event, err := conf.TestGetOutputEvent(100 * time.Millisecond); assert.NoError(err) {
			goglog.Logger.Infof("%#v / %v", event, err)
			require.NotNil(event.Timestamp.UnixNano())
			require.Contains(event.Message, "this is a test log")
		}
	}
}
