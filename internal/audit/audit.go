// Package audit provides a simple append-only audit log that records every
// drift event detected by driftwatch, writing structured JSON lines to a
// configurable file path.
package audit

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Entry is a single record written to the audit log.
type Entry struct {
	Timestamp time.Time `json:"timestamp"`
	Path      string    `json:"path"`
	OldHash   string    `json:"old_hash"`
	NewHash   string    `json:"new_hash"`
	Message   string    `json:"message,omitempty"`
}

// Logger writes audit entries to an underlying writer.
type Logger struct {
	mu  sync.Mutex
	out io.Writer
}

// New creates a Logger that appends to the file at logPath.
// The file is created if it does not exist.
func New(logPath string) (*Logger, error) {
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o640)
	if err != nil {
		return nil, fmt.Errorf("audit: open log file: %w", err)
	}
	return &Logger{out: f}, nil
}

// NewWithWriter creates a Logger backed by an arbitrary writer (useful in tests).
func NewWithWriter(w io.Writer) *Logger {
	return &Logger{out: w}
}

// Record encodes entry as a JSON line and appends it to the log.
func (l *Logger) Record(e Entry) error {
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now().UTC()
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if err := json.NewEncoder(l.out).Encode(e); err != nil {
		return fmt.Errorf("audit: write entry: %w", err)
	}
	return nil
}
