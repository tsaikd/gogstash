package logevent

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsaikd/KDGoLib/jsonex"
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
}

func Test_Tags(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)

	logevent := LogEvent{
		Extra: map[string]interface{}{
			"int": 123,
		},
	}

	logevent.ParseTags(interface{}([]interface{}{"foo", "bar"}))
	assert.Len(logevent.Tags, 2)
	assert.Equal([]string{"foo", "bar"}, logevent.Tags)

	logevent.AddTag("tag1", "tag2", "tag3")
	assert.Len(logevent.Tags, 5)
	assert.Contains(logevent.Tags, "tag1")

	logevent.AddTag("tag1", "tag%{int}")
	assert.Len(logevent.Tags, 6)
	assert.Contains(logevent.Tags, "tag123")

	logevent.RemoveTag("foo", "bar")
	assert.Len(logevent.Tags, 4)
	logevent.RemoveTag("notfoundtag")
	assert.Len(logevent.Tags, 4)
	assert.Equal([]string{"tag1", "tag2", "tag3", "tag123"}, logevent.Tags)
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
					"response_code":  200,
					"remote_ip_list": []string{"1.1.1.1"},
				},
			},
		},
	}

	responseCode, ok := event.GetValue("nginx.access.response_code")
	assert.True(ok)
	assert.Equal(200, responseCode)

	ip, ok := event.GetValue("nginx.access.remote_ip_list[0]")
	assert.True(ok)
	assert.Equal("1.1.1.1", ip)
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

func Test_Remove(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	eventTime := time.Date(2017, time.April, 5, 17, 41, 12, 345, time.UTC)
	event := LogEvent{
		Timestamp: eventTime,
		Message:   "Test Message",
		Extra: map[string]interface{}{
			"foo": "bar",
			"map": map[string]interface{}{
				"foo2": "bar2",
			},
		},
	}

	assert.True(event.Remove("foo"))
	require.Equal(LogEvent{
		Timestamp: eventTime,
		Message:   "Test Message",
		Extra: map[string]interface{}{
			"map": map[string]interface{}{
				"foo2": "bar2",
			},
		},
	}, event)

	event = LogEvent{
		Timestamp: eventTime,
		Message:   "Test Message",
		Extra: map[string]interface{}{
			"foo": "bar",
			"map": map[string]interface{}{
				"foo2": "bar2",
			},
		},
	}

	assert.True(event.Remove("map.foo2"))
	require.Equal(LogEvent{
		Timestamp: eventTime,
		Message:   "Test Message",
		Extra: map[string]interface{}{
			"foo": "bar",
			"map": map[string]interface{}{},
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
	assert.NotContains(string(d), `"tags":[`)

	event.AddTag("test_tag")

	d, err = json.Marshal(event)
	assert.NoError(err)
	assert.Contains(string(d), `"tags":["test_tag"]`)

	d, err = event.MarshalIndent()
	assert.NoError(err)
	assert.Contains(string(d), "\n\t\"")
}

var benchEvent = LogEvent{
	Timestamp: time.Now(),
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

func Benchmark_Marshal_JSONEx(b *testing.B) {
	jsonMap := benchEvent.getJSONMap()
	d, err := jsonex.Marshal(jsonMap)
	if err != nil {
		b.FailNow()
	}
	b.SetBytes(int64(len(d)))
	for n := 0; n < b.N; n++ {
		//nolint: errcheck
		jsonex.Marshal(jsonMap)
	}
}

func Benchmark_Marshal_JSONIter(b *testing.B) {
	jsonMap := benchEvent.getJSONMap()
	d, err := jsoniter.Marshal(jsonMap)
	if err != nil {
		b.FailNow()
	}
	b.SetBytes(int64(len(d)))
	for n := 0; n < b.N; n++ {
		//nolint: errcheck
		jsoniter.Marshal(jsonMap)
	}
}

func Benchmark_Marshal_StdJSON(b *testing.B) {
	jsonMap := benchEvent.getJSONMap()
	d, err := json.Marshal(jsonMap)
	if err != nil {
		b.FailNow()
	}
	b.SetBytes(int64(len(d)))
	for n := 0; n < b.N; n++ {
		//nolint: errcheck
		json.Marshal(jsonMap)
	}
}

func TestChangeMessage(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)

	newMessage := "changed"

	event := LogEvent{
		Message: "Test Message",
	}

	event.SetValue("message", newMessage)

	assert.Equal(newMessage, event.Message)
	assert.Equal(newMessage, event.Get("message"))
	assert.Equal(newMessage, event.GetString("message"))

}
