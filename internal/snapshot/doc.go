// Package snapshot records SHA-256 hashes of watched files and exposes
// helpers to detect when a file's content has drifted from its last
// known state.
//
// Typical usage:
//
//	store := snapshot.New()
//
//	// Seed the initial baseline.
//	for _, path := range watchPaths {
//		store.Record(path)
//	}
//
//	// On each poll interval:
//	changed, newHash, err := store.Changed(path)
//	if changed {
//		// emit drift alert containing newHash
//	}
package snapshot
