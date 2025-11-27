package ttlmap

import (
	"sync"
	"time"
)

// Item represents a record in the map
type Item struct {
	sync.RWMutex
	updateMutex sync.RWMutex
	isUpdating  bool
	data        interface{}
	deadline    time.Time
	ttl         time.Duration
	expires     *time.Time
	onDelete    func(*Item)
}

func newItem(value interface{}, duration time.Duration, deadline time.Time, onDelete func(*Item)) *Item {
	i := &Item{
		data:     value,
		ttl:      duration,
		deadline: deadline,
		onDelete: onDelete,
	}
	expiry := time.Now().Add(duration)
	i.expires = &expiry
	return i
}

// Touch increases the expiry time on the item by the TTL
func (i *Item) Touch() {
	i.Lock()
	expiration := time.Now().Add(i.ttl)
	i.expires = &expiration
	i.Unlock()
}

// Expired returns if the item has passed its expiry time
func (i *Item) Expired() bool {
	var value bool
	i.RLock()
	if i.expires == nil || i.deadline.Before(time.Now()) {
		value = true
	} else {
		value = i.expires.Before(time.Now())
	}
	i.RUnlock()
	return value
}

// GetValue represents the value of the item in the map
func (i *Item) GetValue() interface{} {
	return i.data
}

func (i *Item) GetExpiry() time.Time {
	if i.expires == nil {
		return i.deadline
	}
	return *i.expires
}

func (i *Item) GetDeadline() time.Time {
	return i.deadline
}
