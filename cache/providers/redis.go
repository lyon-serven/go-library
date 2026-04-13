package providers

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
)

// AlertLevel 告警级别
type AlertLevel int

const (
	AlertLevelWarn    AlertLevel = iota // 警告：延迟过高
	AlertLevelError                     // 错误：连接断开
	AlertLevelRecover                   // 恢复：连接重新建立
)

// AlertEvent 告警事件
type AlertEvent struct {
	Level     AlertLevel
	Message   string
	Err       error
	Latency   time.Duration
	Timestamp time.Time
}

// AlertCallback 告警回调函数类型
type AlertCallback func(event AlertEvent)

// RedisOptions 定义 Redis 缓存的选项
type RedisOptions struct {
	Addresses            []string      // Redis 服务器地址列表
	Username             string        // ACL 用户名（Redis 6.0+，阿里云/腾讯云等需要）
	Password             string        // 认证密码
	DB                   int           // 数据库编号
	DialTimeout          time.Duration // 连接超时
	ReadTimeout          time.Duration // 读取超时
	WriteTimeout         time.Duration // 写入超时
	PoolSize             int           // 连接池大小
	KeyPrefix            string        // 键前缀
	EnableHealthCheck    bool          // 是否开启心跳监测
	HealthCheckInterval  time.Duration // 心跳检测间隔，默认 30s
	HealthCheckTimeout   time.Duration // 单次 Ping 超时，默认 3s
	LatencyWarnThreshold time.Duration // 延迟告警阈值，默认 200ms
	OnAlert              AlertCallback // 告警回调（nil 时不回调）
}

// DefaultRedisOptions 返回默认选项
func DefaultRedisOptions() *RedisOptions {
	return &RedisOptions{
		Addresses:            []string{"localhost:6379"},
		DialTimeout:          5 * time.Second,
		ReadTimeout:          3 * time.Second,
		WriteTimeout:         3 * time.Second,
		PoolSize:             10,
		HealthCheckInterval:  30 * time.Second,
		HealthCheckTimeout:   3 * time.Second,
		LatencyWarnThreshold: 200 * time.Millisecond,
	}
}

// PipelineSetItem 表示 Pipeline 批量写入的单个条目
type PipelineSetItem struct {
	Key        string
	Value      []byte
	Expiration time.Duration
}

// RedisCache 使用 go-redis 实现缓存接口
type RedisCache struct {
	options     *RedisOptions
	client      *redis.Client
	healthy     atomic.Bool
	stopHealthy chan struct{}
}

// NewRedisCache 创建 Redis 缓存提供者
func NewRedisCache(options *RedisOptions) (*RedisCache, error) {
	if options == nil {
		options = DefaultRedisOptions()
	}
	client := redis.NewClient(&redis.Options{
		Addr:         options.Addresses[0],
		Username:     options.Username,
		Password:     options.Password,
		DB:           options.DB,
		DialTimeout:  options.DialTimeout,
		ReadTimeout:  options.ReadTimeout,
		WriteTimeout: options.WriteTimeout,
		PoolSize:     options.PoolSize,
	})
	ctx, cancel := context.WithTimeout(context.Background(), options.DialTimeout)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, err
	}
	rc := &RedisCache{options: options, client: client, stopHealthy: make(chan struct{})}
	if options.EnableHealthCheck {
		rc.healthy.Store(true)
		go rc.healthCheckLoop()
	}
	return rc, nil
}

// ============================================
// ICacheProvider 实现
// ============================================

func (rc *RedisCache) Name() string { return "redis" }

func (rc *RedisCache) GetRaw(ctx context.Context, key string) ([]byte, error) {
	data, err := rc.client.Get(ctx, rc.buildKey(key)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return data, err
}

func (rc *RedisCache) SetRaw(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	return rc.client.Set(ctx, rc.buildKey(key), value, expiration).Err()
}

func (rc *RedisCache) Remove(ctx context.Context, key string) error {
	return rc.client.Del(ctx, rc.buildKey(key)).Err()
}

func (rc *RedisCache) RemoveByPattern(ctx context.Context, pattern string) error {
	keys, err := rc.client.Keys(ctx, rc.buildKey(pattern)).Result()
	if err != nil || len(keys) == 0 {
		return err
	}
	return rc.client.Del(ctx, keys...).Err()
}

func (rc *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	n, err := rc.client.Exists(ctx, rc.buildKey(key)).Result()
	return n > 0, err
}

func (rc *RedisCache) Clear(ctx context.Context) error {
	return rc.client.FlushDB(ctx).Err()
}

func (rc *RedisCache) Close() error {
	select {
	case <-rc.stopHealthy:
	default:
		close(rc.stopHealthy)
	}
	return rc.client.Close()
}

// ============================================
// IPipelineProvider 实现 —— 批量操作
// ============================================

// PipelineSet 使用 Pipeline 批量写入，一次网络往返完成所有写入
func (rc *RedisCache) PipelineSet(ctx context.Context, items []PipelineSetItem) error {
	if len(items) == 0 {
		return nil
	}
	pipe := rc.client.Pipeline()
	for _, item := range items {
		pipe.Set(ctx, rc.buildKey(item.Key), item.Value, item.Expiration)
	}
	_, err := pipe.Exec(ctx)
	return err
}

// PipelineRemove 使用 Pipeline 批量删除，一次网络往返完成所有删除
func (rc *RedisCache) PipelineRemove(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}
	pipe := rc.client.Pipeline()
	for _, key := range keys {
		pipe.Del(ctx, rc.buildKey(key))
	}
	_, err := pipe.Exec(ctx)
	return err
}

