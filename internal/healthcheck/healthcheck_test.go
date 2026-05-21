package healthcheck_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/driftwatch/internal/healthcheck"
)

func TestHealthz_ReturnsOK(t *testing.T) {
	h := healthcheck.New()
	h.SetWatchedFiles(5)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var s healthcheck.Status
	if err := json.NewDecoder(rec.Body).Decode(&s); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !s.OK {
		t.Error("expected ok=true")
	}
	if s.WatchedFiles != 5 {
		t.Errorf("expected watched_files=5, got %d", s.WatchedFiles)
	}
	if s.StartedAt.IsZero() {
		t.Error("expected non-zero started_at")
	}
	if s.Uptime == "" {
		t.Error("expected non-empty uptime")
	}
}

func TestHealthz_UptimeIncreases(t *testing.T) {
	h := healthcheck.New()
	time.Sleep(10 * time.Millisecond)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	var s healthcheck.Status
	if err := json.NewDecoder(rec.Body).Decode(&s); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if time.Since(s.StartedAt) < 10*time.Millisecond {
		t.Error("expected started_at to be at least 10ms ago")
	}
}

func TestHealthz_UnknownPathReturns404(t *testing.T) {
	h := healthcheck.New()

	req := httptest.NewRequest(http.MethodGet, "/unknown", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestHealthz_WatchedFilesDefault(t *testing.T) {
	h := healthcheck.New()

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	var s healthcheck.Status
	_ = json.NewDecoder(rec.Body).Decode(&s)

	if s.WatchedFiles != 0 {
		t.Errorf("expected default watched_files=0, got %d", s.WatchedFiles)
	}
}
