package main

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestNewFIFOCache(t *testing.T) {
	cache := NewFIFOCache(10)
	
	if cache.maxSize != 10 {
		t.Errorf("Expected maxSize 10, got %d", cache.maxSize)
	}
	
	if cache.Size() != 0 {
		t.Errorf("Expected empty cache, got size %d", cache.Size())
	}
	
	if cache.cache == nil {
		t.Error("Cache map should not be nil")
	}
	
	if cache.list == nil {
		t.Error("Cache list should not be nil")
	}
	
	if cache.mutex == nil {
		t.Error("Cache mutex should not be nil")
	}
}

func TestFIFOCacheSet(t *testing.T) {
	cache := NewFIFOCache(3)
	
	// Test adding new items
	if !cache.Set("url1") {
		t.Error("Expected Set to return true for new item")
	}
	
	if cache.Size() != 1 {
		t.Errorf("Expected size 1, got %d", cache.Size())
	}
	
	// Test adding duplicate item
	if cache.Set("url1") {
		t.Error("Expected Set to return false for duplicate item")
	}
	
	if cache.Size() != 1 {
		t.Errorf("Expected size to remain 1, got %d", cache.Size())
	}
	
	// Test adding more items
	cache.Set("url2")
	cache.Set("url3")
	
	if cache.Size() != 3 {
		t.Errorf("Expected size 3, got %d", cache.Size())
	}
	
	// Test FIFO behavior - adding 4th item should remove first
	cache.Set("url4")
	
	if cache.Size() != 3 {
		t.Errorf("Expected size to remain 3, got %d", cache.Size())
	}
	
	// url1 should be evicted, so adding it again should return true
	if !cache.Set("url1") {
		t.Error("Expected Set to return true for previously evicted item")
	}
}

func TestFIFOCacheFIFOOrdering(t *testing.T) {
	cache := NewFIFOCache(2)
	
	// Add two items
	cache.Set("first")
	cache.Set("second")
	
	// Add third item, should evict "first" (oldest)
	cache.Set("third")
	
	// "first" should be evicted, so adding it again should return true
	if !cache.Set("first") {
		t.Error("Expected 'first' to be evicted and Set to return true")
	}
	
	// At this point cache contains: "third", "first" (second was evicted when first was re-added)
	// "second" should no longer be in cache
	if !cache.Set("second") {
		t.Error("Expected 'second' to be evicted and Set to return true")
	}
	
	// Verify cache contains "first" and "second" now
	if cache.Set("first") {
		t.Error("Expected 'first' to be in cache and Set to return false")
	}
}

func TestFIFOCacheSize(t *testing.T) {
	cache := NewFIFOCache(5)
	
	urls := []string{"url1", "url2", "url3", "url4", "url5"}
	
	for i, url := range urls {
		cache.Set(url)
		expectedSize := i + 1
		if cache.Size() != expectedSize {
			t.Errorf("After adding %s, expected size %d, got %d", url, expectedSize, cache.Size())
		}
	}
	
	// Adding one more should not increase size beyond maxSize
	cache.Set("url6")
	if cache.Size() != 5 {
		t.Errorf("Expected size to remain 5, got %d", cache.Size())
	}
}

func TestFIFOCacheClear(t *testing.T) {
	cache := NewFIFOCache(5)
	
	// Add some items
	cache.Set("url1")
	cache.Set("url2")
	cache.Set("url3")
	
	if cache.Size() != 3 {
		t.Errorf("Expected size 3 before clear, got %d", cache.Size())
	}
	
	// Clear cache
	cache.Clear()
	
	if cache.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", cache.Size())
	}
	
	// Items should be addable again
	if !cache.Set("url1") {
		t.Error("Expected Set to return true for item after clear")
	}
}

func TestFIFOCacheConcurrency(t *testing.T) {
	cache := NewFIFOCache(100)
	numGoroutines := 10
	itemsPerGoroutine := 50
	
	var wg sync.WaitGroup
	
	// Test concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(routineID int) {
			defer wg.Done()
			for j := 0; j < itemsPerGoroutine; j++ {
				url := fmt.Sprintf("routine%d-url%d", routineID, j)
				cache.Set(url)
			}
		}(i)
	}
	
	wg.Wait()
	
	// Check that cache doesn't exceed max size
	if cache.Size() > 100 {
		t.Errorf("Cache size %d exceeds max size 100", cache.Size())
	}
	
	// Test concurrent reads while writing
	done := make(chan bool)
	go func() {
		for i := 0; i < 100; i++ {
			cache.Set(fmt.Sprintf("writer-url%d", i))
			time.Sleep(time.Microsecond)
		}
		done <- true
	}()
	
	for i := 0; i < 100; i++ {
		go func() {
			cache.Size() // Just read the size
		}()
	}
	
	<-done
}

