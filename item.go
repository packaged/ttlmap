package ttlmap

import (
	"sync"
	"time"
)

// Item represents a record in the map
type Item struct {
	sync.RWMutex
	data     interface{}
	deadline time.Time
	ttl      time.Duration
	expires  *time.Time
}

func newItem(value interface{}, duration time.Duration, deadline time.Time) *Item {
	i := &Item{
		data:     value,
		ttl:      duration,
		deadline: deadline,
	}
	expiry := time.Now().Add(duration)
	i.expires = &expiry
	return i
}

//Touch increases the expiry time on the item by the TTL
func (i *Item) Touch() {
	i.Lock()
	defer i.Unlock()
	expiration := time.Now().Add(i.ttl)
	i.expires = &expiration
}

//Expired returns if the item has passed its expiry time
func (i *Item) Expired() bool {
	var value bool
	i.RLock()
	defer i.RUnlock()
	if i.expires == nil || i.deadline.Before(time.Now()) {
		value = true
	} else {
		value = i.expires.Before(time.Now())
	}
	return value
}

//GetValue represents the value of the item in the map
func (i *Item) GetValue() interface{} {
	return i.data
}
