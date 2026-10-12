// Package cache provides a generic, concurrent-safe TTL cache for expensive
// read-only lookups used by the API (statistics, account lookups, etc.).
package cache

import (
	"sync"
	"time"
)

// Item is the cached value.
type Item[K comparable, V any] struct {
	Value     V
	ExpiresAt time.Time
}

// Cache is a generic thread-safe TTL cache.
type Cache[K comparable, V any] struct {
	mu       sync.Mutex
	items    map[K]Item[K, V]
	defaultTTL time.Duration
}

// New creates a new Cache with the given default TTL.
func New[K comparable, V any](defaultTTL time.Duration) *Cache[K, V] {
	return &Cache[K, V]{
		items:      make(map[K]Item[K, V]),
		defaultTTL: defaultTTL,
	}
}

// Get returns the value for key if present and not expired, along with whether
// it was found.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, ok := c.items[key]
	if !ok {
		var zero V
		return zero, false
	}
	if time.Now().After(item.ExpiresAt) {
		delete(c.items, key)
		var zero V
		return zero, false
	}
	return item.Value, true
}

// Set stores a value for key with the cache's default TTL.
func (c *Cache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = Item[K, V]{
		Value:     value,
		ExpiresAt: time.Now().Add(c.defaultTTL),
	}
}

// SetWithTTL stores a value for key with the given TTL.
func (c *Cache[K, V]) SetWithTTL(key K, value V, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = Item[K, V]{
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
	}
}

// Delete removes a key from the cache.
func (c *Cache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

// Clear removes all entries.
func (c *Cache[K, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[K]Item[K, V])
}

// Size returns the number of entries in the cache.
func (c *Cache[K, V]) Size() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return len(c.items)
}
