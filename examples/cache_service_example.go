package main

import 	// 1. 创建L1: 内存缓存
	memoryOptions := &providers.MemoryCacheOptions{
		MaxSize:           10000,              // 最多缓存10000个用户
		DefaultExpiration: time.Minute * 5,    // 默认5分钟过期
		CleanupInterval:   time.Minute,        // 每分钟清理一次过期数据
		EnableLRU:         true,               // 启用LRU淘汰
	}
	memoryProvider := providers.NewMemoryCache(memoryOptions)

	// 2. 创建L2: Redis缓存
	redisProvider := providers.NewRedisCache(redisAddr, "", 0)t"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"gitee.com/wangsoft/go-library/cache"
	"gitee.com/wangsoft/go-library/cache/providers"
	"gitee.com/wangsoft/go-library/cache/serializers"
)

// CacheService 封装了多级缓存的应用服务
type CacheService struct {
	manager   *cache.CacheManager
	db        *sql.DB
	userCache cache.ICache
}

// NewCacheService 创建缓存服务
func NewCacheService(db *sql.DB, redisAddr string) (*CacheService, error) {
	// 1. 创建L1: 内存缓存
	memoryOptions := &providers.MemoryCacheOptions{
		MaxSize:           10000,           // 最多缓存10000个用户
		DefaultExpiration: time.Minute * 5, // 默认5分钟过期
		CleanupInterval:   time.Minute,     // 每分钟清理一次过期数据
		EnableLRU:         true,            // 启用LRU淘汰
	}
	memoryProvider := providers.NewMemoryCacheProvider(memoryOptions)

	// 2. 创建L2: Redis缓存
	redisProvider := providers.NewRedisCacheProvider(redisAddr, "", 0)

	// 3. 配置多级缓存选项
	multiLevelOptions := &providers.MultiLevelCacheOptions{
		EnableAsyncWrite: true,            // 异步写入低级缓存
		EnableAutoSync:   false,           // 不启用自动同步
		WriteDownLevels:  0,               // 写入所有层级
		EnableMetrics:    true,            // 启用性能监控
		L1TTL:            time.Minute * 5, // L1缓存5分钟
		L2TTL:            time.Hour * 1,   // L2缓存1小时
		EnableWriteBack:  true,            // 启用写回策略（高性能）
	}

	service := &CacheService{db: db}

	// 4. 创建数据库加载器
	dbLoader := func(ctx context.Context, key string) ([]byte, error) {
		return service.loadFromDatabase(ctx, key)
	}

	// 5. 创建多级缓存提供者
	multiLevelProvider := providers.NewMultiLevelCacheProvider(multiLevelOptions, dbLoader)
	multiLevelProvider.AddLevel("L1-Memory", memoryProvider, 1)
	multiLevelProvider.AddLevel("L2-Redis", redisProvider, 2)

	// 6. 创建缓存管理器
	manager := cache.NewCacheManager()
	manager.RegisterProvider("multilevel", multiLevelProvider)
	manager.RegisterSerializer("json", serializers.NewJSONSerializer())
	manager.SetDefaultProvider("multilevel")
	manager.SetDefaultSerializer("json")

	service.manager = manager
	service.userCache = manager.GetCache("user-cache")

	return service, nil
}

// loadFromDatabase 从数据库加载数据
func (s *CacheService) loadFromDatabase(ctx context.Context, key string) ([]byte, error) {
	// 这里实现实际的数据库查询逻辑
	// key格式: "users:user:123"

	// 示例实现
	var data interface{}
	// query := "SELECT * FROM users WHERE id = ?"
	// err := s.db.QueryRowContext(ctx, query, id).Scan(&data)

	// 序列化数据
	return json.Marshal(data)
}

// GetUser 获取用户信息（自动使用多级缓存）
func (s *CacheService) GetUser(ctx context.Context, userID int) (*User, error) {
	key := cache.NewCacheKey(fmt.Sprintf("user:%d", userID), "users")
	
	value, err := s.userCache.Get(ctx, key)
	if err != nil || value == nil {
		// 缓存未命中，从数据库加载
		return s.getUserFromDB(ctx, userID)
	}
	
	// 类型断言
	if userMap, ok := value.(map[string]interface{}); ok {
		user := &User{
			ID:       int(userMap["id"].(float64)),
			Username: userMap["username"].(string),
			Email:    userMap["email"].(string),
		}
		return user, nil
	}
	
	return nil, fmt.Errorf("invalid user data type")
}// GetUserWithFallback 带降级策略的用户获取
func (s *CacheService) GetUserWithFallback(ctx context.Context, userID int) (*User, error) {
	key := cache.NewCacheKey(fmt.Sprintf("user:%d", userID), "users")

	// 使用GetOrSet确保只加载一次
	result, err := s.userCache.GetOrSet(ctx, key, func() (interface{}, error) {
		return s.getUserFromDB(ctx, userID)
	}, nil)

	if err != nil {
		return nil, err
	}

	if user, ok := result.(*User); ok {
		return user, nil
	}

	return nil, fmt.Errorf("invalid user data type")
}

