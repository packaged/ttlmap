package ttlmap_test

import (
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/packaged/ttlmap"
)

// helper to prepopulate cache with n keys
func prepopulate(c ttlmap.CacheMap, n int) []string {
	keys := make([]string, n)
	for i := 0; i < n; i++ {
		k := "k" + strconv.Itoa(i)
		c.Set(k, i, nil)
		keys[i] = k
	}
	return keys
}

func BenchmarkSet(b *testing.B) {
	cache := ttlmap.New(ttlmap.WithCleanupDuration(time.Hour)) // disable cleanup overhead
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set("key"+strconv.Itoa(i), i, nil)
	}
}

func BenchmarkGetHit(b *testing.B) {
	cache := ttlmap.New(ttlmap.WithCleanupDuration(time.Hour))
	keys := prepopulate(cache, 1024)
	// rotate through keys
	b.ResetTimer()
	idx := 0
	for i := 0; i < b.N; i++ {
		if idx == len(keys) {
			idx = 0
		}
		cache.Get(keys[idx])
		idx++
	}
}

func BenchmarkGetMiss(b *testing.B) {
	cache := ttlmap.New(ttlmap.WithCleanupDuration(time.Hour))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get("missing-" + strconv.Itoa(i))
	}
}

func BenchmarkParallelSet(b *testing.B) {
	cache := ttlmap.New(ttlmap.WithCleanupDuration(time.Hour))
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		// local counter to avoid contention on atomics
		j := 0
		for pb.Next() {
			cache.Set("pset-"+strconv.Itoa(j), j, nil)
			j++
		}
	})
}

func BenchmarkParallelGetHit(b *testing.B) {
	cache := ttlmap.New(ttlmap.WithCleanupDuration(time.Hour))
	keys := prepopulate(cache, 4096)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		for pb.Next() {
			cache.Get(keys[r.Intn(len(keys))])
		}
	})
}

func BenchmarkMixedRW(b *testing.B) {
	cache := ttlmap.New(ttlmap.WithCleanupDuration(time.Hour))
	keys := prepopulate(cache, 2048)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		for pb.Next() {
			if r.Intn(10) == 0 { // ~10% writes
				k := keys[r.Intn(len(keys))]
				cache.Set(k, r.Int(), nil)
			} else {
				cache.Get(keys[r.Intn(len(keys))])
			}
		}
	})
}
