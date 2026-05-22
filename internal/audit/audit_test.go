package audit_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/yourusername/driftwatch/internal/audit"
)

func TestRecord_WritesJSONLine(t *testing.T) {
	var buf bytes.Buffer
	l := audit.NewWithWriter(&buf)

	now := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	err := l.Record(audit.Entry{
		Timestamp: now,
		Path:      "/etc/app/config.yaml",
		OldHash:   "abc123",
		NewHash:   "def456",
		Message:   "drift detected",
	})
	if err != nil {
		t.Fatalf("Record: unexpected error: %v", err)
	}

	var got audit.Entry
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if got.Path != "/etc/app/config.yaml" {
		t.Errorf("path: got %q, want %q", got.Path, "/etc/app/config.yaml")
	}
	if got.OldHash != "abc123" {
		t.Errorf("old_hash: got %q, want %q", got.OldHash, "abc123")
	}
	if got.NewHash != "def456" {
		t.Errorf("new_hash: got %q, want %q", got.NewHash, "def456")
	}
}

func TestRecord_MultipleEntriesAreNewlineSeparated(t *testing.T) {
	var buf bytes.Buffer
	l := audit.NewWithWriter(&buf)

	for i := 0; i < 3; i++ {
		if err := l.Record(audit.Entry{Path: "/file", OldHash: "a", NewHash: "b"}); err != nil {
			t.Fatalf("Record[%d]: %v", i, err)
		}
	}

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}
	for i, line := range lines {
		var e audit.Entry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			t.Errorf("line %d is not valid JSON: %v", i, err)
		}
	}
}

func TestRecord_SetsTimestampWhenZero(t *testing.T) {
	var buf bytes.Buffer
	l := audit.NewWithWriter(&buf)

	before := time.Now().UTC()
	if err := l.Record(audit.Entry{Path: "/x", OldHash: "a", NewHash: "b"}); err != nil {
		t.Fatalf("Record: %v", err)
	}
	after := time.Now().UTC()

	var got audit.Entry
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Timestamp.Before(before) || got.Timestamp.After(after) {
		t.Errorf("timestamp %v not in expected range [%v, %v]", got.Timestamp, before, after)
	}
}
