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

// 定义多种不同的类型
type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type Product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

type Order struct {
	ID     int     `json:"id"`
	UserID int     `json:"user_id"`
	Total  float64 `json:"total"`
}

func main() {
	fmt.Println("========================================")
	fmt.Println("使用 GetAs 方法 - 简单直接")
	fmt.Println("========================================\n")

	ctx := context.Background()

	// 创建 Manager
	manager := cache.NewCacheManager()
	manager.RegisterProvider("memory", providers.NewMemoryCache(providers.DefaultMemoryCacheOptions()))
	manager.RegisterSerializer("json", serializers.NewJSONSerializer())

	// 只获取一次 Cache
	myCache := manager.GetCache("default")

	// ========================================
	// 存储不同类型的数据
	// ========================================
	fmt.Println("【存储数据】")
	fmt.Println("----------------------------------------")

	myCache.Set(ctx, cache.K("user:1"), &User{ID: 1, Name: "张三", Age: 25}, nil)
	myCache.Set(ctx, cache.K("product:1"), &Product{ID: 1, Name: "笔记本", Price: 5999}, nil)
	myCache.Set(ctx, cache.K("order:1"), &Order{ID: 1, UserID: 1, Total: 5999}, nil)

	fmt.Println("✓ 已存储 User、Product、Order 三种类型")

	// ========================================
	// 使用 GetAs 获取数据
	// ========================================
	fmt.Println("\n【使用 GetAs 获取数据】")
	fmt.Println("----------------------------------------")

	// 获取 User
	var user User
	err := myCache.GetAs(ctx, cache.K("user:1"), &user)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ 用户: %s (年龄: %d)\n", user.Name, user.Age)

	// 获取 Product
	var product Product
	err = myCache.GetAs(ctx, cache.K("product:1"), &product)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ 产品: %s (价格: %.2f)\n", product.Name, product.Price)

	// 获取 Order
	var order Order
	err = myCache.GetAs(ctx, cache.K("order:1"), &order)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ 订单: ID=%d (总额: %.2f)\n", order.ID, order.Total)

	// ========================================
	// 使用字符串键
	// ========================================
	fmt.Println("\n【使用字符串键】")
	fmt.Println("----------------------------------------")

	myCache.Set(ctx, cache.K("user:2"), &User{ID: 2, Name: "李四", Age: 30}, nil)

	var user2 User
	err = myCache.GetAs(ctx, cache.K("user:2"), &user2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ 用户: %s (年龄: %d)\n", user2.Name, user2.Age)

	// ========================================
	// GetOrSetAs 模式
	// ========================================
	fmt.Println("\n【GetOrSetAs 模式】")
	fmt.Println("----------------------------------------")

	fmt.Println("第一次获取用户 3:")
	var user3 User
	err = myCache.GetOrSetAs(ctx, cache.K("user:3"), &user3, func() (interface{}, error) {
		fmt.Println("  📥 从数据库加载...")
		time.Sleep(100 * time.Millisecond)
		return &User{ID: 3, Name: "王五", Age: 28}, nil
	}, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ 用户: %s (年龄: %d)\n", user3.Name, user3.Age)

	fmt.Println("\n第二次获取用户 3:")
	var user3Again User
	err = myCache.GetOrSetAs(ctx, cache.K("user:3"), &user3Again, func() (interface{}, error) {
		fmt.Println("  ⚠️  不应该看到这条消息")
		return nil, nil
	}, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ 从缓存读取: %s\n", user3Again.Name)

	// ========================================
	// 实际业务场景
	// ========================================
	fmt.Println("\n【实际业务场景】MultiTypeService")
	fmt.Println("----------------------------------------")

	service := NewMultiTypeService(manager)

	// 获取用户
	user4, err := service.GetUser(ctx, 4)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ 用户: %s (年龄: %d)\n", user4.Name, user4.Age)

	// 获取产品
	product2, err := service.GetProduct(ctx, 2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ 产品: %s (价格: %.2f)\n", product2.Name, product2.Price)

	// 获取订单
	order2, err := service.GetOrder(ctx, 2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ 订单: ID=%d (总额: %.2f)\n", order2.ID, order2.Total)

	// ========================================
	// 总结
	// ========================================
	fmt.Println("\n\n========================================")
	fmt.Println("【总结】")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("✓ 一个 Cache 实例可以存储多种不同类型")
	fmt.Println("✓ 只需要 GetCache 一次")
	fmt.Println("✓ 使用 GetAs 传入指针参数")
	fmt.Println("✓ 简单直接，无需泛型")
	fmt.Println()
	fmt.Println("推荐用法:")
	fmt.Println()
	fmt.Println("  // 1. 获取 Cache")
	fmt.Println("  cache := manager.GetCache(\"default\")")
	fmt.Println()
	fmt.Println("  // 2. 存储不同类型")
	fmt.Println("  cache.Set(ctx, cache.K(\"user:1\"), &User{...}, nil)")
	fmt.Println("  cache.Set(ctx, cache.K(\"product:1\"), &Product{...}, nil)")
	fmt.Println()
	fmt.Println("  // 3. 获取时传入指针")
	fmt.Println("  var user User")
	fmt.Println("  cache.GetAs(ctx, cache.K(\"user:1\"), &user)")
	fmt.Println()
	fmt.Println("  var product Product")
	fmt.Println("  cache.GetAs(ctx, cache.K(\"product:1\"), &product)")
	fmt.Println()
	fmt.Println("========================================\n")
}

// ========================================
// MultiTypeService - 实际业务场景
// ========================================

type MultiTypeService struct {
	cache cache.ICache
}

func NewMultiTypeService(manager *cache.CacheManager) *MultiTypeService {
	// 只获取一次 Cache
	c := manager.GetCache("business")
	return &MultiTypeService{cache: c}
}

func (s *MultiTypeService) GetUser(ctx context.Context, userID int) (*User, error) {
	key := cache.K(fmt.Sprintf("user:%d", userID))

	var user User
	err := s.cache.GetOrSetAs(ctx, key, &user, func() (interface{}, error) {
		fmt.Printf("  📥 从数据库加载用户 %d...\n", userID)
		time.Sleep(50 * time.Millisecond)
		return &User{ID: userID, Name: fmt.Sprintf("用户_%d", userID), Age: 20 + userID}, nil
	}, nil)

	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *MultiTypeService) GetProduct(ctx context.Context, productID int) (*Product, error) {
	key := cache.K(fmt.Sprintf("product:%d", productID))

	var product Product
	err := s.cache.GetOrSetAs(ctx, key, &product, func() (interface{}, error) {
		fmt.Printf("  📥 从数据库加载产品 %d...\n", productID)
		time.Sleep(50 * time.Millisecond)
		return &Product{ID: productID, Name: fmt.Sprintf("产品_%d", productID), Price: 99.99}, nil
	}, nil)

	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (s *MultiTypeService) GetOrder(ctx context.Context, orderID int) (*Order, error) {
	key := cache.K(fmt.Sprintf("order:%d", orderID))

	var order Order
	err := s.cache.GetOrSetAs(ctx, key, &order, func() (interface{}, error) {
		fmt.Printf("  📥 从数据库加载订单 %d...\n", orderID)
		time.Sleep(50 * time.Millisecond)
		return &Order{ID: orderID, UserID: orderID, Total: 199.99}, nil
	}, nil)

	if err != nil {
		return nil, err
	}
	return &order, nil
}
