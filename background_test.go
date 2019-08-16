package ttlmap_test

import (
	"github.com/packaged/ttlmap"
	"log"
	"testing"
	"time"
)

func TestBackgroundUpdate(t *testing.T) {
	cache := ttlmap.New(ttlmap.WithDefaultTTL(10*time.Millisecond), ttlmap.WithMaxLifetime(time.Second))

	key := "test"
	updates := 0

	cache.Set(key, time.Now().UnixNano(), nil)
	for {
		if itm, ok := cache.GetItem(key); ok {
			if time.Now().Add(time.Millisecond * 6).After(itm.GetExpiry()) {
				if itm.Expired() {
					log.Print("Item Expired")
				} else {
					log.Print("Item Expiring")
				}
				go cache.BackgroundUpdate(key, func() (interface{}, error) {
					log.Print("Background Updating")
					time.Sleep(time.Millisecond * 5)
					updates++
					log.Print("Background Updated")
					return time.Now().UnixNano(), nil
				})
			} else {
				log.Print("Item All Good ", itm.GetValue())
			}
			time.Sleep(time.Millisecond)
			if time.Now().After(itm.GetDeadline()) {
				break
			}
			if updates > 10 {
				break
			}
		}
	}
	if updates == 0 {
		t.Error("No updates were executed")
	}
}
