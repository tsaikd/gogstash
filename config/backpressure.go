package config

import (
	"github.com/tsaikd/gogstash/config/goglog"
	"sync/atomic"
)

/*
	To handle backpressure an output needs a mechanism to signal that it cannot process messages. And the inputs has to
	stop receiving messages when such signal arrives. Backpressure is the way to do this. Outputs can inform when they
	cannot process more messages and Backpressure will then signal all inputs that have registered that they can stop their inputs.

	Inputs has to implement Pause() and Resume(). Those calls are invoked from another thread and the developer has to
	take care to make sure their code is thread-safe. The inputs init() will have to register the input like this:

	BackpressureFactory().RegisterInput(&Input)

	The output init will register a copy of the Backpressure object:
	output.bpf = BackpressureFactory()

	When the Event() method sees something wrong it will inform about this:
	goglog.logger.Error("output x cannot process events because of...")
	o.bpf.Pause()

	The output MUST also continue to check if the error condition is resolved and then call
	o.bpf.Resume()
*/

var backPressureObject *Backpressure // our global object

// CanPause is used for objects that can be paused
type CanPause interface {
	Pause()
	Resume()
}

// BackpressureFactory returns the global Backpressure object
func BackpressureFactory() *Backpressure {
	if backPressureObject == nil {
		backPressureObject = &Backpressure{
			msgChange: make(chan int32),
		}
		go backPressureObject.backgroundtask()
	}
	return backPressureObject
}

// Backpressure handles all work for stopping inputs. Both inputs and outputs needs to get the global object and hook onto it.
type Backpressure struct {
	numBlocks int32      // number of outputs that has requested to block processing
	isPaused  bool       // set to true if we already are paused
	msgChange chan int32 // internal channel used for messaging changes
	inputs    []CanPause // a list of inputs that can be paused
}

// RegisterInput registers an input that can be paused
func (b *Backpressure) RegisterInput(input CanPause) {
	b.inputs = append(b.inputs, input)
}

// RequestPause is called by an output that wants to stop incoming messages. The output has to log this event themselves.
func (b *Backpressure) RequestPause() {
	b.msgChange <- atomic.AddInt32(&b.numBlocks, 1)
}

// RequestResume is called by an output when it is ready to accept new messages. The output has to log this event themselves.
func (b *Backpressure) RequestResume() {
	b.msgChange <- atomic.AddInt32(&b.numBlocks, -1)
}

// backgroundtask is the background process that receives messages on the channel and stops all registered inputs when
// there are one or more outputs that has requested to pause.
func (b *Backpressure) backgroundtask() {
	for {
		select {
		case num := <-b.msgChange:
			switch num {
			case 0:
				goglog.Logger.Info("Resuming inputs")
				b.unpause()
			case 1:
				if !b.isPaused {
					goglog.Logger.Info("Pausing inputs")
					b.pause()
				}
			default:
				if num < 0 {
					goglog.Logger.Error("backpressure: outputs counted wrong - create issue and notify developers")
					atomic.StoreInt32(&b.numBlocks, 0)
				}
			}
		}
	}
}

// pause is called when the first output has requested to pause
func (b *Backpressure) pause() {
	for _, v := range b.inputs {
		v.Pause()
	}
	b.isPaused = true
}

// unpause is called when the last output has requested to resume
func (b *Backpressure) unpause() {
	for _, v := range b.inputs {
		v.Resume()
	}
	b.isPaused = false
}
