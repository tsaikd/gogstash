package config

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsaikd/gogstash/config/goglog"
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

	// default codec, should be ok
	codec, err := GetCodec(context.TODO(), ConfigRaw{})
	require.NoError(err)
	assert.NotNil(codec)

	// shorthand codec config method, should be ok
	codec, err = GetCodec(context.TODO(), ConfigRaw{"codec": DefaultCodecName})
	require.NoError(err)
	assert.NotNil(codec)

	// undefined codec, should not exists
	codec, err = GetCodec(context.TODO(), ConfigRaw{"codec": map[string]interface{}{"type": "undefined"}})
	require.Error(err)
	assert.Nil(codec)
}
