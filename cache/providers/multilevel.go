package providers

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// CacheLevel 表示多级缓存中的单个级别
type CacheLevel struct {
	Name     string
	Provider ICacheProvider
	Priority int // 数字越小 = 优先级越高（L1=1, L2=2, L3=3）
}

// ICacheProvider 接口（应与 cache 包中的接口匹配）
type ICacheProvider interface {
	GetRaw(ctx context.Context, key string) ([]byte, error)
	SetRaw(ctx context.Context, key string, value []byte, expiration time.Duration) error
	Remove(ctx context.Context, key string) error
	RemoveByPattern(ctx context.Context, pattern string) error
	Exists(ctx context.Context, key string) (bool, error)
	Clear(ctx context.Context) error
	Name() string
	Close() error
}

// MultiLevelCacheProvider 实现多级缓存策略
// L1：本地内存缓存（最快）
// L2：Redis 缓存（跨实例共享）
// L3：数据库/持久化存储（最慢，带加载器函数）
type MultiLevelCacheProvider struct {
	mu              sync.RWMutex
	levels          []CacheLevel
	options         *MultiLevelCacheOptions
	metrics         *MultiLevelMetrics
	dbLoader        DatabaseLoader // Function to load data from database
	asyncWriteQueue chan writeTask
	stopChan        chan struct{}
	wg              sync.WaitGroup
}

// DatabaseLoader 是从数据库加载数据的函数类型
type DatabaseLoader func(ctx context.Context, key string) ([]byte, error)

// MultiLevelCacheOptions 定义多级缓存的选项
type MultiLevelCacheOptions struct {
	// EnableAsyncWrite 启用向较低级别的异步写入
	EnableAsyncWrite bool
	// EnableAutoSync 启用级别之间的自动同步
	EnableAutoSync bool
	// SyncInterval 自动同步的间隔时间
	SyncInterval time.Duration
	// WriteDownLevels 控制向下写入多少级别（0 = 所有级别）
	WriteDownLevels int
	// EnableMetrics 启用性能指标收集
	EnableMetrics bool
	// L1TTL L1 缓存的 TTL（内存）
	L1TTL time.Duration
	// L2TTL L2 缓存的 TTL（Redis）
	L2TTL time.Duration
	// EnableWriteBack 启用回写策略（先写入 L1，然后异步写入其他级别）
	EnableWriteBack bool
}

// DefaultMultiLevelCacheOptions 返回默认选项
func DefaultMultiLevelCacheOptions() *MultiLevelCacheOptions {
	return &MultiLevelCacheOptions{
		EnableAsyncWrite: true,
		EnableAutoSync:   false,
		SyncInterval:     time.Minute * 5,
		WriteDownLevels:  0, // 写入所有级别
		EnableMetrics:    true,
		L1TTL:            time.Minute * 5, // L1：5 分钟
		L2TTL:            time.Hour * 1,   // L2：1 小时
		EnableWriteBack:  true,
	}
}

// MultiLevelMetrics 跟踪缓存性能指标
type MultiLevelMetrics struct {
	mu             sync.RWMutex
	L1Hits         int64
	L2Hits         int64
	L3Hits         int64
	Misses         int64
	TotalRequests  int64
	L1WriteCount   int64
	L2WriteCount   int64
	L3WriteCount   int64
	PromotionCount int64 // 数据提升到更高级别的次数
}

// writeTask 表示一个异步写入任务
type writeTask struct {
	ctx        context.Context
	key        string
	value      []byte
	expiration time.Duration
	levels     []int // 要写入的目标级别
}

