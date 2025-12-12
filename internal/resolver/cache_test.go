package resolver

import (
	"testing"
	"time"
)

func TestCacheGetSet(t *testing.T) {
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
	cache := NewCache(1 * time.Minute)

	// Add multiple entries
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")

	// Verify they exist
	if _, ok := cache.Get("key1"); !ok {
		t.Error("key1 should exist before clear")
	}

	// Clear cache
	cache.Clear()

	// Verify all cleared
	if _, ok := cache.Get("key1"); ok {
		t.Error("key1 should not exist after clear")
	}
	if _, ok := cache.Get("key2"); ok {
		t.Error("key2 should not exist after clear")
	}
	if _, ok := cache.Get("key3"); ok {
		t.Error("key3 should not exist after clear")
	}
}

func TestCacheConcurrency(t *testing.T) {
	cache := NewCache(1 * time.Minute)

	// Concurrent writes
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			cache.Set("key", "value")
			cache.Get("key")
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should not panic (test for data races)
	val, ok := cache.Get("key")
	if !ok || val != "value" {
		t.Errorf("Get() after concurrent access = %s, %v, want value, true", val, ok)
	}
}
