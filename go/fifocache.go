package main

import (
	"container/list"
	"sync"
)

// FIFOCache represents a FIFO cache for deduplicating URLs
type FIFOCache struct {
	cache    map[string]struct{}
	list     *list.List
	mutex    *sync.RWMutex
	maxSize  int
}

// NewFIFOCache creates a new FIFO cache with specified maximum size
func NewFIFOCache(maxSize int) *FIFOCache {
	return &FIFOCache{
		cache:   make(map[string]struct{}),
		list:    list.New(),
		mutex:   &sync.RWMutex{},
		maxSize: maxSize,
	}
}

// Set adds a URL to the cache and returns true if it's new, false if it already exists
func (c *FIFOCache) Set(url string) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	_, exists := c.cache[url]
	if exists {
		return false
	}

	c.cache[url] = struct{}{}
	c.list.PushFront(url)

	// Remove oldest items if cache exceeds max size
	for c.list.Len() > c.maxSize {
		item := c.list.Back()
		delete(c.cache, item.Value.(string))
		c.list.Remove(item)
	}
	
	return true
}

// Size returns the current size of the cache
func (c *FIFOCache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.cache)
}

// Clear removes all items from the cache
func (c *FIFOCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.cache = make(map[string]struct{})
	c.list = list.New()
}

// Global cache instance for backward compatibility
var (
	globalCache = NewFIFOCache(10)
)

// cache_set is a backward compatibility function
func cache_set(url string) bool {
	return globalCache.Set(url)
}
