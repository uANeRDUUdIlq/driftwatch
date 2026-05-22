package notifier_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/example/driftwatch/internal/alert"
	"github.com/example/driftwatch/internal/notifier"
	"github.com/example/driftwatch/internal/ratelimit"
)

func newAlerterAndServer(t *testing.T, calls *int32) (*alert.Alerter, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(calls, 1)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"ok": "1"})
	}))
	t.Cleanup(srv.Close)
	al := alert.New(alert.Config{WebhookURL: srv.URL})
	return al, srv
}

func TestNotifier_SendsAlert(t *testing.T) {
	var calls int32
	al, _ := newAlerterAndServer(t, &calls)
	n := notifier.New(notifier.Config{Alerter: al})

	ok := n.Send(context.Background(), notifier.Event{
		Path:       "/etc/app.conf",
		OldHash:    "aaa",
		NewHash:    "bbb",
		DetectedAt: time.Now(),
	})

	if !ok {
		t.Fatal("expected Send to return true")
	}
	time.Sleep(20 * time.Millisecond)
	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected 1 alert call, got %d", calls)
	}
}

func TestNotifier_RateLimitBlocks(t *testing.T) {
	var calls int32
	al, _ := newAlerterAndServer(t, &calls)
	rl := ratelimit.New(ratelimit.Config{MaxEvents: 1, Window: time.Minute})
	n := notifier.New(notifier.Config{Alerter: al, Limiter: rl})

	ctx := context.Background()
	ev := notifier.Event{Path: "/etc/app.conf", OldHash: "a", NewHash: "b", DetectedAt: time.Now()}

	n.Send(ctx, ev)
	time.Sleep(20 * time.Millisecond)

	ok := n.Send(ctx, ev)
	if ok {
		t.Fatal("expected second Send to be rate-limited")
	}
	time.Sleep(20 * time.Millisecond)
	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected 1 alert call, got %d", calls)
	}
}

func TestNotifier_NilPipelineStages(t *testing.T) {
	var calls int32
	al, _ := newAlerterAndServer(t, &calls)
	// All optional stages nil — should not panic.
	n := notifier.New(notifier.Config{
		Alerter:     al,
		Limiter:     nil,
		Suppression: nil,
		Debouncer:   nil,
		Auditor:     nil,
	})
	n.Send(context.Background(), notifier.Event{
		Path: "/etc/app.conf", OldHash: "x", NewHash: "y", DetectedAt: time.Now(),
	})
	time.Sleep(20 * time.Millisecond)
	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected 1 alert call, got %d", calls)
	}
}
