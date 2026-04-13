package cache

import (
	"context"
	"fmt"
	"time"
)

// Cache 实现 ICache 接口
type Cache struct {
	name           string
	provider       ICacheProvider
	serializer     ICacheSerializer
	defaultOptions *CacheOptions
}

// NewCache 创建一个新的缓存实例
func NewCache(name string, provider ICacheProvider, serializer ICacheSerializer, defaultOptions *CacheOptions) *Cache {
	return &Cache{
		name:           name,
		provider:       provider,
		serializer:     serializer,
		defaultOptions: defaultOptions,
	}
}

// Get 从缓存中获取值
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

// GetAs 从缓存中获取值并反序列化到指定类型的目标对象
// 这样可以避免 JSON 反序列化时类型丢失的问题
// 使用示例：var user User; err := cache.GetAs(ctx, key, &user)
func (c *Cache) GetAs(ctx context.Context, key CacheKey, target interface{}) error {
	data, err := c.provider.GetRaw(ctx, key.String())
	if err != nil {
		return err
	}

	if data == nil {
		return fmt.Errorf("key '%s' not found in cache", key.String())
	}

	err = c.serializer.Deserialize(data, target)
	if err != nil {
		return fmt.Errorf("failed to deserialize cache value: %w", err)
	}

	return nil
}

// GetAsync 异步从缓存中获取值
func (c *Cache) GetAsync(ctx context.Context, key CacheKey) <-chan CacheResult {
	result := make(chan CacheResult, 1)

	go func() {
		defer close(result)
		value, err := c.Get(ctx, key)
		result <- CacheResult{Value: value, Error: err}
	}()

	return result
}

// Set 将值存储到缓存中，带有选项参数
func (c *Cache) Set(ctx context.Context, key CacheKey, value interface{}, options *CacheOptions) error {
	if options == nil {
		options = c.defaultOptions
	}

	data, err := c.serializer.Serialize(value)
	if err != nil {
		return fmt.Errorf("failed to serialize cache value: %w", err)
	}

	// 计算过期时间
	var expiration time.Duration
	if options.AbsoluteExpiration != nil {
		expiration = time.Until(*options.AbsoluteExpiration)
	} else if options.SlidingExpiration != nil {
		expiration = *options.SlidingExpiration
	}

	return c.provider.SetRaw(ctx, key.String(), data, expiration)
}

// SetAsync 异步将值存储到缓存中
func (c *Cache) SetAsync(ctx context.Context, key CacheKey, value interface{}, options *CacheOptions) <-chan error {
	result := make(chan error, 1)

	go func() {
		defer close(result)
		result <- c.Set(ctx, key, value, options)
	}()

	return result
}

// Remove 从缓存中移除值
func (c *Cache) Remove(ctx context.Context, key CacheKey) error {
	return c.provider.Remove(ctx, key.String())
}

// RemoveByPattern 移除所有匹配模式的键
func (c *Cache) RemoveByPattern(ctx context.Context, pattern string) error {
	return c.provider.RemoveByPattern(ctx, pattern)
}

// Exists 检查键是否存在于缓存中
func (c *Cache) Exists(ctx context.Context, key CacheKey) (bool, error) {
	return c.provider.Exists(ctx, key.String())
}

// Clear 清空所有缓存条目
func (c *Cache) Clear(ctx context.Context) error {
	return c.provider.Clear(ctx)
}

// GetOrSet 从缓存中获取值，如果不存在则使用提供的工厂函数设置
func (c *Cache) GetOrSet(ctx context.Context, key CacheKey, factory func() (interface{}, error), options *CacheOptions) (interface{}, error) {
	// 尝试获取现有值
	value, err := c.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	// 如果找到现有值，则返回
	if value != nil {
		return value, nil
	}

	// 使用工厂函数生成新值
	value, err = factory()
	if err != nil {
		return nil, err
	}

	// 将新值设置到缓存中
	err = c.Set(ctx, key, value, options)
	if err != nil {
		// 即使缓存失败也返回值
		return value, nil
	}

	return value, nil
}

// GetOrSetAs 类型安全的 GetOrSet，将值反序列化到指定类型的目标对象
// 使用示例：var user User; err := cache.GetOrSetAs(ctx, key, &user, factory, options)
func (c *Cache) GetOrSetAs(ctx context.Context, key CacheKey, target interface{}, factory func() (interface{}, error), options *CacheOptions) error {
	// 尝试从缓存获取
	err := c.GetAs(ctx, key, target)
	if err == nil {
		return nil // 缓存命中
	}

	// 缓存未命中，使用工厂函数生成新值
	value, err := factory()
	if err != nil {
		return err
	}

	// 保存到缓存
	err = c.Set(ctx, key, value, options)
	if err != nil {
		// 即使缓存失败也继续，只是记录错误
		// 可以选择记录日志
	}

	// 将 factory 返回的值反序列化到 target
	// 需要先序列化再反序列化以确保类型一致
	data, err := c.serializer.Serialize(value)
	if err != nil {
		return fmt.Errorf("failed to serialize factory value: %w", err)
	}

	err = c.serializer.Deserialize(data, target)
	if err != nil {
		return fmt.Errorf("failed to deserialize factory value: %w", err)
	}

	return nil
}