// ============================================
// 心跳监测
// ============================================

// Ping 发送 PING，返回延迟
func (rc *RedisCache) Ping(ctx context.Context) (time.Duration, error) {
	start := time.Now()
	err := rc.client.Ping(ctx).Err()
	return time.Since(start), err
}

// IsHealthy 返回连接是否健康
func (rc *RedisCache) IsHealthy() bool { return rc.healthy.Load() }

func (rc *RedisCache) healthCheckLoop() {
	ticker := time.NewTicker(rc.options.HealthCheckInterval)
	defer ticker.Stop()
	for {
		select {
		case <-rc.stopHealthy:
			return
		case <-ticker.C:
			rc.doHealthCheck()
		}
	}
}

func (rc *RedisCache) doHealthCheck() {
	ctx, cancel := context.WithTimeout(context.Background(), rc.options.HealthCheckTimeout)
	defer cancel()
	latency, err := rc.Ping(ctx)
	if err != nil {
		if rc.healthy.Swap(false) {
			rc.notify(AlertEvent{Level: AlertLevelError, Message: "Redis 连接断开", Err: err, Timestamp: time.Now()})
		}
		return
	}
	if !rc.healthy.Swap(true) {
		rc.notify(AlertEvent{Level: AlertLevelRecover, Message: "Redis 连接已恢复", Timestamp: time.Now()})
	}
	if rc.options.LatencyWarnThreshold > 0 && latency > rc.options.LatencyWarnThreshold {
		rc.notify(AlertEvent{Level: AlertLevelWarn, Message: "Redis 延迟过高", Latency: latency, Timestamp: time.Now()})
	}
}

func (rc *RedisCache) notify(event AlertEvent) {
	if rc.options.OnAlert != nil {
		rc.options.OnAlert(event)
	}
}

func (rc *RedisCache) buildKey(key string) string {
	if rc.options.KeyPrefix == "" {
		return key
	}
	return rc.options.KeyPrefix + ":" + key
}

// ============================================
// MockRedisCache —— 用于单元测试
// ============================================

// MockRedisCache 内存模拟，无需真实 Redis
type MockRedisCache struct {
	mu   sync.RWMutex
	data map[string][]byte
}

func NewMockRedisCache() *MockRedisCache {
	return &MockRedisCache{data: make(map[string][]byte)}
}

func (mrc *MockRedisCache) Name() string { return "mock-redis" }

func (mrc *MockRedisCache) GetRaw(_ context.Context, key string) ([]byte, error) {
	mrc.mu.RLock()
	defer mrc.mu.RUnlock()
	if data, ok := mrc.data[key]; ok {
		cp := make([]byte, len(data))
		copy(cp, data)
		return cp, nil
	}
	return nil, nil
}

func (mrc *MockRedisCache) SetRaw(_ context.Context, key string, value []byte, _ time.Duration) error {
	mrc.mu.Lock()
	defer mrc.mu.Unlock()
	cp := make([]byte, len(value))
	copy(cp, value)
	mrc.data[key] = cp
	return nil
}

func (mrc *MockRedisCache) Remove(_ context.Context, key string) error {
	mrc.mu.Lock()
	defer mrc.mu.Unlock()
	delete(mrc.data, key)
	return nil
}

func (mrc *MockRedisCache) RemoveByPattern(_ context.Context, pattern string) error {
	mrc.mu.Lock()
	defer mrc.mu.Unlock()
	if pattern == "*" {
		mrc.data = make(map[string][]byte)
		return nil
	}
	needle := strings.Trim(pattern, "*")
	for key := range mrc.data {
		if strings.Contains(key, needle) {
			delete(mrc.data, key)
		}
	}
	return nil
}

func (mrc *MockRedisCache) Exists(_ context.Context, key string) (bool, error) {
	mrc.mu.RLock()
	defer mrc.mu.RUnlock()
	_, ok := mrc.data[key]
	return ok, nil
}

func (mrc *MockRedisCache) Clear(_ context.Context) error {
	mrc.mu.Lock()
	defer mrc.mu.Unlock()
	mrc.data = make(map[string][]byte)
	return nil
}

func (mrc *MockRedisCache) Close() error { return nil }

func (mrc *MockRedisCache) PipelineSet(_ context.Context, items []PipelineSetItem) error {
	mrc.mu.Lock()
	defer mrc.mu.Unlock()
	for _, item := range items {
		cp := make([]byte, len(item.Value))
		copy(cp, item.Value)
		mrc.data[item.Key] = cp
	}
	return nil
}

func (mrc *MockRedisCache) PipelineRemove(_ context.Context, keys []string) error {
	mrc.mu.Lock()
	defer mrc.mu.Unlock()
	for _, key := range keys {
		delete(mrc.data, key)
	}
	return nil
}
