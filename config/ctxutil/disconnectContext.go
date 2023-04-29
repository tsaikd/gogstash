// https://rodaine.com/2020/07/break-context-cancellation-chain/

package ctxutil

import (
	"context"
	"time"
)

func DisconnectContext(parent context.Context) context.Context {
	return disconnectedContext{parent: parent}
}

func DisconnectContextWithCancel(parent context.Context) (context.Context, context.CancelFunc) {
	ctx := disconnectedContext{parent: parent}

	return context.WithCancel(ctx)
}

func DisconnectContextWithTimeout(
	parent context.Context,
	timeout time.Duration,
) (context.Context, context.CancelFunc) {
	ctx := disconnectedContext{parent: parent}

	return context.WithTimeout(ctx, timeout)
}

// disconnectedContext looks very similar to the nonexported context.emptyCtx
// implementation from the standard library, with the exception of the parent's
// Value method being the only feature propagated.
type disconnectedContext struct {
	parent context.Context
}

// Deadline will erase any actual deadline from the parent, returning ok==false
func (ctx disconnectedContext) Deadline() (deadline time.Time, ok bool) {
	return
}

// Done will stop propagation of the parent context's done channel. Receiving
// on a nil channel will block forever.
func (ctx disconnectedContext) Done() <-chan struct{} {
	return nil
}

// Err will always return nil since there is no longer any cancellation
func (ctx disconnectedContext) Err() error {
	return nil
}

// Value behaves as normal, continuing up the chain to find a matching
// key-value pair.
func (ctx disconnectedContext) Value(key any) any {
	return ctx.parent.Value(key)
}
