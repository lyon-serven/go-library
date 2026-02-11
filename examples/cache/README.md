# Cache Examples - 缓存使用示例

本目录包含了 go-library 缓存模块的各种使用示例。

## 📁 示例目录

### 基础示例

- **cache_basic_example.go** - 基础缓存操作（Set/Get/Remove）
- **cache_convenient_example.go** - 便捷方法示例（GetOrSet）
- **main.go** - 多种缓存提供者示例（Memory/Redis/Multilevel）

### 类型安全示例

- **type_safe_example.go** - GetAs/GetOrSetAs 方法示例（适用所有 Go 版本）
- **generic_example.go** - GetTyped/TypedCache 泛型方法示例（Go 1.18+，推荐）

### 性能测试报告

- **serializer_bench/** - 📊 **JSON vs Gob 序列化器性能对比报告**
  - 详细的性能测试数据
  - 存储大小对比
  - "大小"概念的详细解释
  - 使用场景建议
  - 查看报告: [serializer_bench/README.md](serializer_bench/README.md)

### 多级缓存示例

- **multilevel_cache_example.go** - 多级缓存使用示例（L1内存 + L2 Redis）

## 🚀 快速开始

### 1. 基础使用

```bash
# 运行基础示例
go run cache_basic_example.go
```

### 2. 类型安全方案（推荐）

```bash
# 泛型方案（Go 1.18+）
go run generic_example.go

# GetAs 方案（所有 Go 版本）
go run type_safe_example.go
```

### 3. 查看性能对比报告

查看 [serializer_bench/README.md](serializer_bench/README.md) 了解 JSON vs Gob 的详细性能对比。

## 📊 序列化器选择建议

根据实际性能测试，我们发现：

### 性能数据（惊人发现！）

| 指标 | JSON | Gob | 胜者 |
|-----|------|-----|-----|
| 序列化速度 | 2.2µs | 6.3µs | **JSON 快 2.9x** ✅ |
| 反序列化速度 | 3.7µs | 19.7µs | **JSON 快 5.3x** ✅ |
| 数据大小 | 266字节 | 278字节 | **JSON 小 4.5%** ✅ |
| 跨语言兼容 | ✓ | ✗ | **JSON** ✅ |
| 人类可读 | ✓ | ✗ | **JSON** ✅ |
| 类型保留 | 用GetTyped | 原生 | 平局 |

### 推荐方案

**🏆 最佳方案：JSON + GetTyped**

```go
// 1. 使用 JSON 序列化器（性能更好）
cache := cache.NewCache(
    cache.WithMemoryProvider(),
    cache.WithJSONSerializer(), // 推荐！
)

// 2. 使用 GetTyped 获取类型安全的结果（Go 1.18+）
user, err := cache.GetTyped[User](ctx, cache, "user:1")
// user 是 *User 类型，不是 map[string]interface{}

// 3. 或者使用 GetAs（所有 Go 版本）
var user User
err := cache.GetAs(ctx, "user:1", &user)
```

**详细性能对比报告：** 查看 [serializer_bench/README.md](serializer_bench/README.md)

## 🎯 示例说明

### 类型丢失问题

当使用 JSON 序列化器时，如果直接 Get，会遇到类型丢失问题：

```go
// ❌ 问题：类型丢失
cache.Set(ctx, "user", user) // 存入 User 对象
val, _ := cache.Get(ctx, "user")
// val 是 map[string]interface{}，不是 User

// ✅ 解决方案 1：GetTyped（Go 1.18+，推荐）
user, _ := cache.GetTyped[User](ctx, cache, "user")

// ✅ 解决方案 2：GetAs（所有 Go 版本）
var user User
cache.GetAs(ctx, "user", &user)

// ✅ 解决方案 3：TypedCache（Go 1.18+，最优雅）
userCache := cache.NewTypedCache[User](cache)
user, _ := userCache.Get(ctx, "user")
```

**深入理解：** 阅读 [cache/WHY_TYPE_LOST.md](../../cache/WHY_TYPE_LOST.md) 或 [cache/WHY_TYPE_LOST_简明版.md](../../cache/WHY_TYPE_LOST_简明版.md)

## 📚 相关文档

### 类型安全相关
- [WHY_TYPE_LOST.md](../../cache/WHY_TYPE_LOST.md) - 类型丢失问题详解（15分钟深度阅读）
- [WHY_TYPE_LOST_简明版.md](../../cache/WHY_TYPE_LOST_简明版.md) - 快速理解（5分钟）
- [GENERIC_SOLUTION.md](../../cache/GENERIC_SOLUTION.md) - 泛型解决方案
- [TYPE_SAFE_FINAL.md](../../cache/TYPE_SAFE_FINAL.md) - 完整解决方案总结

### 性能相关
- [serializer_bench/README.md](serializer_bench/README.md) - 序列化器性能对比报告

### 快速参考
- [QUICK_REFERENCE.md](../../cache/QUICK_REFERENCE.md) - API 快速参考手册
- [文档索引.md](../../cache/文档索引.md) - 完整文档导航

## 🏃 运行所有示例

```bash
# 在 examples/cache 目录下运行
cd examples/cache

# 基础示例
go run cache_basic_example.go
go run cache_convenient_example.go

# 类型安全示例
go run type_safe_example.go    # GetAs 方案
go run generic_example.go       # GetTyped 方案（推荐）

# 多级缓存
go run multilevel_cache_example.go
```

## ❓ 常见问题

### Q1: 为什么推荐 JSON 而不是 Gob？

A: 经过实际测试发现：
- JSON 序列化速度快 2.9倍
- JSON 反序列化速度快 5.3倍
- JSON 数据更小（小 4.5%）
- JSON 跨语言兼容、人类可读

唯一的"问题"是类型丢失，但用 GetTyped 可以完美解决。

详见：[serializer_bench/README.md](serializer_bench/README.md)

### Q2: GetTyped 和 GetAs 有什么区别？

A: 
- **GetTyped**：Go 1.18+ 泛型，直接返回正确类型
  ```go
  user, err := cache.GetTyped[User](ctx, cache, key)
  ```
- **GetAs**：兼容所有 Go 版本，需要传入目标对象
  ```go
  var user User
  err := cache.GetAs(ctx, key, &user)
  ```

推荐使用 GetTyped（如果你的 Go 版本 >= 1.18）

### Q3: 什么时候使用 TypedCache？

A: 当你的整个缓存只存一种类型时：

```go
// 创建一个专门存 User 的缓存
userCache := cache.NewTypedCache[User](baseCache)

// 所有操作都是类型安全的
userCache.Set(ctx, "user:1", user)
user, _ := userCache.Get(ctx, "user:1")  // 返回 *User
```

### Q4: 多级缓存如何使用？

A: 
```go
cache := cache.NewCache(
    cache.WithMultilevelProvider(
        cache.NewMemoryProvider(),  // L1: 内存
        cache.NewRedisProvider(),    // L2: Redis
    ),
)
```

详见：[multilevel_cache_example.go](multilevel_cache_example.go)

## 🔗 相关链接

- [GitHub Repository](https://github.com/yourusername/go-library)
- [Cache Module Documentation](../../cache/README.md)
- [Complete Documentation Index](../../cache/文档索引.md)
