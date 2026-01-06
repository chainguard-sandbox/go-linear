package resolver

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/codeGROOVE-dev/multicache"
	"github.com/codeGROOVE-dev/multicache/pkg/store/localfs"
)

// Cache provides a file-backed TTL cache for name→ID resolution.
// Uses multicache with local filesystem persistence to survive across
// MCP subprocess calls (ophis spawns new process per tool invocation).
type Cache struct {
	tiered   *multicache.TieredCache[string, string]
	inmemory *multicache.Cache[string, string]
	ttl      time.Duration
}

// NewCache creates a new cache with the specified TTL.
// Uses multicache with local filesystem persistence.
func NewCache(ttl time.Duration) *Cache {
	c := &Cache{ttl: ttl}

	// Determine cache directory
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		cacheDir = os.TempDir()
	}
	cacheDir = filepath.Join(cacheDir, "go-linear")

	// Create filesystem store for persistence
	store, err := localfs.New[string, string]("resolver", cacheDir)
	if err != nil {
		// Fall back to in-memory only if filesystem fails
		c.inmemory = multicache.New[string, string](
			multicache.Size(1000),
			multicache.TTL(ttl),
		)
		return c
	}

	// Create tiered cache: memory + filesystem
	c.tiered, err = multicache.NewTiered(store,
		multicache.Size(1000),
		multicache.TTL(ttl),
	)
	if err != nil {
		// Fall back to in-memory only
		c.inmemory = multicache.New[string, string](
			multicache.Size(1000),
			multicache.TTL(ttl),
		)
	}

	return c
}

// Get retrieves a value from the cache.
// Returns empty string if not found or expired.
func (c *Cache) Get(key string) (string, bool) {
	ctx := context.Background()

	if c.tiered != nil {
		val, ok, err := c.tiered.Get(ctx, key)
		if err != nil || !ok {
			return "", false
		}
		return val, true
	}

	if c.inmemory != nil {
		return c.inmemory.Get(key)
	}

	return "", false
}

// Set stores a value in the cache with TTL.
// Uses synchronous write to ensure reliability for name→ID mappings.
func (c *Cache) Set(key, value string) {
	ctx := context.Background()

	if c.tiered != nil {
		// Use synchronous write - resolver mappings are critical and the
		// performance impact is minimal since these are small KV pairs
		_ = c.tiered.Set(ctx, key, value)
		// Ignore errors - cache is best-effort, multicache handles logging
		return
	}

	if c.inmemory != nil {
		c.inmemory.Set(key, value)
	}
}

// Clear removes all entries from the cache.
func (c *Cache) Clear() {
	ctx := context.Background()

	if c.tiered != nil {
		// TieredCache.Flush syncs to disk; use Store.Flush to clear disk storage
		_, _ = c.tiered.Store.Flush(ctx)
		return
	}

	if c.inmemory != nil {
		c.inmemory.Flush()
	}
}
