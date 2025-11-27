package ttlmap

import "time"

func (m CacheMap) Flush() {
	for i := 0; i < m.options.shardCount; i++ {
		m.items[i].Flush()
	}
}

func (ms *CacheMapShared) Flush() {
	ms.Lock()
	ms.items = make(map[string]*Item)
	ms.Unlock()
}

// Cleanup removes any expired items from the cache map
func (ms *CacheMapShared) Cleanup() {
	ms.Lock()
	for key, item := range ms.items {
		if item.Expired() {
			ms.remove(key)
		}
	}
	ms.Unlock()
}

func (ms *CacheMapShared) initCleanup(dur time.Duration) {
	// Initialize shutdown channel once
	if ms.shutdown == nil {
		ms.shutdown = make(chan struct{})
	}
	// Use NewTicker so we can Stop it later to avoid ticker leaks
	ms.ticker = time.NewTicker(dur)
	go func() {
		for {
			select {
			case <-ms.shutdown:
				return
			case <-ms.ticker.C:
				ms.Cleanup()
			}
		}
	}()
}
