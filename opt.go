package ttlmap

import "time"

type cacheOptions struct {
	cleanupDuration      time.Duration
	defaultCacheDuration time.Duration
	maxLifetime          time.Duration
	shardCount           int
}

func defaultCacheOptions() cacheOptions {
	return cacheOptions{
		cleanupDuration:      time.Minute,
		defaultCacheDuration: time.Hour,
		maxLifetime:          365 * (24 * time.Hour),
		shardCount:           32,
	}
}

// CacheOption configures how we set up the cache map
type CacheOption func(options *cacheOptions)

// WithShardSize With a custom sub map shard size
func WithShardSize(shardSize int) CacheOption {
	return func(o *cacheOptions) {
		o.shardCount = shardSize
	}
}

// WithDefaultTTL Sets the default duration for items stored
func WithDefaultTTL(ttl time.Duration) CacheOption {
	return func(o *cacheOptions) {
		o.defaultCacheDuration = ttl
	}
}

// WithCleanupDuration Sets how frequently to cleanup expired items
func WithCleanupDuration(ttl time.Duration) CacheOption {
	return func(o *cacheOptions) {
		o.cleanupDuration = ttl
	}
}

// WithMaxLifetime Sets the maximum amount of time an item can exist within the cache
func WithMaxLifetime(ttl time.Duration) CacheOption {
	return func(o *cacheOptions) {
		o.maxLifetime = ttl
	}
}
