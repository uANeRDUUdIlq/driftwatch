package watcher_test

import (
	"os"
	"testing"
	"time"

	"github.com/yourusername/driftwatch/internal/config"
	"github.com/yourusername/driftwatch/internal/watcher"
)

func tempFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "driftwatch-*.txt")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func newTestConfig(paths []string, intervalSec int) *config.Config {
	return &config.Config{
		WatchPaths: paths,
		Interval:   intervalSec,
		Alerts:     config.Alerts{WebhookURL: "http://example.com"},
	}
}

func TestWatcher_NoDriftWhenUnchanged(t *testing.T) {
	path := tempFile(t, "stable content")
	cfg := newTestConfig([]string{path}, 1)
	w := watcher.New(cfg)

	go func() {
		time.Sleep(150 * time.Millisecond)
		w.Stop()
	}()

	if err := w.Start(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	select {
	case ev := <-w.Events:
		t.Errorf("unexpected drift event: %+v", ev)
	default:
	}
}

func TestWatcher_DetectsDrift(t *testing.T) {
	path := tempFile(t, "original")
	cfg := newTestConfig([]string{path}, 1)
	w := watcher.New(cfg)

	go func() {
		time.Sleep(80 * time.Millisecond)
		if err := os.WriteFile(path, []byte("modified"), 0o644); err != nil {
			t.Errorf("write file: %v", err)
		}
		time.Sleep(1200 * time.Millisecond)
		w.Stop()
	}()

	if err := w.Start(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	select {
	case ev := <-w.Events:
		if ev.Path != path {
			t.Errorf("expected path %s, got %s", path, ev.Path)
		}
		if ev.OldChecksum == ev.NewChecksum {
			t.Error("checksums should differ")
		}
	default:
		t.Error("expected a drift event but got none")
	}
}

func TestWatcher_InvalidPathFails(t *testing.T) {
	cfg := newTestConfig([]string{"/nonexistent/path/file.cfg"}, 1)
	w := watcher.New(cfg)
	if err := w.Start(); err == nil {
		t.Error("expected error for missing file, got nil")
		w.Stop()
	}
}
