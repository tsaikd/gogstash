package queue

import (
	"context"

	"github.com/tsaikd/KDGoLib/errutil"

	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
)

// QueueReceiver defines an interface for using a queue when incoming events has to pause.
// This should be the output object, implementing OutputEvent and TypeCommonConfig.
type QueueReceiver interface {
	config.TypeCommonConfig

	// OutputEvent is the event that is called from the queue when an event is to be processed. This
	// event typically replaces your existing Output() and is called when an event is ready to be sent out on the output.
	OutputEvent(ctx context.Context, event logevent.LogEvent) (err error)
}

// Queue allows for queueing of data into the queue
type Queue interface {
	config.TypeCommonConfig                                    // has to be here to be a supported TypeOutputConfig.
	Output(ctx context.Context, event logevent.LogEvent) error // has to be here to be a supported TypeOutputConfig.
	Queue(ctx context.Context, event any) error                // allows the output to queue an event, also pausing the input if needed. Thread safe.
	Resume(ctx context.Context) error                          // informs that the output is working again - can be called multiple times and is thread safe.
}

const (
	StatusDelivering = iota // if we are in running mode - delivering messages
	StatusPaused            // if we have paused the inputs
)

var (
	ErrContextCancelled = errutil.NewFactory("context canceled")
)
