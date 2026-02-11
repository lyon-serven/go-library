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

// User 用户结构体
type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Age       int       `json:"age"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// Product 产品结构体
type Product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Stock int     `json:"stock"`
}

func main() {
	fmt.Println("========================================")
	fmt.Println("泛型缓存示例 - 最优雅的类型安全方案")
	fmt.Println("========================================\n")

	// 创建缓存
	cacheInstance := setupCache()
	ctx := context.Background()

	// ========================================
	// 对比：三种方式的使用体验
	// ========================================
	fmt.Println("【对比】三种方式的使用体验")
	fmt.Println("----------------------------------------")

	user := &User{
		ID:        1,
		Name:      "张三",
		Age:       25,
		Email:     "zhangsan@example.com",
		CreatedAt: time.Now(),
	}

	// 存储用户
	cacheInstance.Set(ctx, cache.K("user:1"), user, nil)

	// ❌ 方式1: Get - 类型丢失
	fmt.Println("\n❌ 方式1: cache.Get() - 类型丢失")
	value1, _ := cacheInstance.Get(ctx, cache.K("user:1"))
	fmt.Printf("   返回类型: %T\n", value1)
	fmt.Println("   问题: 需要类型断言和转换")

	// ✅ 方式2: GetAs - 需要预创建对象
	fmt.Println("\n✅ 方式2: cache.GetAs() - 需要预创建对象")
	var user2 User
	cacheInstance.GetAs(ctx, cache.K("user:1"), &user2)
	fmt.Printf("   var user User\n")
	fmt.Printf("   cache.GetAs(ctx, key, &user)\n")
	fmt.Printf("   结果: %s (年龄: %d)\n", user2.Name, user2.Age)
	fmt.Println("   问题: 需要先声明变量")

	// ⭐ 方式3: GetTyped - 直接返回正确类型（最推荐）
	fmt.Println("\n⭐ 方式3: cache.GetTyped[User]() - 直接返回（最推荐）")
	user3, _ := cache.GetTyped[User](ctx, cacheInstance, cache.K("user:1"))
	fmt.Printf("   user, _ := cache.GetTyped[User](ctx, cache, key)\n")
	fmt.Printf("   结果: %s (年龄: %d)\n", user3.Name, user3.Age)
	fmt.Println("   ✓ 无需预创建对象")
	fmt.Println("   ✓ 直接返回正确类型")
	fmt.Println("   ✓ 代码最简洁")

	// ========================================
	// 示例 1: GetTyped - 泛型获取
	// ========================================
	fmt.Println("\n\n【示例 1】GetTyped - 泛型获取")
	fmt.Println("----------------------------------------")

	// 直接返回，无需预先创建对象
	retrievedUser, err := cache.GetTyped[User](ctx, cacheInstance, cache.K("user:1"))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("✓ 用户信息:\n")
	fmt.Printf("  ID: %d\n", retrievedUser.ID)
	fmt.Printf("  Name: %s\n", retrievedUser.Name)
	fmt.Printf("  Age: %d\n", retrievedUser.Age)
	fmt.Printf("  Email: %s\n", retrievedUser.Email)

	// ========================================
	// 示例 2: GetOrSetTyped - 泛型 GetOrSet
	// ========================================
	fmt.Println("\n\n【示例 2】GetOrSetTyped - 泛型 GetOrSet")
	fmt.Println("----------------------------------------")

	// 第一次获取（从"数据库"加载）
	fmt.Println("第一次获取用户 2:")
	user2Result, err := cache.GetOrSetTyped[User](ctx, cacheInstance, cache.K("user:2"),
		func() (*User, error) {
			fmt.Println("  📥 从数据库加载...")
			time.Sleep(100 * time.Millisecond) // 模拟数据库延迟
			return &User{
				ID:        2,
				Name:      "李四",
				Age:       30,
				Email:     "lisi@example.com",
				CreatedAt: time.Now(),
			}, nil
		}, nil)

	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ 用户: %s (ID: %d)\n", user2Result.Name, user2Result.ID)

	// 第二次获取（从缓存读取）
	fmt.Println("\n第二次获取用户 2:")
	user2Again, _ := cache.GetOrSetTyped[User](ctx, cacheInstance, cache.K("user:2"),
		func() (*User, error) {
			fmt.Println("  ⚠️  不应该看到这条消息")
			return nil, nil
		}, nil)
	fmt.Printf("✓ 从缓存读取: %s\n", user2Again.Name)

	// ========================================
	// 示例 3: GetTypedS - 字符串键的泛型方法
	// ========================================
	fmt.Println("\n\n【示例 3】GetTypedS - 字符串键的泛型方法")
	fmt.Println("----------------------------------------")

	product := &Product{
		ID:    100,
		Name:  "笔记本电脑",
		Price: 5999.99,
		Stock: 50,
	}

	// 使用字符串键存储
	cacheInstance.Set(ctx, cache.K("product:100"), product, nil)

	// 使用字符串键和泛型获取
	retrievedProduct, err := cache.GetTypedS[Product](ctx, cacheInstance, "product:100")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("✓ 产品信息:\n")
	fmt.Printf("  ID: %d\n", retrievedProduct.ID)
	fmt.Printf("  Name: %s\n", retrievedProduct.Name)
	fmt.Printf("  Price: %.2f\n", retrievedProduct.Price)
	fmt.Printf("  Stock: %d\n", retrievedProduct.Stock)

	// ========================================
	// 示例 4: TypedCache - 类型化缓存包装器
	// ========================================
	fmt.Println("\n\n【示例 4】TypedCache - 类型化缓存包装器")
	fmt.Println("----------------------------------------")
	fmt.Println("适用场景: 某个缓存实例只存储一种类型")

	// 创建类型化的用户缓存
	userCache := cache.NewTypedCache[User](cacheInstance)

	// 所有操作都是类型安全的
	newUser := &User{
		ID:    3,
		Name:  "王五",
		Age:   28,
		Email: "wangwu@example.com",
	}

	// Set - 类型安全
	userCache.Set(ctx, cache.K("3"), newUser, nil)

	// Get - 直接返回正确类型
	user3Result, err := userCache.Get(ctx, cache.K("3"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ Get 返回: %s (年龄: %d)\n", user3Result.Name, user3Result.Age)

	// GetS - 使用字符串键
	user4 := &User{ID: 4, Name: "赵六", Age: 35}
	userCache.SetS(ctx, "4", user4, nil)

	user4Result, _ := userCache.GetS(ctx, "4")
	fmt.Printf("✓ GetS 返回: %s (年龄: %d)\n", user4Result.Name, user4Result.Age)

	// GetOrSet - 类型安全的 GetOrSet
	fmt.Println("\nGetOrSet 测试:")
	user5, _ := userCache.GetOrSet(ctx, cache.K("5"), func() (*User, error) {
		fmt.Println("  📥 创建新用户...")
		return &User{ID: 5, Name: "孙七", Age: 22}, nil
	}, nil)
	fmt.Printf("✓ 第一次: %s\n", user5.Name)

	user5Again, _ := userCache.GetOrSet(ctx, cache.K("5"), func() (*User, error) {
		fmt.Println("  ⚠️  不应该看到这条消息")
		return nil, nil
	}, nil)
	fmt.Printf("✓ 第二次(缓存): %s\n", user5Again.Name)

	// ========================================
	// 示例 5: 实际业务场景 - UserService
	// ========================================
	fmt.Println("\n\n【示例 5】实际业务场景 - UserService")
	fmt.Println("----------------------------------------")

	userService := NewUserService(cacheInstance)

	// 获取用户（带缓存）
	fmt.Println("第一次获取用户 1001:")
	user1001, err := userService.GetUser(ctx, 1001)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ 用户: %s (ID: %d, Age: %d)\n", user1001.Name, user1001.ID, user1001.Age)

	fmt.Println("\n第二次获取用户 1001:")
	user1001Again, err := userService.GetUser(ctx, 1001)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ 用户: %s (从缓存读取)\n", user1001Again.Name)

	// ========================================
	// 总结
	// ========================================
	fmt.Println("\n\n========================================")
	fmt.Println("【总结】使用建议")
	fmt.Println("========================================")
	fmt.Println("\n⭐ 推荐使用泛型方法（需要 Go 1.18+）：")
	fmt.Println()
	fmt.Println("1️⃣  快速获取:")
	fmt.Println("   user, err := cache.GetTyped[User](ctx, cache, key)")
	fmt.Println("   ✓ 直接返回正确类型")
	fmt.Println("   ✓ 无需预创建对象")
	fmt.Println()
	fmt.Println("2️⃣  GetOrSet 模式:")
	fmt.Println("   user, err := cache.GetOrSetTyped[User](ctx, cache, key, factory, nil)")
	fmt.Println("   ✓ 类型安全")
	fmt.Println("   ✓ 代码简洁")
	fmt.Println()
	fmt.Println("3️⃣  字符串键:")
	fmt.Println("   user, err := cache.GetTypedS[User](ctx, cache, \"user:1\")")
	fmt.Println("   ✓ 最简洁的写法")
	fmt.Println()
	fmt.Println("4️⃣  单一类型缓存:")
	fmt.Println("   userCache := cache.NewTypedCache[User](baseCache)")
	fmt.Println("   user, err := userCache.Get(ctx, key)")
	fmt.Println("   ✓ 完整的类型安全封装")
	fmt.Println()
	fmt.Println("💡 如果使用 Go 1.18+，强烈推荐使用泛型方法！")
	fmt.Println("========================================\n")
}

// ========================================
// 辅助函数
// ========================================

func setupCache() cache.ICache {
	manager := cache.NewCacheManager()

	// 注册内存提供者
	memoryProvider := providers.NewMemoryCache(providers.DefaultMemoryCacheOptions())
	manager.RegisterProvider("memory", memoryProvider)

	// 注册 JSON 序列化器
	jsonSerializer := serializers.NewJSONSerializer()
	manager.RegisterSerializer("json", jsonSerializer)

	return manager.GetCache("demo")
}

// ========================================
// UserService - 实际业务场景
// ========================================

type UserService struct {
	cache cache.ICache
}

func NewUserService(c cache.ICache) *UserService {
	return &UserService{cache: c}
}

// GetUser 获取用户（带缓存）- 使用泛型方法
func (s *UserService) GetUser(ctx context.Context, userID int) (*User, error) {
	key := cache.K(fmt.Sprintf("user:%d", userID))

	// ⭐ 使用泛型 GetOrSetTyped - 直接返回正确类型
	return cache.GetOrSetTyped[User](ctx, s.cache, key, func() (*User, error) {
		return s.loadUserFromDB(userID)
	}, &cache.CacheOptions{
		SlidingExpiration: ptrDuration(10 * time.Minute),
	})
}

func (s *UserService) loadUserFromDB(userID int) (*User, error) {
	fmt.Printf("  📥 从数据库加载用户 %d...\n", userID)
	time.Sleep(100 * time.Millisecond) // 模拟数据库延迟

	return &User{
		ID:        userID,
		Name:      fmt.Sprintf("用户_%d", userID),
		Age:       20 + userID%50,
		Email:     fmt.Sprintf("user%d@example.com", userID),
		CreatedAt: time.Now(),
	}, nil
}

func ptrDuration(d time.Duration) *time.Duration {
	return &d
}
