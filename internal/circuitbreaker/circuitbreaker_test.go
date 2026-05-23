package circuitbreaker

import (
	"testing"
	"time"
)

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestBreaker_InitiallyAllows(t *testing.T) {
	br := New(3, 10*time.Second)
	if err := br.Allow(); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestBreaker_OpensAfterThreshold(t *testing.T) {
	br := New(3, 10*time.Second)
	for i := 0; i < 3; i++ {
		br.RecordFailure()
	}
	if err := br.Allow(); err == nil {
		t.Fatal("expected circuit open error, got nil")
	}
	if br.StateSnapshot() != StateOpen {
		t.Fatalf("expected StateOpen, got %v", br.StateSnapshot())
	}
}

func TestBreaker_RemainsOpenBeforeCooldown(t *testing.T) {
	now := time.Now()
	br := New(2, 30*time.Second)
	br.now = fixedClock(now)
	br.RecordFailure()
	br.RecordFailure()

	// advance only 5 seconds — still within cooldown
	br.now = fixedClock(now.Add(5 * time.Second))
	if err := br.Allow(); err == nil {
		t.Fatal("expected circuit to remain open")
	}
}

func TestBreaker_HalfOpenAfterCooldown(t *testing.T) {
	now := time.Now()
	br := New(2, 10*time.Second)
	br.now = fixedClock(now)
	br.RecordFailure()
	br.RecordFailure()

	br.now = fixedClock(now.Add(11 * time.Second))
	if err := br.Allow(); err != nil {
		t.Fatalf("expected half-open to allow, got %v", err)
	}
	if br.StateSnapshot() != StateHalfOpen {
		t.Fatalf("expected StateHalfOpen, got %v", br.StateSnapshot())
	}
}

func TestBreaker_SuccessCloses(t *testing.T) {
	br := New(2, 10*time.Second)
	br.RecordFailure()
	br.RecordFailure()
	br.RecordSuccess()
	if br.StateSnapshot() != StateClosed {
		t.Fatalf("expected StateClosed after success, got %v", br.StateSnapshot())
	}
	if err := br.Allow(); err != nil {
		t.Fatalf("expected allow after close, got %v", err)
	}
}

func TestBreaker_FailureInHalfOpenReopens(t *testing.T) {
	now := time.Now()
	br := New(2, 5*time.Second)
	br.now = fixedClock(now)
	br.RecordFailure()
	br.RecordFailure()

	// move past cooldown to enter half-open
	br.now = fixedClock(now.Add(6 * time.Second))
	_ = br.Allow() // transitions to half-open

	br.RecordFailure() // single failure re-opens
	if br.StateSnapshot() != StateOpen {
		t.Fatalf("expected StateOpen after half-open failure, got %v", br.StateSnapshot())
	}
}