// NewMultiLevelCacheProvider 创建一个新的多级缓存提供者
func NewMultiLevelCacheProvider(options *MultiLevelCacheOptions, dbLoader DatabaseLoader) *MultiLevelCacheProvider {
	if options == nil {
		options = DefaultMultiLevelCacheOptions()
	}

	mlcp := &MultiLevelCacheProvider{
		levels:          make([]CacheLevel, 0),
		options:         options,
		metrics:         &MultiLevelMetrics{},
		dbLoader:        dbLoader,
		asyncWriteQueue: make(chan writeTask, 1000),
		stopChan:        make(chan struct{}),
	}

	// 如果启用，启动异步写入工作协程
	if options.EnableAsyncWrite {
		mlcp.startAsyncWriteWorkers(3) // 3 worker goroutines
	}

	return mlcp
}

// AddLevel 向多级缓存添加一个缓存级别
func (mlcp *MultiLevelCacheProvider) AddLevel(name string, provider ICacheProvider, priority int) error {
	mlcp.mu.Lock()
	defer mlcp.mu.Unlock()

	// 检查级别是否已存在
	for _, level := range mlcp.levels {
		if level.Priority == priority {
			return fmt.Errorf("cache level with priority %d already exists", priority)
		}
		if level.Name == name {
			return fmt.Errorf("cache level with name '%s' already exists", name)
		}
	}

	// 添加新级别
	mlcp.levels = append(mlcp.levels, CacheLevel{
		Name:     name,
		Provider: provider,
		Priority: priority,
	})

	// 按优先级排序（升序）
	mlcp.sortLevels()

	return nil
}

// sortLevels 按优先级对缓存级别排序
func (mlcp *MultiLevelCacheProvider) sortLevels() {
	// 冒泡排序（简单，适用于少量级别）
	for i := 0; i < len(mlcp.levels); i++ {
		for j := i + 1; j < len(mlcp.levels); j++ {
			if mlcp.levels[i].Priority > mlcp.levels[j].Priority {
				mlcp.levels[i], mlcp.levels[j] = mlcp.levels[j], mlcp.levels[i]
			}
		}
	}
}

// GetRaw 从缓存中获取数据，按顺序尝试每个级别
func (mlcp *MultiLevelCacheProvider) GetRaw(ctx context.Context, key string) ([]byte, error) {
	if mlcp.options.EnableMetrics {
		mlcp.metrics.mu.Lock()
		mlcp.metrics.TotalRequests++
		mlcp.metrics.mu.Unlock()
	}

	mlcp.mu.RLock()
	levels := mlcp.levels
	mlcp.mu.RUnlock()

	// 按顺序尝试每个级别（L1 -> L2 -> L3 -> DB）
	for i, level := range levels {
		data, err := level.Provider.GetRaw(ctx, key)
		if err == nil && data != nil {
			// 缓存命中 - 更新指标
			mlcp.recordHit(i + 1)

			// 将数据提升到更高级别（缓存预热）
			if i > 0 {
				mlcp.promoteToHigherLevels(ctx, key, data, i)
			}

			return data, nil
		}
	}

	// 所有缓存级别未命中，尝试数据库加载器
	if mlcp.dbLoader != nil {
		data, err := mlcp.dbLoader(ctx, key)
		if err == nil && data != nil {
			// 记录 L3（数据库）命中
			mlcp.recordHit(3)

			// 回写到所有缓存级别
			mlcp.writeToAllLevels(ctx, key, data)

			return data, nil
		}
	}

	// 完全未命中
	if mlcp.options.EnableMetrics {
		mlcp.metrics.mu.Lock()
		mlcp.metrics.Misses++
		mlcp.metrics.mu.Unlock()
	}

	return nil, fmt.Errorf("cache miss: key '%s' not found in any level", key)
}

