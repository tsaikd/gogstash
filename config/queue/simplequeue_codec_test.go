package queue

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tsaikd/gogstash/config/logevent"
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
	q := NewSimpleQueue(ctx, control, o, o.ch, 5, 1)
	o.queue = q
	// now send a message that should go through
	event := logevent.LogEvent{Timestamp: time.Now()}
	go o.background()
	err := q.Output(ctx, event)
	if err == nil {
		time.Sleep(time.Second)
		if atomic.LoadUint32(&o.numSuccess) != 1 {
			t.Error("Sent one messages, did not get through")
		}
	} else {
		t.Error(err)
	}
	// now send a message that should be queued
	err = q.Output(ctx, event)
	if err != nil {
		t.Error(err)
		return
	}
	time.Sleep(900 * time.Millisecond)
	if atomic.LoadUint32(&o.numQueued) != 1 || atomic.LoadUint32(&o.numReceived) != 2 || atomic.LoadUint32(&o.numSuccess) != 1 {
		t.Errorf("Queueing does not work, %v", *o)
		return
	}
	time.Sleep(900 * time.Millisecond)
	if atomic.LoadUint32(&o.numSuccess) != 2 {
		t.Error("Two messages not delivered")
	}
}
