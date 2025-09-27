package metrics

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"abc-runner/app/core/interfaces"
)

// RingBuffer 内存安全的环形缓冲区
type RingBuffer[T any] struct {
	buffer []T
	size   int
	head   int64
	tail   int64
	count  int64
	mutex  sync.RWMutex
}

// NewRingBuffer 创建环形缓冲区
func NewRingBuffer[T any](size int) *RingBuffer[T] {
	if size <= 0 {
		size = 1024
	}
	return &RingBuffer[T]{
		buffer: make([]T, size),
		size:   size,
	}
}

// Push 添加元素（线程安全）
func (rb *RingBuffer[T]) Push(item T) {
	head := atomic.LoadInt64(&rb.head)
	next := (head + 1) % int64(rb.size)
	
	rb.mutex.Lock()
	rb.buffer[head] = item
	atomic.StoreInt64(&rb.head, next)
	
	count := atomic.LoadInt64(&rb.count)
	if count < int64(rb.size) {
		atomic.AddInt64(&rb.count, 1)
	} else {
		// 缓冲区满时，移动tail指针
		atomic.StoreInt64(&rb.tail, (atomic.LoadInt64(&rb.tail)+1)%int64(rb.size))
	}
	rb.mutex.Unlock()
}

// ToSlice 转换为切片（创建副本，线程安全）
func (rb *RingBuffer[T]) ToSlice() []T {
	rb.mutex.RLock()
	defer rb.mutex.RUnlock()
	
	count := atomic.LoadInt64(&rb.count)
	if count == 0 {
		return []T{}
	}
	
	result := make([]T, count)
	tail := atomic.LoadInt64(&rb.tail)
	
	for i := int64(0); i < count; i++ {
		index := (tail + i) % int64(rb.size)
		result[i] = rb.buffer[index]
	}
	
	return result
}

// Size 获取当前元素数量
func (rb *RingBuffer[T]) Size() int {
	return int(atomic.LoadInt64(&rb.count))
}

// Capacity 获取缓冲区容量
func (rb *RingBuffer[T]) Capacity() int {
	return rb.size
}

// Clear 清空缓冲区
func (rb *RingBuffer[T]) Clear() {
	rb.mutex.Lock()
	defer rb.mutex.Unlock()
	
	atomic.StoreInt64(&rb.head, 0)
	atomic.StoreInt64(&rb.tail, 0)
	atomic.StoreInt64(&rb.count, 0)
	
	// 清零切片内容以帮助GC
	var zero T
	for i := range rb.buffer {
		rb.buffer[i] = zero
	}
}

// TimeWindow 时间窗口统计器
type TimeWindow struct {
	windowSize     time.Duration
	updateInterval time.Duration
	buckets        []int64
	bucketCount    int
	currentBucket  int64
	lastUpdate     time.Time
	mutex          sync.RWMutex
}

// NewTimeWindow 创建时间窗口
func NewTimeWindow(windowSize, updateInterval time.Duration) *TimeWindow {
	bucketCount := int(windowSize / updateInterval)
	if bucketCount <= 0 {
		bucketCount = 60 // 默认60个桶
	}
	
	return &TimeWindow{
		windowSize:     windowSize,
		updateInterval: updateInterval,
		buckets:        make([]int64, bucketCount),
		bucketCount:    bucketCount,
		lastUpdate:     time.Now(),
	}
}

// Record 记录事件
func (tw *TimeWindow) Record(count int64) {
	tw.updateBuckets()
	
	bucket := atomic.LoadInt64(&tw.currentBucket)
	atomic.AddInt64(&tw.buckets[bucket], count)
}

// GetRate 获取速率（每秒）
func (tw *TimeWindow) GetRate() float64 {
	tw.updateBuckets()
	
	tw.mutex.RLock()
	defer tw.mutex.RUnlock()
	
	var total int64
	for _, bucket := range tw.buckets {
		total += atomic.LoadInt64(&bucket)
	}
	
	return float64(total) / tw.windowSize.Seconds()
}

// Reset 重置时间窗口
func (tw *TimeWindow) Reset() {
	tw.mutex.Lock()
	defer tw.mutex.Unlock()
	
	for i := range tw.buckets {
		atomic.StoreInt64(&tw.buckets[i], 0)
	}
	atomic.StoreInt64(&tw.currentBucket, 0)
	tw.lastUpdate = time.Now()
}

// updateBuckets 更新桶位置
func (tw *TimeWindow) updateBuckets() {
	now := time.Now()
	elapsed := now.Sub(tw.lastUpdate)
	
	if elapsed < tw.updateInterval {
		return
	}
	
	tw.mutex.Lock()
	defer tw.mutex.Unlock()
	
	// 计算需要移动的桶数
	bucketsToMove := int(elapsed / tw.updateInterval)
	if bucketsToMove <= 0 {
		return
	}
	
	// 限制移动数量不超过总桶数
	if bucketsToMove >= tw.bucketCount {
		bucketsToMove = tw.bucketCount
		// 清空所有桶
		for i := range tw.buckets {
			atomic.StoreInt64(&tw.buckets[i], 0)
		}
	} else {
		// 清空过期的桶
		currentBucket := atomic.LoadInt64(&tw.currentBucket)
		for i := 0; i < bucketsToMove; i++ {
			nextBucket := (currentBucket + int64(i) + 1) % int64(tw.bucketCount)
			atomic.StoreInt64(&tw.buckets[nextBucket], 0)
		}
	}
	
	// 更新当前桶位置
	newBucket := (atomic.LoadInt64(&tw.currentBucket) + int64(bucketsToMove)) % int64(tw.bucketCount)
	atomic.StoreInt64(&tw.currentBucket, newBucket)
	
	tw.lastUpdate = now
}

