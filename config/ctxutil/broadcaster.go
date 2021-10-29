package ctxutil

import (
	"context"

	"github.com/subchen/go-trylock/v2"
)

type Broadcaster struct {
	mutex   trylock.TryLocker
	channel chan struct{}
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		mutex:   trylock.New(),
		channel: make(chan struct{}),
	}
}

func (t *Broadcaster) Wait(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return context.DeadlineExceeded
	case <-t.Channel():
		return nil
	}
}

func (t *Broadcaster) Channel() <-chan struct{} {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	return t.channel
}

// Signal wakes one goroutine waiting on broadcaster, if there is any.
func (t *Broadcaster) Signal(ctx context.Context) error {
	if !t.mutex.RTryLock(ctx) {
		return context.DeadlineExceeded
	}
	defer t.mutex.RUnlock()

	select {
	case <-ctx.Done():
		return context.DeadlineExceeded
	case t.channel <- struct{}{}:
	default:
	}

	return nil
}

// Broadcast wakes all goroutines waiting on broadcaster, if there is any.
func (t *Broadcaster) Broadcast(ctx context.Context) error {
	newChannel := make(chan struct{})

	if !t.mutex.TryLock(ctx) {
		return context.DeadlineExceeded
	}
	channel := t.channel
	t.channel = newChannel
	t.mutex.Unlock()

	// send broadcast signal
	close(channel)

	return nil
}
