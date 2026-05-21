package suppression

import (
	"testing"
	"time"
)

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestIsSuppressed_NotSuppressed(t *testing.T) {
	s := New()
	if s.IsSuppressed("/etc/app/config.yaml") {
		t.Fatal("expected path to not be suppressed")
	}
}

func TestIsSuppressed_ActiveWindow(t *testing.T) {
	base := time.Now()
	s := New()
	s.now = fixedClock(base)

	s.Suppress("/etc/app/config.yaml", 10*time.Minute)
	s.now = fixedClock(base.Add(5 * time.Minute))

	if !s.IsSuppressed("/etc/app/config.yaml") {
		t.Fatal("expected path to be suppressed within window")
	}
}

func TestIsSuppressed_ExpiredWindow(t *testing.T) {
	base := time.Now()
	s := New()
	s.now = fixedClock(base)

	s.Suppress("/etc/app/config.yaml", 5*time.Minute)
	s.now = fixedClock(base.Add(10 * time.Minute))

	if s.IsSuppressed("/etc/app/config.yaml") {
		t.Fatal("expected suppression to have expired")
	}
}

func TestSuppress_ExtendsWindow(t *testing.T) {
	base := time.Now()
	s := New()
	s.now = fixedClock(base)

	s.Suppress("/etc/hosts", 5*time.Minute)
	s.Suppress("/etc/hosts", 20*time.Minute)
	s.now = fixedClock(base.Add(15 * time.Minute))

	if !s.IsSuppressed("/etc/hosts") {
		t.Fatal("expected extended suppression to still be active")
	}
}

func TestLift_RemovesSuppression(t *testing.T) {
	s := New()
	s.Suppress("/etc/hosts", time.Hour)
	s.Lift("/etc/hosts")

	if s.IsSuppressed("/etc/hosts") {
		t.Fatal("expected suppression to be lifted")
	}
}

func TestActiveCount(t *testing.T) {
	base := time.Now()
	s := New()
	s.now = fixedClock(base)

	s.Suppress("/a", 10*time.Minute)
	s.Suppress("/b", 10*time.Minute)
	s.Suppress("/c", 2*time.Minute)

	s.now = fixedClock(base.Add(5 * time.Minute))

	if got := s.ActiveCount(); got != 2 {
		t.Fatalf("expected 2 active suppressions, got %d", got)
	}
}