// Refresh 刷新缓存项的过期时间
func (c *Cache) Refresh(ctx context.Context, key CacheKey) error {
	// 获取当前值
	value, err := c.Get(ctx, key)
	if err != nil {
		return err
	}

	if value == nil {
		return fmt.Errorf("key '%s' not found in cache", key.String())
	}

	// 使用默认选项重新设置以刷新过期时间
	return c.Set(ctx, key, value, c.defaultOptions)
}

// ============================================
// 字符串键便捷方法
// ============================================

// GetS 使用字符串键从缓存中获取值
func (c *Cache) GetS(ctx context.Context, key string) (interface{}, error) {
	return c.Get(ctx, K(key))
}

// GetAsS 使用字符串键从缓存中获取值并反序列化到指定类型
func (c *Cache) GetAsS(ctx context.Context, key string, target interface{}) error {
	return c.GetAs(ctx, K(key), target)
}

// SetS 使用字符串键将值存储到缓存中
func (c *Cache) SetS(ctx context.Context, key string, value interface{}, options *CacheOptions) error {
	return c.Set(ctx, K(key), value, options)
}

// RemoveS 使用字符串键从缓存中移除值
func (c *Cache) RemoveS(ctx context.Context, key string) error {
	return c.Remove(ctx, K(key))
}

// ExistsS 使用字符串键检查键是否存在
func (c *Cache) ExistsS(ctx context.Context, key string) (bool, error) {
	return c.Exists(ctx, K(key))
}

// GetOrSetS 使用字符串键从缓存获取值，如果不存在则使用工厂函数设置
func (c *Cache) GetOrSetS(ctx context.Context, key string, factory func() (interface{}, error), options *CacheOptions) (interface{}, error) {
	return c.GetOrSet(ctx, K(key), factory, options)
}

// GetOrSetAsS 使用字符串键的类型安全 GetOrSet
func (c *Cache) GetOrSetAsS(ctx context.Context, key string, target interface{}, factory func() (interface{}, error), options *CacheOptions) error {
	return c.GetOrSetAs(ctx, K(key), target, factory, options)
}

// RefreshS 使用字符串键刷新过期时间
func (c *Cache) RefreshS(ctx context.Context, key string) error {
	return c.Refresh(ctx, K(key))
}

// ============================================
// Pipeline 批量操作
// ============================================

// PipelineItem 高层 Pipeline 批量写入条目（包含 CacheKey + 未序列化的 Value）
type PipelineItem struct {
	Key        CacheKey
	Value      interface{}
	Expiration time.Duration
}

// PipelineSet 批量序列化并写入多个键值对
// 底层会检查 Provider 是否实现了 IPipelineProvider，是则使用 Pipeline，否则逐个写入
//
// 示例：
//
//	err := myCache.PipelineSet(ctx, []cache.PipelineItem{
//	    {Key: cache.K("user:1"), Value: &user1, Expiration: 5 * time.Minute},
//	    {Key: cache.K("user:2"), Value: &user2, Expiration: 5 * time.Minute},
//	})
func (c *Cache) PipelineSet(ctx context.Context, items []PipelineItem) error {
	if len(items) == 0 {
		return nil
	}

	// 序列化所有 item → PipelineRawItem
	rawItems := make([]PipelineRawItem, 0, len(items))
	for _, item := range items {
		data, err := c.serializer.Serialize(item.Value)
		if err != nil {
			return fmt.Errorf("pipeline set: failed to serialize key '%s': %w", item.Key.String(), err)
		}
		rawItems = append(rawItems, PipelineRawItem{
			Key:        item.Key.String(),
			Value:      data,
			Expiration: item.Expiration,
		})
	}

	// 尝试使用 Pipeline 接口（Provider 实现了 IPipelineProvider）
	if pp, ok := c.provider.(IPipelineProvider); ok {
		return pp.PipelineSet(ctx, rawItems)
	}

	// 降级：逐个写入
	for _, raw := range rawItems {
		if err := c.provider.SetRaw(ctx, raw.Key, raw.Value, raw.Expiration); err != nil {
			return fmt.Errorf("pipeline set fallback: key '%s': %w", raw.Key, err)
		}
	}
	return nil
}

// PipelineRemove 批量删除多个键
// 底层会检查 Provider 是否实现了 IPipelineProvider，是则使用 Pipeline，否则逐个删除
//
// 示例：
//
//	err := myCache.PipelineRemove(ctx, []cache.CacheKey{
//	    cache.K("user:1"),
//	    cache.K("user:2"),
//	})
func (c *Cache) PipelineRemove(ctx context.Context, keys []CacheKey) error {
	if len(keys) == 0 {
		return nil
	}

	strKeys := make([]string, len(keys))
	for i, k := range keys {
		strKeys[i] = k.String()
	}

	// 尝试使用 Pipeline 接口
	if pp, ok := c.provider.(IPipelineProvider); ok {
		return pp.PipelineRemove(ctx, strKeys)
	}

	// 降级：逐个删除
	for _, key := range strKeys {
		if err := c.provider.Remove(ctx, key); err != nil {
			return fmt.Errorf("pipeline remove fallback: key '%s': %w", key, err)
		}
	}
	return nil
}

// PipelineRemoveS 批量删除多个字符串键（PipelineRemove 的字符串键版本）
func (c *Cache) PipelineRemoveS(ctx context.Context, keys []string) error {
	cacheKeys := make([]CacheKey, len(keys))
	for i, k := range keys {
		cacheKeys[i] = K(k)
	}
	return c.PipelineRemove(ctx, cacheKeys)
}
