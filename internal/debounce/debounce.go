// Package debounce provides a simple debouncer that delays processing
// of rapid successive events, emitting only the last one after a quiet period.
package debounce

import (
	"sync"
	"time"
)

// Debouncer delays calls to fn until after wait has elapsed since the last call.
type Debouncer struct {
	wait  time.Duration
	fn    func(key string)
	mu    sync.Mutex
	timers map[string]*time.Timer
}

// New creates a Debouncer that will call fn at most once per wait duration per key.
func New(wait time.Duration, fn func(key string)) *Debouncer {
	return &Debouncer{
		wait:   wait,
		fn:     fn,
		timers: make(map[string]*time.Timer),
	}
}

// Trigger schedules fn(key) to be called after the debounce wait period.
// If Trigger is called again for the same key before the timer fires,
// the timer is reset, effectively coalescing rapid events.
func (d *Debouncer) Trigger(key string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if t, ok := d.timers[key]; ok {
		t.Stop()
	}

	d.timers[key] = time.AfterFunc(d.wait, func() {
		d.mu.Lock()
		delete(d.timers, key)
		d.mu.Unlock()
		d.fn(key)
	})
}

// Stop cancels all pending debounced calls.
func (d *Debouncer) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()

	for key, t := range d.timers {
		t.Stop()
		delete(d.timers, key)
	}
}