// SetRaw 将数据存储到缓存中，写入所有配置的级别
func (mlcp *MultiLevelCacheProvider) SetRaw(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	mlcp.mu.RLock()
	levels := mlcp.levels
	writeDownLevels := mlcp.options.WriteDownLevels
	mlcp.mu.RUnlock()

	if len(levels) == 0 {
		return fmt.Errorf("no cache levels configured")
	}

	// 确定要写入哪些级别
	targetLevels := len(levels)
	if writeDownLevels > 0 && writeDownLevels < targetLevels {
		targetLevels = writeDownLevels
	}

	// 回写策略：立即写入 L1，其他异步写入
	if mlcp.options.EnableWriteBack && len(levels) > 0 {
		// 立即写入 L1
		err := levels[0].Provider.SetRaw(ctx, key, value, mlcp.getTTLForLevel(1, expiration))
		if err != nil {
			return fmt.Errorf("failed to write to L1 cache: %w", err)
		}
		mlcp.recordWrite(1)

		// 异步写入其他级别
		if mlcp.options.EnableAsyncWrite && targetLevels > 1 {
			levelIndices := make([]int, targetLevels-1)
			for i := 1; i < targetLevels; i++ {
				levelIndices[i-1] = i
			}

			select {
			case mlcp.asyncWriteQueue <- writeTask{
				ctx:        context.Background(), // Use background context for async writes
				key:        key,
				value:      value,
				expiration: expiration,
				levels:     levelIndices,
			}:
			default:
				// 队列已满，回退到同步写入
				mlcp.writeSyncToLevels(ctx, key, value, expiration, levelIndices)
			}
		}

		return nil
	}

	// 直写策略：同步写入所有级别
	var firstError error
	for i := 0; i < targetLevels; i++ {
		ttl := mlcp.getTTLForLevel(i+1, expiration)
		err := levels[i].Provider.SetRaw(ctx, key, value, ttl)
		if err != nil && firstError == nil {
			firstError = err
		} else {
			mlcp.recordWrite(i + 1)
		}
	}

	return firstError
}

