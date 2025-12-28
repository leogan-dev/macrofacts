package foods

import (
	"sync"
	"time"
)

// barcodeCache is a tiny, boring in-memory TTL cache.
// It caches both hits and misses (negative caching) so barcode lookups don't hammer Mongo.
// Not an LRU: we keep it simple and cap size with coarse eviction.
type barcodeCache struct {
	mu sync.Mutex
	max int
	posTTL time.Duration
	negTTL time.Duration
	m map[string]cacheEntry
}

type cacheEntry struct {
	dto      *FoodDTO
	expires  time.Time
}

func newBarcodeCache(max int, posTTL, negTTL time.Duration) *barcodeCache {
	if max <= 0 {
		max = 10000
	}
	return &barcodeCache{max: max, posTTL: posTTL, negTTL: negTTL, m: make(map[string]cacheEntry, max)}
}

func (c *barcodeCache) get(code string) (*FoodDTO, bool) {
	if c == nil {
		return nil, false
	}
	now := time.Now()
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.m[code]
	if !ok {
		return nil, false
	}
	if now.After(e.expires) {
		delete(c.m, code)
		return nil, false
	}
	// dto can be nil to represent a cached "not found".
	return e.dto, true
}

func (c *barcodeCache) set(code string, dto *FoodDTO) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.evictIfNeededLocked()
	c.m[code] = cacheEntry{dto: dto, expires: time.Now().Add(c.posTTL)}
}

func (c *barcodeCache) setNotFound(code string) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.evictIfNeededLocked()
	c.m[code] = cacheEntry{dto: nil, expires: time.Now().Add(c.negTTL)}
}

func (c *barcodeCache) evictIfNeededLocked() {
	// coarse eviction: if size exceeds max, clear ~10% oldest-ish by simply deleting arbitrary keys.
	if len(c.m) < c.max {
		return
	}
	n := c.max / 10
	if n < 100 {
		n = 100
	}
	for k := range c.m {
		delete(c.m, k)
		n--
		if n <= 0 {
			break
		}
	}
}
