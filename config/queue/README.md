# queue

This queue implements queueing and control logic to handle backpressure in case an output cannot deliver events for any reason.
The control logic that an input can listen to is implemented in [control.go](../control.go).
For this guide I assume you already have some understanding in how gogstash works internally.

The queue is designed to make it easier for developers to write outputs that does backpressurehandling and queueing, without rewriting all code every time.

Later down in this guide you will see how to rewrite existing outputs to support backpressure.

## Components of the queue

There are [two interfaces](queue.go) that you need to understand.

* QueueReceiver - defines all gogstash common types and OutputEvent(). The only method you need to implement in your code for the queue to work is OutputEvent().
* Queue - defines all gogstash common types, Queue() and Resume(). Queue() is used to signal that you have an issue and put the event on the queue. Resume() is called every time the output delivers an event and makes sure that the inputs are resumed.

## simpleQueue

This implementation of a queue retries any events at a specified interval. If there are more than one event in the queue and the queue is paused, then one event will be sent out. The output module has to queue it back if it fails delivery.
If the queue is in normal state then all events in the queue will be sent immediately.

## Customizing existing outputs

The output [http](../../output/http) is implementing using the steps in this guide and can be used as a reference on how to implement queueing.

For this guide it is assumed that the existing output only handles events synchronously; ie the current Output() method does not return before the event either was delivered or failed delivery.

First you need to edit the modules OutputConfig and add the queue:
```go
type OutputConfig struct {
	// existing defs
  queue queue.Queue
}
```

During InitHandler() you will need to create a new queue object and pass it back.
```go
func InitHandler(
	ctx context.Context,
	raw config.ConfigRaw,
	control config.Control,
) (config.TypeOutputConfig, error) {
	conf := DefaultOutputConfig()
	// all your custom init code
  conf.queue = queue.NewSimpleQueue(ctx, control, &conf, 1, 30) // last values are queue size and retry interval in seconds
  return conf.queue, nil
}
```

The next thing you need to do is to find the handlers existing Output() handler. Change its name to OutputEvent(), this way you satisfy the implementation of queue.QueueReceiver.

At last, rewrite the OutputEvent() handler. You need to make two changes:
1. If the event was delivered successfully, exit the handler with ```return t.queue.Resume(ctx)```.
2. If you failed delivery and this is something that you want to retry at a later time then call ```t.queue.Queue(ctx, event)``` and handle any errors. Exit with an error appropriate for why the output failed.

The developer should drop any events that have fatal errors - where the receiver is actively failing the event because of errors that likely will not go away.
Examples on such errors are page not found (404) and access denied (401).
