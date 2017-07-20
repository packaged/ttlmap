package ttlmap_test

import (
	"testing"
	"github.com/packaged/ttlmap"
	"time"
)

func TestGet(t *testing.T) {
	cache := ttlmap.New(ttlmap.WithDefaultTTL(300*time.Millisecond), ttlmap.WithCleanupDuration(time.Millisecond*100))

	data, exists := cache.Get("hello")
	if exists || data != nil {
		t.Errorf("Expected empty cache to return no data")
	}

	cache.Set("hello", "world", nil)
	data, exists = cache.Get("hello")
	if !exists {
		t.Errorf("Expected cache to return data for `hello`")
	}
	if data.(string) != "world" {
		t.Errorf("Expected cache to return `world` for `hello`")
	}

	//Check to see if cleanup is clearing unexpired items
	time.Sleep(time.Millisecond * 200)
	data, exists = cache.Get("hello")
	if !exists || data == nil {
		t.Errorf("Expected cache to return data")
	}

	//Check Cache is re-touching after a get
	time.Sleep(time.Millisecond * 200)
	data, exists = cache.Get("hello")
	if !exists || data == nil {
		t.Errorf("Expected cache to return data")
	}

	//Check Cache is optionally re-touching after a get
	time.Sleep(time.Millisecond * 200)
	data, exists = cache.TouchGet("hello", false)
	if !exists || data == nil {
		t.Errorf("Expected cache to return data")
	}

	//Make sure cache clears after expiry
	time.Sleep(time.Millisecond * 200)
	data, exists = cache.Get("hello")
	if exists || data != nil {
		t.Errorf("Expected empty cache to return no data")
	}
}