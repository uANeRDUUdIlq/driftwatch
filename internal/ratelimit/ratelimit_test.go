package ratelimit

import (
	"testing"
	"time"
)

func fixedNow(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestAllow_UnderLimit(t *testing.T) {
	l := New(3, time.Minute)
	for i := 0; i < 3; i++ {
		if !l.Allow("file.conf") {
			t.Fatalf("expected allow on call %d", i+1)
		}
	}
}

func TestAllow_ExceedsLimit(t *testing.T) {
	l := New(2, time.Minute)
	l.Allow("k")
	l.Allow("k")
	if l.Allow("k") {
		t.Fatal("expected deny on third call")
	}
}

func TestAllow_WindowExpiry(t *testing.T) {
	base := time.Now()
	l := New(2, time.Minute)
	l.nowFunc = fixedNow(base)

	l.Allow("k")
	l.Allow("k")

	// Advance time past the window — old events should be pruned.
	l.nowFunc = fixedNow(base.Add(61 * time.Second))
	if !l.Allow("k") {
		t.Fatal("expected allow after window expiry")
	}
}

func TestAllow_IndependentKeys(t *testing.T) {
	l := New(1, time.Minute)
	if !l.Allow("a") {
		t.Fatal("expected allow for key a")
	}
	if !l.Allow("b") {
		t.Fatal("expected allow for key b")
	}
	if l.Allow("a") {
		t.Fatal("expected deny for key a on second call")
	}
}

func TestReset(t *testing.T) {
	l := New(1, time.Minute)
	l.Allow("x")
	if l.Allow("x") {
		t.Fatal("expected deny before reset")
	}
	l.Reset("x")
	if !l.Allow("x") {
		t.Fatal("expected allow after reset")
	}
}
