package providers

import (
	"context"
	"fmt"
	"regexp"
	"sync"
	"time"
)

// MemoryCacheItem 表示存储在内存缓存中的项
type MemoryCacheItem struct {
	Data       []byte
	Expiration time.Time
	LastAccess time.Time
}

// IsExpired 检查缓存项是否已过期
func (item *MemoryCacheItem) IsExpired() bool {
	if item.Expiration.IsZero() {
		return false
	}
	return time.Now().After(item.Expiration)
}

// Touch 更新最后访问时间
func (item *MemoryCacheItem) Touch() {
	item.LastAccess = time.Now()
}

// MemoryCacheOptions 定义内存缓存的选项
type MemoryCacheOptions struct {
	// MaxSize 是要存储的最大项数
	MaxSize int64
	// DefaultExpiration 是默认的过期时间
	DefaultExpiration time.Duration
	// CleanupInterval 是清理过期项的频率
	CleanupInterval time.Duration
	// EnableLRU 启用 LRU 驱逐策略（当达到最大大小时）
	EnableLRU bool
}

// DefaultMemoryCacheOptions 返回内存缓存的默认选项
func DefaultMemoryCacheOptions() *MemoryCacheOptions {
	return &MemoryCacheOptions{
		MaxSize:           10000,
		DefaultExpiration: time.Hour,
		CleanupInterval:   time.Minute * 10,
		EnableLRU:         true,
	}
}

// MemoryCache 使用内存存储实现 ICacheProvider 接口
type MemoryCache struct {
	mu      sync.RWMutex
	items   map[string]*MemoryCacheItem
	options *MemoryCacheOptions
	stopCh  chan struct{}
	stopped bool
}

// NewMemoryCache 创建一个新的内存缓存提供者
func NewMemoryCache(options *MemoryCacheOptions) *MemoryCache {
	if options == nil {
		options = DefaultMemoryCacheOptions()
	}

	mc := &MemoryCache{
		items:   make(map[string]*MemoryCacheItem),
		options: options,
		stopCh:  make(chan struct{}),
	}

	// 启动清理协程
	go mc.startCleanup()

	return mc
}

// Name 返回提供者名称
func (mc *MemoryCache) Name() string {
	return "memory"
}

// GetRaw 从缓存中获取原始数据
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

	// 更新最后访问时间以支持 LRU
	item.Touch()

	// 返回数据的副本
	data := make([]byte, len(item.Data))
	copy(data, item.Data)
	return data, nil
}

// SetRaw 将原始数据存储到缓存中
func (mc *MemoryCache) SetRaw(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	// 检查是否需要驱逐项
	if int64(len(mc.items)) >= mc.options.MaxSize {
		if mc.options.EnableLRU {
			mc.evictLRU()
		} else {
			return fmt.Errorf("cache is full (max size: %d)", mc.options.MaxSize)
		}
	}

	// 计算过期时间
	var exp time.Time
	if expiration > 0 {
		exp = time.Now().Add(expiration)
	} else if mc.options.DefaultExpiration > 0 {
		exp = time.Now().Add(mc.options.DefaultExpiration)
	}

	// 存储数据的副本
	data := make([]byte, len(value))
	copy(data, value)

	mc.items[key] = &MemoryCacheItem{
		Data:       data,
		Expiration: exp,
		LastAccess: time.Now(),
	}

	return nil
}

// Remove 从缓存中移除数据
func (mc *MemoryCache) Remove(ctx context.Context, key string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	delete(mc.items, key)
	return nil
}

// RemoveByPattern 移除所有匹配模式的键
func (mc *MemoryCache) RemoveByPattern(ctx context.Context, pattern string) error {
	// 编译正则表达式模式
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid pattern: %w", err)
	}

	mc.mu.Lock()
	defer mc.mu.Unlock()

	// 查找要移除的键
	keysToRemove := make([]string, 0)
	for key := range mc.items {
		if regex.MatchString(key) {
			keysToRemove = append(keysToRemove, key)
		}
	}

	// 移除匹配的键
	for _, key := range keysToRemove {
		delete(mc.items, key)
	}

	return nil
}

// Exists 检查键是否存在
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

// Clear 清空所有数据
func (mc *MemoryCache) Clear(ctx context.Context) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.items = make(map[string]*MemoryCacheItem)
	return nil
}

// Close 关闭提供者并释放资源
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

// GetStats 返回缓存统计信息
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
		"hit_rate":      0, // 可以通过计数器实现
	}
}

// startCleanup 启动清理协程
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

// cleanup 移除过期的项
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

// removeExpiredItem 移除单个过期项（从协程调用）
func (mc *MemoryCache) removeExpiredItem(key string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if item, exists := mc.items[key]; exists && item.IsExpired() {
		delete(mc.items, key)
	}
}

// evictLRU 驱逐最近最少使用的项
func (mc *MemoryCache) evictLRU() {
	if len(mc.items) == 0 {
		return
	}

	var oldestKey string
	var oldestTime time.Time

	// 查找最旧的最后访问时间的项
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
