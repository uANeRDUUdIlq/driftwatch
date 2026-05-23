// Package circuitbreaker implements a simple circuit breaker to prevent
// repeated alert delivery attempts when a downstream target is unhealthy.
package circuitbreaker

import (
	"fmt"
	"sync"
	"time"
)

// State represents the current circuit breaker state.
type State int

const (
	StateClosed   State = iota // normal operation
	StateOpen                  // failing, requests rejected
	StateHalfOpen              // probe request allowed
)

// Breaker is a per-key circuit breaker.
type Breaker struct {
	mu           sync.Mutex
	failures     int
	threshold    int
	cooldown     time.Duration
	openedAt     time.Time
	state        State
	now          func() time.Time
}

// New returns a Breaker that opens after threshold consecutive failures
// and attempts recovery after cooldown elapses.
func New(threshold int, cooldown time.Duration) *Breaker {
	return &Breaker{
		threshold: threshold,
		cooldown:  cooldown,
		now:       time.Now,
	}
}

// Allow reports whether a request should be attempted.
// Returns an error if the circuit is open.
func (b *Breaker) Allow() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch b.state {
	case StateClosed:
		return nil
	case StateOpen:
		if b.now().Sub(b.openedAt) >= b.cooldown {
			b.state = StateHalfOpen
			return nil
		}
		return fmt.Errorf("circuit open: too many failures, retry after %s", b.openedAt.Add(b.cooldown).Format(time.RFC3339))
	case StateHalfOpen:
		return nil
	}
	return nil
}

// RecordSuccess resets the breaker to closed state.
func (b *Breaker) RecordSuccess() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failures = 0
	b.state = StateClosed
}

// RecordFailure increments the failure counter and may open the circuit.
func (b *Breaker) RecordFailure() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failures++
	if b.failures >= b.threshold {
		b.state = StateOpen
		b.openedAt = b.now()
	}
}

// StateSnapshot returns the current state for observability.
func (b *Breaker) StateSnapshot() State {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.state
}
