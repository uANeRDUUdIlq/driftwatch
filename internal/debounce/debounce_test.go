package debounce_test

import (
	"sync"
	"testing"
	"time"

	"github.com/yourorg/driftwatch/internal/debounce"
)

func TestDebounce_FiresAfterQuiet(t *testing.T) {
	var mu sync.Mutex
	called := 0

	db := debounce.New(50*time.Millisecond, func(key string) {
		mu.Lock()
		called++
		mu.Unlock()
	})

	db.Trigger("file.conf")
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	if called != 1 {
		t.Errorf("expected 1 call, got %d", called)
	}
	mu.Unlock()
}

func TestDebounce_CoalescesRapidTriggers(t *testing.T) {
	var mu sync.Mutex
	called := 0

	db := debounce.New(80*time.Millisecond, func(key string) {
		mu.Lock()
		called++
		mu.Unlock()
	})

	// Fire rapidly — only one call should result.
	for i := 0; i < 5; i++ {
		db.Trigger("file.conf")
		time.Sleep(10 * time.Millisecond)
	}

	time.Sleep(150 * time.Millisecond)

	mu.Lock()
	if called != 1 {
		t.Errorf("expected 1 coalesced call, got %d", called)
	}
	mu.Unlock()
}

func TestDebounce_IndependentKeys(t *testing.T) {
	var mu sync.Mutex
	keys := map[string]int{}

	db := debounce.New(50*time.Millisecond, func(key string) {
		mu.Lock()
		keys[key]++
		mu.Unlock()
	})

	db.Trigger("a.conf")
	db.Trigger("b.conf")
	time.Sleep(120 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if keys["a.conf"] != 1 || keys["b.conf"] != 1 {
		t.Errorf("expected 1 call per key, got %v", keys)
	}
}

func TestDebounce_StopCancelsPending(t *testing.T) {
	called := 0

	db := debounce.New(100*time.Millisecond, func(key string) {
		called++
	})

	db.Trigger("file.conf")
	db.Stop()
	time.Sleep(150 * time.Millisecond)

	if called != 0 {
		t.Errorf("expected 0 calls after Stop, got %d", called)
	}
}
