package ctxutil

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCancelGroup(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	var count1 int64
	var count2 int64
	cg1 := NewCancelGroup(context.Background())
	defer func() {
		require.NoError(cg1.Close())
	}()
	cg2 := NewCancelGroup(context.Background())
	defer func() {
		require.Error(cg2.Close())
	}()

	cg1.Go(func(ctx context.Context) error {
		time.Sleep(400 * time.Millisecond)
		atomic.AddInt64(&count1, 1)
		cg1.Cancel()

		return nil
	})
	cg1.Go(func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Second):
			atomic.AddInt64(&count1, 1)
		}

		return nil
	})

	cg2.Go(func(ctx context.Context) error {
		time.Sleep(200 * time.Millisecond)
		atomic.AddInt64(&count2, 1)

		return context.DeadlineExceeded
	})
	cg2.Go(func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Second):
			atomic.AddInt64(&count2, 1)
		}

		return nil
	})

	var firstClose string
	select {
	case <-cg1.Done():
		firstClose = "cg1"
	case <-cg2.Done():
		firstClose = "cg2"
	}

	assert.EqualValues(0, atomic.LoadInt64(&count1))
	assert.EqualValues(1, atomic.LoadInt64(&count2))
	assert.EqualValues("cg2", firstClose)
}

func TestCancelGroupDone(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	var count1 int64
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	cg1 := NewCancelGroup(ctx)
	defer func() {
		require.NoError(cg1.Close())
	}()

	cg1.Go(func(ctx context.Context) error {
		time.Sleep(100 * time.Millisecond)
		atomic.AddInt64(&count1, 1)

		return nil
	})

	var err error
	var path string
	select {
	case <-ctx.Done():
		path = "ctx"
	case err = <-cg1.Done():
		path = "cg1"
	}

	assert.EqualValues(1, atomic.LoadInt64(&count1))
	assert.EqualValues("cg1", path)
	assert.NoError(err)
	assert.NoError(cg1.Wait())
}

func TestCancelGroupFork(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	const step = 300 * time.Millisecond
	var timeGo1 time.Time
	var timeGo2 time.Time
	var timeGo3 time.Time
	var timeGo4 time.Time
	var timeFork1 time.Time
	var timeFork2 time.Time
	var timeFork3 time.Time
	var timeFork4 time.Time
	ctx, cancel := context.WithTimeout(context.Background(), 3*step)
	defer cancel()
	cg := NewCancelGroup(ctx)
	defer func() {
		require.NoError(cg.Close())
	}()
	start := time.Now()

	cg.Go(func(ctx context.Context) error {
		Sleep(ctx, 1*step)
		timeGo1 = time.Now()

		return nil
	})
	cg.Go(func(ctx context.Context) error {
		Sleep(ctx, 5*step)
		timeGo2 = time.Now()

		return nil
	})
	cancelGo3 := cg.GoCancel(func(ctx context.Context) error {
		Sleep(ctx, 2*step)
		timeGo3 = time.Now()

		return nil
	})
	cancelGo3()
	cg.GoTimeout(1*step, func(ctx context.Context) error {
		Sleep(ctx, 5*step)
		timeGo4 = time.Now()

		return nil
	})

	cg.Fork(func(ctx context.Context) error {
		Sleep(ctx, 7*step)
		timeFork1 = time.Now()

		return nil
	})
	cancelFork2 := cg.ForkCancel(func(ctx context.Context) error {
		<-ctx.Done()
		timeFork2 = time.Now()

		return nil
	})
	time.AfterFunc(9*step, cancelFork2)
	cg.ForkTimeout(1*step, func(ctx context.Context) error {
		<-ctx.Done()
		timeFork3 = time.Now()

		return nil
	})
	cg.ForkTimeout(5*step, func(ctx context.Context) error {
		<-ctx.Done()
		timeFork4 = time.Now()

		return nil
	})

	assert.NoError(cg.Wait())
	const delta = step >> 1
	assert.InDelta(1*step, timeGo1.Sub(start), float64(delta))
	assert.InDelta(3*step, timeGo2.Sub(start), float64(delta))
	assert.InDelta(0*step, timeGo3.Sub(start), float64(delta))
	assert.InDelta(1*step, timeGo4.Sub(start), float64(delta))
	assert.InDelta(7*step, timeFork1.Sub(start), float64(delta))
	assert.InDelta(9*step, timeFork2.Sub(start), float64(delta))
	assert.InDelta(1*step, timeFork3.Sub(start), float64(delta))
	assert.InDelta(5*step, timeFork4.Sub(start), float64(delta))
	assert.InDelta(9*step, time.Since(start), float64(delta))
}
