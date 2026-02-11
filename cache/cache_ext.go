package cache

import "context"

// cacheExtWrapper 实现 ICacheExt 接口，为 ICache 添加字符串键的便捷方法
type cacheExtWrapper struct {
	ICache
}

// NewCacheExt 将普通的 ICache 包装为支持扩展方法的 ICacheExt
func NewCacheExt(cache ICache) ICacheExt {
	return &cacheExtWrapper{ICache: cache}
}

// GetS 使用字符串键从缓存中获取值
func (c *cacheExtWrapper) GetS(ctx context.Context, key string) (interface{}, error) {
	return c.Get(ctx, K(key))
}

// GetAsS 使用字符串键从缓存中获取值并反序列化到指定类型
func (c *cacheExtWrapper) GetAsS(ctx context.Context, key string, target interface{}) error {
	return c.GetAs(ctx, K(key), target)
}

// SetS 使用字符串键将值存储到缓存中
func (c *cacheExtWrapper) SetS(ctx context.Context, key string, value interface{}, options *CacheOptions) error {
	return c.Set(ctx, K(key), value, options)
}

// RemoveS 使用字符串键从缓存中移除值
func (c *cacheExtWrapper) RemoveS(ctx context.Context, key string) error {
	return c.Remove(ctx, K(key))
}

// ExistsS 使用字符串键检查键是否存在
func (c *cacheExtWrapper) ExistsS(ctx context.Context, key string) (bool, error) {
	return c.Exists(ctx, K(key))
}

// GetOrSetS 使用字符串键从缓存获取值，如果不存在则使用工厂函数设置
func (c *cacheExtWrapper) GetOrSetS(ctx context.Context, key string, factory func() (interface{}, error), options *CacheOptions) (interface{}, error) {
	return c.GetOrSet(ctx, K(key), factory, options)
}

// GetOrSetAsS 使用字符串键的类型安全 GetOrSet
func (c *cacheExtWrapper) GetOrSetAsS(ctx context.Context, key string, target interface{}, factory func() (interface{}, error), options *CacheOptions) error {
	return c.GetOrSetAs(ctx, K(key), target, factory, options)
}

// RefreshS 使用字符串键刷新缓存项的过期时间
func (c *cacheExtWrapper) RefreshS(ctx context.Context, key string) error {
	return c.Refresh(ctx, K(key))
}

// ============================================
// TypedCacheExt - 泛型缓存扩展包装器（Go 1.18+）
// ============================================

// TypedCacheExt 提供类型安全的缓存扩展操作，支持字符串键
// 适用于某个缓存实例只存储一种类型的场景，并且希望使用字符串键
//
// 使用示例：
//
//	userCache := cache.NewTypedCacheExt[User](manager.GetCache("users"))
//	user, err := userCache.Get(ctx, "1")           // 直接用字符串键
//	user, err := userCache.GetOrSet(ctx, "2", loadUser, nil)
type TypedCacheExt[T any] struct {
	cache ICache
}

// NewTypedCacheExt 创建一个类型化的缓存扩展包装器
// 相比 TypedCache，这个版本的所有方法都直接使用字符串键，更加便捷
func NewTypedCacheExt[T any](cache ICache) *TypedCacheExt[T] {
	return &TypedCacheExt[T]{cache: cache}
}

// Get 使用字符串键获取值，返回类型 T
func (tc *TypedCacheExt[T]) Get(ctx context.Context, key string) (*T, error) {
	return GetTyped[T](ctx, tc.cache, K(key))
}

// Set 使用字符串键设置值
func (tc *TypedCacheExt[T]) Set(ctx context.Context, key string, value *T, options *CacheOptions) error {
	return tc.cache.Set(ctx, K(key), value, options)
}

// GetOrSet 使用字符串键从缓存获取或设置值
func (tc *TypedCacheExt[T]) GetOrSet(ctx context.Context, key string, factory func() (*T, error), options *CacheOptions) (*T, error) {
	return GetOrSetTyped[T](ctx, tc.cache, K(key), factory, options)
}

// Remove 使用字符串键删除值
func (tc *TypedCacheExt[T]) Remove(ctx context.Context, key string) error {
	return tc.cache.Remove(ctx, K(key))
}

// Exists 使用字符串键检查是否存在
func (tc *TypedCacheExt[T]) Exists(ctx context.Context, key string) (bool, error) {
	return tc.cache.Exists(ctx, K(key))
}

// Clear 清空所有缓存
func (tc *TypedCacheExt[T]) Clear(ctx context.Context) error {
	return tc.cache.Clear(ctx)
}

// Refresh 使用字符串键刷新过期时间
func (tc *TypedCacheExt[T]) Refresh(ctx context.Context, key string) error {
	return tc.cache.Refresh(ctx, K(key))
}

// RemoveByPattern 移除所有匹配模式的键
func (tc *TypedCacheExt[T]) RemoveByPattern(ctx context.Context, pattern string) error {
	return tc.cache.RemoveByPattern(ctx, pattern)
}
