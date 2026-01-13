package resolver

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/codeGROOVE-dev/fido"
	"github.com/codeGROOVE-dev/fido/pkg/store/localfs"
)

// Cache provides a file-backed TTL cache for name→ID resolution.
// Uses fido with local filesystem persistence to survive across
// MCP subprocess calls (ophis spawns new process per tool invocation).
type Cache struct {
	mu       sync.RWMutex
	tiered   *fido.TieredCache[string, string]
	inmemory *fido.Cache[string, string]
	ttl      time.Duration
}

// NewCache creates a new cache with the specified TTL.
// Uses fido with local filesystem persistence.
func NewCache(ttl time.Duration) *Cache {
	// Fido's TTL resolution is per-second
	// Only do this if we specifically set a TTL
	// to allow the default of 0 to take precedence
	if ttl != 0 {
		ttl = max(ttl, 1*time.Second)
	}
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
		c.inmemory = fido.New[string, string](
			fido.Size(1000),
			fido.TTL(ttl),
		)
		return c
	}

	// Create tiered cache: memory + filesystem
	c.tiered, err = fido.NewTiered(store,
		fido.Size(1000),
		fido.TTL(ttl),
	)
	if err != nil {
		// Fall back to in-memory only
		c.inmemory = fido.New[string, string](
			fido.Size(1000),
			fido.TTL(ttl),
		)
	}

	return c
}

// Get retrieves a value from the cache.
// Returns empty string if not found or expired.
func (c *Cache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

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
	c.mu.Lock()
	defer c.mu.Unlock()

	ctx := context.Background()

	if c.tiered != nil {
		// Use synchronous write - resolver mappings are critical and the
		// performance impact is minimal since these are small KV pairs
		_ = c.tiered.Set(ctx, key, value)
		// Ignore errors - cache is best-effort, fido handles logging
		return
	}

	if c.inmemory != nil {
		c.inmemory.Set(key, value)
	}
}

// Clear removes all entries from the cache.
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

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
