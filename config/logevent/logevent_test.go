package logevent

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func Test_GetValue(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)

	eventTime := time.Date(2017, time.April, 5, 17, 41, 12, 345, time.UTC)
	event := LogEvent{
		Timestamp: eventTime,
		Message:   "Test Message",
		Extra: map[string]interface{}{
			"nginx": map[string]interface{}{
				"access": map[string]interface{}{
					"response_code": 200,
				},
			},
		},
	}

	responseCode, ok := event.GetValue("nginx.access.response_code")
	assert.True(ok)
	assert.Equal(200, responseCode)
}

func Test_SetValue(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	eventTime := time.Date(2017, time.April, 5, 17, 41, 12, 345, time.UTC)
	event := LogEvent{
		Timestamp: eventTime,
		Message:   "Test Message",
	}

	assert.True(event.SetValue("nginx.access.remote_ip", "1.1.1.1"))

	require.Equal(LogEvent{
		Timestamp: eventTime,
		Message:   "Test Message",
		Extra: map[string]interface{}{
			"nginx": map[string]interface{}{
				"access": map[string]interface{}{
					"remote_ip": "1.1.1.1",
				},
			},
		},
	}, event)
}

func Test_MarshalJSON(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)

	eventTime := time.Date(2017, time.April, 5, 17, 41, 12, 345, time.UTC)
	event := LogEvent{
		Timestamp: eventTime,
		Message:   "Test Message",
	}

	d, err := json.Marshal(event)
	assert.NoError(err)
	assert.Equal(`{"@timestamp":"2017-04-05T17:41:12.000000345Z","message":"Test Message"}`, string(d))

	event.AddTag("test_tag")

	d, err = json.Marshal(event)
	assert.NoError(err)
	assert.Equal(`{"@timestamp":"2017-04-05T17:41:12.000000345Z","message":"Test Message","tags":["test_tag"]}`, string(d))

	event.Extra = map[string]interface{}{
		"tags": nil,
	}

	d, err = json.Marshal(event)
	assert.NoError(err)
	assert.Equal(`{"@timestamp":"2017-04-05T17:41:12.000000345Z","message":"Test Message","tags":["test_tag"]}`, string(d))

	event.Extra = map[string]interface{}{
		"tags": []interface{}{"original_tag"},
	}

	d, err = json.Marshal(event)
	assert.NoError(err)
	assert.Equal(`{"@timestamp":"2017-04-05T17:41:12.000000345Z","message":"Test Message","tags":["original_tag","test_tag"]}`, string(d))

	// failed to treated `tags` as a string array
	event.Extra = map[string]interface{}{
		"tags": []interface{}{"original_tag", 1},
	}

	d, err = json.Marshal(event)
	assert.NoError(err)
	assert.Equal(`{"@timestamp":"2017-04-05T17:41:12.000000345Z","message":"Test Message","tags":["original_tag",1]}`, string(d))
}
