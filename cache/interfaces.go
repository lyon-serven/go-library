// Package cache 提供了一个灵活的缓存系统，支持依赖注入
package cache

import (
	"context"
	"time"
)

// CacheKey 表示一个支持命名空间的缓存键
type CacheKey struct {
	Key       string // 键名
	Namespace string // 命名空间（可选）
}

// String 返回完整的缓存键字符串
func (ck CacheKey) String() string {
	if ck.Namespace == "" {
		return ck.Key
	}
	return ck.Namespace + ":" + ck.Key
}

// NewCacheKey 创建一个新的缓存键
func NewCacheKey(key, namespace string) CacheKey {
	return CacheKey{Key: key, Namespace: namespace}
}

// K 是创建缓存键的快捷方法（只有键名）
func K(key string) CacheKey {
	return CacheKey{Key: key}
}

// NK 是创建带命名空间的缓存键的快捷方法
func NK(namespace, key string) CacheKey {
	return CacheKey{Key: key, Namespace: namespace}
}

// CacheItem 表示存储在缓存中的项
type CacheItem struct {
	Key               CacheKey       // 缓存键
	Value             interface{}    // 缓存值
	Expiration        time.Time      // 过期时间
	SlidingExpiration *time.Duration // 滑动过期时间
}

// IsExpired 检查缓存项是否已过期
func (ci *CacheItem) IsExpired() bool {
	if ci.Expiration.IsZero() {
		return false
	}
	return time.Now().After(ci.Expiration)
}

// UpdateSlidingExpiration 更新滑动过期时间
func (ci *CacheItem) UpdateSlidingExpiration() {
	if ci.SlidingExpiration != nil {
		ci.Expiration = time.Now().Add(*ci.SlidingExpiration)
	}
}

// CacheOptions 定义缓存操作的选项
type CacheOptions struct {
	AbsoluteExpiration *time.Time     // 绝对过期时间
	SlidingExpiration  *time.Duration // 滑动过期时间
	Priority           CachePriority  // 缓存优先级
}

// CachePriority 表示缓存项的优先级
type CachePriority int

const (
	// Low 低优先级缓存项在内存不足时会被优先移除
	Low CachePriority = iota
	// Normal 普通优先级缓存项（默认）
	Normal
	// High 高优先级缓存项不太可能被移除
	High
	// NeverRemove 永不移除的缓存项，不会被自动清理
	NeverRemove
)

// ICache 定义主要的缓存接口
type ICache interface {
	// Get 从缓存中获取值（返回 interface{}，JSON 序列化时会是 map[string]interface{}）
	Get(ctx context.Context, key CacheKey) (interface{}, error)

	// GetAs 从缓存中获取值并反序列化到指定类型的目标对象
	// 使用示例：var user User; err := cache.GetAs(ctx, key, &user)
	GetAs(ctx context.Context, key CacheKey, target interface{}) error

	// GetAsync 异步从缓存中获取值
	GetAsync(ctx context.Context, key CacheKey) <-chan CacheResult

	// Set 将值存储到缓存中（带选项）
	Set(ctx context.Context, key CacheKey, value interface{}, options *CacheOptions) error

	// SetAsync 异步将值存储到缓存中
	SetAsync(ctx context.Context, key CacheKey, value interface{}, options *CacheOptions) <-chan error

	// Remove 从缓存中移除指定的键
	Remove(ctx context.Context, key CacheKey) error

	// RemoveByPattern 移除所有匹配模式的键
	RemoveByPattern(ctx context.Context, pattern string) error

	// Exists 检查键是否存在于缓存中
	Exists(ctx context.Context, key CacheKey) (bool, error)

	// Clear 清空所有缓存条目
	Clear(ctx context.Context) error

	// GetOrSet 从缓存获取值，如果不存在则使用工厂函数设置
	GetOrSet(ctx context.Context, key CacheKey, factory func() (interface{}, error), options *CacheOptions) (interface{}, error)

	// GetOrSetAs 类型安全的 GetOrSet，将值反序列化到指定类型的目标对象
	// 使用示例：var user User; err := cache.GetOrSetAs(ctx, key, &user, factory, options)
	GetOrSetAs(ctx context.Context, key CacheKey, target interface{}, factory func() (interface{}, error), options *CacheOptions) error

	// Refresh 刷新缓存项的过期时间
	Refresh(ctx context.Context, key CacheKey) error
}

