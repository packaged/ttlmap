package ttlmap

import (
	"errors"
	"time"
)

// ErrTypeMismatch is returned when the cached value cannot be cast to the requested generic type.
var ErrTypeMismatch = errors.New("ttlmap: cached value has different type")

// Fetch returns a strictly typed value from the cache, fetching from the provided source function when missing.
func Fetch[T any](m CacheMap, key string, source func(string) (T, error)) (T, error) {
	var zero T
	var returnValue T
	var okCast bool

	shard := m.GetShard(key)
	shard.RLock()
	itm, ok := shard.items[key]
	if ok {
		returnValue, okCast = itm.GetValue().(T)
		shard.RUnlock()
		if !okCast {
			return zero, ErrTypeMismatch
		}

		if !itm.Expired() {
			return returnValue, nil
		}

		if !itm.isUpdating && itm.updateMutex.TryLock() {
			itm.isUpdating = true
			value, err := source(key)
			if err == nil {
				m.Set(key, value, nil)
			}
			itm.updateMutex.Unlock()
			itm.isUpdating = false
			return value, nil
		}

		// Item has expired, but another thread is updateMutex
		return returnValue, nil
	}
	shard.RUnlock()
	shard.Lock()
	defer shard.Unlock()

	itm, ok = shard.items[key]
	if ok {
		// check the value was not already processed when waiting for the lock
		returnValue, okCast = itm.GetValue().(T)
		if !okCast {
			return zero, ErrTypeMismatch
		}
		return returnValue, nil
	}

	value, err := source(key)
	if err == nil {
		itm = newItem(value, m.options.defaultCacheDuration, time.Now().Add(m.options.maxLifetime), nil)
		shard.items[key] = itm
	}
	return value, err
}
