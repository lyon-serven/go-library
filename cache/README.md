# Cache Management System - 缓存管理系统

基于依赖注入的缓存管理系统设计理念，支持多种缓存提供程序和序列化方式。

## ⭐ 重要：类型安全问题与解决方案

### ❓ 常见问题

> "为什么我 Set 存入一个 struct，但 Get 返回的是 map[string]interface{}？"

```go
// 存入 User 结构体
cache.Set(ctx, key, &User{ID: 1, Name: "张三"}, nil)

// ❌ 获取时变成了 map[string]interface{}
value, _ := cache.Get(ctx, key)
fmt.Printf("%T\n", value)  // map[string]interface {}
```

**原因**：JSON 序列化时会丢失类型信息，反序列化到 `interface{}` 时只能根据 JSON 结构推断类型。

📖 **详细解释**：查看 [WHY_TYPE_LOST.md](./WHY_TYPE_LOST.md) 或 [简明版](./WHY_TYPE_LOST_简明版.md)

### ✅ 解决方案（四选一）

#### 方案 1: GetAs（兼容所有 Go 版本）

```go
var user User
err := cache.GetAs(ctx, cache.K("user:1"), &user)
// user 现在是正确的 User 类型
```

#### 方案 2: GetTyped 泛型（Go 1.18+）

```go
// ✅ 直接返回正确类型，无需预创建对象
user, err := cache.GetTyped[User](ctx, cache, cache.K("user:1"))
fmt.Printf("用户: %s, 年龄: %d\n", user.Name, user.Age)
```

#### 方案 3: TypedCache 类型化包装器（Go 1.18+）

```go
// 创建类型化的用户缓存
userCache := cache.NewTypedCache[User](baseCache)

// 所有操作都是类型安全的
user, _ := userCache.Get(ctx, cache.K("1"))
user, _ := userCache.GetOrSet(ctx, cache.K("2"), loadUser, nil)
```

#### 方案 4: TypedCacheExt - 最便捷方案（Go 1.18+）⭐⭐⭐

```go
// 创建类型化的用户缓存（直接用字符串键）
userCache := cache.NewTypedCacheExt[User](baseCache)

// 🚀 直接使用字符串键，无需 cache.K()
user, _ := userCache.Get(ctx, "user:1")  // 最简洁！
user, _ := userCache.GetOrSet(ctx, "user:2", loadUser, nil)
```

**TypedCacheExt 是最推荐的方案**：类型安全 + 字符串键 + 最简洁的 API

📖 **详细说明**：查看 [TYPED_CACHE_EXT.md](./TYPED_CACHE_EXT.md)

📖 **完整文档**：
- [类型安全完整解决方案](./TYPE_SAFE_FINAL.md)
- [快速参考](./QUICK_REFERENCE.md)
- [为什么类型会丢失](./WHY_TYPE_LOST.md)

---

## ✨ 新特性：便捷的键使用方式

### 🔑 三种使用方式

**方式1: 使用快捷函数 K() 和 NK() （推荐）**

```go
// 简单键
cache.Get(ctx, cache.K("user:123"))

// 带命名空间的键
cache.Get(ctx, cache.NK("users", "123"))
```

**方式2: 使用字符串方法 ICacheExt**

```go
// 包装为支持字符串的缓存
extCache := cache.NewCacheExt(myCache)

// 直接使用字符串
extCache.GetS(ctx, "user:123")
extCache.SetS(ctx, "user:123", value, nil)
```

**方式3: 显式创建 CacheKey**

```go
key := cache.NewCacheKey("123", "users")
cache.Get(ctx, key)
```

📖 **详细使用指南**：查看 [USAGE_GUIDE.md](./USAGE_GUIDE.md)

---

## 🏗️ 架构设计

### 核心接口

- **ICacheManager**: 缓存管理器，负责依赖注入和配置管理
- **ICache**: 缓存操作接口，提供 Get/Set/Remove 等操作
- **ICacheExt**: 扩展接口，支持字符串键的便捷方法
- **ICacheProvider**: 缓存提供程序接口，支持不同的存储后端
- **ICacheSerializer**: 序列化器接口，支持不同的数据序列化方式

### 依赖注入模式

```go
// 1. 创建缓存管理器
manager := cache.NewCacheManager()

// 2. 注册提供程序
memoryProvider := providers.NewMemoryCache(nil)
manager.RegisterProvider("memory", memoryProvider)

// 3. 注册序列化器
jsonSerializer := serializers.NewJSONSerializer()
manager.RegisterSerializer("json", jsonSerializer)

// 4. 配置特定缓存
manager.Configure("user-cache", "memory", "json")

// 5. 获取配置好的缓存实例
cache := manager.GetCache("user-cache")
```

