package cache

import (
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	// Create a new cache with 1 second TTL
	cache := NewCache(1 * time.Second)

	// Test setting and getting data
	testKey := "test-key"
	testData := []byte("test-data")

	// Set data
	cache.Set(testKey, testData)

	// Get data immediately (should hit)
	data, found := cache.Get(testKey)
	if !found {
		t.Error("Expected to find data in cache")
	}
	if string(data) != string(testData) {
		t.Errorf("Expected %s, got %s", string(testData), string(data))
	}

	// Test TTL expiration
	time.Sleep(2 * time.Second)
	_, found = cache.Get(testKey)
	if found {
		t.Error("Expected data to be expired")
	}
}
