# 多级缓存使用指南

## 📚 概述

多级缓存（Multi-Level Cache）提供了一个优雅的解决方案，将本地内存缓存、Redis缓存和数据库查询组合在一起，形成一个透明的缓存层次结构。

## 🏗️ 架构设计

### 缓存层次结构

```
┌─────────────────────────────────────────────────────────────┐
│                     应用程序请求                              │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                  多级缓存提供者                               │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  L1: 本地内存缓存 (最快, 5分钟TTL)                    │   │
│  │  • 单机私有                                           │   │
│  │  • 纳秒级响应                                         │   │
│  │  • LRU淘汰策略                                        │   │
│  └──────────────────────────────────────────────────────┘   │
│                            ↓ (缓存未命中)                    │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  L2: Redis缓存 (快速, 1小时TTL)                       │   │
│  │  • 多实例共享                                         │   │
│  │  • 微秒级响应                                         │   │
│  │  • 分布式缓存                                         │   │
│  └──────────────────────────────────────────────────────┘   │
│                            ↓ (缓存未命中)                    │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  L3: 数据库查询 (慢, 通过Loader加载)                  │   │
│  │  • 持久化存储                                         │   │
│  │  • 毫秒级响应                                         │   │
│  │  • 数据源头                                           │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

## ✨ 核心特性

### 1. **自动数据提升（Cache Warming）**
当从低级缓存命中时，自动将数据提升到高级缓存，提高后续访问速度。

```
读取流程：L1未命中 → L2命中 → 自动提升数据到L1 → 下次直接从L1读取
```

### 2. **灵活的写入策略**

#### Write-Back（写回策略）- 推荐
- 先写入L1，立即返回
- 异步写入L2和L3
- 性能最优，适合高并发场景

#### Write-Through（写穿策略）
- 同步写入所有缓存层
- 数据一致性最强
- 适合数据一致性要求高的场景

### 3. **异步写入队列**
- 后台Worker异步处理写入任务
- 不阻塞主流程
- 可配置队列大小和Worker数量

### 4. **性能指标收集**
- L1/L2/L3命中率统计
- 缓存未命中率
- 写入次数统计
- 数据提升次数统计

### 5. **可配置的TTL策略**
- 不同缓存层设置不同的过期时间
- L1: 短期缓存（分钟级）
- L2: 中期缓存（小时级）
- L3: 长期/永久存储

## 🚀 快速开始

### 1. 基本使用

```go
package main

import (
    "context"
    "time"
    
    "github.com/lyon-serven/go-library/cache"
    "github.com/lyon-serven/go-library/cache/providers"
    "github.com/lyon-serven/go-library/cache/serializers"
)

func main() {
    // 创建L1: 内存缓存
    memoryProvider := providers.NewMemoryCacheProvider(
        providers.DefaultMemoryCacheOptions(),
    )
    
    // 创建L2: Redis缓存
    redisProvider := providers.NewRedisCacheProvider(
        "localhost:6379", 
        "", // password
        0,  // database
    )
    
    // 创建多级缓存提供者
    options := providers.DefaultMultiLevelCacheOptions()
    options.L1TTL = time.Minute * 5  // L1缓存5分钟
    options.L2TTL = time.Hour * 1    // L2缓存1小时
    
    multiLevel := providers.NewMultiLevelCacheProvider(
        options,
        loadFromDatabase, // 数据库加载函数
    )
    
    // 添加缓存层
    multiLevel.AddLevel("L1-Memory", memoryProvider, 1)
    multiLevel.AddLevel("L2-Redis", redisProvider, 2)
    
    // 创建缓存管理器
    manager := cache.NewCacheManager()
    manager.RegisterProvider("multilevel", multiLevel)
    manager.RegisterSerializer("json", serializers.NewJSONSerializer())
    
    // 使用缓存
    cache := manager.GetCache("my-cache")
    
    // 读取数据（自动从L1 → L2 → DB查找）
    var data MyData
    err := cache.Get(context.Background(), 
        cache.NewCacheKey("key1", "namespace"), 
        &data,
    )
    
    defer manager.Close()
}

// 数据库加载函数
func loadFromDatabase(ctx context.Context, key string) ([]byte, error) {
    // 从数据库加载数据
    // ...
    return data, nil
}
```

### 2. 完整示例

参考 `examples/multilevel_cache_example.go`

## ⚙️ 配置选项

### MultiLevelCacheOptions

```go
type MultiLevelCacheOptions struct {
    // 启用异步写入
    EnableAsyncWrite bool
    
    // 启用自动同步（定期同步不同层级）
    EnableAutoSync bool
    
    // 同步间隔
    SyncInterval time.Duration
    
    // 写入层数（0表示写入所有层）
    WriteDownLevels int
    
    // 启用性能指标
    EnableMetrics bool
    
    // L1缓存TTL
    L1TTL time.Duration
    
    // L2缓存TTL
    L2TTL time.Duration
    
    // 启用写回策略
    EnableWriteBack bool
}
```

### 默认配置

```go
options := providers.DefaultMultiLevelCacheOptions()
// EnableAsyncWrite: true
// EnableAutoSync: false
// WriteDownLevels: 0 (所有层)
// EnableMetrics: true
// L1TTL: 5分钟
// L2TTL: 1小时
// EnableWriteBack: true
```

## 📊 性能优化建议

### 1. TTL设置策略

| 缓存层 | 推荐TTL | 适用场景 |
|--------|---------|----------|
| L1 (Memory) | 1-10分钟 | 热点数据、高频访问 |
| L2 (Redis) | 10分钟-2小时 | 共享数据、中频访问 |
| L3 (Database) | - | 冷数据、持久化存储 |

### 2. 写入策略选择

| 策略 | 性能 | 一致性 | 适用场景 |
|------|------|--------|----------|
| Write-Back | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | 高并发、读多写少 |
| Write-Through | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | 金融、订单等强一致性场景 |

### 3. 缓存层数建议

```
2层（Memory + Redis）：适合大多数场景
3层（Memory + Redis + DB）：需要数据库缓冲的场景
```

## 📈 性能指标

### 查看性能指标

```go
// 获取指标
metrics := multiLevelProvider.GetMetrics()

