// export_test.go exposes internal helpers to the snapshot_test package
// without polluting the public API.
package snapshot

// StoreSize returns the number of paths currently tracked by the store.
// Intended for use in tests only.
func (s *Store) StoreSize() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.hashes)
}

// Reset clears all recorded hashes from the store.
// Intended for use in tests only.
func (s *Store) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hashes = make(map[string]string)
}
