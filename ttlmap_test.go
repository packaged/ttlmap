package ttlmap_test

import (
	"sync"
	"testing"
	"time"

	"github.com/packaged/ttlmap"
)

func TestHasAndExpiry(t *testing.T) {
	cache := ttlmap.New(ttlmap.WithDefaultTTL(50*time.Millisecond), ttlmap.WithCleanupDuration(10*time.Millisecond))
	if cache.Has("missing") {
		t.Fatalf("expected Has to be false for missing key")
	}
	cache.Set("a", 1, nil)
	if !cache.Has("a") {
		t.Fatalf("expected Has to be true right after Set")
	}
	time.Sleep(70 * time.Millisecond)
	if cache.Has("a") {
		t.Fatalf("expected Has to be false after TTL expiry")
	}
}

func TestGetExpiryAndGetItemCopy(t *testing.T) {
	ttl := 80 * time.Millisecond
	cache := ttlmap.New(ttlmap.WithDefaultTTL(ttl), ttlmap.WithCleanupDuration(10*time.Millisecond))
	cache.Set("k", "v", &ttl)

	exp1 := cache.GetExpiry("k")
	if exp1 == nil {
		t.Fatalf("expected expiry for existing key")
	}
	// Missing key should return nil
	if exp := cache.GetExpiry("missing"); exp != nil {
		t.Fatalf("expected nil expiry for missing key")
	}

	// Get a copy and Touch it; the underlying stored item should not change expiry
	itm, ok := cache.GetItem("k")
	if !ok || itm == nil {
		t.Fatalf("expected GetItem to succeed")
	}
	// Touch the returned copy
	itm.Touch()
	exp2 := cache.GetExpiry("k")
	if exp2 == nil || !exp2.Equal(*exp1) {
		t.Fatalf("expected stored item expiry to remain unchanged after touching copy")
	}
}

func TestFlush(t *testing.T) {
	cache := ttlmap.New(ttlmap.WithDefaultTTL(500 * time.Millisecond))
	cache.Set("x", 1, nil)
	cache.Set("y", 2, nil)
	if !cache.Has("x") || !cache.Has("y") {
		t.Fatalf("expected keys to exist before Flush")
	}
	cache.Flush()
	if cache.Has("x") || cache.Has("y") {
		t.Fatalf("expected keys to be removed after Flush")
	}
}

func TestCloseIdempotent(t *testing.T) {
	cache := ttlmap.New(ttlmap.WithCleanupDuration(5 * time.Millisecond))
	// Should not panic on repeated Close
	cache.Close()
	cache.Close()
}

func TestBackgroundUpdateSingleFlight(t *testing.T) {
	cache := ttlmap.New(ttlmap.WithDefaultTTL(20 * time.Millisecond))
	key := "sf"
	cache.Set(key, 1, nil)

	var calls int
	var mu sync.Mutex
	start := make(chan struct{})
	var wg sync.WaitGroup
	// Launch many contenders; only one updater should run at a time for the key
	n := 50
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			<-start
			cache.BackgroundUpdate(key, func() (interface{}, error) {
				mu.Lock()
				calls++
				mu.Unlock()
				time.Sleep(5 * time.Millisecond)
				return 2, nil
			})
		}()
	}
	close(start)
	wg.Wait()

	// Only one should have executed; the rest return immediately because the key is locked
	if calls != 1 {
		t.Fatalf("expected exactly one updater call, got %d", calls)
	}
}
