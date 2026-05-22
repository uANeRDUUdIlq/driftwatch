package notifier_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/example/driftwatch/internal/debounce"
	"github.com/example/driftwatch/internal/notifier"
)

func TestNotifier_DebounceCoalesces(t *testing.T) {
	var calls int32
	al, _ := newAlerterAndServer(t, &calls)
	deb := debounce.New()

	n := notifier.New(notifier.Config{
		Alerter:      al,
		Debouncer:    deb,
		DebounceWait: 80 * time.Millisecond,
	})

	ctx := context.Background()
	ev := notifier.Event{Path: "/etc/nginx.conf", OldHash: "1", NewHash: "2", DetectedAt: time.Now()}

	// Fire three rapid events; only one alert should be dispatched.
	n.Send(ctx, ev)
	n.Send(ctx, ev)
	n.Send(ctx, ev)

	// Wait for debounce window to pass and alert to fire.
	time.Sleep(200 * time.Millisecond)

	got := atomic.LoadInt32(&calls)
	if got != 1 {
		t.Fatalf("expected 1 coalesced alert call, got %d", got)
	}
}

func TestNotifier_DebounceDistinctPaths(t *testing.T) {
	var calls int32
	al, _ := newAlerterAndServer(t, &calls)
	deb := debounce.New()

	n := notifier.New(notifier.Config{
		Alerter:      al,
		Debouncer:    deb,
		DebounceWait: 60 * time.Millisecond,
	})

	ctx := context.Background()
	n.Send(ctx, notifier.Event{Path: "/etc/a.conf", OldHash: "1", NewHash: "2", DetectedAt: time.Now()})
	n.Send(ctx, notifier.Event{Path: "/etc/b.conf", OldHash: "3", NewHash: "4", DetectedAt: time.Now()})

	time.Sleep(200 * time.Millisecond)

	got := atomic.LoadInt32(&calls)
	if got != 2 {
		t.Fatalf("expected 2 alert calls for distinct paths, got %d", got)
	}
}
