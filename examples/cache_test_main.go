package main

import (
	"context"
	"fmt"
	"log"

	"gitee.com/wangsoft/go-library/cache"
	"gitee.com/wangsoft/go-library/cache/providers"
	"gitee.com/wangsoft/go-library/cache/serializers"
)

func main() {
	fmt.Println("=== MyLib Cache System Quick Test ===")

	// 创建缓存管理�?
	manager := cache.NewCacheManager()
	defer manager.Close()

	// 注册内存提供程序
	memoryProvider := providers.NewMemoryCache(providers.DefaultMemoryCacheOptions())
	if err := manager.RegisterProvider("memory", memoryProvider); err != nil {
		log.Fatal("Failed to register memory provider:", err)
	}

	// 注册JSON序列化器
	jsonSerializer := serializers.NewJSONSerializer()
	if err := manager.RegisterSerializer("json", jsonSerializer); err != nil {
		log.Fatal("Failed to register JSON serializer:", err)
	}

	// 配置缓存
	manager.Configure("test-cache", "memory", "json")

	// 获取缓存实例
	testCache := manager.GetCache("test-cache")

	ctx := context.Background()

	// 测试基本操作
	fmt.Println("\n1. 测试基本�?Set/Get 操作:")
	key := cache.NewCacheKey("user:123", "users")
	userData := map[string]interface{}{
		"id":    123,
		"name":  "张三",
		"email": "zhangsan@example.com",
		"roles": []string{"admin", "user"},
	}

	err := testCache.Set(ctx, key, userData, cache.DefaultCacheOptions())
	if err != nil {
		log.Printf("Set 操作失败: %v", err)
	} else {
		fmt.Printf("�?成功存储用户数据，键: %s\n", key.String())
	}

	value, err := testCache.Get(ctx, key)
	if err != nil {
		log.Printf("Get 操作失败: %v", err)
	} else if value != nil {
		fmt.Printf("�?成功获取用户数据: %+v\n", value)
	} else {
		fmt.Println("�?未找到数�?)
	}

	// 测试 Exists
	fmt.Println("\n2. 测试 Exists 操作:")
	exists, err := testCache.Exists(ctx, key)
	if err != nil {
		log.Printf("Exists 操作失败: %v", err)
	} else {
		fmt.Printf("�?键是否存�? %v\n", exists)
	}

	// 测试 GetOrSet
	fmt.Println("\n3. 测试 GetOrSet 模式:")
	calcKey := cache.NewCacheKey("calculation:result", "math")

	result, err := testCache.GetOrSet(ctx, calcKey, func() (interface{}, error) {
		fmt.Println("  �?执行复杂计算...")
		return map[string]interface{}{
			"result":    42,
			"operation": "6 * 7",
		}, nil
	}, cache.DefaultCacheOptions())

	if err != nil {
		log.Printf("GetOrSet 操作失败: %v", err)
	} else {
		fmt.Printf("�?计算结果: %+v\n", result)
	}

	// 再次调用，应该从缓存获取
	result2, err := testCache.GetOrSet(ctx, calcKey, func() (interface{}, error) {
		fmt.Println("  �?这不应该被执行（从缓存获取）")
		return nil, nil
	}, cache.DefaultCacheOptions())

	if err != nil {
		log.Printf("第二�?GetOrSet 操作失败: %v", err)
	} else {
		fmt.Printf("�?从缓存获取的结果: %+v\n", result2)
	}

	// 测试不同的序列化�?
	fmt.Println("\n4. 测试不同的序列化�?")

	// 字符串序列化�?
	stringSerializer := serializers.NewStringSerializer()
	manager.RegisterSerializer("string", stringSerializer)
	manager.Configure("string-cache", "memory", "string")

	stringCache := manager.GetCache("string-cache")
	stringKey := cache.NewCacheKey("message", "strings")

	err = stringCache.Set(ctx, stringKey, "Hello, 缓存世界!", nil)
	if err != nil {
		log.Printf("字符串缓存设置失�? %v", err)
	} else {
		fmt.Printf("�?字符串缓存设置成功\n")

		if value, _ := stringCache.Get(ctx, stringKey); value != nil {
			fmt.Printf("�?字符串缓存获�? %v\n", value)
		}
	}

	// 测试多个缓存实例
	fmt.Println("\n5. 测试多个缓存实例:")

	cache1 := manager.GetCache("cache-1")
	cache2 := manager.GetCache("cache-2")

	key1 := cache.NewCacheKey("data", "cache1")
	key2 := cache.NewCacheKey("data", "cache2")

	cache1.Set(ctx, key1, "这是缓存1的数�?, nil)
	cache2.Set(ctx, key2, "这是缓存2的数�?, nil)

	val1, _ := cache1.Get(ctx, key1)
	val2, _ := cache2.Get(ctx, key2)

	fmt.Printf("�?缓存1数据: %v\n", val1)
	fmt.Printf("�?缓存2数据: %v\n", val2)

	fmt.Println("\n�?所有测试完成！缓存系统工作正常�?)
}
