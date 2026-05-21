// Package snapshot provides functionality for capturing and comparing
// file content hashes to detect configuration drift.
package snapshot

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"sync"
)

// Store holds a thread-safe map of file paths to their last known SHA-256 hashes.
type Store struct {
	mu     sync.RWMutex
	hashes map[string]string
}

// New returns an initialised, empty Store.
func New() *Store {
	return &Store{
		hashes: make(map[string]string),
	}
}

// Hash computes the SHA-256 hex digest of the file at path.
func Hash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("snapshot: open %q: %w", path, err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("snapshot: hash %q: %w", path, err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// Record stores the current hash for path, returning the hash.
func (s *Store) Record(path string) (string, error) {
	hash, err := Hash(path)
	if err != nil {
		return "", err
	}
	s.mu.Lock()
	s.hashes[path] = hash
	s.mu.Unlock()
	return hash, nil
}

// Changed returns true and the new hash if the file at path differs from
// the stored snapshot. If no snapshot exists the file is recorded and
// Changed returns false.
func (s *Store) Changed(path string) (bool, string, error) {
	current, err := Hash(path)
	if err != nil {
		return false, "", err
	}

	s.mu.RLock()
	prev, exists := s.hashes[path]
	s.mu.RUnlock()

	if !exists {
		s.mu.Lock()
		s.hashes[path] = current
		s.mu.Unlock()
		return false, current, nil
	}

	if current != prev {
		s.mu.Lock()
		s.hashes[path] = current
		s.mu.Unlock()
		return true, current, nil
	}
	return false, current, nil
}

// Get returns the stored hash for path and whether it exists.
func (s *Store) Get(path string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	h, ok := s.hashes[path]
	return h, ok
}
