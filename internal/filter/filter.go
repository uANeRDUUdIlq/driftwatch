// Package filter provides path-based filtering for driftwatch,
// allowing users to exclude specific files or glob patterns from drift detection.
package filter

import (
	"path/filepath"
	"strings"
)

// Filter holds compiled exclusion patterns and decides whether a given
// file path should be ignored during drift checks.
type Filter struct {
	patterns []string
}

// New returns a Filter configured with the provided glob patterns.
// Patterns follow filepath.Match semantics (e.g. "*.tmp", "/etc/hosts").
func New(patterns []string) *Filter {
	normalized := make([]string, 0, len(patterns))
	for _, p := range patterns {
		p = strings.TrimSpace(p)
		if p != "" {
			normalized = append(normalized, p)
		}
	}
	return &Filter{patterns: normalized}
}

// Ignored reports whether path matches any of the exclusion patterns.
// A match against the full path or the base name of the path is sufficient
// to trigger exclusion.
func (f *Filter) Ignored(path string) bool {
	base := filepath.Base(path)
	for _, pattern := range f.patterns {
		// Try matching against the full path first.
		if matched, err := filepath.Match(pattern, path); err == nil && matched {
			return true
		}
		// Fall back to matching against just the file name.
		if matched, err := filepath.Match(pattern, base); err == nil && matched {
			return true
		}
	}
	return false
}

// Patterns returns a copy of the configured exclusion patterns.
func (f *Filter) Patterns() []string {
	out := make([]string, len(f.patterns))
	copy(out, f.patterns)
	return out
}
