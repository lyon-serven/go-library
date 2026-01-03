package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"gitee.com/wangsoft/go-library/cache"
	"gitee.com/wangsoft/go-library/cache/providers"
	"gitee.com/wangsoft/go-library/cache/serializers"
)

// User represents a sample data structure
type User struct {
	ID       int
	Username string
	Email    string
	Created  time.Time
}

// DatabaseLoader simulates loading user from database
func loadUserFromDatabase(ctx context.Context, key string) ([]byte, error) {
	fmt.Printf("📊 Loading from database: %s\n", key)

	// Simulate database query
	time.Sleep(100 * time.Millisecond)

	// Simulate user data
	user := User{
		ID:       123,
		Username: "john_doe",
		Email:    "john@example.com",
		Created:  time.Now(),
	}

	// Serialize user
	serializer := serializers.NewJSONSerializer()
	data, err := serializer.Serialize(user)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func main() {
	fmt.Println("=== 多级缓存示例 ===\n")

	// 1. 创建各级缓存提供者

	// L1: 内存缓存（最快）
	memoryProvider := providers.NewMemoryCache(providers.DefaultMemoryCacheOptions())
	fmt.Println("✅ L1 (Memory Cache) 已创建")

	// L2: Redis缓存（共享缓存，这里用内存模拟）
	// 实际使用时应该是：
	// redisProvider := providers.NewRedisCache("localhost:6379", "", 0)
	redisProvider := providers.NewMemoryCache(&providers.MemoryCacheOptions{
		MaxSize:           5000,
		DefaultExpiration: time.Hour,
		CleanupInterval:   time.Minute * 15,
		EnableLRU:         true,
	})
	fmt.Println("✅ L2 (Redis Cache) 已创建")

	// 2. 创建多级缓存提供者
	multiLevelOptions := providers.DefaultMultiLevelCacheOptions()
	multiLevelOptions.EnableAsyncWrite = true // 启用异步写入
	multiLevelOptions.EnableWriteBack = true  // 启用写回策略
	multiLevelOptions.EnableMetrics = true    // 启用性能指标
	multiLevelOptions.L1TTL = time.Minute * 5 // L1缓存5分钟
	multiLevelOptions.L2TTL = time.Hour * 1   // L2缓存1小时

	multiLevelProvider := providers.NewMultiLevelCacheProvider(
		multiLevelOptions,
		loadUserFromDatabase, // 数据库加载器
	)

	// 添加缓存级别
	multiLevelProvider.AddLevel("L1-Memory", memoryProvider, 1)
	multiLevelProvider.AddLevel("L2-Redis", redisProvider, 2)
	fmt.Println("✅ 多级缓存提供者已配置\n")

	// 3. 创建缓存管理器
	cacheManager := cache.NewCacheManager()

	// 注册提供者和序列化器
	cacheManager.RegisterProvider("multilevel", multiLevelProvider)
	cacheManager.RegisterSerializer("json", serializers.NewJSONSerializer())

	// 设置默认提供者
	cacheManager.SetDefaultProvider("multilevel")
	cacheManager.SetDefaultSerializer("json")

	// 4. 获取缓存实例
	userCache := cacheManager.GetCache("user-cache")

	ctx := context.Background()
	userKey := cache.NewCacheKey("user:123", "users")

	// 5. 第一次获取（会从数据库加载）
	fmt.Println("【第1次请求】从数据库加载...")
	value1, err := userCache.Get(ctx, userKey)
	if err != nil {
		log.Printf("获取失败: %v", err)
	} else if value1 != nil {
		fmt.Printf("✅ 获取到数据: %+v\n\n", value1)
	}

	// 6. 第二次获取（应该从L1缓存获取）
	fmt.Println("【第2次请求】从L1缓存获取...")
	value2, err := userCache.Get(ctx, userKey)
	if err != nil {
		log.Printf("获取失败: %v", err)
	} else if value2 != nil {
		fmt.Printf("✅ 快速获取到数据: %+v\n\n", value2)
	}

	// 7. 使用GetOrSet模式
	fmt.Println("【使用GetOrSet】...")
	productKey := cache.NewCacheKey("product:456", "products")

	product, err := userCache.GetOrSet(ctx, productKey, func() (interface{}, error) {
		fmt.Println("  → 执行数据加载逻辑...")
		return map[string]interface{}{
			"id":    456,
			"name":  "商品A",
			"price": 99.99,
		}, nil
	}, cache.DefaultCacheOptions())

	if err != nil {
		log.Printf("GetOrSet失败: %v", err)
	} else {
		fmt.Printf("✅ 获取到产品: %+v\n\n", product)
	}

	// 再次获取，应该从缓存
	fmt.Println("【再次GetOrSet】从缓存获取...")
	product2, err := userCache.GetOrSet(ctx, productKey, func() (interface{}, error) {
		fmt.Println("  → 这不应该被执行")
		return nil, nil
	}, cache.DefaultCacheOptions())

	if err != nil {
		log.Printf("GetOrSet失败: %v", err)
	} else {
		fmt.Printf("✅ 从缓存获取产品: %+v\n\n", product2)
	}

	// 8. 显示性能指标
	fmt.Println("【性能指标】")
	if metrics := multiLevelProvider.GetMetrics(); metrics != nil {
		fmt.Printf("L1命中: %d, L2命中: %d, L3命中: %d\n",
			metrics.L1Hits, metrics.L2Hits, metrics.L3Hits)
		fmt.Printf("总请求: %d, 未命中: %d\n",
			metrics.TotalRequests, metrics.Misses)

		if metrics.TotalRequests > 0 {
			hitRate := float64(metrics.L1Hits+metrics.L2Hits+metrics.L3Hits) / float64(metrics.TotalRequests) * 100
			fmt.Printf("整体命中率: %.2f%%\n", hitRate)
		}
	}

	// 9. 测试删除操作
	fmt.Println("\n【测试删除】")
	err = userCache.Delete(ctx, userKey)
	if err != nil {
		log.Printf("删除失败: %v", err)
	} else {
		fmt.Println("✅ 成功删除键")
	}

	exists, _ := userCache.Exists(ctx, userKey)
	fmt.Printf("删除后键是否存在: %v\n", exists)

	fmt.Println("\n✅ 多级缓存示例完成！")
}

/*
运行方式：
go run multilevel_cache_example.go

输出示例：
=== 多级缓存示例 ===

✅ L1 (Memory Cache) 已创建
✅ L2 (Redis Cache) 已创建
✅ 多级缓存提供者已配置

【第1次请求】从数据库加载...
📊 Loading from database: users:user:123
✅ 获取到数据: map[Created:2025-12-31 10:20:30 Email:john@example.com ID:123 Username:john_doe]

【第2次请求】从L1缓存获取...
✅ 快速获取到数据: map[Created:2025-12-31 10:20:30 Email:john@example.com ID:123 Username:john_doe]

【使用GetOrSet】...
  → 执行数据加载逻辑...
✅ 获取到产品: map[id:456 name:商品A price:99.99]

【再次GetOrSet】从缓存获取...
✅ 从缓存获取产品: map[id:456 name:商品A price:99.99]

【性能指标】
L1命中: 3, L2命中: 0, L3命中: 0
总请求: 4, 未命中: 1
整体命中率: 75.00%

【测试删除】
✅ 成功删除键
删除后键是否存在: false

✅ 多级缓存示例完成！

说明：
- 第1次请求会从数据库加载（L3）
- 第2次请求直接从L1内存缓存获取，非常快
- GetOrSet模式在缓存未命中时执行加载函数
- 多级缓存提供了优秀的性能和灵活性
*/
