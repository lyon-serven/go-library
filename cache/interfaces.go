// Package cache provides a flexible caching system with dependency injection support,
// inspired by ABP vNext's cache design patterns.
package cache

import (
	"context"
	"time"
)

// CacheKey represents a cache key with namespace support
type CacheKey struct {
	Key       string
	Namespace string
}

// String returns the full cache key as a string
func (ck CacheKey) String() string {
	if ck.Namespace == "" {
		return ck.Key
	}
	return ck.Namespace + ":" + ck.Key
}

// NewCacheKey creates a new cache key
func NewCacheKey(key, namespace string) CacheKey {
	return CacheKey{Key: key, Namespace: namespace}
}

// CacheItem represents an item stored in cache
type CacheItem struct {
	Key               CacheKey
	Value             interface{}
	Expiration        time.Time
	SlidingExpiration *time.Duration
}

// IsExpired checks if the cache item has expired
func (ci *CacheItem) IsExpired() bool {
	if ci.Expiration.IsZero() {
		return false
	}
	return time.Now().After(ci.Expiration)
}

// UpdateSlidingExpiration updates the expiration time for sliding expiration
func (ci *CacheItem) UpdateSlidingExpiration() {
	if ci.SlidingExpiration != nil {
		ci.Expiration = time.Now().Add(*ci.SlidingExpiration)
	}
}

// CacheOptions defines options for cache operations
type CacheOptions struct {
	AbsoluteExpiration *time.Time
	SlidingExpiration  *time.Duration
	Priority           CachePriority
}

// CachePriority represents the priority of a cache item
type CachePriority int

const (
	// Low priority cache items may be removed first when memory is needed
	Low CachePriority = iota
	// Normal priority cache items (default)
	Normal
	// High priority cache items are less likely to be removed
	High
	// NeverRemove cache items should never be removed automatically
	NeverRemove
)

// ICache defines the main cache interface
type ICache interface {
	// Get retrieves a value from cache
	Get(ctx context.Context, key CacheKey) (interface{}, error)

	// GetAsync retrieves a value from cache asynchronously
	GetAsync(ctx context.Context, key CacheKey) <-chan CacheResult

	// Set stores a value in cache with options
	Set(ctx context.Context, key CacheKey, value interface{}, options *CacheOptions) error

	// SetAsync stores a value in cache asynchronously
	SetAsync(ctx context.Context, key CacheKey, value interface{}, options *CacheOptions) <-chan error

	// Remove removes a value from cache
	Remove(ctx context.Context, key CacheKey) error

	// RemoveByPattern removes all keys matching the pattern
	RemoveByPattern(ctx context.Context, pattern string) error

	// Exists checks if a key exists in cache
	Exists(ctx context.Context, key CacheKey) (bool, error)

	// Clear clears all cache entries
	Clear(ctx context.Context) error

	// GetOrSet gets a value from cache or sets it using the provided factory
	GetOrSet(ctx context.Context, key CacheKey, factory func() (interface{}, error), options *CacheOptions) (interface{}, error)

	// Refresh refreshes the expiration time of a cache item
	Refresh(ctx context.Context, key CacheKey) error
}

// CacheResult represents the result of an async cache operation
type CacheResult struct {
	Value interface{}
	Error error
}

// ICacheProvider defines the interface for cache providers (Redis, Memory, etc.)
type ICacheProvider interface {
	// Get retrieves raw data from the cache provider
	GetRaw(ctx context.Context, key string) ([]byte, error)

	// Set stores raw data in the cache provider
	SetRaw(ctx context.Context, key string, value []byte, expiration time.Duration) error

	// Remove removes data from the cache provider
	Remove(ctx context.Context, key string) error

	// RemoveByPattern removes all keys matching the pattern
	RemoveByPattern(ctx context.Context, pattern string) error

	// Exists checks if a key exists
	Exists(ctx context.Context, key string) (bool, error)

	// Clear clears all data
	Clear(ctx context.Context) error

	// Name returns the provider name
	Name() string

	// Close closes the provider and releases resources
	Close() error
}

// ICacheSerializer defines the interface for cache serializers
type ICacheSerializer interface {
	// Serialize converts an object to bytes
	Serialize(value interface{}) ([]byte, error)

	// Deserialize converts bytes back to an object
	Deserialize(data []byte, target interface{}) error

	// Name returns the serializer name
	Name() string
}

// ICacheManager defines the interface for cache manager with dependency injection
type ICacheManager interface {
	// GetCache gets or creates a cache instance with the specified name
	GetCache(name string) ICache

	// RegisterProvider registers a cache provider
	RegisterProvider(name string, provider ICacheProvider) error

	// RegisterSerializer registers a cache serializer
	RegisterSerializer(name string, serializer ICacheSerializer) error

	// SetDefaultProvider sets the default cache provider
	SetDefaultProvider(name string) error

	// SetDefaultSerializer sets the default cache serializer
	SetDefaultSerializer(name string) error

	// Configure configures a specific cache with custom provider and serializer
	Configure(cacheName string, providerName string, serializerName string) error

	// Close closes all providers and releases resources
	Close() error
}

// CacheConfiguration holds configuration for a specific cache
type CacheConfiguration struct {
	Name           string
	ProviderName   string
	SerializerName string
	DefaultOptions *CacheOptions
}

// DefaultCacheOptions returns default cache options
func DefaultCacheOptions() *CacheOptions {
	return &CacheOptions{
		Priority: Normal,
	}
}

// WithAbsoluteExpiration sets absolute expiration time
func (opts *CacheOptions) WithAbsoluteExpiration(expiration time.Time) *CacheOptions {
	opts.AbsoluteExpiration = &expiration
	return opts
}

// WithSlidingExpiration sets sliding expiration duration
func (opts *CacheOptions) WithSlidingExpiration(duration time.Duration) *CacheOptions {
	opts.SlidingExpiration = &duration
	return opts
}

// WithPriority sets cache priority
func (opts *CacheOptions) WithPriority(priority CachePriority) *CacheOptions {
	opts.Priority = priority
	return opts
}
