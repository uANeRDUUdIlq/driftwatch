// Package reporter provides a periodic drift reporting summary
// that aggregates drift events and emits a digest on a configurable schedule.
package reporter

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// DriftEvent represents a single detected drift occurrence.
type DriftEvent struct {
	Path      string
	OldHash   string
	NewHash   string
	DetectedAt time.Time
}

// Sink is the destination for a drift report digest.
type Sink func(events []DriftEvent) error

// Reporter accumulates drift events and flushes them to a Sink periodically.
type Reporter struct {
	mu       sync.Mutex
	events   []DriftEvent
	interval time.Duration
	sink     Sink
}

// New creates a Reporter that flushes to sink every interval.
func New(interval time.Duration, sink Sink) *Reporter {
	return &Reporter{
		interval: interval,
		sink:     sink,
	}
}

// Record appends a drift event to the internal buffer.
func (r *Reporter) Record(e DriftEvent) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, e)
}

// flush drains the buffer and calls the sink. Safe to call concurrently.
func (r *Reporter) flush() {
	r.mu.Lock()
	if len(r.events) == 0 {
		r.mu.Unlock()
		return
	}
	batch := make([]DriftEvent, len(r.events))
	copy(batch, r.events)
	r.events = r.events[:0]
	r.mu.Unlock()

	if err := r.sink(batch); err != nil {
		log.Printf("reporter: sink error: %v", err)
	}
}

// Run starts the periodic flush loop and blocks until ctx is cancelled.
func (r *Reporter) Run(ctx context.Context) {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			r.flush()
		case <-ctx.Done():
			r.flush() // final flush on shutdown
			return
		}
	}
}

// FormatDigest produces a human-readable summary of drift events.
func FormatDigest(events []DriftEvent) string {
	if len(events) == 0 {
		return "No drift detected."
	}
	out := fmt.Sprintf("Drift digest: %d event(s)\n", len(events))
	for _, e := range events {
		out += fmt.Sprintf("  [%s] %s  old=%s new=%s\n",
			e.DetectedAt.Format(time.RFC3339), e.Path, e.OldHash[:8], e.NewHash[:8])
	}
	return out
}
