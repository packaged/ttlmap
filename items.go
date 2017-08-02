package ttlmap

// Returns all items as map[string]interface{}
func (m CacheMap) Items() map[string]interface{} {
	tmp := make(map[string]interface{})

	for i := 0; i < m.options.shardCount; i++ {
		shard := m.items[i]
		shard.RLock()
		for key, itm := range shard.items {
			if !itm.Expired() {
				tmp[key] = itm.GetValue()
			}
		}
		shard.RUnlock()
	}
	return tmp
}