// SetUser 设置用户信息（写入多级缓存）
func (s *CacheService) SetUser(ctx context.Context, user *User) error {
	key := cache.NewCacheKey(fmt.Sprintf("user:%d", user.ID), "users")

	options := cache.DefaultCacheOptions()
	options.WithSlidingExpiration(time.Minute * 10) // 滑动过期10分钟

	return s.userCache.Set(ctx, key, user, options)
}

// UpdateUser 更新用户信息（更新数据库并清除缓存）
func (s *CacheService) UpdateUser(ctx context.Context, user *User) error {
	// 1. 更新数据库
	err := s.updateUserInDB(ctx, user)
	if err != nil {
		return err
	}

	// 2. 清除所有层级的缓存（确保数据一致性）
	key := cache.NewCacheKey(fmt.Sprintf("user:%d", user.ID), "users")
	err = s.userCache.Remove(ctx, key)
	if err != nil {
		// 记录日志，但不影响更新操作
		fmt.Printf("警告: 清除缓存失败: %v\n", err)
	}

	// 3. 可选：立即写入新数据到缓存
	options := cache.DefaultCacheOptions()
	s.userCache.Set(ctx, key, user, options)

	return nil
}

// DeleteUser 删除用户（删除数据库并清除缓存）
func (s *CacheService) DeleteUser(ctx context.Context, userID int) error {
	// 1. 删除数据库记录
	err := s.deleteUserFromDB(ctx, userID)
	if err != nil {
		return err
	}

	// 2. 清除缓存
	key := cache.NewCacheKey(fmt.Sprintf("user:%d", userID), "users")
	return s.userCache.Remove(ctx, key)
}

// BatchGetUsers 批量获取用户（利用缓存）
func (s *CacheService) BatchGetUsers(ctx context.Context, userIDs []int) ([]*User, error) {
	users := make([]*User, 0, len(userIDs))

	for _, id := range userIDs {
		user, err := s.GetUser(ctx, id)
		if err != nil {
			// 可以选择跳过错误或返回错误
			continue
		}
		users = append(users, user)
	}

	return users, nil
}

// WarmupCache 预热缓存
func (s *CacheService) WarmupCache(ctx context.Context, userIDs []int) error {
	for _, id := range userIDs {
		key := cache.NewCacheKey(fmt.Sprintf("user:%d", id), "users")

		// 使用GetOrSet避免重复加载
		_, err := s.userCache.GetOrSet(ctx, key, func() (interface{}, error) {
			return s.getUserFromDB(ctx, id)
		}, nil)

		if err != nil {
			fmt.Printf("预热缓存失败 (user %d): %v\n", id, err)
		}
	}

	return nil
}

// ClearUserCache 清除用户缓存
func (s *CacheService) ClearUserCache(ctx context.Context) error {
	return s.userCache.RemoveByPattern(ctx, "user:*")
}

// Close 关闭缓存服务
func (s *CacheService) Close() error {
	return s.manager.Close()
}

// 辅助方法（实际实现需要根据数据库类型调整）

func (s *CacheService) getUserFromDB(ctx context.Context, userID int) (*User, error) {
	var user User
	query := "SELECT id, username, email, created_at FROM users WHERE id = ?"
	err := s.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Created,
	)
	return &user, err
}

func (s *CacheService) updateUserInDB(ctx context.Context, user *User) error {
	query := "UPDATE users SET username = ?, email = ? WHERE id = ?"
	_, err := s.db.ExecContext(ctx, query, user.Username, user.Email, user.ID)
	return err
}

func (s *CacheService) deleteUserFromDB(ctx context.Context, userID int) error {
	query := "DELETE FROM users WHERE id = ?"
	_, err := s.db.ExecContext(ctx, query, userID)
	return err
}

// User 用户模型
type User struct {
	ID       int       `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Created  time.Time `json:"created"`
}

// === 在HTTP Handler中使用示例 ===

/*
// 在您的HTTP服务中使用

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
    userID, _ := strconv.Atoi(r.URL.Query().Get("id"))

    // 自动使用多级缓存
    user, err := h.cacheService.GetUser(r.Context(), userID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }

    json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
    var user User
    json.NewDecoder(r.Body).Decode(&user)

    // 更新数据库并自动清除缓存
    err := h.cacheService.UpdateUser(r.Context(), &user)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
}
*/
