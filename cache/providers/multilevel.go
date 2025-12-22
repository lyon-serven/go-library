package providers

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// CacheLevel represents a single level in the multi-level cache
type CacheLevel struct {
	Name     string
	Provider ICacheProvider
	Priority int // Lower number = higher priority (L1=1, L2=2, L3=3)
}

// ICacheProvider interface (should match the one in cache package)
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

// MultiLevelCacheProvider implements a multi-level cache strategy
// L1: Local Memory Cache (fastest)
// L2: Redis Cache (shared across instances)
// L3: Database/Persistent Storage (slowest, with loader function)
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

// DatabaseLoader is a function type for loading data from database
type DatabaseLoader func(ctx context.Context, key string) ([]byte, error)

// MultiLevelCacheOptions defines options for multi-level cache
type MultiLevelCacheOptions struct {
	// EnableAsyncWrite enables asynchronous writes to lower levels
	EnableAsyncWrite bool
	// EnableAutoSync enables automatic synchronization between levels
	EnableAutoSync bool
	// SyncInterval is the interval for automatic synchronization
	SyncInterval time.Duration
	// WriteDownLevels controls how many levels down to write (0 = all levels)
	WriteDownLevels int
	// EnableMetrics enables performance metrics collection
	EnableMetrics bool
	// L1TTL is the TTL for L1 cache (memory)
	L1TTL time.Duration
	// L2TTL is the TTL for L2 cache (redis)
	L2TTL time.Duration
	// EnableWriteBack enables write-back strategy (write to L1 first, then async to others)
	EnableWriteBack bool
}

// DefaultMultiLevelCacheOptions returns default options
func DefaultMultiLevelCacheOptions() *MultiLevelCacheOptions {
	return &MultiLevelCacheOptions{
		EnableAsyncWrite: true,
		EnableAutoSync:   false,
		SyncInterval:     time.Minute * 5,
		WriteDownLevels:  0, // Write to all levels
		EnableMetrics:    true,
		L1TTL:            time.Minute * 5, // L1: 5 minutes
		L2TTL:            time.Hour * 1,   // L2: 1 hour
		EnableWriteBack:  true,
	}
}

// MultiLevelMetrics tracks cache performance metrics
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
	PromotionCount int64 // Count of data promoted to higher levels
}

// writeTask represents an async write task
type writeTask struct {
	ctx        context.Context
	key        string
	value      []byte
	expiration time.Duration
	levels     []int // Target levels to write to
}

// NewMultiLevelCacheProvider creates a new multi-level cache provider
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

	// Start async write workers if enabled
	if options.EnableAsyncWrite {
		mlcp.startAsyncWriteWorkers(3) // 3 worker goroutines
	}

	return mlcp
}

// AddLevel adds a cache level to the multi-level cache
func (mlcp *MultiLevelCacheProvider) AddLevel(name string, provider ICacheProvider, priority int) error {
	mlcp.mu.Lock()
	defer mlcp.mu.Unlock()

	// Check if level already exists
	for _, level := range mlcp.levels {
		if level.Priority == priority {
			return fmt.Errorf("cache level with priority %d already exists", priority)
		}
		if level.Name == name {
			return fmt.Errorf("cache level with name '%s' already exists", name)
		}
	}

	// Add new level
	mlcp.levels = append(mlcp.levels, CacheLevel{
		Name:     name,
		Provider: provider,
		Priority: priority,
	})

	// Sort levels by priority (ascending)
	mlcp.sortLevels()

	return nil
}

// sortLevels sorts cache levels by priority
func (mlcp *MultiLevelCacheProvider) sortLevels() {
	// Bubble sort (simple, sufficient for small number of levels)
	for i := 0; i < len(mlcp.levels); i++ {
		for j := i + 1; j < len(mlcp.levels); j++ {
			if mlcp.levels[i].Priority > mlcp.levels[j].Priority {
				mlcp.levels[i], mlcp.levels[j] = mlcp.levels[j], mlcp.levels[i]
			}
		}
	}
}

