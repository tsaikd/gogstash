package config

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

func TestGetCodec(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	// default codec, should be ok
	codec, err := GetCodecOrDefault(ctx, ConfigRaw{})
	require.NoError(err)
	require.NotNil(codec)
	require.EqualValues(DefaultCodecName, codec.GetType())

	// shorthand codec config method, should be ok
	codec, err = GetCodecOrDefault(ctx, ConfigRaw{"codec": DefaultCodecName})
	require.NoError(err)
	require.NotNil(codec)
	require.EqualValues(DefaultCodecName, codec.GetType())

	// undefined codec, should not exists
	codec, err = GetCodecOrDefault(ctx, ConfigRaw{"codec": map[string]interface{}{"type": "undefined"}})
	require.Error(err)
	require.Nil(codec)
}

func TestDefaultCodecDecode(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	codec, err := DefaultCodecInitHandler(ctx, nil)
	require.NoError(err)

	msgChan := make(chan logevent.LogEvent, 1)

	ok, err := codec.Decode(ctx, []byte("foobar"), nil, []string{}, msgChan)
	require.NoError(err)
	assert.True(ok)
	require.Len(msgChan, 1)
	event := <-msgChan
	assert.Equal("foobar", event.Message)

	// string should be ok
	ok, err = codec.Decode(ctx, "johnsmith", nil, []string{}, msgChan)
	require.NoError(err)
	assert.True(ok)
	require.Len(msgChan, 1)
	event = <-msgChan
	assert.Equal("johnsmith", event.Message)

	// ok will be true, as message sent
	ok, err = codec.Decode(ctx, 114514, nil, []string{}, msgChan)
	require.Error(err) // fail to decode
	assert.True(ok)
	require.Len(msgChan, 1)
	event = <-msgChan
	assert.Equal("", event.Message)
}
