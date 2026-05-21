// Package suppression provides a time-based alert suppression mechanism.
// It allows callers to silence alerts for specific paths during maintenance
// windows or after an acknowledged drift event.
package suppression

import (
	"sync"
	"time"
)

// Suppressor tracks suppressed paths and their expiry times.
type Suppressor struct {
	mu      sync.Mutex
	entries map[string]time.Time
	now     func() time.Time
}

// New returns a new Suppressor.
func New() *Suppressor {
	return &Suppressor{
		entries: make(map[string]time.Time),
		now:     time.Now,
	}
}

// Suppress silences alerts for path for the given duration.
// Calling Suppress again on an already-suppressed path extends the window.
func (s *Suppressor) Suppress(path string, d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[path] = s.now().Add(d)
}

// IsSuppressed reports whether alerts for path are currently suppressed.
func (s *Suppressor) IsSuppressed(path string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	expiry, ok := s.entries[path]
	if !ok {
		return false
	}
	if s.now().After(expiry) {
		delete(s.entries, path)
		return false
	}
	return true
}

// Lift removes any active suppression for path immediately.
func (s *Suppressor) Lift(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, path)
}

// ActiveCount returns the number of currently suppressed paths.
func (s *Suppressor) ActiveCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	count := 0
	for path, expiry := range s.entries {
		if now.After(expiry) {
			delete(s.entries, path)
		} else {
			count++
		}
	}
	return count
}