// GetRaw retrieves data from cache, trying each level in order
func (mlcp *MultiLevelCacheProvider) GetRaw(ctx context.Context, key string) ([]byte, error) {
	if mlcp.options.EnableMetrics {
		mlcp.metrics.mu.Lock()
		mlcp.metrics.TotalRequests++
		mlcp.metrics.mu.Unlock()
	}

	mlcp.mu.RLock()
	levels := mlcp.levels
	mlcp.mu.RUnlock()

	// Try each level in order (L1 -> L2 -> L3 -> DB)
	for i, level := range levels {
		data, err := level.Provider.GetRaw(ctx, key)
		if err == nil && data != nil {
			// Cache hit - update metrics
			mlcp.recordHit(i + 1)

			// Promote data to higher levels (cache warming)
			if i > 0 {
				mlcp.promoteToHigherLevels(ctx, key, data, i)
			}

			return data, nil
		}
	}

	// All cache levels missed, try database loader
	if mlcp.dbLoader != nil {
		data, err := mlcp.dbLoader(ctx, key)
		if err == nil && data != nil {
			// Record L3 (database) hit
			mlcp.recordHit(3)

			// Write back to all cache levels
			mlcp.writeToAllLevels(ctx, key, data)

			return data, nil
		}
	}

	// Complete miss
	if mlcp.options.EnableMetrics {
		mlcp.metrics.mu.Lock()
		mlcp.metrics.Misses++
		mlcp.metrics.mu.Unlock()
	}

	return nil, fmt.Errorf("cache miss: key '%s' not found in any level", key)
}

// SetRaw stores data in cache, writing to all configured levels
func (mlcp *MultiLevelCacheProvider) SetRaw(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	mlcp.mu.RLock()
	levels := mlcp.levels
	writeDownLevels := mlcp.options.WriteDownLevels
	mlcp.mu.RUnlock()

	if len(levels) == 0 {
		return fmt.Errorf("no cache levels configured")
	}

	// Determine which levels to write to
	targetLevels := len(levels)
	if writeDownLevels > 0 && writeDownLevels < targetLevels {
		targetLevels = writeDownLevels
	}

	// Write-back strategy: write to L1 immediately, others async
	if mlcp.options.EnableWriteBack && len(levels) > 0 {
		// Write to L1 immediately
		err := levels[0].Provider.SetRaw(ctx, key, value, mlcp.getTTLForLevel(1, expiration))
		if err != nil {
			return fmt.Errorf("failed to write to L1 cache: %w", err)
		}
		mlcp.recordWrite(1)

		// Write to other levels asynchronously
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
				// Queue is full, write synchronously as fallback
				mlcp.writeSyncToLevels(ctx, key, value, expiration, levelIndices)
			}
		}

		return nil
	}

	// Write-through strategy: write to all levels synchronously
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

// Remove removes data from all cache levels
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

// RemoveByPattern removes data matching pattern from all cache levels
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

// Exists checks if key exists in any cache level
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

// Clear clears all cache levels
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

// Close closes all cache levels and workers
func (mlcp *MultiLevelCacheProvider) Close() error {
	// Stop async workers
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

// GetMetrics returns cache performance metrics
func (mlcp *MultiLevelCacheProvider) GetMetrics() *MultiLevelMetrics {
	mlcp.metrics.mu.RLock()
	defer mlcp.metrics.mu.RUnlock()

	// Return a copy
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

// PrintMetrics prints cache metrics (for debugging)
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

// recordHit records a cache hit for the specified level
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

// recordWrite records a cache write for the specified level
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

// promoteToHigherLevels writes data to higher priority cache levels
func (mlcp *MultiLevelCacheProvider) promoteToHigherLevels(ctx context.Context, key string, data []byte, currentLevel int) {
	if mlcp.options.EnableMetrics {
		mlcp.metrics.mu.Lock()
		mlcp.metrics.PromotionCount++
		mlcp.metrics.mu.Unlock()
	}

	// Promote to all higher levels
	for i := 0; i < currentLevel; i++ {
		ttl := mlcp.getTTLForLevel(i+1, 0)
		if mlcp.options.EnableAsyncWrite {
			// Async promotion
			select {
			case mlcp.asyncWriteQueue <- writeTask{
				ctx:        context.Background(),
				key:        key,
				value:      data,
				expiration: ttl,
				levels:     []int{i},
			}:
			default:
				// Queue full, skip promotion
			}
		} else {
			// Sync promotion
			mlcp.levels[i].Provider.SetRaw(ctx, key, data, ttl)
			mlcp.recordWrite(i + 1)
		}
	}
}

// writeToAllLevels writes data to all cache levels
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

// writeSyncToLevels writes data synchronously to specified levels
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

// getTTLForLevel returns the appropriate TTL for a cache level
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
		return time.Hour * 24 // Default for L3+
	}
}

// startAsyncWriteWorkers starts worker goroutines for async writes
func (mlcp *MultiLevelCacheProvider) startAsyncWriteWorkers(numWorkers int) {
	for i := 0; i < numWorkers; i++ {
		mlcp.wg.Add(1)
		go mlcp.asyncWriteWorker()
	}
}

// asyncWriteWorker processes async write tasks
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

// processWriteTask processes a single write task
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
