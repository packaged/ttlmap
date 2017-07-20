## usage

Import the package:

```go
import (
    "github.com/packaged/ttlmap"
)

```

```bash
go get "github.com/packaged/ttlmap"
```

## example

```go

	// Create a new map.
	cache := ttlmap.New()
	
	// Sets item within map, sets "bar" under key "foo" expiring after the default cache time (1 hour)
	cache.Set("foo", "bar", nil)
	
	// Sets item within map, sets "custom" under key "xyz" expiring after 3 minutes
	cache.Set("xyz", "custom", time.Minute*3)

	// Retrieve item from map.
	if tmp, ok := cache.Get("foo"); ok {
		bar := tmp.(string)
	}

	// Removes item under key "foo"
	cache.Remove("foo")

```
