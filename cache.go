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
	shutdown     chan struct{}
	ticker       *time.Ticker
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

// Close stops the cleanup background goroutine and ticker for this shard
func (ms *CacheMapShared) Close() {
	// Protect against double-close: close is safe only once; recover if already closed.
	if ms.shutdown != nil {
		select {
		case <-ms.shutdown:
			// already closed
		default:
			close(ms.shutdown)
		}
	}
	if ms.ticker != nil {
		ms.ticker.Stop()
	}
}

// Returns shard under given key
func (m CacheMap) GetShard(key string) *CacheMapShared {
	return m.items[uint(fnv32(key))%uint(m.options.shardCount)]
}

func (m CacheMap) MSet(data map[string]interface{}, duration time.Duration) {
	for key, value := range data {
		shard := m.GetShard(key)
		shard.Lock()
		shard.items[key] = newItem(value, duration, time.Now().Add(m.options.maxLifetime), nil)
		shard.Unlock()
	}
}

func (m CacheMap) SetWithCleanup(key string, value interface{}, duration *time.Duration, cleanup func(*Item)) {
	// Get map shard.
	shard := m.GetShard(key)
	shard.Lock()
	if duration == nil {
		duration = &m.options.defaultCacheDuration
	}
	itm := newItem(value, *duration, time.Now().Add(m.options.maxLifetime), cleanup)
	shard.items[key] = itm
	shard.Unlock()
}

// Sets the given value under the specified key
func (m CacheMap) Set(key string, value interface{}, duration *time.Duration) {
	m.SetWithCleanup(key, value, duration, nil)
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

// Retrieves an item from the map with the given key, and increase its expiry time if found
func (m CacheMap) GetItem(key string) (*Item, bool) {
	shard := m.GetShard(key)
	shard.RLock()
	defer shard.RUnlock()
	if val, ok := shard.items[key]; ok {
		return &Item{
			data:     val.data,
			deadline: val.deadline,
			ttl:      val.ttl,
			expires:  val.expires,
		}, true
	}
	return nil, false
}

// Removes an element from the map
func (m CacheMap) Remove(key string) {
	shard := m.GetShard(key)
	if shard != nil {
		shard.Remove(key)
	}
}

// Removes an element from the map
func (ms *CacheMapShared) Remove(key string) {
	ms.Lock()
	ms.remove(key)
	ms.Unlock()
}

// Removes an element from the map
func (ms *CacheMapShared) remove(key string) {
	if itm, ok := ms.items[key]; ok && itm.onDelete != nil {
		itm.onDelete(itm)
	}

	delete(ms.items, key)
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
