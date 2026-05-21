package reporter_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/yourorg/driftwatch/internal/reporter"
)

func makeEvent(path string) reporter.DriftEvent {
	return reporter.DriftEvent{
		Path:       path,
		OldHash:    "aabbccdd1122",
		NewHash:    "11223344aabb",
		DetectedAt: time.Now(),
	}
}

func TestReporter_FlushesOnInterval(t *testing.T) {
	var mu sync.Mutex
	var received []reporter.DriftEvent

	sink := func(events []reporter.DriftEvent) error {
		mu.Lock()
		defer mu.Unlock()
		received = append(received, events...)
		return nil
	}

	r := reporter.New(50*time.Millisecond, sink)
	r.Record(makeEvent("/etc/app/config.yaml"))
	r.Record(makeEvent("/etc/app/db.yaml"))

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	go r.Run(ctx)

	<-ctx.Done()

	mu.Lock()
	defer mu.Unlock()
	if len(received) != 2 {
		t.Errorf("expected 2 events, got %d", len(received))
	}
}

func TestReporter_FinalFlushOnShutdown(t *testing.T) {
	var mu sync.Mutex
	var received []reporter.DriftEvent

	sink := func(events []reporter.DriftEvent) error {
		mu.Lock()
		defer mu.Unlock()
		received = append(received, events...)
		return nil
	}

	r := reporter.New(10*time.Second, sink) // long interval — won't tick
	r.Record(makeEvent("/etc/nginx/nginx.conf"))

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		r.Run(ctx)
		close(done)
	}()
	cancel()
	<-done

	mu.Lock()
	defer mu.Unlock()
	if len(received) != 1 {
		t.Errorf("expected 1 event on shutdown flush, got %d", len(received))
	}
}

func TestFormatDigest_Empty(t *testing.T) {
	out := reporter.FormatDigest(nil)
	if out != "No drift detected." {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestFormatDigest_Events(t *testing.T) {
	events := []reporter.DriftEvent{makeEvent("/etc/hosts")}
	out := reporter.FormatDigest(events)
	if len(out) == 0 {
		t.Error("expected non-empty digest output")
	}
}
