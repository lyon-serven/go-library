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

// RefreshS 使用字符串键刷新过期时间
func (c *Cache) RefreshS(ctx context.Context, key string) error {
	return c.Refresh(ctx, K(key))
}

// ============================================
// 泛型函数 - 类型安全的缓存操作（Go 1.18+）
// ============================================
// 注意：由于 Go 的限制，泛型只能是函数，不能是方法
// 但这些函数设计为与 Cache 紧密配合使用

// GetTyped 使用泛型从缓存中获取值，直接返回指定类型
// 不需要预先创建对象，使用更加直观
//
// 使用示例：
//
//	user, err := cache.GetTyped[User](ctx, myCache, cache.K("user:1"))
//	fmt.Printf("用户: %s\n", user.Name)
func GetTyped[T any](ctx context.Context, c *Cache, key CacheKey) (*T, error) {
	// 获取原始数据
	data, err := c.provider.GetRaw(ctx, key.String())
	if err != nil {
		return nil, err
	}

	if data == nil {
		return nil, fmt.Errorf("key '%s' not found in cache", key.String())
	}

	// 反序列化到指定类型
	var result T
	err = c.serializer.Deserialize(data, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize cache value: %w", err)
	}

	return &result, nil
}

// GetTypedS 使用字符串键和泛型从缓存中获取值
//
// 使用示例：
//
//	user, err := cache.GetTypedS[User](ctx, myCache, "user:1")
func GetTypedS[T any](ctx context.Context, c *Cache, key string) (*T, error) {
	return GetTyped[T](ctx, c, K(key))
}

// GetOrSetTyped 使用泛型的 GetOrSet，直接返回指定类型
// 如果缓存中没有值，则调用工厂函数生成并缓存
//
// 使用示例：
//
//	user, err := cache.GetOrSetTyped[User](ctx, myCache, cache.K("user:1"),
//	    func() (*User, error) {
//	        return loadUserFromDB(1)
//	    }, nil)
func GetOrSetTyped[T any](ctx context.Context, c *Cache, key CacheKey, factory func() (*T, error), options *CacheOptions) (*T, error) {
	// 尝试从缓存获取
	result, err := GetTyped[T](ctx, c, key)
	if err == nil {
		return result, nil // 缓存命中
	}

	// 缓存未命中，使用工厂函数生成
	value, err := factory()
	if err != nil {
		return nil, err
	}

	// 保存到缓存
	err = c.Set(ctx, key, value, options)
	if err != nil {
		// 即使缓存失败也返回值
		// 可以选择记录日志
	}

	return value, nil
}

// GetOrSetTypedS 使用字符串键和泛型的 GetOrSet
//
// 使用示例：
//
//	user, err := cache.GetOrSetTypedS[User](ctx, myCache, "user:1",
//	    func() (*User, error) {
//	        return loadUserFromDB(1)
//	    }, nil)
func GetOrSetTypedS[T any](ctx context.Context, c *Cache, key string, factory func() (*T, error), options *CacheOptions) (*T, error) {
	return GetOrSetTyped[T](ctx, c, K(key), factory, options)
}
