// Package ratelimit provides a simple token-bucket rate limiter
// to prevent alert flooding when many files drift simultaneously.
package ratelimit

import (
	"sync"
	"time"
)

// Limiter controls how many alerts can be sent within a given window.
type Limiter struct {
	mu       sync.Mutex
	max      int
	window   time.Duration
	buckets  map[string][]time.Time
	nowFunc  func() time.Time
}

// New creates a Limiter that allows at most max events per window per key.
func New(max int, window time.Duration) *Limiter {
	return &Limiter{
		max:     max,
		window:  window,
		buckets: make(map[string][]time.Time),
		nowFunc: time.Now,
	}
}

// Allow reports whether an event for the given key is within the rate limit.
// It records the event if allowed.
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.nowFunc()
	cutoff := now.Add(-l.window)

	times := l.buckets[key]
	filtered := times[:0]
	for _, t := range times {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}

	if len(filtered) >= l.max {
		l.buckets[key] = filtered
		return false
	}

	l.buckets[key] = append(filtered, now)
	return true
}

// Reset clears the event history for the given key.
func (l *Limiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.buckets, key)
}
