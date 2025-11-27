package ttlmap_test

import (
	"log"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/packaged/ttlmap"
)

func TestFetch_MissSingleflight(t *testing.T) {
	cache := ttlmap.New(ttlmap.WithDefaultTTL(100 * time.Millisecond))

	var calls int32
	source := func(key string) (int, error) {
		time.Sleep(10 * time.Millisecond)
		atomic.AddInt32(&calls, 1)
		return 123, nil
	}

	// Launch many concurrent callers for a missing key
	const n = 32
	var wg sync.WaitGroup
	wg.Add(n)
	results := make([]int, n)
	for i := 0; i < n; i++ {
		go func(ix int) {
			defer wg.Done()
			v, err := ttlmap.Fetch[int](cache, "k1", source)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			results[ix] = v
		}(i)
	}
	wg.Wait()

	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected single source call, got %d", calls)
	}
	for i := 0; i < n; i++ {
		if results[i] != 123 {
			t.Fatalf("unexpected result at %d: %d", i, results[i])
		}
	}
}

func TestFetch_TypeMismatch(t *testing.T) {
	cache := ttlmap.New(ttlmap.WithDefaultTTL(100 * time.Millisecond))
	cache.Set("k", "hello", nil)

	var called int32
	src := func(key string) (int, error) {
		atomic.AddInt32(&called, 1)
		return 42, nil
	}
	_, err := ttlmap.Fetch[int](cache, "k", src)
	if err == nil {
		t.Fatalf("expected ErrTypeMismatch, got nil")
	}
	if err != ttlmap.ErrTypeMismatch {
		t.Fatalf("expected ErrTypeMismatch, got %v", err)
	}
	if atomic.LoadInt32(&called) != 0 {
		t.Fatalf("source should not be called on type mismatch hit")
	}
}

func TestFetch_StaleWhileRevalidate(t *testing.T) {
	// Short TTL to trigger expiry
	cache := ttlmap.New(ttlmap.WithDefaultTTL(time.Second), ttlmap.WithMaxLifetime(2*time.Second))
	ttl := 20 * time.Millisecond
	cache.Set("sk", 1, &ttl)

	src := func(key string) (int, error) {
		time.Sleep(15 * time.Millisecond)
		return 2, nil
	}
	// First call should return stale value 1 and trigger background refresh
	v1, err := ttlmap.Fetch[int](cache, "sk", src)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if v1 != 1 {
		t.Fatalf("expected stale value 1, got %d", v1)
	}

	wg := sync.WaitGroup{}
	for i := 0; i < 100000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, fetchErr := ttlmap.Fetch[int](cache, "sk", src)
			if fetchErr != nil {
				log.Print(fetchErr.Error())
			}
		}()
	}
	wg.Wait()

	time.Sleep(10 * time.Millisecond)
	// Next call should observe refreshed value 2
	v2, err := ttlmap.Fetch[int](cache, "sk", src)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if v2 != 2 {
		t.Fatalf("expected refreshed value 2, got %d", v2)
	}
}
