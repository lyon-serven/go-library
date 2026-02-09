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

func main2() {
	fmt.Println("=== Cache 便捷方法使用示例 ===\n")

	// 初始化缓存
	manager := cache.NewCacheManager()
	defer manager.Close()

	memoryProvider := providers.NewMemoryCache(nil)
	manager.RegisterProvider("memory", memoryProvider)
	manager.RegisterSerializer("json", serializers.NewJSONSerializer())
	manager.Configure("test-cache", "memory", "json")

	baseCache := manager.GetCache("test-cache")
	ctx := context.Background()

	// ========== 方式1: 使用 K() 和 NK() 快捷函数（推荐） ==========
	fmt.Println("【方式1】使用 K() 和 NK() 快捷函数")
	fmt.Println("-----------------------------------")

	// 1.1 使用 K() 创建简单键
	fmt.Println("1. 使用 K() 创建简单键:")
	userData := map[string]interface{}{
		"id":    123,
		"name":  "张三",
		"email": "zhangsan@example.com",
	}

	err := baseCache.Set(ctx, cache.K("user:123"), userData, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("  ✅ 存储成功: cache.K(\"user:123\")")

	value, _ := baseCache.Get(ctx, cache.K("user:123"))
	fmt.Printf("  ✅ 获取成功: %+v\n", value)

	// 1.2 使用 NK() 创建带命名空间的键
	fmt.Println("\n2. 使用 NK() 创建带命名空间的键:")
	productData := map[string]interface{}{
		"id":    456,
		"name":  "iPhone 15",
		"price": 5999.00,
	}

	err = baseCache.Set(ctx, cache.NK("products", "456"), productData, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("  ✅ 存储成功: cache.NK(\"products\", \"456\") → 生成键: products:456")

	value, _ = baseCache.Get(ctx, cache.NK("products", "456"))
	fmt.Printf("  ✅ 获取成功: %+v\n", value)

	// 1.3 使用 GetOrSet 模式
	fmt.Println("\n3. 使用 K() 配合 GetOrSet:")
	orderData, err := baseCache.GetOrSet(
		ctx,
		cache.K("order:789"),
		func() (interface{}, error) {
			fmt.Println("  → 首次调用，从数据库加载订单...")
			return map[string]interface{}{
				"id":     789,
				"total":  1299.00,
				"status": "paid",
			}, nil
		},
		&cache.CacheOptions{
			SlidingExpiration: ptrDuration(5 * time.Minute),
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("  ✅ 订单数据: %+v\n", orderData)

	// 再次调用 - 应该从缓存获取
	orderData2, _ := baseCache.GetOrSet(
		ctx,
		cache.K("order:789"),
		func() (interface{}, error) {
			fmt.Println("  → 这不应该执行（从缓存获取）")
			return nil, nil
		},
		nil,
	)
	fmt.Printf("  ✅ 从缓存获取: %+v\n", orderData2)

	// ========== 方式2: 使用字符串方法 ICacheExt ==========
	fmt.Println("\n\n【方式2】使用字符串方法 ICacheExt")
	fmt.Println("-----------------------------------")

	// 将基础缓存包装为支持字符串方法的缓存
	extCache := cache.NewCacheExt(baseCache)

	// 2.1 使用 SetS/GetS 字符串方法
	fmt.Println("1. 使用 SetS/GetS 字符串方法:")
	commentData := map[string]interface{}{
		"id":      101,
		"content": "这个产品真不错！",
		"author":  "李四",
	}

	err = extCache.SetS(ctx, "comment:101", commentData, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("  ✅ 使用 SetS 存储: \"comment:101\"")

	value, _ = extCache.GetS(ctx, "comment:101")
	fmt.Printf("  ✅ 使用 GetS 获取: %+v\n", value)

	// 2.2 使用 ExistsS 检查存在
	fmt.Println("\n2. 使用 ExistsS 检查键是否存在:")
	exists, _ := extCache.ExistsS(ctx, "comment:101")
	fmt.Printf("  ✅ comment:101 存在: %v\n", exists)

	exists, _ = extCache.ExistsS(ctx, "comment:999")
	fmt.Printf("  ✅ comment:999 存在: %v\n", exists)

	// 2.3 使用 GetOrSetS
	fmt.Println("\n3. 使用 GetOrSetS 字符串方法:")
	settingData, err := extCache.GetOrSetS(
		ctx,
		"settings:app",
		func() (interface{}, error) {
			fmt.Println("  → 从配置文件加载设置...")
			return map[string]interface{}{
				"theme":    "dark",
				"language": "zh-CN",
				"version":  "1.0.0",
			}, nil
		},
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("  ✅ 应用设置: %+v\n", settingData)

	// 2.4 使用 RemoveS 删除
	fmt.Println("\n4. 使用 RemoveS 删除键:")
	err = extCache.RemoveS(ctx, "comment:101")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("  ✅ 成功删除: comment:101")

	exists, _ = extCache.ExistsS(ctx, "comment:101")
	fmt.Printf("  ✅ 验证删除: comment:101 存在 = %v\n", exists)

	// ========== 方式3: 显式创建 CacheKey ==========
	fmt.Println("\n\n【方式3】显式创建 CacheKey")
	fmt.Println("-----------------------------------")

	// 3.1 直接创建 CacheKey 结构体
	fmt.Println("1. 直接创建 CacheKey 结构体:")
	key1 := cache.CacheKey{Key: "cart:999"}
	cartData := map[string]interface{}{
		"items": []string{"iPhone", "AirPods"},
		"total": 7999.00,
	}

	err = baseCache.Set(ctx, key1, cartData, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("  ✅ 存储成功，键: %s\n", key1.String())

	value, _ = baseCache.Get(ctx, key1)
	fmt.Printf("  ✅ 获取成功: %+v\n", value)

	// 3.2 使用 NewCacheKey 工厂函数
	fmt.Println("\n2. 使用 NewCacheKey 工厂函数:")
	key2 := cache.NewCacheKey("555", "wishlists")
	wishlistData := map[string]interface{}{
		"user_id": 555,
		"items":   []string{"MacBook", "iPad"},
	}

	err = baseCache.Set(ctx, key2, wishlistData, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("  ✅ 存储成功，键: %s\n", key2.String())

	value, _ = baseCache.Get(ctx, key2)
	fmt.Printf("  ✅ 获取成功: %+v\n", value)

	// ========== 对比三种方式 ==========
	fmt.Println("\n\n【对比总结】")
	fmt.Println("-----------------------------------")
	fmt.Println("✅ 方式1: K()/NK()     - 最推荐，简洁且类型安全")
	fmt.Println("✅ 方式2: 字符串方法    - 极简风格，适合快速开发")
	fmt.Println("✅ 方式3: 显式CacheKey  - 完整方式，适合复杂场景")
	fmt.Println("\n选择适合你的方式，开始使用吧！🚀")

	// ========== 实际业务示例 ==========
	fmt.Println("\n\n【实际业务示例】使用 K() 实现用户服务")
	fmt.Println("-----------------------------------")
	userService := NewUserService(baseCache)

	// 获取用户（首次从数据库加载）
	user, err := userService.GetUser(ctx, 1001)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("用户信息: ID=%d, Name=%s\n", user.ID, user.Name)

	// 再次获取（从缓存获取）
	user2, err := userService.GetUser(ctx, 1001)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("从缓存获取: ID=%d, Name=%s\n", user2.ID, user2.Name)

	// 更新用户缓存
	user.Name = "张三（已更新）"
	err = userService.SetUser(ctx, user)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("✅ 用户缓存已更新")

	// 删除用户缓存
	err = userService.DeleteUser(ctx, 1001)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("✅ 用户缓存已删除")
}

// ========== 辅助函数和类型 ==========

func ptrDuration(d time.Duration) *time.Duration {
	return &d
}

// UserService 用户服务（演示 K() 函数的实际用法）
type UserService struct {
	cache cache.ICache
}

func NewUserService(cache cache.ICache) *UserService {
	return &UserService{cache: cache}
}

// GetUser 获取用户（使用 K() 函数）
func (s *UserService) GetUser(ctx context.Context, userID int) (*User, error) {
	// 使用 K() 创建键 - 非常简洁！
	key := cache.K(fmt.Sprintf("user:%d", userID))

	value, err := s.cache.GetOrSet(
		ctx,
		key,
		func() (interface{}, error) {
			fmt.Printf("  → 从数据库加载用户 %d...\n", userID)
			return &User{
				ID:    userID,
				Name:  "张三",
				Email: "zhangsan@example.com",
			}, nil
		},
		&cache.CacheOptions{
			SlidingExpiration: ptrDuration(10 * time.Minute),
		},
	)

	if err != nil {
		return nil, err
	}

	return value.(*User), nil
}

// SetUser 设置用户缓存
func (s *UserService) SetUser(ctx context.Context, user *User) error {
	key := cache.K(fmt.Sprintf("user:%d", user.ID))
	return s.cache.Set(ctx, key, user, &cache.CacheOptions{
		SlidingExpiration: ptrDuration(10 * time.Minute),
	})
}

// DeleteUser 删除用户缓存
func (s *UserService) DeleteUser(ctx context.Context, userID int) error {
	key := cache.K(fmt.Sprintf("user:%d", userID))
	return s.cache.Remove(ctx, key)
}
