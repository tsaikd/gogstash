package queue

import (
	"container/list"
	"context"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
	"sync/atomic"
	"time"
)

// simpleQueue is an implementation of QueueReceiver with an easy retry-logic
type simpleQueue struct {
	RetryInterval uint `json:"retry_interval" yaml:"retry_interval"` // seconds before a new retry in case on error
	MaxQueueSize  int  `json:"max_queue_size" yaml:"max_queue_size"` // max size of queue before deleting events (-1=no limit, 0=disable)

	ctx       context.Context
	output    QueueReceiver // the output
	control   config.Control
	isInPause uint32           // set to either StatusDelivering or StatusPaused
	queue     chan interface{} // channel to send events to the internal queue
	codecCh   chan []byte      // channel to push a coded message onto
}

// Resume informs that the output is working again - can be called multiple times and is thread safe.
// Should be called after each successfully delivery by the output.
func (t *simpleQueue) Resume(ctx context.Context) error {
	if atomic.CompareAndSwapUint32(&t.isInPause, StatusPaused, StatusDelivering) {
		goglog.Logger.Debugf("queue %s: requesting resume", t.GetType())
		return t.control.RequestResume(ctx)
	}
	return nil
}

// Queue queues an event into the queue, blocking if necessary until cancelled. Queue is used from the output to put something into the queue.
// A call to add an event onto the queue will also pause the input.
func (t *simpleQueue) Queue(ctx context.Context, event interface{}) error {
	if atomic.CompareAndSwapUint32(&t.isInPause, StatusDelivering, StatusPaused) {
		goglog.Logger.Debugf("queue %s is requesting pause", t.GetType())
		err := t.control.RequestPause(ctx)
		if err != nil {
			goglog.Logger.Errorf("queue %s: %s", t.GetType(), err.Error())
		}
	}
	select {
	case t.queue <- event:
		return nil
	case <-ctx.Done():
		return ErrContextCancelled.New(nil)
	}
}

// NewSimpleQueue returns a new queue using a simple retry/queuing mechanism. receiver is your Output object. queueSize is the number of events that will be queued before dropping events from the queue; with a value of -1 the sky is the limit. retryInterval is the time in seconds between each retry.
func NewSimpleQueue(ctx context.Context, control config.Control, receiver QueueReceiver, outch chan []byte, queueSize int, retryInterval uint) Queue {
	var chanSize int
	if queueSize < 1 {
		chanSize = 1
	} else {
		chanSize = queueSize
	}
	s := &simpleQueue{
		RetryInterval: retryInterval,
		MaxQueueSize:  queueSize,
		ctx:           ctx,
		output:        receiver,
		control:       control,
		isInPause:     StatusDelivering,
		queue:         make(chan interface{}, chanSize),
		codecCh:       outch,
	}
	go s.backgroundtask()
	return s
}

// GetType satisfies TypeCommonConfig
func (t *simpleQueue) GetType() string {
	return t.output.GetType()
}

// Output satisfies TypeOutputConfig and is handling incoming events
func (t *simpleQueue) Output(ctx context.Context, event logevent.LogEvent) (err error) {
	// see if output has requested pause and if so just queue the event instead of trying
	if atomic.LoadUint32(&t.isInPause) == StatusPaused {
		select {
		case t.queue <- event:
			return nil
		case <-ctx.Done():
			return ErrContextCancelled.New(nil)
		case <-t.ctx.Done():
			return ErrContextCancelled.New(nil)
		}
	}
	// If we are not in pause mode then call the sender method
	return t.output.OutputEvent(ctx, event)
}

// backgroundtask is running in the background and adds new events to the queue, tries to send them out on the RetryInterval.
func (t *simpleQueue) backgroundtask() {
	var retryqueue list.List // list of queued messages; the list is not multithreading safe, and must only be accessed from backgroundtask()

	goglog.Logger.Debug("backgroundtask started for ", t.GetType())
	dur := time.Duration(t.RetryInterval) * time.Second
	ticker := time.NewTicker(dur)
	defer ticker.Stop()
	for {
		select {
		case event := <-t.queue:
			if (retryqueue.Len() < t.MaxQueueSize) || t.MaxQueueSize == -1 {
				retryqueue.PushBack(event)
			}
		case <-t.ctx.Done():
			goglog.Logger.Debugf("queue %s closing", t.GetType())
			return
		case <-ticker.C:
			// We have reached a RetryInterval. If there are any events in the queue, lets send one back.
			// If we are still in pause mode we will send one, if we are in normal mode we will send all back.
			if e := retryqueue.Front(); e != nil {
				goglog.Logger.Debugf("%s simplequeue trying to deliver, has %v events in queue", t.GetType(), retryqueue.Len())
				if atomic.LoadUint32(&t.isInPause) == StatusPaused {
					msg := e.Value
					retryqueue.Remove(e)
					switch v := msg.(type) {
					case logevent.LogEvent:
						go func() {
							ctx, cancel := context.WithTimeout(context.Background(), dur)
							err := t.output.OutputEvent(ctx, v)
							if err != nil {
								goglog.Logger.Errorf("queue: %ssendone %s", t.GetType(), err.Error())
							}
							cancel()
						}()
					case []byte:
						go func() {
							select {
							case <-t.ctx.Done():
								goglog.Logger.Debugf("queue %s closing", t.GetType())
								return
							case t.codecCh <- v:
							}
						}()
					default:
						goglog.Logger.Errorf("Invalid type %T", v)
					}
				} else {
					// we are not in pause mode and will queue all the events in the queue for sending.
					// First we need to empty the queue and get all events to send.
					myList := []interface{}{}
					for {
						e := retryqueue.Front()
						if e == nil {
							break
						}
						event := e.Value
						myList = append(myList, event)
						retryqueue.Remove(e)
					}
					// Now we have to send them all out
					go func() {
						for x := range myList {
							ctx, cancel := context.WithTimeout(context.Background(), dur)
							defer cancel()
							switch v := (myList[x]).(type) {
							case []byte:
								select {
								case <-t.ctx.Done():
									return
								case t.codecCh <- v:
								}
							case logevent.LogEvent:
								err := t.Output(ctx, v)
								if err != nil {
									goglog.Logger.Errorf("queue: %s sendall %s", t.GetType(), err.Error())
								}
							default:
								goglog.Logger.Errorf("Invalid type %T", v)
							}
						}
					}()
				}
			}
		}
	}
}
