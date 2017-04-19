package dockertool

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringExist(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	stringExist := struct {
		StringExist
	}{
		StringExist: NewStringExist(),
	}

	require.False(stringExist.Exist("foo"))

	stringExist.Add("foo")
	require.True(stringExist.Exist("foo"))
	require.False(stringExist.Exist("bar"))

	stringExist.Remove("foo")
	require.False(stringExist.Exist("foo"))
	require.False(stringExist.Exist("bar"))
}
