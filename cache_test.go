package ttlmap_test

import (
	"github.com/packaged/ttlmap"
	"testing"
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

func TestMaxLifetime(t *testing.T) {
	cache := ttlmap.New(ttlmap.WithMaxLifetime(time.Millisecond * 100))

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

	//Check to see if max lifetime has killed the item
	time.Sleep(time.Millisecond * 200)
	data, exists = cache.Get("hello")
	if exists || data != nil {
		t.Errorf("Expected empty cache to return no data")
	}
}

func TestItems(t *testing.T) {
	dur := time.Millisecond * 100

	cache := ttlmap.New()
	cache.Set("item1", "one", nil)
	cache.Set("item2", "two", nil)
	cache.Set("item3", "three", &dur)
	cache.Set("item4", "four", nil)

	if len(cache.Items()) != 4 {
		t.Errorf("Expected cache to return 4 items")
	}

	time.Sleep(dur)

	if len(cache.Items()) != 3 {
		t.Errorf("Expected cache to return 3 items after cache expiry")
	}
}
