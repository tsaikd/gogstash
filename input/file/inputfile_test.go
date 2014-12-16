package inputfile

import (
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/tsaikd/gogstash/config"
)

func Test_main(t *testing.T) {
	var (
		assert   = assert.New(t)
		err      error
		conftest config.Config
	)

	log.SetLevel(log.DebugLevel)

	conftest, err = config.LoadConfig("config_test.json")
	assert.NoError(err)

	inputs := conftest.Input()
	assert.Len(inputs, 1)
	if len(inputs) > 0 {
		input := inputs[0].(*InputConfig)
		assert.IsType(&InputConfig{}, input)
		assert.Equal("file", input.Type())
		assert.Equal("/tmp/log/syslog", input.Path)

		eventChan := make(chan config.LogEvent, 10)
		go func() {
			for {
				<-eventChan
			}
		}()
		err = input.Event(eventChan)
		assert.NoError(err)

		waitsec := 10
		log.Debugf("Wait for %d seconds", waitsec)
		time.Sleep(time.Duration(waitsec) * time.Second)
	}
}
