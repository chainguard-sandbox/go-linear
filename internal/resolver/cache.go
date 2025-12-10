package resolver

import (
	"sync"
	"time"
)

// cacheEntry represents a cached value with expiration.
type cacheEntry struct {
	value      string
	expiration time.Time
}

// Cache provides a simple TTL-based cache for name→ID resolution.
type Cache struct {
	mu      sync.RWMutex
	entries map[string]cacheEntry
	ttl     time.Duration
}

// NewCache creates a new cache with the specified TTL.
func NewCache(ttl time.Duration) *Cache {
	return &Cache{
		entries: make(map[string]cacheEntry),
		ttl:     ttl,
	}
}

// Get retrieves a value from the cache.
// Returns empty string if not found or expired.
func (c *Cache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.entries[key]
	if !ok {
		return "", false
	}

	// Check if expired
	if time.Now().After(entry.expiration) {
		return "", false
	}

	return entry.value, true
}

// Set stores a value in the cache with TTL.
func (c *Cache) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = cacheEntry{
		value:      value,
		expiration: time.Now().Add(c.ttl),
	}
}

// Clear removes all entries from the cache.
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]cacheEntry)
}
