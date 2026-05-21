package snapshot_test

import (
	"os"
	"testing"

	"github.com/yourusername/driftwatch/internal/snapshot"
)

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "snap-*.txt")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestHash_Deterministic(t *testing.T) {
	path := writeTempFile(t, "hello world")
	h1, err := snapshot.Hash(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	h2, err := snapshot.Hash(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h1 != h2 {
		t.Errorf("expected identical hashes, got %q and %q", h1, h2)
	}
}

func TestHash_MissingFile(t *testing.T) {
	_, err := snapshot.Hash("/nonexistent/path/file.txt")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestChanged_FirstCallNoChange(t *testing.T) {
	store := snapshot.New()
	path := writeTempFile(t, "initial content")

	changed, _, err := store.Changed(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if changed {
		t.Error("expected no change on first call")
	}
}

func TestChanged_DetectsDrift(t *testing.T) {
	store := snapshot.New()
	path := writeTempFile(t, "original")

	if _, err := store.Record(path); err != nil {
		t.Fatalf("record: %v", err)
	}

	if err := os.WriteFile(path, []byte("modified"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	changed, newHash, err := store.Changed(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !changed {
		t.Error("expected drift to be detected")
	}
	if newHash == "" {
		t.Error("expected non-empty new hash")
	}
}

func TestChanged_NoDriftWhenSame(t *testing.T) {
	store := snapshot.New()
	path := writeTempFile(t, "stable content")

	store.Record(path) //nolint:errcheck

	changed, _, err := store.Changed(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if changed {
		t.Error("expected no drift for unchanged file")
	}
}

func TestGet_ReturnsStoredHash(t *testing.T) {
	store := snapshot.New()
	path := writeTempFile(t, "data")

	hash, err := store.Record(path)
	if err != nil {
		t.Fatalf("record: %v", err)
	}

	got, ok := store.Get(path)
	if !ok {
		t.Fatal("expected hash to exist in store")
	}
	if got != hash {
		t.Errorf("expected %q, got %q", hash, got)
	}
}
