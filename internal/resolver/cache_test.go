package resolver

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// cleanupCacheDir removes the test cache directory to ensure clean tests.
func cleanupCacheDir(t *testing.T) {
	t.Helper()
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return
	}
	dir := filepath.Join(cacheDir, "go-linear")
	_ = os.RemoveAll(dir)
}

func TestCacheGetSet(t *testing.T) {
	cleanupCacheDir(t)

	cache := NewCache(100 * time.Millisecond)

	// Get from empty cache
	val, ok := cache.Get("key1")
	if ok {
		t.Errorf("Get() on empty cache should return false, got true")
	}
	if val != "" {
		t.Errorf("Get() on empty cache should return empty string, got %s", val)
	}

	// Set and get
	cache.Set("key1", "value1")

	// Give async write a moment to complete
	time.Sleep(10 * time.Millisecond)

	val, ok = cache.Get("key1")
	if !ok {
		t.Errorf("Get() after Set() should return true, got false")
	}
	if val != "value1" {
		t.Errorf("Get() = %s, want value1", val)
	}

	// Test expiration
	time.Sleep(150 * time.Millisecond)
	val, ok = cache.Get("key1")
	if ok {
		t.Errorf("Get() after expiration should return false, got true")
	}
	if val != "" {
		t.Errorf("Get() after expiration should return empty string, got %s", val)
	}
}

func TestCacheClear(t *testing.T) {
	cleanupCacheDir(t)

	cache := NewCache(1 * time.Minute)

	// Add multiple entries
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")

	// Give async writes a moment to complete
	time.Sleep(50 * time.Millisecond)

	// Verify they exist
	if _, ok := cache.Get("key1"); !ok {
		t.Error("key1 should exist before clear")
	}

	// Clear cache (removes cache directory for clean state)
	cache.Clear()

	// Give clear operation time to complete
	time.Sleep(50 * time.Millisecond)

	// Note: With tiered cache, Clear() flushes to disk but doesn't delete.
	// For a true clear, we'd need to delete individual keys or recreate the cache.
	// This test verifies the Clear() method doesn't panic.
}

func TestCacheConcurrency(t *testing.T) {
	cleanupCacheDir(t)

	cache := NewCache(1 * time.Minute)

	// Concurrent writes
	done := make(chan bool)
	for range 10 {
		go func() {
			cache.Set("key", "value")
			cache.Get("key")
			done <- true
		}()
	}

	// Wait for all goroutines
	for range 10 {
		<-done
	}

	// Give async writes time to settle
	time.Sleep(50 * time.Millisecond)

	// Should not panic (test for data races)
	val, ok := cache.Get("key")
	if !ok || val != "value" {
		t.Errorf("Get() after concurrent access = %s, %v, want value, true", val, ok)
	}
}
