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

// PipelineRemoveS 使用字符串键批量删除多个键
func (c *cacheExtWrapper) PipelineRemoveS(ctx context.Context, keys []string) error {
	cacheKeys := make([]CacheKey, len(keys))
	for i, k := range keys {
		cacheKeys[i] = K(k)
	}
	return c.PipelineRemove(ctx, cacheKeys)
}
