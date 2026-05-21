package filter_test

import (
	"testing"

	"github.com/yourorg/driftwatch/internal/filter"
)

func TestIgnored_ExactPath(t *testing.T) {
	f := filter.New([]string{"/etc/hosts"})
	if !f.Ignored("/etc/hosts") {
		t.Error("expected /etc/hosts to be ignored")
	}
}

func TestIgnored_GlobExtension(t *testing.T) {
	f := filter.New([]string{"*.tmp"})
	if !f.Ignored("/var/run/state.tmp") {
		t.Error("expected *.tmp glob to match state.tmp")
	}
	if f.Ignored("/etc/nginx/nginx.conf") {
		t.Error("expected nginx.conf not to be ignored")
	}
}

func TestIgnored_BaseNameMatch(t *testing.T) {
	f := filter.New([]string{".DS_Store"})
	if !f.Ignored("/some/deep/path/.DS_Store") {
		t.Error("expected base name match to trigger exclusion")
	}
}

func TestIgnored_NoPatterns(t *testing.T) {
	f := filter.New(nil)
	if f.Ignored("/etc/passwd") {
		t.Error("expected no patterns to ignore nothing")
	}
}

func TestIgnored_EmptyAndWhitespacePatterns(t *testing.T) {
	f := filter.New([]string{"", "   ", "\t"})
	if f.Ignored("/etc/passwd") {
		t.Error("blank patterns should be stripped and ignore nothing")
	}
	if len(f.Patterns()) != 0 {
		t.Errorf("expected 0 patterns, got %d", len(f.Patterns()))
	}
}

func TestIgnored_MultiplePatterns(t *testing.T) {
	f := filter.New([]string{"*.bak", "/tmp/*", "secret.conf"})
	cases := []struct {
		path    string
		want    bool
	}{
		{"/etc/nginx/nginx.conf.bak", true},
		{"/etc/app/secret.conf", true},
		{"/etc/app/normal.conf", false},
	}
	for _, tc := range cases {
		got := f.Ignored(tc.path)
		if got != tc.want {
			t.Errorf("Ignored(%q) = %v, want %v", tc.path, got, tc.want)
		}
	}
}

func TestPatterns_ReturnsCopy(t *testing.T) {
	f := filter.New([]string{"*.log"})
	p := f.Patterns()
	p[0] = "mutated"
	if f.Patterns()[0] != "*.log" {
		t.Error("Patterns() should return a copy, not expose internal slice")
	}
}
