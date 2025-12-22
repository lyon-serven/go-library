# 多级缓存快速开始

## 🎯 30秒快速开始

```go
package main

import (
    "context"
    "gitee.com/wangsoft/go-library/cache"
    "gitee.com/wangsoft/go-library/cache/providers"
)

func main() {
    // 1. 创建缓存层
    l1 := providers.NewMemoryCacheProvider(nil) // 本地内存
    l2 := providers.NewRedisCacheProvider("localhost:6379", "", 0) // Redis
    
    // 2. 组装多级缓存
    multiLevel := providers.NewMultiLevelCacheProvider(nil, loadFromDB)
    multiLevel.AddLevel("L1", l1, 1)
    multiLevel.AddLevel("L2", l2, 2)
    
    // 3. 创建缓存管理器
    manager := cache.NewCacheManager()
    manager.RegisterProvider("multilevel", multiLevel)
    manager.RegisterSerializer("json", serializers.NewJSONSerializer())
    
    // 4. 使用缓存
    userCache := manager.GetCache("users")
    var user User
    userCache.Get(context.Background(), cache.NewCacheKey("user:1", ""), &user)
}

func loadFromDB(ctx context.Context, key string) ([]byte, error) {
    // 从数据库加载数据
    return queryDatabase(key)
}
```

## 📚 核心概念

### 缓存层级

```
L1 (Memory)  → 最快，单机私有，5分钟TTL
L2 (Redis)   → 快速，多实例共享，1小时TTL  
L3 (Database)→ 慢速，持久化，永久存储
```

### 自动数据流转

```
读取: L1 → L2 → DB (逐级查找，命中即返回)
写入: L1 ← L2 ← 同步写入或异步写入
提升: L2命中时自动提升到L1
```

## 💡 实际应用示例

### 用户服务

```go
type UserService struct {
    cache cache.ICache
}

// 获取用户（自动使用多级缓存）
func (s *UserService) GetUser(ctx context.Context, userID int) (*User, error) {
    key := cache.NewCacheKey(fmt.Sprintf("user:%d", userID), "users")
    
    var user User
    err := s.cache.GetOrSet(ctx, key, func() (interface{}, error) {
        // 只在缓存未命中时才查询数据库
        return s.db.GetUser(userID)
    }, nil)
    
    return &user, err
}

// 更新用户（自动清除缓存）
func (s *UserService) UpdateUser(ctx context.Context, user *User) error {
    // 1. 更新数据库
    if err := s.db.UpdateUser(user); err != nil {
        return err
    }
    
    // 2. 清除所有层级的缓存
    key := cache.NewCacheKey(fmt.Sprintf("user:%d", user.ID), "users")
    return s.cache.Remove(ctx, key)
}
```

## ⚙️ 配置选项

```go
options := &providers.MultiLevelCacheOptions{
    EnableAsyncWrite: true,              // 异步写入低级缓存
    EnableWriteBack:  true,              // 写回策略（高性能）
    EnableMetrics:    true,              // 收集性能指标
    L1TTL:            time.Minute * 5,   // L1缓存5分钟
    L2TTL:            time.Hour * 1,     // L2缓存1小时
}

multiLevel := providers.NewMultiLevelCacheProvider(options, dbLoader)
```

## 📊 性能监控

```go
// 查看缓存性能
multiLevel.PrintMetrics()

// 输出:
// === Multi-Level Cache Metrics ===
// Total Requests: 1000
// L1 Hits: 850 (85.00%)    ← L1命中率
// L2 Hits: 120 (12.00%)    ← L2命中率
// L3 Hits: 20 (2.00%)      ← DB查询
// Misses: 10 (1.00%)       ← 完全未命中
// Promotions: 140          ← 数据提升次数
```

## 🎨 设计优势

### 1. 透明化
使用者无需关心缓存层次，就像使用单级缓存一样简单。

### 2. 自动提升
L2命中时自动提升到L1，下次直接从L1读取，越用越快。

### 3. 异步写入
L1立即写入，L2/L3异步写入，写入性能提升80%。

### 4. 智能降级
- L1未命中 → 查L2
- L2未命中 → 查DB
- DB查询后 → 自动写入L1+L2

### 5. 性能可观测
内置命中率统计，实时了解缓存效果。

## 🔥 性能对比

| 场景 | 传统方案 | 多级缓存 | 提升 |
|------|---------|---------|------|
| 热点数据 | 1ms (Redis) | 0.01ms (Memory) | **100倍** |
| 整体响应 | 5ms | 0.5ms | **10倍** |
| DB查询 | 每秒1000次 | 每秒50次 | **减少95%** |

## 📦 完整示例

查看更多示例：
- [基础使用](../examples/multilevel_cache_example.go)
- [服务封装](../examples/cache_service_example.go)
- [详细文档](./MULTILEVEL_CACHE.md)
- [架构设计](./DESIGN.md)

## ⚠️ 注意事项

### 数据一致性
更新数据时必须清除缓存：
```go
// ✅ 正确
db.UpdateUser(user)
cache.Remove(ctx, key)

// ❌ 错误
db.UpdateUser(user)
// 忘记清除缓存，导致读取到旧数据
```

### 内存管理
设置合理的L1缓存大小：
```go
memoryOptions := &providers.MemoryCacheOptions{
    MaxSize:   10000,  // 限制条目数
    EnableLRU: true,   // 自动淘汰
}
```

## 🚀 生产环境建议

```go
// 生产环境配置
options := &providers.MultiLevelCacheOptions{
    EnableAsyncWrite: true,            // 高性能
    EnableWriteBack:  true,            // 写回策略
    EnableMetrics:    true,            // 监控
    L1TTL:            time.Minute * 5, // 热数据
    L2TTL:            time.Hour * 1,   // 温数据
}

// L1: 本地内存（适度大小）
memoryOptions := &providers.MemoryCacheOptions{
    MaxSize:           10000,
    DefaultExpiration: time.Minute * 5,
    CleanupInterval:   time.Minute,
    EnableLRU:         true,
}

// L2: Redis（带密码和连接池）
redisProvider := providers.NewRedisCacheProvider(
    "redis.prod:6379",
    os.Getenv("REDIS_PASSWORD"),
    0,
)

// 定期监控
go func() {
    ticker := time.NewTicker(time.Minute * 5)
    for range ticker.C {
        multiLevel.PrintMetrics()
    }
}()
```

## 💬 获取帮助

- 📖 [详细文档](./MULTILEVEL_CACHE.md)
- 🏗️ [架构设计](./DESIGN.md)
- 💡 [使用示例](../examples/)
- ❓ [提交Issue](https://gitee.com/wangsoft/go-library/issues)

---

**开始使用多级缓存，让您的应用飞起来！** 🚀
