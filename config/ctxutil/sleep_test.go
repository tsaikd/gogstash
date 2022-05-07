package ctxutil

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSleep(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	start := time.Now()
	require.False(Sleep(ctx, 0))
	require.WithinDuration(start, time.Now(), 100*time.Millisecond)

	require.False(Sleep(ctx, 100*time.Millisecond))
	require.WithinDuration(start.Add(100*time.Millisecond), time.Now(), 100*time.Millisecond)

	require.True(Sleep(ctx, 2000*time.Millisecond))
	require.WithinDuration(start.Add(1000*time.Millisecond), time.Now(), 100*time.Millisecond)
}
