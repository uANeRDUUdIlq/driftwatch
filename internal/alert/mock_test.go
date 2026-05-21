package alert_test

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/yourorg/driftwatch/internal/alert"
)

// TestSend_BothTargets verifies that when both webhook and Slack URLs are
// configured, the notifier posts to each exactly once.
func TestSend_BothTargets(t *testing.T) {
	var webhookHits, slackHits atomic.Int32

	webhook := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		webhookHits.Add(1)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer webhook.Close()

	slack := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slackHits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer slack.Close()

	n := alert.New(webhook.URL, slack.URL)
	if err := n.Send(newEvent()); err != nil {
		t.Fatalf("Send: %v", err)
	}

	if webhookHits.Load() != 1 {
		t.Errorf("expected 1 webhook hit, got %d", webhookHits.Load())
	}
	if slackHits.Load() != 1 {
		t.Errorf("expected 1 slack hit, got %d", slackHits.Load())
	}
}

// TestSend_PartialFailure verifies that errors from one target do not
// suppress delivery attempts to the other.
func TestSend_PartialFailure(t *testing.T) {
	var slackHits atomic.Int32

	slack := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slackHits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer slack.Close()

	// Use a non-listening address to force a webhook failure.
	n := alert.New("http://127.0.0.1:1", slack.URL)
	err := n.Send(newEvent())
	if err == nil {
		t.Fatal("expected error due to bad webhook URL")
	}

	if slackHits.Load() != 1 {
		t.Errorf("expected Slack to still receive alert despite webhook failure, got %d hits", slackHits.Load())
	}
}