## 🔧 缓存提供程序

### 1. 内存缓存 (Memory Cache)

```go
// 基本配置
memoryCache := providers.NewMemoryCache(nil) // 使用默认配置

// 自定义配置
options := &providers.MemoryCacheOptions{
    MaxSize:           10000,                    // 最大条目数
    DefaultExpiration: time.Hour,               // 默认过期时间
    CleanupInterval:   time.Minute * 10,        // 清理间隔
    EnableLRU:         true,                    // 启用LRU淘汰
}
memoryCache := providers.NewMemoryCache(options)
```
```

**特性:**

- ✅ TTL (生存时间) 支持
- ✅ LRU (最近最少使用) 淘汰策略
- ✅ 自动清理过期条目
- ✅ 线程安全
- ✅ 内存使用统计

### 2. Redis 缓存 (Redis Cache)

```go
// Redis 配置
options := &providers.RedisOptions{
    Addresses:    []string{"localhost:6379"},
    Password:     "your-password",
    DB:           0,
    DialTimeout:  time.Second * 5,
    KeyPrefix:    "myapp",
}
redisCache, err := providers.NewRedisCache(options)

// Mock Redis (用于测试)
mockRedis := providers.NewMockRedisCache()
```

**特性:**

- ✅ 分布式缓存支持
- ✅ 键前缀支持
- ✅ 连接池管理
- ✅ 认证支持
- ✅ 多数据库支持

## 🔄 序列化器

### 1. JSON 序列化器

```go
jsonSerializer := serializers.NewJSONSerializer()
```

- 适用于: 结构体、映射、切片等可JSON化的数据
- 优点: 人类可读、跨语言兼容
- 缺点: 性能较低、体积较大

### 2. Gob 序列化器

```go
gobSerializer := serializers.NewGobSerializer()
```

- 适用于: Go 原生数据结构
- 优点: 高性能、支持复杂类型
- 缺点: 仅Go语言支持

### 3. 字符串序列化器

```go
stringSerializer := serializers.NewStringSerializer()
```

- 适用于: 简单字符串和字节数据
- 优点: 最小开销
- 缺点: 仅支持基本类型

### 4. 二进制序列化器

```go
binarySerializer := serializers.NewBinarySerializer()
```

- 适用于: 原始字节数据
- 优点: 零序列化开销
- 缺点: 仅支持 []byte 和 string

## 🚀 使用示例

### 基本用法

```go
import (
    "context"
    "mylib/cache"
    "mylib/cache/providers"
    "mylib/cache/serializers"
)

func main() {
    // 创建管理器
    manager := cache.NewCacheManager()
    defer manager.Close()
  
    // 注册组件
    manager.RegisterProvider("memory", providers.NewMemoryCache(nil))
    manager.RegisterSerializer("json", serializers.NewJSONSerializer())
  
    // 获取缓存
    myCache := manager.GetCache("default")
  
    // 基本操作
    ctx := context.Background()
    key := cache.NewCacheKey("user:123", "users")
  
    // 存储数据
    user := map[string]interface{}{
        "id":   123,
        "name": "John Doe",
        "email": "john@example.com",
    }
  
    options := cache.DefaultCacheOptions().WithSlidingExpiration(time.Hour)
    err := myCache.Set(ctx, key, user, options)
  
    // 读取数据
    value, err := myCache.Get(ctx, key)
    if err == nil && value != nil {
        fmt.Printf("User: %+v\n", value)
    }
}
```

### 高级模式

#### 1. GetOrSet 模式 (缓存穿透保护)

```go
result, err := myCache.GetOrSet(ctx, key, func() (interface{}, error) {
    // 这里执行昂贵的操作（如数据库查询）
    return fetchUserFromDatabase(userID), nil
}, options)
```

#### 2. 多缓存配置

```go
// 为不同用途配置不同的缓存
manager.Configure("user-cache", "memory", "json")      // 用户数据
manager.Configure("session-cache", "memory", "string") // 会话数据
manager.Configure("file-cache", "redis", "binary")     // 文件缓存

userCache := manager.GetCache("user-cache")
sessionCache := manager.GetCache("session-cache")
fileCache := manager.GetCache("file-cache")
```

#### 3. 异步操作

```go
// 异步获取
resultChan := myCache.GetAsync(ctx, key)
select {
case result := <-resultChan:
    if result.Error != nil {
        log.Printf("Error: %v", result.Error)
    } else {
        fmt.Printf("Value: %v", result.Value)
    }
case <-time.After(time.Second):
    fmt.Println("Timeout")
}