// ICacheExt 扩展的缓存接口，支持字符串键的便捷方法
type ICacheExt interface {
	ICache

	// GetS 使用字符串键从缓存中获取值
	GetS(ctx context.Context, key string) (interface{}, error)

	// GetAsS 使用字符串键从缓存中获取值并反序列化到指定类型
	GetAsS(ctx context.Context, key string, target interface{}) error

	// SetS 使用字符串键将值存储到缓存中
	SetS(ctx context.Context, key string, value interface{}, options *CacheOptions) error

	// RemoveS 使用字符串键从缓存中移除值
	RemoveS(ctx context.Context, key string) error

	// ExistsS 使用字符串键检查键是否存在
	ExistsS(ctx context.Context, key string) (bool, error)

	// GetOrSetS 使用字符串键从缓存获取值，如果不存在则使用工厂函数设置
	GetOrSetS(ctx context.Context, key string, factory func() (interface{}, error), options *CacheOptions) (interface{}, error)

	// GetOrSetAsS 使用字符串键的类型安全 GetOrSet
	GetOrSetAsS(ctx context.Context, key string, target interface{}, factory func() (interface{}, error), options *CacheOptions) error

	// RefreshS 使用字符串键刷新缓存项的过期时间
	RefreshS(ctx context.Context, key string) error
}

// CacheResult 表示异步缓存操作的结果
type CacheResult struct {
	Value interface{} // 缓存值
	Error error       // 错误信息
}

// ICacheProvider 定义缓存提供者的接口（Redis、内存等）
type ICacheProvider interface {
	// GetRaw 从缓存提供者获取原始数据
	GetRaw(ctx context.Context, key string) ([]byte, error)

	// SetRaw 将原始数据存储到缓存提供者
	SetRaw(ctx context.Context, key string, value []byte, expiration time.Duration) error

	// Remove 从缓存提供者移除数据
	Remove(ctx context.Context, key string) error

	// RemoveByPattern 移除所有匹配模式的键
	RemoveByPattern(ctx context.Context, pattern string) error

	// Exists 检查键是否存在
	Exists(ctx context.Context, key string) (bool, error)

	// Clear 清空所有数据
	Clear(ctx context.Context) error

	// Name 返回提供者名称
	Name() string

	// Close 关闭提供者并释放资源
	Close() error
}

// ICacheSerializer 定义缓存序列化器的接口
type ICacheSerializer interface {
	// Serialize 将对象转换为字节数组
	Serialize(value interface{}) ([]byte, error)

	// Deserialize 将字节数组转换回对象
	Deserialize(data []byte, target interface{}) error

	// Name 返回序列化器名称
	Name() string
}

// ICacheManager 定义支持依赖注入的缓存管理器接口
type ICacheManager interface {
	// GetCache 获取或创建指定名称的缓存实例
	GetCache(name string) ICache

	// RegisterProvider 注册缓存提供者
	RegisterProvider(name string, provider ICacheProvider) error

	// RegisterSerializer 注册缓存序列化器
	RegisterSerializer(name string, serializer ICacheSerializer) error

	// SetDefaultProvider 设置默认缓存提供者
	SetDefaultProvider(name string) error

	// SetDefaultSerializer 设置默认缓存序列化器
	SetDefaultSerializer(name string) error

	// Configure 配置特定缓存的自定义提供者和序列化器
	Configure(cacheName string, providerName string, serializerName string) error

	// Close 关闭所有提供者并释放资源
	Close() error
}

// CacheConfiguration 保存特定缓存的配置
type CacheConfiguration struct {
	Name           string        // 缓存名称
	ProviderName   string        // 提供者名称
	SerializerName string        // 序列化器名称
	DefaultOptions *CacheOptions // 默认选项
}

// DefaultCacheOptions 返回默认的缓存选项
func DefaultCacheOptions() *CacheOptions {
	return &CacheOptions{
		Priority: Normal,
	}
}

// WithAbsoluteExpiration 设置绝对过期时间
func (opts *CacheOptions) WithAbsoluteExpiration(expiration time.Time) *CacheOptions {
	opts.AbsoluteExpiration = &expiration
	return opts
}

// WithSlidingExpiration 设置滑动过期时间
func (opts *CacheOptions) WithSlidingExpiration(duration time.Duration) *CacheOptions {
	opts.SlidingExpiration = &duration
	return opts
}

// WithPriority 设置缓存优先级
func (opts *CacheOptions) WithPriority(priority CachePriority) *CacheOptions {
	opts.Priority = priority
	return opts
}