// Remove 从所有缓存级别移除数据
func (mlcp *MultiLevelCacheProvider) Remove(ctx context.Context, key string) error {
	mlcp.mu.RLock()
	levels := mlcp.levels
	mlcp.mu.RUnlock()

	var errors []error
	for _, level := range levels {
		if err := level.Provider.Remove(ctx, key); err != nil {
			errors = append(errors, fmt.Errorf("level %s: %w", level.Name, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors removing from cache levels: %v", errors)
	}

	return nil
}

// RemoveByPattern 从所有缓存级别移除匹配模式的数据
func (mlcp *MultiLevelCacheProvider) RemoveByPattern(ctx context.Context, pattern string) error {
	mlcp.mu.RLock()
	levels := mlcp.levels
	mlcp.mu.RUnlock()

	var errors []error
	for _, level := range levels {
		if err := level.Provider.RemoveByPattern(ctx, pattern); err != nil {
			errors = append(errors, fmt.Errorf("level %s: %w", level.Name, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors removing pattern from cache levels: %v", errors)
	}

	return nil
}

// Exists 检查键是否存在于任何缓存级别
func (mlcp *MultiLevelCacheProvider) Exists(ctx context.Context, key string) (bool, error) {
	mlcp.mu.RLock()
	levels := mlcp.levels
	mlcp.mu.RUnlock()

	for _, level := range levels {
		exists, err := level.Provider.Exists(ctx, key)
		if err != nil {
			continue
		}
		if exists {
			return true, nil
		}
	}

	return false, nil
}

// Clear 清空所有缓存级别
func (mlcp *MultiLevelCacheProvider) Clear(ctx context.Context) error {
	mlcp.mu.RLock()
	levels := mlcp.levels
	mlcp.mu.RUnlock()

	var errors []error
	for _, level := range levels {
		if err := level.Provider.Clear(ctx); err != nil {
			errors = append(errors, fmt.Errorf("level %s: %w", level.Name, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors clearing cache levels: %v", errors)
	}

	return nil
}

// Name returns the provider name
func (mlcp *MultiLevelCacheProvider) Name() string {
	return "MultiLevel"
}

// Close 关闭所有缓存级别和工作协程
func (mlcp *MultiLevelCacheProvider) Close() error {
	// 停止异步工作协程
	close(mlcp.stopChan)
	mlcp.wg.Wait()
	close(mlcp.asyncWriteQueue)

	mlcp.mu.Lock()
	defer mlcp.mu.Unlock()

	var errors []error
	for _, level := range mlcp.levels {
		if err := level.Provider.Close(); err != nil {
			errors = append(errors, fmt.Errorf("level %s: %w", level.Name, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors closing cache levels: %v", errors)
	}

	return nil
}

// GetMetrics 返回缓存性能指标
func (mlcp *MultiLevelCacheProvider) GetMetrics() *MultiLevelMetrics {
	mlcp.metrics.mu.RLock()
	defer mlcp.metrics.mu.RUnlock()

	// 返回副本
	return &MultiLevelMetrics{
		L1Hits:         mlcp.metrics.L1Hits,
		L2Hits:         mlcp.metrics.L2Hits,
		L3Hits:         mlcp.metrics.L3Hits,
		Misses:         mlcp.metrics.Misses,
		TotalRequests:  mlcp.metrics.TotalRequests,
		L1WriteCount:   mlcp.metrics.L1WriteCount,
		L2WriteCount:   mlcp.metrics.L2WriteCount,
		L3WriteCount:   mlcp.metrics.L3WriteCount,
		PromotionCount: mlcp.metrics.PromotionCount,
	}
}

// PrintMetrics 打印缓存指标（用于调试）
func (mlcp *MultiLevelCacheProvider) PrintMetrics() {
	metrics := mlcp.GetMetrics()
	total := metrics.TotalRequests
	if total == 0 {
		fmt.Println("No cache requests yet")
		return
	}

	l1HitRate := float64(metrics.L1Hits) / float64(total) * 100
	l2HitRate := float64(metrics.L2Hits) / float64(total) * 100
	l3HitRate := float64(metrics.L3Hits) / float64(total) * 100
	missRate := float64(metrics.Misses) / float64(total) * 100

	fmt.Printf("=== Multi-Level Cache Metrics ===\n")
	fmt.Printf("Total Requests: %d\n", total)
	fmt.Printf("L1 Hits: %d (%.2f%%)\n", metrics.L1Hits, l1HitRate)
	fmt.Printf("L2 Hits: %d (%.2f%%)\n", metrics.L2Hits, l2HitRate)
	fmt.Printf("L3 Hits: %d (%.2f%%)\n", metrics.L3Hits, l3HitRate)
	fmt.Printf("Misses: %d (%.2f%%)\n", metrics.Misses, missRate)
	fmt.Printf("L1 Writes: %d\n", metrics.L1WriteCount)
	fmt.Printf("L2 Writes: %d\n", metrics.L2WriteCount)
	fmt.Printf("L3 Writes: %d\n", metrics.L3WriteCount)
	fmt.Printf("Promotions: %d\n", metrics.PromotionCount)
	fmt.Printf("================================\n")
}

// Helper methods

// recordHit 记录指定级别的缓存命中
func (mlcp *MultiLevelCacheProvider) recordHit(level int) {
	if !mlcp.options.EnableMetrics {
		return
	}

	mlcp.metrics.mu.Lock()
	defer mlcp.metrics.mu.Unlock()

	switch level {
	case 1:
		mlcp.metrics.L1Hits++
	case 2:
		mlcp.metrics.L2Hits++
	case 3:
		mlcp.metrics.L3Hits++
	}
}

// recordWrite 记录指定级别的缓存写入
func (mlcp *MultiLevelCacheProvider) recordWrite(level int) {
	if !mlcp.options.EnableMetrics {
		return
	}

	mlcp.metrics.mu.Lock()
	defer mlcp.metrics.mu.Unlock()

	switch level {
	case 1:
		mlcp.metrics.L1WriteCount++
	case 2:
		mlcp.metrics.L2WriteCount++
	case 3:
		mlcp.metrics.L3WriteCount++
	}
}

// promoteToHigherLevels 将数据写入优先级更高的缓存级别
func (mlcp *MultiLevelCacheProvider) promoteToHigherLevels(ctx context.Context, key string, data []byte, currentLevel int) {
	if mlcp.options.EnableMetrics {
		mlcp.metrics.mu.Lock()
		mlcp.metrics.PromotionCount++
		mlcp.metrics.mu.Unlock()
	}

	// 提升到所有更高级别
	for i := 0; i < currentLevel; i++ {
		ttl := mlcp.getTTLForLevel(i+1, 0)
		if mlcp.options.EnableAsyncWrite {
			// 异步提升
			select {
			case mlcp.asyncWriteQueue <- writeTask{
				ctx:        context.Background(),
				key:        key,
				value:      data,
				expiration: ttl,
				levels:     []int{i},
			}:
			default:
				// 队列已满，跳过提升
			}
		} else {
			// 同步提升
			mlcp.levels[i].Provider.SetRaw(ctx, key, data, ttl)
			mlcp.recordWrite(i + 1)
		}
	}
}

// writeToAllLevels 将数据写入所有缓存级别
func (mlcp *MultiLevelCacheProvider) writeToAllLevels(ctx context.Context, key string, data []byte) {
	mlcp.mu.RLock()
	levels := mlcp.levels
	mlcp.mu.RUnlock()

	for i, level := range levels {
		ttl := mlcp.getTTLForLevel(i+1, 0)
		if err := level.Provider.SetRaw(ctx, key, data, ttl); err == nil {
			mlcp.recordWrite(i + 1)
		}
	}
}

// writeSyncToLevels 将数据同步写入指定级别
func (mlcp *MultiLevelCacheProvider) writeSyncToLevels(ctx context.Context, key string, value []byte, expiration time.Duration, levelIndices []int) {
	mlcp.mu.RLock()
	levels := mlcp.levels
	mlcp.mu.RUnlock()

	for _, idx := range levelIndices {
		if idx < len(levels) {
			ttl := mlcp.getTTLForLevel(idx+1, expiration)
			if err := levels[idx].Provider.SetRaw(ctx, key, value, ttl); err == nil {
				mlcp.recordWrite(idx + 1)
			}
		}
	}
}

// getTTLForLevel 返回缓存级别的适当 TTL
func (mlcp *MultiLevelCacheProvider) getTTLForLevel(level int, defaultTTL time.Duration) time.Duration {
	if defaultTTL > 0 {
		return defaultTTL
	}

	switch level {
	case 1:
		return mlcp.options.L1TTL
	case 2:
		return mlcp.options.L2TTL
	default:
		return time.Hour * 24 // L3+ 的默认值
	}
}

// startAsyncWriteWorkers 启动用于异步写入的工作协程
func (mlcp *MultiLevelCacheProvider) startAsyncWriteWorkers(numWorkers int) {
	for i := 0; i < numWorkers; i++ {
		mlcp.wg.Add(1)
		go mlcp.asyncWriteWorker()
	}
}

// asyncWriteWorker 处理异步写入任务
func (mlcp *MultiLevelCacheProvider) asyncWriteWorker() {
	defer mlcp.wg.Done()

	for {
		select {
		case task, ok := <-mlcp.asyncWriteQueue:
			if !ok {
				return
			}
			mlcp.processWriteTask(task)
		case <-mlcp.stopChan:
			return
		}
	}
}

// processWriteTask 处理单个写入任务
func (mlcp *MultiLevelCacheProvider) processWriteTask(task writeTask) {
	mlcp.mu.RLock()
	levels := mlcp.levels
	mlcp.mu.RUnlock()

	for _, levelIdx := range task.levels {
		if levelIdx < len(levels) {
			ttl := mlcp.getTTLForLevel(levelIdx+1, task.expiration)
			if err := levels[levelIdx].Provider.SetRaw(task.ctx, task.key, task.value, ttl); err == nil {
				mlcp.recordWrite(levelIdx + 1)
			}
		}
	}
}
