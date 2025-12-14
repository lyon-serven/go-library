package cache

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// CacheManager implements ICacheManager with dependency injection support
type CacheManager struct {
	mu                sync.RWMutex
	providers         map[string]ICacheProvider
	serializers       map[string]ICacheSerializer
	caches            map[string]ICache
	configurations    map[string]*CacheConfiguration
	defaultProvider   string
	defaultSerializer string
}

// NewCacheManager creates a new cache manager
func NewCacheManager() *CacheManager {
	return &CacheManager{
		providers:      make(map[string]ICacheProvider),
		serializers:    make(map[string]ICacheSerializer),
		caches:         make(map[string]ICache),
		configurations: make(map[string]*CacheConfiguration),
	}
}

// RegisterProvider registers a cache provider
func (cm *CacheManager) RegisterProvider(name string, provider ICacheProvider) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, exists := cm.providers[name]; exists {
		return fmt.Errorf("provider '%s' already registered", name)
	}

	cm.providers[name] = provider

	// Set as default if it's the first provider
	if cm.defaultProvider == "" {
		cm.defaultProvider = name
	}

	return nil
}

// RegisterSerializer registers a cache serializer
func (cm *CacheManager) RegisterSerializer(name string, serializer ICacheSerializer) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, exists := cm.serializers[name]; exists {
		return fmt.Errorf("serializer '%s' already registered", name)
	}

	cm.serializers[name] = serializer

	// Set as default if it's the first serializer
	if cm.defaultSerializer == "" {
		cm.defaultSerializer = name
	}

	return nil
}

// SetDefaultProvider sets the default cache provider
func (cm *CacheManager) SetDefaultProvider(name string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, exists := cm.providers[name]; !exists {
		return fmt.Errorf("provider '%s' not found", name)
	}

	cm.defaultProvider = name
	return nil
}

// SetDefaultSerializer sets the default cache serializer
func (cm *CacheManager) SetDefaultSerializer(name string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, exists := cm.serializers[name]; !exists {
		return fmt.Errorf("serializer '%s' not found", name)
	}

	cm.defaultSerializer = name
	return nil
}

// Configure configures a specific cache with custom provider and serializer
func (cm *CacheManager) Configure(cacheName string, providerName string, serializerName string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Validate provider
	if _, exists := cm.providers[providerName]; !exists {
		return fmt.Errorf("provider '%s' not found", providerName)
	}

	// Validate serializer
	if _, exists := cm.serializers[serializerName]; !exists {
		return fmt.Errorf("serializer '%s' not found", serializerName)
	}

	// Store configuration
	cm.configurations[cacheName] = &CacheConfiguration{
		Name:           cacheName,
		ProviderName:   providerName,
		SerializerName: serializerName,
		DefaultOptions: DefaultCacheOptions(),
	}

	// Remove existing cache instance to force recreation with new config
	delete(cm.caches, cacheName)

	return nil
}

// GetCache gets or creates a cache instance with the specified name
func (cm *CacheManager) GetCache(name string) ICache {
	cm.mu.RLock()
	cache, exists := cm.caches[name]
	cm.mu.RUnlock()

	if exists {
		return cache
	}

	// Create new cache instance
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Double-check locking pattern
	if cache, exists := cm.caches[name]; exists {
		return cache
	}

	// Get configuration or use defaults
	config := cm.configurations[name]
	if config == nil {
		config = &CacheConfiguration{
			Name:           name,
			ProviderName:   cm.defaultProvider,
			SerializerName: cm.defaultSerializer,
			DefaultOptions: DefaultCacheOptions(),
		}
	}

	// Get provider and serializer
	provider := cm.providers[config.ProviderName]
	serializer := cm.serializers[config.SerializerName]

	if provider == nil {
		panic(fmt.Sprintf("no provider available for cache '%s'", name))
	}

	if serializer == nil {
		panic(fmt.Sprintf("no serializer available for cache '%s'", name))
	}

	// Create cache instance
	cache = NewCache(name, provider, serializer, config.DefaultOptions)
	cm.caches[name] = cache

	return cache
}