func TestFIFOCacheEdgeCases(t *testing.T) {
	// Test with maxSize 0
	t.Run("ZeroMaxSize", func(t *testing.T) {
		cache := NewFIFOCache(0)
		
		// Should always return true since nothing stays in cache
		if !cache.Set("url1") {
			t.Error("Expected Set to return true for zero-size cache")
		}
		
		if cache.Size() != 0 {
			t.Errorf("Expected size 0 for zero-size cache, got %d", cache.Size())
		}
	})
	
	// Test with maxSize 1
	t.Run("MaxSizeOne", func(t *testing.T) {
		cache := NewFIFOCache(1)
		
		cache.Set("url1")
		if cache.Size() != 1 {
			t.Errorf("Expected size 1, got %d", cache.Size())
		}
		
		cache.Set("url2")
		if cache.Size() != 1 {
			t.Errorf("Expected size to remain 1, got %d", cache.Size())
		}
		
		// url1 should be evicted
		if !cache.Set("url1") {
			t.Error("Expected url1 to be evicted and Set to return true")
		}
	})
	
	// Test with empty strings
	t.Run("EmptyString", func(t *testing.T) {
		cache := NewFIFOCache(5)
		
		if !cache.Set("") {
			t.Error("Expected Set to return true for empty string")
		}
		
		if cache.Set("") {
			t.Error("Expected Set to return false for duplicate empty string")
		}
	})
}

func TestGlobalCacheBackwardCompatibility(t *testing.T) {
	// Test the global cache_set function
	globalCache.Clear() // Start with clean cache
	
	if !cache_set("test-url") {
		t.Error("Expected cache_set to return true for new URL")
	}
	
	if cache_set("test-url") {
		t.Error("Expected cache_set to return false for duplicate URL")
	}
	
	if globalCache.Size() != 1 {
		t.Errorf("Expected global cache size 1, got %d", globalCache.Size())
	}
}

// Benchmark tests
func BenchmarkFIFOCacheSet(b *testing.B) {
	cache := NewFIFOCache(1000)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(fmt.Sprintf("url%d", i))
	}
}

func BenchmarkFIFOCacheSetDuplicates(b *testing.B) {
	cache := NewFIFOCache(1000)
	cache.Set("duplicate-url")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set("duplicate-url")
	}
}

func BenchmarkFIFOCacheSize(b *testing.B) {
	cache := NewFIFOCache(1000)
	for i := 0; i < 100; i++ {
		cache.Set(fmt.Sprintf("url%d", i))
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Size()
	}
}

func BenchmarkFIFOCacheClear(b *testing.B) {
	cache := NewFIFOCache(1000)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Add some items
		for j := 0; j < 10; j++ {
			cache.Set(fmt.Sprintf("url%d-%d", i, j))
		}
		cache.Clear()
	}
}

func BenchmarkFIFOCacheConcurrent(b *testing.B) {
	cache := NewFIFOCache(1000)
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			cache.Set(fmt.Sprintf("url%d", i))
			i++
		}
	})
}

// Test memory usage and growth
func TestFIFOCacheMemoryBounds(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory test in short mode")
	}
	
	cache := NewFIFOCache(1000)
	
	// Add many items
	for i := 0; i < 10000; i++ {
		cache.Set(fmt.Sprintf("url%d", i))
	}
	
	// Size should not exceed maxSize
	if cache.Size() > 1000 {
		t.Errorf("Cache size %d exceeds max size 1000", cache.Size())
	}
	
	// Cache should contain the most recent 1000 items
	// Test that old items are properly evicted
	if cache.Set("url0") != true {
		t.Error("Old item should have been evicted")
	}
	
	if cache.Set("url9999") != false {
		t.Error("Most recent item should still be in cache")
	}
}
