package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/lyon-serven/go-library/cache"
	"github.com/lyon-serven/go-library/cache/providers"
	"github.com/lyon-serven/go-library/cache/serializers"
)

// RedisUser 用于演示结构体缓存
type RedisUser struct {
	ID    int      `json:"id"`
	Name  string   `json:"name"`
	Email string   `json:"email"`
	Age   int      `json:"age"`
	Roles []string `json:"roles"`
}

// RedisProduct 用于演示多类型缓存
type RedisProduct struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Stock int     `json:"stock"`
}

func mainRedis() {
	fmt.Println("=== Cache System Redis 示例 ===")
	fmt.Println("请确保 Redis 服务已启动（默认 localhost:6379）")
	fmt.Println()

	// ============================================
	// 1. 初始化 Redis Provider
	// ============================================
	fmt.Println("1. 初始化 Redis Provider")

	redisProvider, err := providers.NewRedisCache(&providers.RedisOptions{
		Addresses:            []string{"172.24.140.239:6379"},
		Password:             "Zc2hmmeOpEjD",         // 无密码留空
		DB:                   0,                      // 使用默认 DB
		PoolSize:             10,                     // 连接池大小
		DialTimeout:          5 * time.Second,        // 连接超时
		ReadTimeout:          3 * time.Second,        // 读取超时
		WriteTimeout:         3 * time.Second,        // 写入超时
		KeyPrefix:            "go-library",           // 键前缀，避免与其他应用冲突
		EnableHealthCheck:    true,                   // 开启心跳监测
		HealthCheckInterval:  30 * time.Second,       // 心跳检测间隔
		HealthCheckTimeout:   3 * time.Second,        // 单次 Ping 超时
		LatencyWarnThreshold: 200 * time.Millisecond, // 延迟告警阈值
		OnAlert: func(event providers.AlertEvent) {
			switch event.Level {
			case providers.AlertLevelWarn:
				log.Printf("⚠️ Redis 延迟过高: %v (延迟: %v)", event.Message, event.Latency)
			case providers.AlertLevelError:
				log.Printf("❌ Redis 连接错误: %v", event.Err)
			case providers.AlertLevelRecover:
				log.Printf("✅ Redis 连接恢复: %v", event.Message)
			}
		},
	})
	if err != nil {
		log.Fatal("❌ 连接 Redis 失败:", err)
	}
	fmt.Println("✅ Redis 连接成功")

	// 创建缓存管理器
	manager := cache.NewCacheManager()
	defer manager.Close()

	// 注册 Provider 和序列化器
	manager.RegisterProvider("redis", redisProvider)
	manager.RegisterSerializer("json", serializers.NewJSONSerializer())
	fmt.Println("✅ Provider 和序列化器注册成功")

	// 配置多个缓存实例（共用同一个 Redis，通过 KeyPrefix 区分）
	manager.Configure("users", "redis", "json")
	manager.Configure("products", "redis", "json")
	manager.Configure("sessions", "redis", "json")

	ctx := context.Background()

	// manager.GetCache 返回 ICache，泛型函数需要 *Cache，做一次类型断言
	mustCache := func(name string) *cache.Cache {
		c, ok := manager.GetCache(name).(*cache.Cache)
		if !ok {
			log.Fatalf("GetCache(%q) 类型断言失败", name)
		}
		return c
	}

	// 过期选项辅助函数
	absExp := func(d time.Duration) *cache.CacheOptions {
		return cache.DefaultCacheOptions().WithAbsoluteExpiration(time.Now().Add(d))
	}

	// ============================================
	// 2. 基本 Set / Get
	// ============================================
	fmt.Println("\n2. 基本 Set / Get")

	userCache := mustCache("users")
	userKey := cache.NK("user", "1001")

	user := &RedisUser{
		ID:    1001,
		Name:  "张三",
		Email: "zhangsan@example.com",
		Age:   28,
		Roles: []string{"admin", "user"},
	}

	if err := userCache.Set(ctx, userKey, user, absExp(10*time.Minute)); err != nil {
		log.Printf("❌ Set 失败: %v", err)
	} else {
		fmt.Printf("✅ 存储用户数据，键: %s\n", userKey.String())
	}

	// Get 返回 interface{}（JSON 反序列化为 map）
	raw, err := userCache.Get(ctx, userKey)
	if err != nil {
		log.Printf("❌ Get 失败: %v", err)
	} else {
		fmt.Printf("✅ Get 原始类型: %T\n", raw) // map[string]interface{}
	}

	// ============================================
	// 3. GetAs —— 类型安全获取
	// ============================================
	fmt.Println("\n3. GetAs —— 类型安全获取")

	var fetched RedisUser
	if err := userCache.GetAs(ctx, userKey, &fetched); err != nil {
		log.Printf("❌ GetAs 失败: %v", err)
	} else {
		fmt.Printf("✅ GetAs: ID=%d, Name=%s, Roles=%v\n", fetched.ID, fetched.Name, fetched.Roles)
	}

	// ============================================
	// 4. GetAsS —— 字符串键类型安全获取
	// ============================================
	fmt.Println("\n4. GetAsS —— 字符串键类型安全获取")

	var typedUser RedisUser
	if err := userCache.GetAsS(ctx, "user:1001", &typedUser); err != nil {
		log.Printf("❌ GetAsS 失败: %v", err)
	} else {
		fmt.Printf("✅ GetAsS: ID=%d, Name=%s, 类型=%T\n", typedUser.ID, typedUser.Name, typedUser)
	}

	// ============================================
	// 5. GetOrSetAsS —— 缓存穿透保护（最常用）
	// ============================================
	fmt.Println("\n5. GetOrSetAsS —— 缓存穿透保护")

	// 第一次：缓存未命中，执行工厂函数
	fmt.Println("   第一次调用（缓存未命中，模拟 DB 查询）:")
	var user2 RedisUser
	err = userCache.GetOrSetAsS(ctx, "user:1002", &user2,
		func() (interface{}, error) {
			fmt.Println("   → 从数据库加载 user:1002 ...")
			time.Sleep(50 * time.Millisecond) // 模拟 DB 耗时
			return &RedisUser{ID: 1002, Name: "李四", Email: "lisi@example.com", Age: 25, Roles: []string{"user"}}, nil
		},
		absExp(5*time.Minute),
	)
	if err != nil {
		log.Printf("❌ GetOrSetAsS 失败: %v", err)
	} else {
		fmt.Printf("✅ 获取: ID=%d, Name=%s\n", user2.ID, user2.Name)
	}

	// 第二次：缓存命中，工厂函数不会执行
	fmt.Println("   第二次调用（缓存命中）:")
	var user2Again RedisUser
	err = userCache.GetOrSetAsS(ctx, "user:1002", &user2Again,
		func() (interface{}, error) {
			fmt.Println("   → ❌ 不应执行此处！")
			return nil, nil
		}, nil,
	)
	if err != nil {
		log.Printf("❌ 第二次 GetOrSetAsS 失败: %v", err)
	} else {
		fmt.Printf("✅ 从缓存获取: ID=%d, Name=%s\n", user2Again.ID, user2Again.Name)
	}

	// ============================================
	// 6. 多类型缓存
	// ============================================
	fmt.Println("\n6. 多类型缓存（Product）")

	productCache := mustCache("products")
	products := []*RedisProduct{
		{ID: 1, Name: "Go 语言编程", Price: 89.9, Stock: 100},
		{ID: 2, Name: "Redis 实战", Price: 79.9, Stock: 50},
		{ID: 3, Name: "分布式系统", Price: 99.9, Stock: 200},
	}

	for _, p := range products {
		key := cache.K(fmt.Sprintf("product:%d", p.ID))
		productCache.Set(ctx, key, p, absExp(30*time.Minute))
	}
	fmt.Printf("✅ 批量存储 %d 个商品\n", len(products))

	for _, p := range products {
		var product RedisProduct
		if err := productCache.GetAsS(ctx, fmt.Sprintf("product:%d", p.ID), &product); err != nil {
			log.Printf("❌ 获取商品 %d 失败: %v", p.ID, err)
		} else {
			fmt.Printf("✅ 商品: ID=%d, Name=%s, Price=%.1f, Stock=%d\n",
				product.ID, product.Name, product.Price, product.Stock)
		}
	}

	// ============================================
	// 7. Exists / Remove
	// ============================================
	fmt.Println("\n7. Exists / Remove")

	testKey := cache.K("test:remove")
	userCache.Set(ctx, testKey, "待删除", nil)

	exists, _ := userCache.Exists(ctx, testKey)
	fmt.Printf("✅ 删除前存在: %v\n", exists)

	userCache.Remove(ctx, testKey)

	exists, _ = userCache.Exists(ctx, testKey)
	fmt.Printf("✅ 删除后存在: %v\n", exists)

	// ============================================
	// 8. 过期时间验证
	// ============================================
	fmt.Println("\n8. 过期时间验证（2 秒过期）")

	sessionCache := mustCache("sessions")
	sessionKey := cache.K("session:abc123")
	sessionCache.Set(ctx, sessionKey, map[string]interface{}{
		"user_id": 1001,
		"token":   "abc123",
	}, absExp(2*time.Second))

	exists, _ = sessionCache.Exists(ctx, sessionKey)
	fmt.Printf("✅ 设置后立即检查，存在: %v\n", exists)

	fmt.Println("   等待 3 秒...")
	time.Sleep(3 * time.Second)

	exists, _ = sessionCache.Exists(ctx, sessionKey)
	fmt.Printf("✅ 3 秒后检查，存在: %v（已过期）\n", exists)

	// ============================================
	// 9. RemoveByPattern 批量删除
	// ============================================
	fmt.Println("\n9. RemoveByPattern 批量删除")

	for i := 1; i <= 5; i++ {
		userCache.Set(ctx, cache.K(fmt.Sprintf("temp:%d", i)), fmt.Sprintf("临时数据%d", i), nil)
	}
	fmt.Println("✅ 写入 5 条 temp:* 数据")

	// 注意：pattern 需要包含 KeyPrefix
	if err := userCache.RemoveByPattern(ctx, "go-library:temp:*"); err != nil {
		log.Printf("❌ RemoveByPattern 失败: %v", err)
	} else {
		exists, _ := userCache.Exists(ctx, cache.K("temp:1"))
		fmt.Printf("✅ 批量删除后 temp:1 存在: %v\n", exists)
	}

	fmt.Println("\n✅ 所有 Redis 缓存测试完成！")
}

/*
运行前置条件：
  启动 Redis（Docker 方式）：
    docker run -d -p 6379:6379 redis

  本地运行：
    cd examples/cache
    go run cache_redis_example.go

  连接远程 Redis，修改 Addresses：
    Addresses: []string{"192.168.1.100:6379"},
    Password:  "yourpassword",
*/
