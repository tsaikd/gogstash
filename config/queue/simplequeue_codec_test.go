package queue

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsaikd/gogstash/config/logevent"
)

const (
	waitTimeout  = 5 * time.Second
	pollInterval = 10 * time.Millisecond
)

// This test will check that we receive messages after a codec ([]byte) has processed the stream.
// At this point we no longer have access to logevent.Logevent, only the coded []byte stream.

type codecOutput struct {
	numReceived uint32          // increments for each message received (as []byte) on channel
	numSuccess  uint32          // successful deliveries
	numIncoming uint32          // increments for each message delivered to OutputEvent()
	numQueued   uint32          // number of messages sent to queue
	queue       Queue           // our queue
	ch          chan []byte     // the channel we will receive on
	ctx         context.Context // to listen for stop signal
	FailMsgID   []uint32        // received message # to fail on
}

// background runs in the background, counting number of successfully received messages
func (c *codecOutput) background() {
	for {
		select {
		case <-c.ctx.Done():
			return
		case msg := <-c.ch:
			id := atomic.AddUint32(&c.numReceived, 1)
			var failed bool
			for x := range c.FailMsgID {
				if c.FailMsgID[x] == id {
					failed = true
					break
				}
			}
			if failed {
				err := c.queue.Queue(c.ctx, msg)
				if err != nil {
					panic(err)
				}
				atomic.AddUint32(&c.numQueued, 1)
			} else {
				atomic.AddUint32(&c.numSuccess, 1)
			}
		}
	}
}

func (c *codecOutput) GetType() string {
	return "codecOutput"
}

func (c *codecOutput) OutputEvent(ctx context.Context, event logevent.LogEvent) error {
	atomic.AddUint32(&c.numIncoming, 1)
	msg, err := event.MarshalJSON()
	if err != nil {
		return err
	}
	c.ch <- msg
	return nil
}

func TestSimpleQueueCodec(t *testing.T) {
	ctx := t.Context()
	o := &codecOutput{
		queue:     nil,
		ch:        make(chan []byte, 10),
		ctx:       ctx,
		FailMsgID: []uint32{2},
	}
	control := newControlCounter()
	const retryInterval = 1
	start := time.Now()
	q := NewSimpleQueue(ctx, control, o, o.ch, 5, retryInterval)
	o.queue = q
	// now send a message that should go through
	event := logevent.LogEvent{Timestamp: time.Now()}
	go o.background()
	require.NoError(t, q.Output(ctx, event))
	require.Eventually(t, func() bool {
		return atomic.LoadUint32(&o.numSuccess) == 1
	}, waitTimeout, pollInterval, "sent one message, did not get through")

	// now send a message that should be queued
	require.NoError(t, q.Output(ctx, event))
	require.Eventually(t, func() bool {
		return atomic.LoadUint32(&o.numQueued) == 1
	}, waitTimeout, pollInterval, "second message was not queued")

	// The retry fires on the queue's own ticker, created by the goroutine
	// NewSimpleQueue starts, so it is not aligned with the test's sends and
	// redelivery may follow queueing at once. It cannot fire before
	// start+retryInterval, though: counters read before that instant must still
	// show the message waiting, and a slower run skips the check rather than
	// racing the ticker.
	received := atomic.LoadUint32(&o.numReceived)
	success := atomic.LoadUint32(&o.numSuccess)
	if time.Since(start) < retryInterval*time.Second {
		assert.EqualValues(t, 2, received, "queued message was redelivered before the first retry tick")
		assert.EqualValues(t, 1, success, "queued message was redelivered before the first retry tick")
	}

	require.Eventually(t, func() bool {
		return atomic.LoadUint32(&o.numSuccess) == 2
	}, waitTimeout, pollInterval, "queued message was not redelivered")
	assert.EqualValues(t, 3, atomic.LoadUint32(&o.numReceived))
	assert.EqualValues(t, 1, atomic.LoadUint32(&o.numQueued))
}
