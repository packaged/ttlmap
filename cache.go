package ttlmap

import (
	"sync"
	"time"
)

// A "thread" safe map of type string:Interface{}
// To avoid lock bottlenecks this map is dived to several (SHARD_COUNT) map shards.
type CacheMap struct {
	items   []*CacheMapShared
	options cacheOptions
}

// A "thread" safe string to anything map
type CacheMapShared struct {
	shutdown     chan bool
	cleanupCycle time.Duration
	items        map[string]*Item
	sync.RWMutex // Read Write mutex, guards access to internal map.
}

// Creates a new cache map
func New(opts ...CacheOption) CacheMap {

	cmp := CacheMap{options: defaultCacheOptions()}

	for _, opt := range opts {
		opt(&cmp.options)
	}

	cmp.items = make([]*CacheMapShared, cmp.options.shardCount)
	for i := 0; i < cmp.options.shardCount; i++ {
		cmp.items[i] = &CacheMapShared{items: make(map[string]*Item)}
		cmp.items[i].initCleanup(cmp.options.cleanupDuration)
	}
	return cmp
}

func (m CacheMap) Close() {
	for i := 0; i < m.options.shardCount; i++ {
		m.items[i].Close()
	}
}

// Close stops the cleanup
func (ms CacheMapShared) Close() {
	ms.shutdown <- true
}

// Returns shard under given key
func (m CacheMap) GetShard(key string) *CacheMapShared {
	return m.items[uint(fnv32(key))%uint(m.options.shardCount)]
}

func (m CacheMap) MSet(data map[string]interface{}, duration time.Duration) {
	for key, value := range data {
		shard := m.GetShard(key)
		shard.Lock()
		shard.items[key] = NewItem(value, duration)
		shard.Unlock()
	}
}

// Sets the given value under the specified key
func (m CacheMap) Set(key string, value interface{}, duration *time.Duration) {
	// Get map shard.
	shard := m.GetShard(key)
	shard.Lock()
	if duration == nil {
		duration = &m.options.defaultCacheDuration
	}
	shard.items[key] = NewItem(value, *duration)
	shard.Unlock()
}

// Retrieves an item from the map with the given key, and optionally increase its expiry time if found
func (m CacheMap) TouchGet(key string, touch bool) (interface{}, bool) {
	shard := m.GetShard(key)
	shard.RLock()
	// Get item from shard.
	val, ok := shard.items[key]
	var ret interface{}
	if ok {
		if val.Expired() {
			ok = false
		} else {
			if touch {
				val.Touch()
			}
			ret = val.GetValue()
		}
	}
	shard.RUnlock()
	return ret, ok
}

// Retrieves an item from the map with the given key, and increase its expiry time if found
func (m CacheMap) Get(key string) (interface{}, bool) {
	return m.TouchGet(key, true)
}

// Removes an element from the map
func (m CacheMap) Remove(key string) {
	shard := m.GetShard(key)
	if shard != nil {
		shard.Remove(key)
	}
}

// Removes an element from the map
func (ms CacheMapShared) Remove(key string) {
	ms.Lock()
	delete(ms.items, key)
	ms.Unlock()
}

// Has checks to see if an item exists
func (m CacheMap) Has(key string) bool {
	shard := m.GetShard(key)
	shard.RLock()
	val, ok := shard.items[key]
	if ok && val.Expired() {
		ok = false
	}
	shard.RUnlock()
	return ok
}

func (m CacheMap) GetExpiry(key string) *time.Time {
	shard := m.GetShard(key)
	shard.RLock()
	var expiry *time.Time
	val, ok := shard.items[key]
	if ok {
		expiry = val.expires
	}
	shard.RUnlock()
	return expiry
}
