package config

import (
	"context"
	"sync/atomic"
)

type Control interface {
	RequestPause(ctx context.Context) error
	RequestResume(ctx context.Context) error
	PauseSignal() <-chan struct{}
	ResumeSignal() <-chan struct{}
}

func (t *Config) RequestPause(ctx context.Context) error {
	if atomic.CompareAndSwapInt32(&t.state, stateNormal, statePause) {
		return t.signalPause.Broadcast(ctx)
	} else {
		return ErrorInvalidState.New(nil)
	}
}
func (t *Config) RequestResume(ctx context.Context) error {
	if atomic.CompareAndSwapInt32(&t.state, statePause, stateNormal) {
		return t.signalResume.Broadcast(ctx)
	} else {
		return ErrorInvalidState.New(nil)
	}
}

func (t *Config) PauseSignal() <-chan struct{}  { return t.signalPause.Channel() }
func (t *Config) ResumeSignal() <-chan struct{} { return t.signalResume.Channel() }

const (
	stateNormal = iota
	statePause
)
