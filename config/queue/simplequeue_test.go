package queue

import (
	"context"
	"github.com/tsaikd/gogstash/config/logevent"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// controlCounter is used to count control messages, so we can check it is as expected
type controlCounter struct {
	lock sync.Mutex // lock for all data

	numPause  uint32 //  number of pauses
	numResume uint32 // number of resumes

	pause  chan struct{}
	resume chan struct{}
}

func newControlCounter() *controlCounter {
	return &controlCounter{resume: make(chan struct{}), pause: make(chan struct{})}
}

func (c *controlCounter) RequestPause(ctx context.Context) error {
	c.lock.Lock()
	c.numPause++
	old := c.resume
	c.resume = make(chan struct{})
	c.lock.Unlock()
	close(old)
	return nil
}

func (c *controlCounter) RequestResume(ctx context.Context) error {
	c.lock.Lock()
	c.numResume++
	old := c.pause
	c.pause = make(chan struct{})
	c.lock.Unlock()
	close(old)
	return nil
}

func (c *controlCounter) PauseSignal() <-chan struct{} {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.pause
}

func (c *controlCounter) ResumeSignal() <-chan struct{} {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.resume
}

type sampleOutput struct {
	numReceived uint32        // increments for each message received
	numSent     uint32        // increments for each message delivered
	queue       Queue         // our queue
	target      uint32        // number of messages that we want to receive
	doneCh      chan struct{} // closed when target # of messages has been received
	FailMsgId   []uint32      // received message # to fail on
}

func (s *sampleOutput) GetType() string {
	return "sampleOutput"
}

const numMessages = 140

func (s *sampleOutput) OutputEvent(ctx context.Context, event logevent.LogEvent) (err error) {
	id := atomic.AddUint32(&s.numReceived, 1)
	var failed bool
	for x := range s.FailMsgId {
		if s.FailMsgId[x] == id {
			failed = true
			break
		}
	}
	if failed {
		_ = s.queue.Queue(ctx, event)
	} else {
		_ = s.queue.Resume(ctx)
		if atomic.AddUint32(&s.numSent, 1) >= s.target {
			close(s.doneCh)
		}
	}
	return nil
}

// check that we receive
func TestNewSimpleQueue1(t *testing.T) {
	control := newControlCounter()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	o := &sampleOutput{doneCh: make(chan struct{}), target: numMessages, FailMsgId: []uint32{1, 2, 100}}
	q := NewSimpleQueue(ctx, control, o, nil, numMessages, 1)
	o.queue = q
	// send four messages
	go func() {
		event := logevent.LogEvent{}
		for x := 0; x < numMessages; x++ {
			event.Message = strconv.Itoa(x)
			_ = q.Output(ctx, event)
		}
	}()
	// wait and see
	select {
	case <-ctx.Done():
		t.Errorf("test timed out after %v events", atomic.LoadUint32(&o.numReceived))
		return
	case <-o.doneCh:
		break
	}
	// check if we got the expected result
	numFails := uint32(len(o.FailMsgId))
	if o.numSent != numMessages {
		t.Errorf("Received %v messages, expected %v messages", o.numSent, numMessages)
	}
	if o.numReceived != uint32(numMessages+numFails) {
		t.Errorf("Sent %v messages, received %v messages", numMessages, o.numReceived)
	}
	if control.numPause == 0 {
		t.Error("control got no pauses, expected at least 1")
	}
	if control.numResume == 0 {
		t.Error("control got no resumes, expected at least 1")
	}
}
