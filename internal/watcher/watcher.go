package watcher

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/yourusername/driftwatch/internal/config"
)

// FileState holds the last known checksum of a watched file.
type FileState struct {
	Path     string
	Checksum string
	ModTime  time.Time
}

// DriftEvent is emitted when a file's checksum changes unexpectedly.
type DriftEvent struct {
	Path        string
	OldChecksum string
	NewChecksum string
	DetectedAt  time.Time
}

// Watcher monitors a set of file paths for drift.
type Watcher struct {
	cfg      *config.Config
	states   map[string]FileState
	mu       sync.RWMutex
	Events   chan DriftEvent
	stopCh   chan struct{}
}

// New creates a new Watcher from the provided config.
func New(cfg *config.Config) *Watcher {
	return &Watcher{
		cfg:    cfg,
		states: make(map[string]FileState),
		Events: make(chan DriftEvent, 16),
		stopCh: make(chan struct{}),
	}
}

// Start begins the polling loop. It blocks until Stop is called.
func (w *Watcher) Start() error {
	if err := w.snapshot(); err != nil {
		return fmt.Errorf("initial snapshot failed: %w", err)
	}
	ticker := time.NewTicker(time.Duration(w.cfg.Interval) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			w.check()
		case <-w.stopCh:
			return nil
		}
	}
}

// Stop signals the watcher to halt.
func (w *Watcher) Stop() { close(w.stopCh) }

func (w *Watcher) snapshot() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, p := range w.cfg.WatchPaths {
		cs, mt, err := checksumFile(p)
		if err != nil {
			return fmt.Errorf("cannot read %s: %w", p, err)
		}
		w.states[p] = FileState{Path: p, Checksum: cs, ModTime: mt}
	}
	return nil
}

func (w *Watcher) check() {
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, p := range w.cfg.WatchPaths {
		cs, mt, err := checksumFile(p)
		if err != nil {
			continue
		}
		prev := w.states[p]
		if cs != prev.Checksum {
			w.Events <- DriftEvent{
				Path:        p,
				OldChecksum: prev.Checksum,
				NewChecksum: cs,
				DetectedAt:  time.Now(),
			}
			w.states[p] = FileState{Path: p, Checksum: cs, ModTime: mt}
		}
	}
}

func checksumFile(path string) (string, time.Time, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", time.Time{}, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return "", time.Time{}, err
	}
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", time.Time{}, err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), info.ModTime(), nil
}
