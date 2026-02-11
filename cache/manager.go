package cache

import (
	"fmt"
	"sync"
)

// CacheManager 实现 ICacheManager 接口，支持依赖注入
type CacheManager struct {
	mu                sync.RWMutex
	providers         map[string]ICacheProvider
	serializers       map[string]ICacheSerializer
	caches            map[string]ICache
	configurations    map[string]*CacheConfiguration
	defaultProvider   string
	defaultSerializer string
}

// NewCacheManager 创建一个新的缓存管理器
func NewCacheManager() *CacheManager {
	return &CacheManager{
		providers:      make(map[string]ICacheProvider),
		serializers:    make(map[string]ICacheSerializer),
		caches:         make(map[string]ICache),
		configurations: make(map[string]*CacheConfiguration),
	}
}

// RegisterProvider 注册一个缓存提供者
func (cm *CacheManager) RegisterProvider(name string, provider ICacheProvider) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, exists := cm.providers[name]; exists {
		return fmt.Errorf("provider '%s' already registered", name)
	}

	cm.providers[name] = provider

	// 如果是第一个提供者，设置为默认提供者
	if cm.defaultProvider == "" {
		cm.defaultProvider = name
	}

	return nil
}

// RegisterSerializer 注册一个缓存序列化器
func (cm *CacheManager) RegisterSerializer(name string, serializer ICacheSerializer) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, exists := cm.serializers[name]; exists {
		return fmt.Errorf("serializer '%s' already registered", name)
	}

	cm.serializers[name] = serializer

	// 如果是第一个序列化器，设置为默认序列化器
	if cm.defaultSerializer == "" {
		cm.defaultSerializer = name
	}

	return nil
}

// SetDefaultProvider 设置默认的缓存提供者
func (cm *CacheManager) SetDefaultProvider(name string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, exists := cm.providers[name]; !exists {
		return fmt.Errorf("provider '%s' not found", name)
	}

	cm.defaultProvider = name
	return nil
}

// SetDefaultSerializer 设置默认的缓存序列化器
func (cm *CacheManager) SetDefaultSerializer(name string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, exists := cm.serializers[name]; !exists {
		return fmt.Errorf("serializer '%s' not found", name)
	}

	cm.defaultSerializer = name
	return nil
}

// Configure 配置特定的缓存，使用自定义的提供者和序列化器
func (cm *CacheManager) Configure(cacheName string, providerName string, serializerName string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 验证提供者
	if _, exists := cm.providers[providerName]; !exists {
		return fmt.Errorf("provider '%s' not found", providerName)
	}

	// 验证序列化器
	if _, exists := cm.serializers[serializerName]; !exists {
		return fmt.Errorf("serializer '%s' not found", serializerName)
	}

	// 存储配置
	cm.configurations[cacheName] = &CacheConfiguration{
		Name:           cacheName,
		ProviderName:   providerName,
		SerializerName: serializerName,
		DefaultOptions: DefaultCacheOptions(),
	}

	// 删除现有的缓存实例，强制使用新配置重新创建
	delete(cm.caches, cacheName)

	return nil
}

// GetCache 获取或创建指定名称的缓存实例
func (cm *CacheManager) GetCache(name string) ICache {
	cm.mu.RLock()
	cache, exists := cm.caches[name]
	cm.mu.RUnlock()

	if exists {
		return cache
	}

	// 创建新的缓存实例
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 双重检查锁定模式
	if cache, exists := cm.caches[name]; exists {
		return cache
	}

	// 获取配置或使用默认配置
	config := cm.configurations[name]
	if config == nil {
		config = &CacheConfiguration{
			Name:           name,
			ProviderName:   cm.defaultProvider,
			SerializerName: cm.defaultSerializer,
			DefaultOptions: DefaultCacheOptions(),
		}
	}

	// 获取提供者和序列化器
	provider := cm.providers[config.ProviderName]
	serializer := cm.serializers[config.SerializerName]

	if provider == nil {
		panic(fmt.Sprintf("no provider available for cache '%s'", name))
	}

	if serializer == nil {
		panic(fmt.Sprintf("no serializer available for cache '%s'", name))
	}

	// 创建缓存实例
	cache = NewCache(name, provider, serializer, config.DefaultOptions)
	cm.caches[name] = cache

	return cache
}

// Close 关闭所有提供者并释放资源
func (cm *CacheManager) Close() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	var errors []error

	// 关闭所有提供者
	for _, provider := range cm.providers {
		if err := provider.Close(); err != nil {
			errors = append(errors, err)
		}
	}

	// 清空所有映射表
	cm.providers = make(map[string]ICacheProvider)
	cm.serializers = make(map[string]ICacheSerializer)
	cm.caches = make(map[string]ICache)
	cm.configurations = make(map[string]*CacheConfiguration)

	if len(errors) > 0 {
		return fmt.Errorf("errors while closing: %v", errors)
	}

	return nil
}