// SystemTracker 系统监控追踪器
type SystemTracker struct {
	config         SystemConfig
	memStats       runtime.MemStats
	lastGCNum      uint32
	lastGCPause    time.Duration
	goroutineCount int
	peakGoroutines int
	cpuUsage       float64
	snapshots      *RingBuffer[SystemSnapshot]
	mutex          sync.RWMutex
}

// SystemSnapshot 系统快照
type SystemSnapshot struct {
	Timestamp      time.Time                  `json:"timestamp"`
	Memory         interfaces.MemoryMetrics   `json:"memory"`
	GC             interfaces.GCMetrics       `json:"gc"`
	GoroutineCount int                        `json:"goroutine"`
	CPU            interfaces.CPUMetrics      `json:"cpu"`
}

// NewSystemTracker 创建系统追踪器
func NewSystemTracker(config SystemConfig) *SystemTracker {
	st := &SystemTracker{
		config:    config,
		snapshots: NewRingBuffer[SystemSnapshot](config.SnapshotRetention),
	}
	
	// 初始化时进行一次更新，确保有初始数据
	st.Update()
	
	return st
}

// Update 更新系统指标
func (st *SystemTracker) Update() {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	
	// 更新内存统计
	runtime.ReadMemStats(&st.memStats)
	
	// 更新协程数量
	currentGoroutines := runtime.NumGoroutine()
	st.goroutineCount = currentGoroutines
	if currentGoroutines > st.peakGoroutines {
		st.peakGoroutines = currentGoroutines
	}
	
	// 计算CPU使用率（简化版本）
	st.cpuUsage = st.calculateCPUUsage()
	
	// 创建快照
	snapshot := SystemSnapshot{
		Timestamp: time.Now(),
		Memory:    st.getMemoryMetrics(),
		GC:        st.getGCMetrics(),
		GoroutineCount: st.goroutineCount,
		CPU: interfaces.CPUMetrics{
			UsagePercent: st.cpuUsage,
			Cores:        runtime.NumCPU(),
		},
	}
	
	st.snapshots.Push(snapshot)
}

// GetMetrics 获取当前系统指标
func (st *SystemTracker) GetMetrics() SystemMetrics {
	st.mutex.RLock()
	defer st.mutex.RUnlock()
	
	return SystemMetrics{
		MemoryUsage:    st.getMemoryMetrics(),
		GCStats:        st.getGCMetrics(),
		GoroutineCount: st.goroutineCount,
		CPUUsage: interfaces.CPUMetrics{
			UsagePercent: st.cpuUsage,
			Cores:        runtime.NumCPU(),
		},
	}
}

// Reset 重置系统统计
func (st *SystemTracker) Reset() {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	
	st.peakGoroutines = runtime.NumGoroutine()
	st.snapshots.Clear()
}

// GetSnapshots 获取历史快照
func (st *SystemTracker) GetSnapshots() []SystemSnapshot {
	return st.snapshots.ToSlice()
}

// getMemoryMetrics 获取内存指标
func (st *SystemTracker) getMemoryMetrics() interfaces.MemoryMetrics {
	memStats := &st.memStats
	
	return interfaces.MemoryMetrics{
		Allocated:  memStats.Alloc,
		InUse:      memStats.Alloc,  // 使用Alloc作为InUse的值
		TotalAlloc: memStats.TotalAlloc,
		Sys:        memStats.Sys,
		GCReleased: 0, // 可以根据需要计算
	}
}

// getGCMetrics 获取GC指标
func (st *SystemTracker) getGCMetrics() interfaces.GCMetrics {
	memStats := &st.memStats
	
	var avgPause time.Duration
	if memStats.NumGC > 0 {
		avgPause = time.Duration(memStats.PauseTotalNs / uint64(memStats.NumGC))
	}
	
	return interfaces.GCMetrics{
		LastGC:       time.Unix(0, int64(memStats.LastGC)),
		NumGC:        memStats.NumGC,
		TotalPause:   time.Duration(memStats.PauseTotalNs),
		AveragePause: avgPause,
	}
}

// calculateCPUUsage 计算CPU使用率（简化版本）
func (st *SystemTracker) calculateCPUUsage() float64 {
	// 基于协程数量的简单估算
	goroutines := float64(st.goroutineCount)
	cores := float64(runtime.NumCPU())
	
	// 简化的CPU使用率估算
	usage := (goroutines / cores) * 10.0 // 粗略估算
	if usage > 100.0 {
		usage = 100.0
	}
	
	return usage
}

// MemoryPool 内存池，用于减少内存分配
type MemoryPool[T any] struct {
	pool sync.Pool
}

// NewMemoryPool 创建内存池
func NewMemoryPool[T any](newFunc func() T) *MemoryPool[T] {
	return &MemoryPool[T]{
		pool: sync.Pool{
			New: func() interface{} {
				return newFunc()
			},
		},
	}
}

// Get 获取对象
func (mp *MemoryPool[T]) Get() T {
	return mp.pool.Get().(T)
}

// Put 归还对象
func (mp *MemoryPool[T]) Put(obj T) {
	mp.pool.Put(obj)
}

