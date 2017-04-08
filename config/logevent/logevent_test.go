package logevent

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_FormatWithEnv(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)

	key := "TESTENV"

	originenv := os.Getenv(key)
	defer func() {
		os.Setenv(key, originenv)
	}()

	err := os.Setenv(key, "Testing ENV")
	assert.NoError(err)

	out := FormatWithEnv("prefix %{TESTENV} suffix")
	assert.Equal("prefix Testing ENV suffix", out)
}

func Test_FormatWithCurrentTime(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)

	out := FormatWithCurrentTime("prefix %{+2006-01-02} suffix")
	nowdatestring := time.Now().Format("2006-01-02")
	assert.Equal("prefix "+nowdatestring+" suffix", out)
}

func Test_FormatWithEventTime(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)

	eventTime := time.Date(2017, time.April, 5, 17, 41, 12, 345, time.UTC)
	out := FormatWithEventTime("prefix %{+@2006-01-02} suffix", eventTime)
	assert.Equal("prefix 2017-04-05 suffix", out)
}

func Test_Format(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)

	eventTime := time.Date(2017, time.April, 5, 17, 41, 12, 345, time.UTC)
	logevent := LogEvent{
		Timestamp: eventTime,
		Message:   "Test Message",
		Extra: map[string]interface{}{
			"int":    123,
			"float":  1.23,
			"string": "Test String",
			"time":   time.Now(),
			"child": map[string]interface{}{
				"childA": "foo",
			},
		},
	}

	out := logevent.Format("%{message}")
	assert.Equal("Test Message", out)

	out = logevent.Format("%{@timestamp}")
	assert.NotEmpty(out)
	assert.NotEqual("%{@timestamp}", out)

	out = logevent.Format("%{int}")
	assert.Equal("123", out)

	out = logevent.Format("%{float}")
	assert.Equal("1.23", out)

	out = logevent.Format("%{string}")
	assert.Equal("Test String", out)

	out = logevent.Format("%{child.childA}")
	assert.Equal("foo", out)

	out = logevent.Format("time string %{+2006-01-02}")
	nowdatestring := time.Now().Format("2006-01-02")
	assert.Equal("time string "+nowdatestring, out)

	out = logevent.Format("time string %{+@2006-01-02}")
	assert.Equal("time string 2017-04-05", out)

	out = logevent.Format("%{null}")
	assert.Equal("%{null}", out)

	logevent.AddTag("tag1", "tag2", "tag3")
	assert.Len(logevent.Tags, 3)
	assert.Contains(logevent.Tags, "tag1")

	logevent.AddTag("tag1", "tag%{int}")
	assert.Len(logevent.Tags, 4)
	assert.Contains(logevent.Tags, "tag123")
}
