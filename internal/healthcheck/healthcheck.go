// Package healthcheck provides a simple HTTP health endpoint for driftwatch.
// It exposes a /healthz route that returns the daemon's current status,
// including uptime and the number of files being watched.
package healthcheck

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
	"time"
)

// Status holds the data returned by the health endpoint.
type Status struct {
	OK           bool      `json:"ok"`
	Uptime       string    `json:"uptime"`
	WatchedFiles int64     `json:"watched_files"`
	StartedAt    time.Time `json:"started_at"`
}

// Handler is an HTTP handler that reports daemon health.
type Handler struct {
	startedAt    time.Time
	watchedFiles atomic.Int64
}

// New creates a new Handler.
func New() *Handler {
	return &Handler{
		startedAt: time.Now(),
	}
}

// SetWatchedFiles updates the count of files currently being monitored.
func (h *Handler) SetWatchedFiles(n int64) {
	h.watchedFiles.Store(n)
}

// ServeHTTP implements http.Handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/healthz" {
		http.NotFound(w, r)
		return
	}

	s := Status{
		OK:           true,
		Uptime:       time.Since(h.startedAt).Round(time.Second).String(),
		WatchedFiles: h.watchedFiles.Load(),
		StartedAt:    h.startedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(s)
}
