package providers

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"
)

// RedisOptions 定义 Redis 缓存的选项
type RedisOptions struct {
	// Addresses Redis 服务器地址列表（用于集群模式）
	Addresses []string
	// Password Redis 认证密码
	Password string
	// DB 数据库编号
	DB int
	// DialTimeout 连接超时时间
	DialTimeout time.Duration
	// ReadTimeout 读取超时时间
	ReadTimeout time.Duration
	// WriteTimeout 写入超时时间
	WriteTimeout time.Duration
	// PoolSize 连接池大小
	PoolSize int
	// KeyPrefix 所有缓存键的前缀
	KeyPrefix string
}

// DefaultRedisOptions 返回 Redis 缓存的默认选项
func DefaultRedisOptions() *RedisOptions {
	return &RedisOptions{
		Addresses:    []string{"localhost:6379"},
		DB:           0,
		DialTimeout:  time.Second * 5,
		ReadTimeout:  time.Second * 3,
		WriteTimeout: time.Second * 3,
		PoolSize:     10,
	}
}

// RedisCache 使用 Redis 实现 ICacheProvider 接口
// 注意：这是一个简化的实现。在生产环境中，应该使用适当的 Redis 客户端，如 go-redis
type RedisCache struct {
	options *RedisOptions
	conn    net.Conn
}

// NewRedisCache 创建一个新的 Redis 缓存提供者
// 注意：这是一个用于演示的基本实现。
// 在生产环境中，请使用适当的 Redis 客户端库，如 github.com/go-redis/redis/v8
func NewRedisCache(options *RedisOptions) (*RedisCache, error) {
	if options == nil {
		options = DefaultRedisOptions()
	}

	// 连接到第一个 Redis 服务器
	addr := options.Addresses[0]
	conn, err := net.DialTimeout("tcp", addr, options.DialTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis at %s: %w", addr, err)
	}

	rc := &RedisCache{
		options: options,
		conn:    conn,
	}

	// 如果提供了密码，进行认证
	if options.Password != "" {
		if err := rc.auth(); err != nil {
			conn.Close()
			return nil, fmt.Errorf("Redis authentication failed: %w", err)
		}
	}

	// 选择数据库
	if options.DB != 0 {
		if err := rc.selectDB(); err != nil {
			conn.Close()
			return nil, fmt.Errorf("failed to select Redis database %d: %w", options.DB, err)
		}
	}

	return rc, nil
}

// Name returns the provider name
func (rc *RedisCache) Name() string {
	return "redis"
}

// GetRaw 从 Redis 获取原始数据
func (rc *RedisCache) GetRaw(ctx context.Context, key string) ([]byte, error) {
	fullKey := rc.buildKey(key)

	// 发送 GET 命令
	cmd := fmt.Sprintf("GET %s\r\n", fullKey)
	if _, err := rc.conn.Write([]byte(cmd)); err != nil {
		return nil, fmt.Errorf("failed to send GET command: %w", err)
	}

	// 读取响应
	response, err := rc.readResponse()
	if err != nil {
		return nil, err
	}

	// 解析响应
	if strings.HasPrefix(response, "$-1") {
		// 键未找到
		return nil, nil
	}

	if strings.HasPrefix(response, "$") {
		// 批量字符串响应
		lines := strings.Split(response, "\r\n")
		if len(lines) >= 2 {
			return []byte(lines[1]), nil
		}
	}

	return nil, fmt.Errorf("unexpected Redis response: %s", response)
}

// SetRaw 将原始数据存储到 Redis
func (rc *RedisCache) SetRaw(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	fullKey := rc.buildKey(key)

	var cmd string
	if expiration > 0 {
		// 带过期时间的 SET
		seconds := int(expiration.Seconds())
		cmd = fmt.Sprintf("SET %s %q EX %d\r\n", fullKey, string(value), seconds)
	} else {
		// 不带过期时间的 SET
		cmd = fmt.Sprintf("SET %s %q\r\n", fullKey, string(value))
	}

	if _, err := rc.conn.Write([]byte(cmd)); err != nil {
		return fmt.Errorf("failed to send SET command: %w", err)
	}

	// 读取响应
	response, err := rc.readResponse()
	if err != nil {
		return err
	}

	if !strings.Contains(response, "OK") {
		return fmt.Errorf("SET command failed: %s", response)
	}

	return nil
}

// Remove 从 Redis 移除数据
func (rc *RedisCache) Remove(ctx context.Context, key string) error {
	fullKey := rc.buildKey(key)

	cmd := fmt.Sprintf("DEL %s\r\n", fullKey)
	if _, err := rc.conn.Write([]byte(cmd)); err != nil {
		return fmt.Errorf("failed to send DEL command: %w", err)
	}

	// 读取响应 (we don't need to check the result for DEL)
	_, err := rc.readResponse()
	return err
}

// RemoveByPattern removes all keys matching the pattern
func (rc *RedisCache) RemoveByPattern(ctx context.Context, pattern string) error {
	fullPattern := rc.buildKey(pattern)

	// 首先，使用 KEYS 命令获取所有匹配的键
	cmd := fmt.Sprintf("KEYS %s\r\n", fullPattern)
	if _, err := rc.conn.Write([]byte(cmd)); err != nil {
		return fmt.Errorf("failed to send KEYS command: %w", err)
	}

	response, err := rc.readResponse()
	if err != nil {
		return err
	}

	// 从响应中解析键并删除它们
	// 这是一个简化的实现
	if strings.Contains(response, "*0") {
		// 未找到键
		return nil
	}

	// 在实际实现中，你应该解析数组响应并删除每个键
	// 现在，我们使用 FLUSHDB 作为简单方法（仅当模式为 * 时）
	if pattern == "*" {
		return rc.Clear(ctx)
	}

	return nil
}

