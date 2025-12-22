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

	// 5. 演示多级缓存工作流程

	fmt.Println("--- 第一次请求（缓存未命中，从数据库加载）---")
	startTime := time.Now()

	value, err := userCache.Get(ctx, userKey)
	if err != nil || value == nil {
		log.Printf("❌ 缓存未命中，设置初始数据\n")
		// 第一次获取会失败，需要设置
		user := User{
			ID:       123,
			Username: "john_doe",
			Email:    "john@example.com",
			Created:  time.Now(),
		}

		options := cache.DefaultCacheOptions()
		options.WithSlidingExpiration(time.Minute * 10)

		err = userCache.Set(ctx, userKey, user, options)
		if err != nil {
			log.Fatalf("设置缓存失败: %v", err)
		}
		fmt.Printf("✅ 数据已写入多级缓存 (耗时: %v)\n\n", time.Since(startTime))
	}

	// 等待异步写入完成
	time.Sleep(100 * time.Millisecond)

	fmt.Println("--- 第二次请求（L1缓存命中）---")
	startTime = time.Now()
	value2, err := userCache.Get(ctx, userKey)
	if err != nil {
		log.Printf("❌ 获取失败: %v\n", err)
	} else if value2 != nil {
		if user2, ok := value2.(map[string]interface{}); ok {
			fmt.Printf("✅ 从缓存获取成功: %+v (耗时: %v)\n\n", user2, time.Since(startTime))
		}
	}

	fmt.Println("--- 清除L1缓存，模拟L2命中 ---")
	memoryProvider.Clear(ctx)

	startTime = time.Now()
	value3, err := userCache.Get(ctx, userKey)
	if err != nil {
		log.Printf("❌ 获取失败: %v\n", err)
	} else if value3 != nil {
		if user3, ok := value3.(map[string]interface{}); ok {
			fmt.Printf("✅ 从L2缓存获取并提升到L1: %+v (耗时: %v)\n\n", user3, time.Since(startTime))
		}
	}

	// 等待数据提升完成
	time.Sleep(100 * time.Millisecond)

	fmt.Println("--- 第三次请求（再次从L1命中，因为数据已提升）---")
	startTime = time.Now()
	value4, err := userCache.Get(ctx, userKey)
	if err != nil {
		log.Printf("❌ 获取失败: %v\n", err)
	} else if value4 != nil {
		if user4, ok := value4.(map[string]interface{}); ok {
			fmt.Printf("✅ 从L1缓存获取: %+v (耗时: %v)\n\n", user4, time.Since(startTime))
		}
	}

	// 6. 显示性能指标
	fmt.Println("\n--- 缓存性能指标 ---")
	multiLevelProvider.PrintMetrics()

	// 7. 清理资源
	fmt.Println("\n--- 清理资源 ---")
	cacheManager.Close()
	fmt.Println("✅ 所有资源已释放")
}
