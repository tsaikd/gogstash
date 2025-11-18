package config

import (
	"sync/atomic"
	"testing"
	"time"
)

// countPauses satisfies CanPause and just counts each occurence
type countPauses struct {
	numPause, numResume int32
}

// Pause count each pause
func (c *countPauses) Pause() {
	atomic.AddInt32(&c.numPause, 1)
}

// Resume count each resume
func (c *countPauses) Resume() {
	atomic.AddInt32(&c.numResume, 1)
}

// Balance returns numPauses - numResumes
func (c *countPauses) Balance() int {
	return int(atomic.LoadInt32(&c.numPause) - atomic.LoadInt32(&c.numResume))
}

func TestBackpressure1(t *testing.T) {
	bp := BackpressureFactory()
	var myCounter countPauses
	bp.RegisterInput(&myCounter)
	bp.RequestPause()
	bp.RequestPause()
	bp.RequestResume()
	if !bp.isPaused {
		t.Error("Backpressure isPaused in wrong state")
	}
	bp.RequestResume()
	time.Sleep(100 * time.Millisecond)
	if myCounter.Balance() != 0 {
		t.Error("Backpressure did not resume CanPause consumer")
	}
	if bp.isPaused {
		t.Error("Backpressure isPaused in wrong state")
	}
}