// Exists 检查键是否存在于 Redis
func (rc *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := rc.buildKey(key)

	cmd := fmt.Sprintf("EXISTS %s\r\n", fullKey)
	if _, err := rc.conn.Write([]byte(cmd)); err != nil {
		return false, fmt.Errorf("failed to send EXISTS command: %w", err)
	}

	response, err := rc.readResponse()
	if err != nil {
		return false, err
	}

	// 检查响应是否表示键存在（存在应为 ":1"，不存在为 ":0"）
	return strings.Contains(response, ":1"), nil
}

// Clear 清空当前数据库中的所有数据
func (rc *RedisCache) Clear(ctx context.Context) error {
	cmd := "FLUSHDB\r\n"
	if _, err := rc.conn.Write([]byte(cmd)); err != nil {
		return fmt.Errorf("failed to send FLUSHDB command: %w", err)
	}

	response, err := rc.readResponse()
	if err != nil {
		return err
	}

	if !strings.Contains(response, "OK") {
		return fmt.Errorf("FLUSHDB command failed: %s", response)
	}

	return nil
}

// Close 关闭 Redis 连接
func (rc *RedisCache) Close() error {
	if rc.conn != nil {
		return rc.conn.Close()
	}
	return nil
}

// buildKey 构建带前缀的完整缓存键
func (rc *RedisCache) buildKey(key string) string {
	if rc.options.KeyPrefix == "" {
		return key
	}
	return rc.options.KeyPrefix + ":" + key
}

// auth 使用提供的密码向 Redis 进行认证
func (rc *RedisCache) auth() error {
	cmd := fmt.Sprintf("AUTH %s\r\n", rc.options.Password)
	if _, err := rc.conn.Write([]byte(cmd)); err != nil {
		return err
	}

	response, err := rc.readResponse()
	if err != nil {
		return err
	}

	if !strings.Contains(response, "OK") {
		return fmt.Errorf("authentication failed: %s", response)
	}

	return nil
}

// selectDB 选择 Redis 数据库
func (rc *RedisCache) selectDB() error {
	cmd := fmt.Sprintf("SELECT %d\r\n", rc.options.DB)
	if _, err := rc.conn.Write([]byte(cmd)); err != nil {
		return err
	}

	response, err := rc.readResponse()
	if err != nil {
		return err
	}

	if !strings.Contains(response, "OK") {
		return fmt.Errorf("database selection failed: %s", response)
	}

	return nil
}

// readResponse 从 Redis 读取响应
// 注意：这是一个用于演示目的的非常基本的 Redis 协议解析器
func (rc *RedisCache) readResponse() (string, error) {
	// 设置读取超时
	if err := rc.conn.SetReadDeadline(time.Now().Add(rc.options.ReadTimeout)); err != nil {
		return "", err
	}

	// 读取响应 (simplified - in production you'd need a proper Redis protocol parser)
	buffer := make([]byte, 4096)
	n, err := rc.conn.Read(buffer)
	if err != nil {
		return "", fmt.Errorf("failed to read Redis response: %w", err)
	}

	return string(buffer[:n]), nil
}

// MockRedisCache 是用于测试的 Redis 缓存模拟实现
type MockRedisCache struct {
	data map[string][]byte
}

// NewMockRedisCache 创建一个新的模拟 Redis 缓存用于测试
func NewMockRedisCache() *MockRedisCache {
	return &MockRedisCache{
		data: make(map[string][]byte),
	}
}

// Name returns the provider name
func (mrc *MockRedisCache) Name() string {
	return "mock-redis"
}

// GetRaw 从模拟 Redis 获取原始数据
func (mrc *MockRedisCache) GetRaw(ctx context.Context, key string) ([]byte, error) {
	if data, exists := mrc.data[key]; exists {
		result := make([]byte, len(data))
		copy(result, data)
		return result, nil
	}
	return nil, nil
}

// SetRaw 将原始数据存储到模拟 Redis
func (mrc *MockRedisCache) SetRaw(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	data := make([]byte, len(value))
	copy(data, value)
	mrc.data[key] = data
	return nil
}

// Remove 从模拟 Redis 移除数据
func (mrc *MockRedisCache) Remove(ctx context.Context, key string) error {
	delete(mrc.data, key)
	return nil
}

// RemoveByPattern 从模拟 Redis 移除所有匹配模式的键
func (mrc *MockRedisCache) RemoveByPattern(ctx context.Context, pattern string) error {
	if pattern == "*" {
		mrc.data = make(map[string][]byte)
		return nil
	}
	// Simplified pattern matching
	for key := range mrc.data {
		if strings.Contains(key, strings.Trim(pattern, "*")) {
			delete(mrc.data, key)
		}
	}
	return nil
}

// Exists 检查键是否存在于模拟 Redis
func (mrc *MockRedisCache) Exists(ctx context.Context, key string) (bool, error) {
	_, exists := mrc.data[key]
	return exists, nil
}

// Clear 清空模拟 Redis 中的所有数据
func (mrc *MockRedisCache) Clear(ctx context.Context) error {
	mrc.data = make(map[string][]byte)
	return nil
}

// Close 关闭模拟 Redis 缓存
func (mrc *MockRedisCache) Close() error {
	return nil
}
