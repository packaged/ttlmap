package ttlmap

import (
	"sync"
	"time"
)

// Item represents a record in the map
type Item struct {
	sync.RWMutex
	data    interface{}
	ttl     time.Duration
	expires *time.Time
}

func NewItem(value interface{}, duration time.Duration) *Item {
	i := &Item{}
	i.data = value
	i.ttl = duration
	expiry := time.Now().Add(duration)
	i.expires = &expiry
	return i
}

//Touch increases the expiry time on the item by the TTL
func (i *Item) Touch() {
	i.Lock()
	expiration := time.Now().Add(i.ttl)
	i.expires = &expiration
	i.Unlock()
}

//Expired returns if the item has passed its expiry time
func (i *Item) Expired() bool {
	var value bool
	i.RLock()
	if i.expires == nil {
		value = true
	} else {
		value = i.expires.Before(time.Now())
	}
	i.RUnlock()
	return value
}

//GetValue represents the value of the item in the map
func (i *Item) GetValue() interface{} {
	return i.data
}
