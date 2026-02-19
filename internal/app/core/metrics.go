package core

import (
	"time"
)

type LatencyTracker struct {
	start time.Time
}

func StartLatency() LatencyTracker {
	return LatencyTracker{start: time.Now()}
}

func (t LatencyTracker) Elapsed() time.Duration {
	if t.start.IsZero() {
		return 0
	}
	return time.Since(t.start)
}