fmt.Printf("L1命中率: %.2f%%\n", 
    float64(metrics.L1Hits)/float64(metrics.TotalRequests)*100)

// 或直接打印
multiLevelProvider.PrintMetrics()
```

### 输出示例

```
=== Multi-Level Cache Metrics ===
Total Requests: 1000
L1 Hits: 850 (85.00%)
L2 Hits: 120 (12.00%)
L3 Hits: 20 (2.00%)
Misses: 10 (1.00%)
L1 Writes: 170
L2 Writes: 170
L3 Writes: 0
Promotions: 140
================================
```

## 🎯 使用场景

### 1. 用户信息缓存

```go
// L1: 5分钟，当前服务实例的活跃用户
// L2: 1小时，所有服务实例共享的用户数据
// L3: 数据库，用户持久化数据

userCache := manager.GetCache("user-cache")
userKey := cache.NewCacheKey(fmt.Sprintf("user:%d", userID), "users")

var user User
err := userCache.Get(ctx, userKey, &user)
```

### 2. 商品信息缓存

```go
// L1: 10分钟，热门商品
// L2: 2小时，所有商品信息
// L3: 数据库，商品持久化数据

productCache := manager.GetCache("product-cache")
productKey := cache.NewCacheKey(fmt.Sprintf("product:%d", productID), "products")
```

### 3. 配置信息缓存

```go
// L1: 5分钟，应用配置
// L2: 1小时，共享配置
// L3: 配置中心，配置数据源

configCache := manager.GetCache("config-cache")
configKey := cache.NewCacheKey("app:config", "configs")
```

## 🔧 高级用法

### 1. 自定义数据库加载器

```go
// 带上下文的数据库加载器
func loadUserFromDB(ctx context.Context, key string) ([]byte, error) {
    // 解析key
    userID := parseUserIDFromKey(key)
    
    // 查询数据库
    var user User
    err := db.WithContext(ctx).
        Where("id = ?", userID).
        First(&user).Error
    if err != nil {
        return nil, err
    }
    
    // 序列化
    serializer := serializers.NewJSONSerializer()
    return serializer.Serialize(user)
}
```

### 2. 批量操作

```go
// 批量预热缓存
func warmupCache(cache ICache, userIDs []int) {
    for _, id := range userIDs {
        key := cache.NewCacheKey(fmt.Sprintf("user:%d", id), "users")
        
        // 触发加载（GetOrSet确保只加载一次）
        cache.GetOrSet(ctx, key, func() (interface{}, error) {
            return loadUserFromDB(ctx, key.String())
        }, nil)
    }
}
```

### 3. 动态调整TTL

```go
// 热点数据使用更短的L1 TTL
hotDataOptions := &CacheOptions{}
hotDataOptions.WithSlidingExpiration(time.Minute * 2)

cache.Set(ctx, hotKey, hotData, hotDataOptions)
```

## ⚠️ 注意事项

### 1. 数据一致性

多级缓存可能存在短暂的数据不一致：
- L1和L2可能有不同的数据版本
- 更新数据时需要主动清除所有层级的缓存

```go
// 更新数据后清除缓存
multiLevelProvider.Remove(ctx, key)

// 或按模式清除
multiLevelProvider.RemoveByPattern(ctx, "user:*")
```

### 2. 内存管理

- L1内存缓存会占用应用内存
- 设置合理的MaxSize避免OOM
- 启用LRU自动淘汰

```go
memoryOptions := &providers.MemoryCacheOptions{
    MaxSize:   10000,  // 最多10000个条目
    EnableLRU: true,   // 启用LRU淘汰
}
```

### 3. 异步写入

- 异步写入可能导致短暂的数据延迟
- 关键数据考虑使用同步写入（禁用EnableWriteBack）

## 🔍 故障排查

### 1. 缓存命中率低

```go
// 检查指标
metrics := multiLevelProvider.GetMetrics()
if float64(metrics.Misses)/float64(metrics.TotalRequests) > 0.1 {
    // 未命中率 > 10%，需要优化
    // - 增加L1 TTL
    // - 增加L1缓存大小
    // - 检查数据访问模式
}
```

### 2. 内存占用过高

```go
// 减小L1缓存大小
memoryOptions.MaxSize = 5000

// 缩短L1 TTL
options.L1TTL = time.Minute * 2
```

### 3. 异步写入队列阻塞

```go
// 增加Worker数量（在创建时修改）
// 或减少异步写入，改用同步写入
options.EnableAsyncWrite = false
```

## 📚 相关文档

- [Cache包基础使用](./README.md)
- [Memory Provider](./providers/memory.go)
- [Redis Provider](./providers/redis.go)
- [完整示例](../examples/multilevel_cache_example.go)

## 🤝 贡献

欢迎提交Issue和PR来改进多级缓存实现！

## 📄 许可证

MIT License
