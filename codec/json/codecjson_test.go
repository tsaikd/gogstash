package codecjson

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
)

func init() {
	goglog.Logger.SetLevel(logrus.DebugLevel)
}

func TestDecode(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	codec, err := InitHandler(ctx, nil)
	require.NoError(err)

	msgChan := make(chan logevent.LogEvent, 1)

	ok, err := codec.Decode(ctx, []byte(`{"foo":"bar"}`), nil, msgChan)
	require.NoError(err)
	assert.True(ok)
	require.Len(msgChan, 1)
	event := <-msgChan
	assert.Equal(map[string]interface{}{"foo": "bar"}, event.Extra)
	assert.Equal("", event.Message)

	// ok will be true, as message sent
	ok, err = codec.Decode(ctx, []byte(`{"foo":"bar"dr.who}`), nil, msgChan)
	require.Error(err) // fail to decode
	assert.True(ok)
	require.Len(msgChan, 1)
	event = <-msgChan
	assert.Equal([]string{ErrorTag}, event.Tags)
	assert.Equal(`{"foo":"bar"dr.who}`, event.Message)

	// message will map to event.Message
	ok, err = codec.Decode(ctx, []byte(`{"message":"hello"}`), nil, msgChan)
	require.NoError(err)
	assert.True(ok)
	require.Len(msgChan, 1)
	event = <-msgChan
	assert.Equal("hello", event.Message)

	// merge & override extra
	ok, err = codec.Decode(ctx, []byte(`{"foo":"bar2"}`), map[string]interface{}{
		"foo": "bar",
		"one": "more thing",
	}, msgChan)
	require.NoError(err)
	assert.True(ok)
	require.Len(msgChan, 1)
	event = <-msgChan
	assert.Equal(map[string]interface{}{
		"foo": "bar2",
		"one": "more thing",
	}, event.Extra)
}