// Close closes all providers and releases resources
func (cm *CacheManager) Close() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	var errors []error

	// Close all providers
	for _, provider := range cm.providers {
		if err := provider.Close(); err != nil {
			errors = append(errors, err)
		}
	}

	// Clear all maps
	cm.providers = make(map[string]ICacheProvider)
	cm.serializers = make(map[string]ICacheSerializer)
	cm.caches = make(map[string]ICache)
	cm.configurations = make(map[string]*CacheConfiguration)

	if len(errors) > 0 {
		return fmt.Errorf("errors while closing: %v", errors)
	}

	return nil
}

// Cache implements ICache interface
type Cache struct {
	name           string
	provider       ICacheProvider
	serializer     ICacheSerializer
	defaultOptions *CacheOptions
}

// NewCache creates a new cache instance
func NewCache(name string, provider ICacheProvider, serializer ICacheSerializer, defaultOptions *CacheOptions) *Cache {
	return &Cache{
		name:           name,
		provider:       provider,
		serializer:     serializer,
		defaultOptions: defaultOptions,
	}
}

// Get retrieves a value from cache
func (c *Cache) Get(ctx context.Context, key CacheKey) (interface{}, error) {
	data, err := c.provider.GetRaw(ctx, key.String())
	if err != nil {
		return nil, err
	}

	if data == nil {
		return nil, nil
	}

	var value interface{}
	err = c.serializer.Deserialize(data, &value)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize cache value: %w", err)
	}

	return value, nil
}

// GetAsync retrieves a value from cache asynchronously
func (c *Cache) GetAsync(ctx context.Context, key CacheKey) <-chan CacheResult {
	result := make(chan CacheResult, 1)

	go func() {
		defer close(result)
		value, err := c.Get(ctx, key)
		result <- CacheResult{Value: value, Error: err}
	}()

	return result
}

// Set stores a value in cache with options
func (c *Cache) Set(ctx context.Context, key CacheKey, value interface{}, options *CacheOptions) error {
	if options == nil {
		options = c.defaultOptions
	}

	data, err := c.serializer.Serialize(value)
	if err != nil {
		return fmt.Errorf("failed to serialize cache value: %w", err)
	}

	// Calculate expiration duration
	var expiration time.Duration
	if options.AbsoluteExpiration != nil {
		expiration = time.Until(*options.AbsoluteExpiration)
	} else if options.SlidingExpiration != nil {
		expiration = *options.SlidingExpiration
	}

	return c.provider.SetRaw(ctx, key.String(), data, expiration)
}

// SetAsync stores a value in cache asynchronously
func (c *Cache) SetAsync(ctx context.Context, key CacheKey, value interface{}, options *CacheOptions) <-chan error {
	result := make(chan error, 1)

	go func() {
		defer close(result)
		result <- c.Set(ctx, key, value, options)
	}()

	return result
}

// Remove removes a value from cache
func (c *Cache) Remove(ctx context.Context, key CacheKey) error {
	return c.provider.Remove(ctx, key.String())
}

// RemoveByPattern removes all keys matching the pattern
func (c *Cache) RemoveByPattern(ctx context.Context, pattern string) error {
	return c.provider.RemoveByPattern(ctx, pattern)
}

// Exists checks if a key exists in cache
func (c *Cache) Exists(ctx context.Context, key CacheKey) (bool, error) {
	return c.provider.Exists(ctx, key.String())
}

// Clear clears all cache entries
func (c *Cache) Clear(ctx context.Context) error {
	return c.provider.Clear(ctx)
}

// GetOrSet gets a value from cache or sets it using the provided factory
func (c *Cache) GetOrSet(ctx context.Context, key CacheKey, factory func() (interface{}, error), options *CacheOptions) (interface{}, error) {
	// Try to get existing value
	value, err := c.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	// Return existing value if found
	if value != nil {
		return value, nil
	}

	// Generate new value using factory
	value, err = factory()
	if err != nil {
		return nil, err
	}

	// Set the new value in cache
	err = c.Set(ctx, key, value, options)
	if err != nil {
		// Return the value even if caching failed
		return value, nil
	}

	return value, nil
}

// Refresh refreshes the expiration time of a cache item
func (c *Cache) Refresh(ctx context.Context, key CacheKey) error {
	// Get current value
	value, err := c.Get(ctx, key)
	if err != nil {
		return err
	}

	if value == nil {
		return fmt.Errorf("key '%s' not found in cache", key.String())
	}

	// Reset with default options to refresh expiration
	return c.Set(ctx, key, value, c.defaultOptions)
}
