package ttlmap

import "time"

func (m CacheMap) Flush() {
	for i := 0; i < m.options.shardCount; i++ {
		m.items[i].Flush()
	}
}

func (ms *CacheMapShared) Flush() {
	ms.Lock()
	defer ms.Unlock()
	ms.items = make(map[string]*Item)
}

//Cleanup removes any expired items from the cache map
func (ms *CacheMapShared) Cleanup() {
	ms.Lock()
	defer ms.Unlock()
	for key, item := range ms.items {
		if item.Expired() {
			ms.remove(key)
		}
	}
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
