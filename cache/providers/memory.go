package providers

import (
	"context"
	"fmt"
	"regexp"
	"sync"
	"time"
)

// MemoryCacheItem represents an item stored in memory cache
type MemoryCacheItem struct {
	Data       []byte
	Expiration time.Time
	LastAccess time.Time
}

// IsExpired checks if the cache item has expired
func (item *MemoryCacheItem) IsExpired() bool {
	if item.Expiration.IsZero() {
		return false
	}
	return time.Now().After(item.Expiration)
}

// Touch updates the last access time
func (item *MemoryCacheItem) Touch() {
	item.LastAccess = time.Now()
}

// MemoryCacheOptions defines options for memory cache
type MemoryCacheOptions struct {
	// MaxSize is the maximum number of items to store
	MaxSize int64
	// DefaultExpiration is the default expiration duration
	DefaultExpiration time.Duration
	// CleanupInterval is how often expired items are cleaned up
	CleanupInterval time.Duration
	// EnableLRU enables LRU eviction when max size is reached
	EnableLRU bool
}

// DefaultMemoryCacheOptions returns default options for memory cache
func DefaultMemoryCacheOptions() *MemoryCacheOptions {
	return &MemoryCacheOptions{
		MaxSize:           10000,
		DefaultExpiration: time.Hour,
		CleanupInterval:   time.Minute * 10,
		EnableLRU:         true,
	}
}

// MemoryCache implements ICacheProvider using in-memory storage
type MemoryCache struct {
	mu      sync.RWMutex
	items   map[string]*MemoryCacheItem
	options *MemoryCacheOptions
	stopCh  chan struct{}
	stopped bool
}

// NewMemoryCache creates a new memory cache provider
func NewMemoryCache(options *MemoryCacheOptions) *MemoryCache {
	if options == nil {
		options = DefaultMemoryCacheOptions()
	}

	mc := &MemoryCache{
		items:   make(map[string]*MemoryCacheItem),
		options: options,
		stopCh:  make(chan struct{}),
	}

	// Start cleanup goroutine
	go mc.startCleanup()

	return mc
}

// Name returns the provider name
func (mc *MemoryCache) Name() string {
	return "memory"
}

// GetRaw retrieves raw data from the cache
func (mc *MemoryCache) GetRaw(ctx context.Context, key string) ([]byte, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	item, exists := mc.items[key]
	if !exists {
		return nil, nil
	}

	if item.IsExpired() {
		// Remove expired item
		go mc.removeExpiredItem(key)
		return nil, nil
	}

	// Update last access time for LRU
	item.Touch()

	// Return a copy of the data
	data := make([]byte, len(item.Data))
	copy(data, item.Data)
	return data, nil
}

// SetRaw stores raw data in the cache
func (mc *MemoryCache) SetRaw(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Check if we need to evict items
	if int64(len(mc.items)) >= mc.options.MaxSize {
		if mc.options.EnableLRU {
			mc.evictLRU()
		} else {
			return fmt.Errorf("cache is full (max size: %d)", mc.options.MaxSize)
		}
	}

	// Calculate expiration time
	var exp time.Time
	if expiration > 0 {
		exp = time.Now().Add(expiration)
	} else if mc.options.DefaultExpiration > 0 {
		exp = time.Now().Add(mc.options.DefaultExpiration)
	}

	// Store a copy of the data
	data := make([]byte, len(value))
	copy(data, value)

	mc.items[key] = &MemoryCacheItem{
		Data:       data,
		Expiration: exp,
		LastAccess: time.Now(),
	}

	return nil
}

// Remove removes data from the cache
func (mc *MemoryCache) Remove(ctx context.Context, key string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	delete(mc.items, key)
	return nil
}

// RemoveByPattern removes all keys matching the pattern
func (mc *MemoryCache) RemoveByPattern(ctx context.Context, pattern string) error {
	// Compile regex pattern
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid pattern: %w", err)
	}

	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Find keys to remove
	keysToRemove := make([]string, 0)
	for key := range mc.items {
		if regex.MatchString(key) {
			keysToRemove = append(keysToRemove, key)
		}
	}

	// Remove matched keys
	for _, key := range keysToRemove {
		delete(mc.items, key)
	}

	return nil
}

// Exists checks if a key exists
func (mc *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	item, exists := mc.items[key]
	if !exists {
		return false, nil
	}

	if item.IsExpired() {
		// Remove expired item
		go mc.removeExpiredItem(key)
		return false, nil
	}

	return true, nil
}

// Clear clears all data
func (mc *MemoryCache) Clear(ctx context.Context) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.items = make(map[string]*MemoryCacheItem)
	return nil
}

// Close closes the provider and releases resources
func (mc *MemoryCache) Close() error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if !mc.stopped {
		close(mc.stopCh)
		mc.stopped = true
	}

	mc.items = nil
	return nil
}

// GetStats returns cache statistics
func (mc *MemoryCache) GetStats() map[string]interface{} {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	totalSize := 0
	expiredCount := 0

	for _, item := range mc.items {
		totalSize += len(item.Data)
		if item.IsExpired() {
			expiredCount++
		}
	}

	return map[string]interface{}{
		"total_items":   len(mc.items),
		"expired_items": expiredCount,
		"total_size":    totalSize,
		"max_size":      mc.options.MaxSize,
		"hit_rate":      0, // Could be implemented with counters
	}
}

// startCleanup starts the cleanup goroutine
func (mc *MemoryCache) startCleanup() {
	ticker := time.NewTicker(mc.options.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			mc.cleanup()
		case <-mc.stopCh:
			return
		}
	}
}

// cleanup removes expired items
func (mc *MemoryCache) cleanup() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	now := time.Now()
	keysToRemove := make([]string, 0)

	for key, item := range mc.items {
		if !item.Expiration.IsZero() && now.After(item.Expiration) {
			keysToRemove = append(keysToRemove, key)
		}
	}

	for _, key := range keysToRemove {
		delete(mc.items, key)
	}
}

// removeExpiredItem removes a single expired item (called from goroutine)
func (mc *MemoryCache) removeExpiredItem(key string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if item, exists := mc.items[key]; exists && item.IsExpired() {
		delete(mc.items, key)
	}
}

// evictLRU evicts the least recently used item
func (mc *MemoryCache) evictLRU() {
	if len(mc.items) == 0 {
		return
	}

	var oldestKey string
	var oldestTime time.Time

	// Find the item with the oldest last access time
	for key, item := range mc.items {
		if oldestKey == "" || item.LastAccess.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.LastAccess
		}
	}

	if oldestKey != "" {
		delete(mc.items, oldestKey)
	}
}
