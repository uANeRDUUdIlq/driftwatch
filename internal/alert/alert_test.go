package alert_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourorg/driftwatch/internal/alert"
)

func newEvent() alert.DriftEvent {
	return alert.DriftEvent{
		Path:       "/etc/app/config.yaml",
		OldHash:    "abc123",
		NewHash:    "def456",
		DetectedAt: time.Now(),
	}
}

func TestSend_Webhook(t *testing.T) {
	var received map[string]any

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := alert.New(ts.URL, "")
	if err := n.Send(newEvent()); err != nil {
		t.Fatalf("Send: %v", err)
	}

	if received["path"] != "/etc/app/config.yaml" {
		t.Errorf("expected path in payload, got %v", received)
	}
}

func TestSend_Slack(t *testing.T) {
	var received map[string]any

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := alert.New("", ts.URL)
	if err := n.Send(newEvent()); err != nil {
		t.Fatalf("Send: %v", err)
	}

	text, ok := received["text"].(string)
	if !ok || text == "" {
		t.Errorf("expected non-empty Slack text, got %v", received)
	}
}

func TestSend_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	n := alert.New(ts.URL, "")
	if err := n.Send(newEvent()); err == nil {
		t.Fatal("expected error on 500 response, got nil")
	}
}

func TestSend_NoTargets(t *testing.T) {
	n := alert.New("", "")
	if err := n.Send(newEvent()); err != nil {
		t.Fatalf("expected no error when no targets configured, got: %v", err)
	}
}
