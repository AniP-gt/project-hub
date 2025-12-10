package app

import (
	"time"
)

// LatencyTracker measures durations for key interactions (e.g., view switch, status move).
type LatencyTracker struct {
	start time.Time
}

// Start creates a tracker starting now.
func StartLatency() LatencyTracker {
	return LatencyTracker{start: time.Now()}
}

// Elapsed returns the duration since start.
func (t LatencyTracker) Elapsed() time.Duration {
	if t.start.IsZero() {
		return 0
	}
	return time.Since(t.start)
}
