package ctxutil_test

import (
	"bytes"
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsaikd/gogstash/config/ctxutil"
)

func TestDisconnectContext(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	type key string

	ctx := context.Background()
	ctx = context.WithValue(ctx, key("foo"), "bar")
	ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()
	ctx1, cancel1 := context.WithCancel(ctx)
	defer cancel1()
	buffer := SyncBuffer{}

	ctx2, cancel2 := ctxutil.DisconnectContextWithTimeout(ctx1, 300*time.Millisecond)
	defer cancel2()
	require.EqualValues("bar", ctx2.Value(key("foo")))
	ctx3, cancel3 := ctxutil.DisconnectContextWithTimeout(ctx1, 700*time.Millisecond)
	defer cancel3()

	buffer.MustWriteString("1")
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		buffer.MustWriteString("3")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx1.Done()
		buffer.MustWriteString("3")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx2.Done()
		buffer.MustWriteString("2")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx3.Done()
		buffer.MustWriteString("4")
	}()

	wg.Wait()

	require.EqualValues("12334", buffer.String())
}

type SyncBuffer struct {
	mutex  sync.Mutex
	buffer bytes.Buffer
}

func (t *SyncBuffer) String() string {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	return t.buffer.String()
}

func (t *SyncBuffer) MustWriteString(text string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	_, err := t.buffer.WriteString(text)
	if err != nil {
		panic(err)
	}
}
