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
	RegistCodecHandler(DefaultCodecName, DefaultCodecInitHandler)
}

func TestGetCodec(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx := context.Background()
	// default codec, should be ok
	codec, err := GetCodec(ctx, ConfigRaw{})
	require.NoError(err)
	assert.NotNil(codec)

	// shorthand codec config method, should be ok
	codec, err = GetCodec(ctx, ConfigRaw{"codec": DefaultCodecName})
	require.NoError(err)
	assert.NotNil(codec)

	// undefined codec, should not exists
	codec, err = GetCodec(ctx, ConfigRaw{"codec": map[string]interface{}{"type": "undefined"}})
	require.Error(err)
	assert.Nil(codec)
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

	ok, err := codec.Decode(ctx, []byte("foobar"), nil, msgChan)
	require.NoError(err)
	assert.True(ok)
	require.Len(msgChan, 1)
	event := <-msgChan
	assert.Equal("foobar", event.Message)

	// string should be ok
	ok, err = codec.Decode(ctx, "johnsmith", nil, msgChan)
	require.NoError(err)
	assert.True(ok)
	require.Len(msgChan, 1)
	event = <-msgChan
	assert.Equal("johnsmith", event.Message)

	// ok will be true, as message sent
	ok, err = codec.Decode(ctx, 114514, nil, msgChan)
	require.Error(err) // fail to decode
	assert.True(ok)
	require.Len(msgChan, 1)
	event = <-msgChan
	assert.Equal("", event.Message)
}
