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
	ticker := time.Tick(dur)
	go (func() {
		for {
			select {
			case <-ms.shutdown:
				return
			case <-ticker:
				ms.Cleanup()
			}
		}
	})()
}
