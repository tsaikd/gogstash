package ctxutil

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBroadcaster(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cg := NewCancelGroup(ctx)
	b := NewBroadcaster()

	var broadcastCount int32
	listenReady := make(chan struct{}, 1)
	for i := 0; i < 5; i++ {
		cg.Go(func(ctx context.Context) error {
			listenReady <- struct{}{}
			select {
			case <-ctx.Done():
				t.Fatal("wait for broadcast signal timeout")
			case <-b.Channel():
				atomic.AddInt32(&broadcastCount, 1)
			}

			return nil
		})
	}
	for i := 0; i < 5; i++ {
		<-listenReady
	}

	require.False(Sleep(ctx, 500*time.Millisecond))
	require.NoError(b.Signal(ctx))
	require.False(Sleep(ctx, 500*time.Millisecond))
	require.EqualValues(1, atomic.LoadInt32(&broadcastCount))
	require.NoError(b.Broadcast(ctx))

	require.NoError(cg.Wait())
	require.EqualValues(5, atomic.LoadInt32(&broadcastCount))
}
