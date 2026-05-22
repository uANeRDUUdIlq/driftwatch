package notifier_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/example/driftwatch/internal/notifier"
	"github.com/example/driftwatch/internal/suppression"
)

func TestNotifier_SuppressionBlocks(t *testing.T) {
	var calls int32
	al, _ := newAlerterAndServer(t, &calls)

	sup := suppression.New()
	sup.Suppress("/etc/cron.conf", time.Now().Add(-time.Hour), time.Now().Add(time.Hour))

	n := notifier.New(notifier.Config{
		Alerter:     al,
		Suppression: sup,
	})

	ok := n.Send(context.Background(), notifier.Event{
		Path:       "/etc/cron.conf",
		OldHash:    "a",
		NewHash:    "b",
		DetectedAt: time.Now(),
	})

	if ok {
		t.Fatal("expected Send to be suppressed")
	}
	time.Sleep(20 * time.Millisecond)
	if atomic.LoadInt32(&calls) != 0 {
		t.Fatalf("expected 0 alert calls during suppression window, got %d", calls)
	}
}

func TestNotifier_SuppressionExpiredAllows(t *testing.T) {
	var calls int32
	al, _ := newAlerterAndServer(t, &calls)

	sup := suppression.New()
	// Window already expired.
	sup.Suppress("/etc/hosts", time.Now().Add(-2*time.Hour), time.Now().Add(-time.Hour))

	n := notifier.New(notifier.Config{
		Alerter:     al,
		Suppression: sup,
	})

	ok := n.Send(context.Background(), notifier.Event{
		Path:       "/etc/hosts",
		OldHash:    "c",
		NewHash:    "d",
		DetectedAt: time.Now(),
	})

	if !ok {
		t.Fatal("expected Send to succeed after suppression window expired")
	}
	time.Sleep(20 * time.Millisecond)
	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected 1 alert call, got %d", calls)
	}
}
