package metrics_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourorg/driftwatch/internal/metrics"
)

func TestMetrics_InitialCountersAreZero(t *testing.T) {
	m := metrics.New()
	c := m.Counters()
	if v := c.DriftsDetected.Load(); v != 0 {
		t.Fatalf("expected 0, got %d", v)
	}
	if v := c.AlertsSent.Load(); v != 0 {
		t.Fatalf("expected 0, got %d", v)
	}
}

func TestMetrics_CountersIncrementCorrectly(t *testing.T) {
	m := metrics.New()
	c := m.Counters()
	c.DriftsDetected.Add(3)
	c.AlertsSent.Add(2)
	c.AlertsFailed.Add(1)

	if v := c.DriftsDetected.Load(); v != 3 {
		t.Fatalf("DriftsDetected: want 3, got %d", v)
	}
	if v := c.AlertsSent.Load(); v != 2 {
		t.Fatalf("AlertsSent: want 2, got %d", v)
	}
	if v := c.AlertsFailed.Load(); v != 1 {
		t.Fatalf("AlertsFailed: want 1, got %d", v)
	}
}

func TestMetrics_HandlerReturnsJSON(t *testing.T) {
	m := metrics.New()
	m.Counters().DriftsDetected.Add(5)
	m.Counters().RateLimited.Add(2)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	m.Handler()(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var payload map[string]int64
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if payload["drifts_detected"] != 5 {
		t.Errorf("drifts_detected: want 5, got %d", payload["drifts_detected"])
	}
	if payload["rate_limited"] != 2 {
		t.Errorf("rate_limited: want 2, got %d", payload["rate_limited"])
	}
}

func TestMetrics_HandlerUnknownPathReturns404(t *testing.T) {
	m := metrics.New()
	req := httptest.NewRequest(http.MethodGet, "/unknown", nil)
	rec := httptest.NewRecorder()
	m.Handler()(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
