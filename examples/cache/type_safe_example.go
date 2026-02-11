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
	fmt.Println("Cache 类型安全示例")
	fmt.Println("========================================\n")

	// 创建缓存管理器
	manager := cache.NewCacheManager()

	// 注册内存提供者
	memoryProvider := providers.NewMemoryCache(providers.DefaultMemoryCacheOptions())
	manager.RegisterProvider("memory", memoryProvider)

	// 注册 JSON 序列化器
	jsonSerializer := serializers.NewJSONSerializer()
	manager.RegisterSerializer("json", jsonSerializer)

	// 获取缓存实例
	userCache := manager.GetCache("users")

	ctx := context.Background()

	// ========================================
	// 示例 1: Get vs GetAs 的区别
	// ========================================
	fmt.Println("【示例 1】Get vs GetAs 的区别")
	fmt.Println("----------------------------------------")

	user := &User{
		ID:        1,
		Name:      "张三",
		Age:       25,
		Email:     "zhangsan@example.com",
		CreatedAt: time.Now(),
	}

	// 存储用户
	userCache.Set(ctx, cache.K("user:1"), user, nil)

	// ❌ 使用 Get - 类型会丢失
	fmt.Println("\n❌ 使用 Get 方法（类型丢失）：")
	value, _ := userCache.Get(ctx, cache.K("user:1"))
	fmt.Printf("   返回类型: %T\n", value)

	if m, ok := value.(map[string]interface{}); ok {
		fmt.Println("   实际是 map[string]interface{}:")
		fmt.Printf("     - id: %v (类型: %T)\n", m["id"], m["id"])          // float64
		fmt.Printf("     - name: %v (类型: %T)\n", m["name"], m["name"])    // string
		fmt.Printf("     - age: %v (类型: %T)\n", m["age"], m["age"])       // float64
		fmt.Printf("     - email: %v (类型: %T)\n", m["email"], m["email"]) // string

		// ❌ 无法直接断言为 *User
		if _, ok := value.(*User); !ok {
			fmt.Println("   ⚠️  无法断言为 *User 类型")
		}
	}

	// ✅ 使用 GetAs - 保持原始类型
	fmt.Println("\n✅ 使用 GetAs 方法（保持类型）：")
	var retrievedUser User
	err := userCache.GetAs(ctx, cache.K("user:1"), &retrievedUser)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   返回类型: User\n")
	fmt.Printf("     - ID: %d (类型: %T)\n", retrievedUser.ID, retrievedUser.ID)       // int
	fmt.Printf("     - Name: %s (类型: %T)\n", retrievedUser.Name, retrievedUser.Name) // string
	fmt.Printf("     - Age: %d (类型: %T)\n", retrievedUser.Age, retrievedUser.Age)    // int
	fmt.Printf("     - Email: %s\n", retrievedUser.Email)
	fmt.Println("   ✓ 类型正确，可以直接使用")

	// ========================================
	// 示例 2: GetOrSet vs GetOrSetAs
	// ========================================
	fmt.Println("\n\n【示例 2】GetOrSet vs GetOrSetAs")
	fmt.Println("----------------------------------------")

	// 模拟从数据库加载用户
	loadUserFromDB := func(id int) (*User, error) {
		fmt.Printf("   📥 从数据库加载用户 %d...\n", id)
		return &User{
			ID:        id,
			Name:      "李四",
			Age:       30,
			Email:     "lisi@example.com",
			CreatedAt: time.Now(),
		}, nil
	}

	// ❌ 使用 GetOrSet - 需要类型转换
	fmt.Println("\n❌ 使用 GetOrSet 方法：")
	value2, _ := userCache.GetOrSet(ctx, cache.K("user:2"), func() (interface{}, error) {
		return loadUserFromDB(2)
	}, nil)

	fmt.Printf("   返回类型: %T\n", value2)
	if m, ok := value2.(map[string]interface{}); ok {
		fmt.Println("   需要从 map 转换：")
		fmt.Printf("     - name: %s\n", m["name"])
		fmt.Printf("     - age: %.0f (需要转换为 int)\n", m["age"])
	}

	// ✅ 使用 GetOrSetAs - 直接得到正确类型
	fmt.Println("\n✅ 使用 GetOrSetAs 方法：")
	var user3 User
	err = userCache.GetOrSetAs(ctx, cache.K("user:3"), &user3, func() (interface{}, error) {
		return loadUserFromDB(3)
	}, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("   返回类型: User\n")
	fmt.Printf("     - ID: %d\n", user3.ID)
	fmt.Printf("     - Name: %s\n", user3.Name)
	fmt.Printf("     - Age: %d\n", user3.Age)
	fmt.Println("   ✓ 无需类型转换，直接使用")

	// 再次获取，应该从缓存读取（不调用工厂函数）
	fmt.Println("\n   第二次获取（缓存命中）：")
	var user3Again User
	userCache.GetOrSetAs(ctx, cache.K("user:3"), &user3Again, func() (interface{}, error) {
		fmt.Println("   ⚠️  不应该看到这条消息")
		return loadUserFromDB(3)
	}, nil)
	fmt.Println("   ✓ 从缓存读取，工厂函数未被调用")

	// ========================================
	// 示例 3: 使用字符串键的便捷方法
	// ========================================
	fmt.Println("\n\n【示例 3】使用字符串键的便捷方法")
	fmt.Println("----------------------------------------")

	// 创建扩展缓存
	extCache := cache.NewCacheExt(userCache)

	product := &Product{
		ID:    100,
		Name:  "笔记本电脑",
		Price: 5999.99,
		Stock: 50,
	}

	// 使用字符串键存储
	extCache.SetS(ctx, "product:100", product, nil)

	// ❌ GetS 返回 map
	fmt.Println("\n❌ 使用 GetS 方法：")
	value3, _ := extCache.GetS(ctx, "product:100")
	fmt.Printf("   返回类型: %T\n", value3)

	// ✅ GetAsS 返回正确类型
	fmt.Println("\n✅ 使用 GetAsS 方法：")
	var retrievedProduct Product
	err = extCache.GetAsS(ctx, "product:100", &retrievedProduct)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   产品信息:\n")
	fmt.Printf("     - ID: %d\n", retrievedProduct.ID)
	fmt.Printf("     - Name: %s\n", retrievedProduct.Name)
	fmt.Printf("     - Price: %.2f\n", retrievedProduct.Price)
	fmt.Printf("     - Stock: %d\n", retrievedProduct.Stock)

	// ✅ GetOrSetAsS - 字符串键 + 类型安全
	fmt.Println("\n✅ 使用 GetOrSetAsS 方法：")
	var product2 Product
	err = extCache.GetOrSetAsS(ctx, "product:101", &product2, func() (interface{}, error) {
		fmt.Println("   📥 从数据库加载产品 101...")
		return &Product{
			ID:    101,
			Name:  "机械键盘",
			Price: 299.99,
			Stock: 100,
		}, nil
	}, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   产品信息: %s, 价格: %.2f\n", product2.Name, product2.Price)

	// ========================================
	// 示例 4: 实际业务场景 - UserService
	// ========================================
	fmt.Println("\n\n【示例 4】实际业务场景 - UserService")
	fmt.Println("----------------------------------------")

	userService := NewUserService(userCache)

	// 第一次获取（从"数据库"加载）
	fmt.Println("\n第一次获取用户:")
	user1001, err := userService.GetUser(ctx, 1001)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ 用户: %s (ID: %d, Age: %d)\n", user1001.Name, user1001.ID, user1001.Age)

	// 第二次获取（从缓存读取）
	fmt.Println("\n第二次获取用户:")
	user1001Again, err := userService.GetUser(ctx, 1001)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ 用户: %s (ID: %d, Age: %d)\n", user1001Again.Name, user1001Again.ID, user1001Again.Age)

	// ========================================
	// 总结
	// ========================================
	fmt.Println("\n\n========================================")
	fmt.Println("【总结】")
	fmt.Println("========================================")
	fmt.Println("✅ GetAs      - 直接获取正确类型，无需转换")
	fmt.Println("✅ GetOrSetAs - 类型安全的缓存或设置模式")
	fmt.Println("✅ GetAsS     - 字符串键 + 类型安全")
	fmt.Println("✅ GetOrSetAsS- 字符串键 + 类型安全 + 缓存或设置")
	fmt.Println("\n推荐：在实际项目中使用 GetAs 系列方法避免类型问题！")
	fmt.Println("========================================\n")
}

// ========================================
// UserService - 实际业务场景示例
// ========================================

// UserService 用户服务
type UserService struct {
	cache cache.ICache
}

// NewUserService 创建用户服务
func NewUserService(cache cache.ICache) *UserService {
	return &UserService{cache: cache}
}

// GetUser 获取用户（带缓存）
func (s *UserService) GetUser(ctx context.Context, userID int) (*User, error) {
	key := cache.K(fmt.Sprintf("user:%d", userID))

	var user User
	err := s.cache.GetOrSetAs(ctx, key, &user, func() (interface{}, error) {
		// 模拟从数据库加载
		return s.loadUserFromDB(userID)
	}, &cache.CacheOptions{
		SlidingExpiration: ptrDuration(10 * time.Minute),
	})

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// loadUserFromDB 模拟从数据库加载用户
func (s *UserService) loadUserFromDB(userID int) (*User, error) {
	fmt.Printf("   📥 从数据库加载用户 %d...\n", userID)
	time.Sleep(100 * time.Millisecond) // 模拟数据库延迟

	return &User{
		ID:        userID,
		Name:      fmt.Sprintf("用户_%d", userID),
		Age:       20 + userID%50,
		Email:     fmt.Sprintf("user%d@example.com", userID),
		CreatedAt: time.Now(),
	}, nil
}

// ptrDuration 辅助函数：创建 Duration 指针
func ptrDuration(d time.Duration) *time.Duration {
	return &d
}
