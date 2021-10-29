package ctxutil

import (
	"context"
	"time"
)

func Sleep(ctx context.Context, duration time.Duration) (done bool) {
	if duration < 1 {
		return false
	}
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return true
	case <-timer.C:
		return false
	}
}
