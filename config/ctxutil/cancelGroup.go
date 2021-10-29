package ctxutil

import (
	"context"
	"sync"
	"time"
)

func NewCancelGroup(parent context.Context) *CancelGroup {
	ctx, cancel := context.WithCancel(parent)

	return &CancelGroup{
		ctx:    ctx,
		cancel: cancel,
	}
}

type CancelGroup struct {
	ctx    context.Context
	cancel func()

	mutex sync.Mutex
	done  chan error

	wg sync.WaitGroup

	errOnce sync.Once
	err     error
}

func (t *CancelGroup) Wait() error {
	t.wg.Wait()
	t.cancel()

	return t.err
}

func (t *CancelGroup) Done() <-chan error {
	t.mutex.Lock()
	if t.done == nil {
		t.done = make(chan error)
		go func() {
			t.wg.Wait()
			t.cancel()
			t.done <- t.err
		}()
	}
	d := t.done
	t.mutex.Unlock()

	return d
}

func (t *CancelGroup) Go(f func(context.Context) error) {
	t.wg.Add(1)

	go func() {
		defer t.wg.Done()

		if err := f(t.ctx); err != nil {
			t.CancelError(err)
		}
	}()
}

// GoCancel go with cancel
func (t *CancelGroup) GoCancel(f func(context.Context) error) context.CancelFunc {
	t.wg.Add(1)

	ctx, cancel := context.WithCancel(t.ctx)

	go func() {
		defer t.wg.Done()

		if err := f(ctx); err != nil {
			t.CancelError(err)
		}
	}()

	return cancel
}

// GoTimeout go with timeout
func (t *CancelGroup) GoTimeout(timeout time.Duration, f func(context.Context) error) context.CancelFunc {
	t.wg.Add(1)

	ctx, cancel := context.WithTimeout(t.ctx, timeout)

	go func() {
		defer t.wg.Done()

		if err := f(ctx); err != nil {
			t.CancelError(err)
		}
	}()

	return cancel
}

// Fork goroutine will disconnect context propagation
func (t *CancelGroup) Fork(f func(context.Context) error) {
	t.wg.Add(1)

	go func() {
		defer t.wg.Done()

		ctx := DisconnectContext(t.ctx)

		if err := f(ctx); err != nil {
			t.CancelError(err)
		}
	}()
}

// ForkTimeout fork with cancel
func (t *CancelGroup) ForkCancel(f func(context.Context) error) context.CancelFunc {
	t.wg.Add(1)

	ctx, cancel := DisconnectContextWithCancel(t.ctx)

	go func() {
		defer t.wg.Done()

		if err := f(ctx); err != nil {
			t.CancelError(err)
		}
	}()

	return cancel
}

// ForkTimeout fork with timeout
func (t *CancelGroup) ForkTimeout(timeout time.Duration, f func(context.Context) error) context.CancelFunc {
	t.wg.Add(1)

	ctx, cancel := DisconnectContextWithTimeout(t.ctx, timeout)

	go func() {
		defer t.wg.Done()

		if err := f(ctx); err != nil {
			t.CancelError(err)
		}
	}()

	return cancel
}

func (t *CancelGroup) Context() context.Context {
	return t.ctx
}

func (t *CancelGroup) Cancel() {
	t.cancel()
}

func (t *CancelGroup) CancelError(err error) {
	t.errOnce.Do(func() {
		t.err = err
		t.cancel()
	})
}

func (t *CancelGroup) Close() (err error) {
	t.cancel()
	t.wg.Wait()

	return t.err
}