// 异步设置
errChan := myCache.SetAsync(ctx, key, value, options)
if err := <-errChan; err != nil {
    log.Printf("Set error: %v", err)
}
```

## 📋 缓存选项

```go
options := cache.DefaultCacheOptions()

// 绝对过期时间
options.WithAbsoluteExpiration(time.Now().Add(time.Hour))

// 滑动过期时间（访问后重新计时）
options.WithSlidingExpiration(time.Minute * 30)

// 缓存优先级
options.WithPriority(cache.High)  // Low, Normal, High, NeverRemove
```

## 🔍 监控和统计

```go
// 内存缓存统计
if memCache, ok := provider.(*providers.MemoryCache); ok {
    stats := memCache.GetStats()
    fmt.Printf("Total items: %d\n", stats["total_items"])
    fmt.Printf("Memory usage: %d bytes\n", stats["total_size"])
    fmt.Printf("Expired items: %d\n", stats["expired_items"])
}
```

## 🧪 测试支持

```go
// 使用 Mock 提供程序进行测试
func TestCacheOperations(t *testing.T) {
    manager := cache.NewCacheManager()
  
    // 注册模拟提供程序
    manager.RegisterProvider("test", providers.NewMockRedisCache())
    manager.RegisterSerializer("test", serializers.NewJSONSerializer())
  
    testCache := manager.GetCache("test-cache")
  
    // 执行测试...
}
```

## 🎯 最佳实践

### 1. 键命名约定

```go
// 使用命名空间组织键
userKey := cache.NewCacheKey("user:123", "users")
sessionKey := cache.NewCacheKey("session:abc", "sessions")
configKey := cache.NewCacheKey("app_config", "configuration")
```

### 2. 错误处理

```go
value, err := myCache.Get(ctx, key)
if err != nil {
    log.Printf("Cache error: %v", err)
    // 降级到原始数据源
    return fetchFromDatabase(key)
}
```

### 3. 缓存预热

```go
func warmUpCache(cache cache.ICache) error {
    // 预加载常用数据
    commonKeys := []string{"config", "menu", "permissions"}
  
    for _, key := range commonKeys {
        cache.GetOrSet(ctx, cache.NewCacheKey(key, "system"), func() (interface{}, error) {
            return loadSystemData(key), nil
        }, nil)
    }
    return nil
}
```

### 4. 资源清理

```go
func main() {
    manager := cache.NewCacheManager()
  
    // 确保程序退出时清理资源
    defer func() {
        if err := manager.Close(); err != nil {
            log.Printf("Error closing cache manager: %v", err)
        }
    }()
  
    // 应用程序逻辑...
}
```

## 🔮 扩展点

### 自定义提供程序

```go
type MyCustomProvider struct {
    // 实现 ICacheProvider 接口
}

func (mcp *MyCustomProvider) GetRaw(ctx context.Context, key string) ([]byte, error) {
    // 自定义获取逻辑
}

func (mcp *MyCustomProvider) SetRaw(ctx context.Context, key string, value []byte, expiration time.Duration) error {
    // 自定义存储逻辑
}

// ... 实现其他接口方法
```

### 自定义序列化器

```go
type MyCustomSerializer struct {
    // 实现 ICacheSerializer 接口
}

func (mcs *MyCustomSerializer) Serialize(value interface{}) ([]byte, error) {
    // 自定义序列化逻辑
}

func (mcs *MyCustomSerializer) Deserialize(data []byte, target interface{}) error {
    // 自定义反序列化逻辑
}
```

## 📊 性能特性

- **内存缓存**: 纳秒级访问速度，支持高并发
- **TTL 管理**: 自动后台清理，不影响主业务逻辑
- **LRU 淘汰**: 智能内存管理，防止内存溢出
- **异步操作**: 支持非阻塞操作，提升应用响应性

## 🔧 配置建议

### 开发环境

```go
options := &providers.MemoryCacheOptions{
    MaxSize:           1000,
    DefaultExpiration: time.Minute * 5,
    CleanupInterval:   time.Minute,
    EnableLRU:         true,
}
```

### 生产环境

```go
options := &providers.MemoryCacheOptions{
    MaxSize:           100000,
    DefaultExpiration: time.Hour,
    CleanupInterval:   time.Minute * 10,
    EnableLRU:         true,
}

// 或使用 Redis 进行分布式缓存
redisOptions := &providers.RedisOptions{
    Addresses:    []string{"redis-cluster-1:6379", "redis-cluster-2:6379"},
    PoolSize:     50,
    KeyPrefix:    "myapp:prod",
}
```

这个缓存系统提供了灵活的架构，支持不同的缓存策略和序列化方式，可以根据具体的业务需求进行定制和扩展。
