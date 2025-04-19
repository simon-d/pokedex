package pokecache

import (
	"sync"
	"time"
)

type cacheEntry struct {
	CreatedAt time.Time
	Val       []byte
}

type Cache struct {
	entries map[string]cacheEntry
	mu      sync.Mutex
}

func NewCache(interval time.Duration) *Cache {
	ticker := time.NewTicker(interval)
	// defer ticker.Stop()

	cache := Cache{
		entries: map[string]cacheEntry{},
	}

	// interval call reap
	go func() {
		for range ticker.C {
			// fmt.Println("reaping...")
			cache.reapLoop(interval)
		}
	}()
	return &cache
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for cKey, cVal := range c.entries {
		if cKey == key {
			// fmt.Printf("Cache hit: %s\n", key)
			return cVal.Val, true
		}
	}

	// fmt.Printf("Cache miss: %s\n", key)
	return nil, false
}

func (c *Cache) Add(key string, val []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = cacheEntry{
		Val:       val,
		CreatedAt: time.Now(),
	}

	// fmt.Printf("Entry added. Key: %s\n", key)
}

func (c *Cache) reapLoop(interval time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, entry := range c.entries {

		now := time.Now()
		diff := now.Sub(entry.CreatedAt)

		// fmt.Printf("Diff: %s\n", diff)
		if diff >= interval {
			delete(c.entries, key)
		}
	}
}
