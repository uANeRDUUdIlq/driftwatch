// Package metrics provides lightweight in-process counters for driftwatch
// operational observability, exposed via a simple HTTP handler.
package metrics

import (
	"encoding/json"
	"net/http"
	"sync"
	"sync/atomic"
)

// Counters holds all runtime metrics tracked by driftwatch.
type Counters struct {
	DriftsDetected  atomic.Int64
	Alertssent      atomic.Int64
	AlertsFailed    atomic.Int64
	FilesWatched    atomic.Int64
	SuppressedAlerts atomic.Int64
	RateLimited     atomic.Int64
}

// Metrics wraps Counters and serves them over HTTP.
type Metrics struct {
	mu       sync.Mutex
	counters *Counters
}

// New creates a new Metrics instance.
func New() *Metrics {
	return &Metrics{counters: &Counters{}}
}

// Counters returns a pointer to the underlying Counters.
func (m *Metrics) Counters() *Counters {
	return m.counters
}

// snapshot returns a map suitable for JSON serialisation.
func (m *Metrics) snapshot() map[string]int64 {
	c := m.counters
	return map[string]int64{
		"drifts_detected":   c.DriftsDetected.Load(),
		"alerts_sent":       c.Alertsent.Load(),
		"alerts_failed":     c.AlertsFailed.Load(),
		"files_watched":     c.FilesWatched.Load(),
		"suppressed_alerts": c.SuppressedAlerts.Load(),
		"rate_limited":      c.RateLimited.Load(),
	}
}

// Handler returns an http.HandlerFunc that serves current metrics as JSON.
func (m *Metrics) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/metrics" {
			http.NotFound(w, r)
			return
		}
		m.mu.Lock()
		snap := m.snapshot()
		m.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(snap)
	}
}
