package metrics

import (
	"context"
	"math"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"abc-runner/app/core/interfaces"
)

// BaseCollector 基础指标收集器实现
type BaseCollector[T any] struct {
	// 配置
	config *MetricsConfig

	// 核心指标收集组件
	operations  *OperationTracker
	latency     *LatencyTracker
	throughput  *ThroughputTracker

	// 系统监控组件
	system *SystemTracker

	// 协议特定指标
	protocol T

	// 状态管理
	startTime   time.Time
	mutex       sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	isRunning   int32

	// 健康检查器
	healthChecker HealthChecker
}

// NewBaseCollector 创建基础收集器
func NewBaseCollector[T any](config *MetricsConfig, protocolData T) *BaseCollector[T] {
	if config == nil {
		config = DefaultMetricsConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	collector := &BaseCollector[T]{
		config:        config,
		operations:    NewOperationTracker(),
		latency:       NewLatencyTracker(config.Latency),
		throughput:    NewThroughputTracker(config.Throughput),
		system:        NewSystemTracker(config.System),
		protocol:      protocolData,
		startTime:     time.Now(),
		ctx:           ctx,
		cancel:        cancel,
		healthChecker: NewHealthChecker(config.System.HealthThresholds),
	}

	// 启动后台监控
	if config.System.Enabled {
		collector.startBackgroundMonitoring()
	}

	atomic.StoreInt32(&collector.isRunning, 1)
	return collector
}

// Record 记录操作结果
func (bc *BaseCollector[T]) Record(result *interfaces.OperationResult) {
	if atomic.LoadInt32(&bc.isRunning) == 0 {
		return
	}

	// 记录操作指标
	bc.operations.Record(result)

	// 记录延迟指标
	bc.latency.Record(result.Duration)

	// 更新吞吐量指标
	bc.throughput.Record(result)
}

// Snapshot 获取当前指标快照
func (bc *BaseCollector[T]) Snapshot() *MetricsSnapshot[T] {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	duration := time.Since(bc.startTime)

	return &MetricsSnapshot[T]{
		Core: CoreMetrics{
			Operations: bc.operations.GetMetrics(),
			Latency:    bc.latency.GetMetrics(),
			Throughput: bc.throughput.GetMetrics(duration),
			Duration:   duration,
		},
		Protocol:  bc.protocol,
		System:    bc.system.GetMetrics(),
		Timestamp: time.Now(),
	}
}

// Reset 重置所有指标
func (bc *BaseCollector[T]) Reset() {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	bc.operations.Reset()
	bc.latency.Reset()
	bc.throughput.Reset()
	bc.system.Reset()
	bc.startTime = time.Now()
}

// Stop 停止收集器
func (bc *BaseCollector[T]) Stop() {
	if atomic.CompareAndSwapInt32(&bc.isRunning, 1, 0) {
		bc.cancel()
	}
}

// UpdateProtocolMetrics 更新协议特定指标
func (bc *BaseCollector[T]) UpdateProtocolMetrics(protocolData T) {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()
	bc.protocol = protocolData
}

// GetHealthStatus 获取健康状态
func (bc *BaseCollector[T]) GetHealthStatus() *HealthCheckResult {
	systemMetrics := bc.system.GetMetrics()
	return bc.healthChecker.Check(bc.ctx, systemMetrics)
}

// startBackgroundMonitoring 启动后台监控
func (bc *BaseCollector[T]) startBackgroundMonitoring() {
	go func() {
		ticker := time.NewTicker(bc.config.System.MonitorInterval)
		defer ticker.Stop()

		for {
			select {
			case <-bc.ctx.Done():
				return
			case <-ticker.C:
				bc.system.Update()
			}
		}
	}()
}

// OperationTracker 操作追踪器
type OperationTracker struct {
	total   int64
	success int64
	failed  int64
	read    int64
	write   int64
	mutex   sync.RWMutex
}

// NewOperationTracker 创建操作追踪器
func NewOperationTracker() *OperationTracker {
	return &OperationTracker{}
}

// Record 记录操作
func (ot *OperationTracker) Record(result *interfaces.OperationResult) {
	atomic.AddInt64(&ot.total, 1)

	if result.Success {
		atomic.AddInt64(&ot.success, 1)
	} else {
		atomic.AddInt64(&ot.failed, 1)
	}

	if result.IsRead {
		atomic.AddInt64(&ot.read, 1)
	} else {
		atomic.AddInt64(&ot.write, 1)
	}
}

// GetMetrics 获取操作指标
func (ot *OperationTracker) GetMetrics() OperationMetrics {
	total := atomic.LoadInt64(&ot.total)
	success := atomic.LoadInt64(&ot.success)
	failed := atomic.LoadInt64(&ot.failed)
	read := atomic.LoadInt64(&ot.read)
	write := atomic.LoadInt64(&ot.write)

	var rate float64
	if total > 0 {
		rate = float64(success) / float64(total) * 100.0
	}

	return OperationMetrics{
		Total:   total,
		Success: success,
		Failed:  failed,
		Read:    read,
		Write:   write,
		Rate:    rate,
	}
}

// Reset 重置操作统计
func (ot *OperationTracker) Reset() {
	atomic.StoreInt64(&ot.total, 0)
	atomic.StoreInt64(&ot.success, 0)
	atomic.StoreInt64(&ot.failed, 0)
	atomic.StoreInt64(&ot.read, 0)
	atomic.StoreInt64(&ot.write, 0)
}

// LatencyTracker 延迟追踪器
type LatencyTracker struct {
	config      LatencyConfig
	buffer      *RingBuffer[time.Duration]
	min         int64 // nanoseconds
	max         int64 // nanoseconds
	total       int64 // nanoseconds
	count       int64
	lastCompute time.Time
	cached      LatencyMetrics
	mutex       sync.RWMutex
}

// NewLatencyTracker 创建延迟追踪器
func NewLatencyTracker(config LatencyConfig) *LatencyTracker {
	return &LatencyTracker{
		config:      config,
		buffer:      NewRingBuffer[time.Duration](config.HistorySize),
		min:         math.MaxInt64,
		max:         0,
		lastCompute: time.Now(),
	}
}

// Record 记录延迟
func (lt *LatencyTracker) Record(duration time.Duration) {
	// 采样检查
	if lt.config.SamplingRate < 1.0 {
		// 简单采样策略：基于随机数
		if time.Now().UnixNano()%1000 > int64(lt.config.SamplingRate*1000) {
			return
		}
	}

	nanos := duration.Nanoseconds()
	
	// 更新基础统计
	atomic.AddInt64(&lt.total, nanos)
	atomic.AddInt64(&lt.count, 1)

	// 更新最小值
	for {
		current := atomic.LoadInt64(&lt.min)
		if nanos >= current || atomic.CompareAndSwapInt64(&lt.min, current, nanos) {
			break
		}
	}

	// 更新最大值
	for {
		current := atomic.LoadInt64(&lt.max)
		if nanos <= current || atomic.CompareAndSwapInt64(&lt.max, current, nanos) {
			break
		}
	}

	// 添加到历史记录
	lt.buffer.Push(duration)
}

// GetMetrics 获取延迟指标
func (lt *LatencyTracker) GetMetrics() LatencyMetrics {
	// 检查是否有数据但缓存为空，强制计算
	count := atomic.LoadInt64(&lt.count)
	if count == 0 {
		return LatencyMetrics{}
	}
	
	// 检查是否需要重新计算或缓存为空
	lt.mutex.RLock()
	cachedIsEmpty := lt.cached.Average == 0 && lt.cached.Min == 0 && lt.cached.Max == 0
	needsRecompute := time.Since(lt.lastCompute) >= lt.config.ComputeInterval || cachedIsEmpty
	if !needsRecompute {
		cached := lt.cached
		lt.mutex.RUnlock()
		return cached
	}
	lt.mutex.RUnlock()

	lt.mutex.Lock()
	defer lt.mutex.Unlock()

	// 再次检查数据量（双重检查）
	if count == 0 {
		return LatencyMetrics{}
	}

	total := atomic.LoadInt64(&lt.total)
	min := atomic.LoadInt64(&lt.min)
	max := atomic.LoadInt64(&lt.max)

	// 修复：当min仍为初始值时，设置为0
	minDuration := time.Duration(min)
	if min == math.MaxInt64 {
		minDuration = 0
	}

	metrics := LatencyMetrics{
		Min:     minDuration,
		Max:     time.Duration(max),
		Average: time.Duration(total / count),
	}

	// 计算分位数
	if data := lt.buffer.ToSlice(); len(data) > 0 {
		percentiles := lt.calculatePercentiles(data)
		metrics.P50 = percentiles[50]
		metrics.P90 = percentiles[90]
		metrics.P95 = percentiles[95]
		metrics.P99 = percentiles[99]
		metrics.StdDeviation = lt.calculateStdDev(data, metrics.Average)
	}

	lt.cached = metrics
	lt.lastCompute = time.Now()
	return metrics
}

// Reset 重置延迟统计
func (lt *LatencyTracker) Reset() {
	atomic.StoreInt64(&lt.total, 0)
	atomic.StoreInt64(&lt.count, 0)
	atomic.StoreInt64(&lt.min, math.MaxInt64)
	atomic.StoreInt64(&lt.max, 0)
	lt.buffer.Clear()

	lt.mutex.Lock()
	lt.cached = LatencyMetrics{}
	lt.lastCompute = time.Now()
	lt.mutex.Unlock()
}

// calculatePercentiles 计算分位数
func (lt *LatencyTracker) calculatePercentiles(data []time.Duration) map[int]time.Duration {
	if len(data) == 0 {
		return make(map[int]time.Duration)
	}

	// 复制并排序数据
	sorted := make([]time.Duration, len(data))
	copy(sorted, data)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	percentiles := make(map[int]time.Duration)
	for _, p := range []int{50, 90, 95, 99} {
		index := int(float64(len(sorted)) * float64(p) / 100.0)
		if index >= len(sorted) {
			index = len(sorted) - 1
		}
		if index < 0 {
			index = 0
		}
		percentiles[p] = sorted[index]
	}

	return percentiles
}

// calculateStdDev 计算标准差
func (lt *LatencyTracker) calculateStdDev(data []time.Duration, mean time.Duration) time.Duration {
	if len(data) <= 1 {
		return 0
	}

	var sum float64
	meanNanos := float64(mean.Nanoseconds())

	for _, d := range data {
		diff := float64(d.Nanoseconds()) - meanNanos
		sum += diff * diff
	}

	variance := sum / float64(len(data)-1)
	return time.Duration(math.Sqrt(variance))
}

// ThroughputTracker 吞吐量追踪器
type ThroughputTracker struct {
	config     ThroughputConfig
	window     *TimeWindow
	readCount  int64
	writeCount int64
	mutex      sync.RWMutex
}

// NewThroughputTracker 创建吞吐量追踪器
func NewThroughputTracker(config ThroughputConfig) *ThroughputTracker {
	return &ThroughputTracker{
		config: config,
		window: NewTimeWindow(config.WindowSize, config.UpdateInterval),
	}
}

// Record 记录操作
func (tt *ThroughputTracker) Record(result *interfaces.OperationResult) {
	tt.window.Record(1)

	if result.IsRead {
		atomic.AddInt64(&tt.readCount, 1)
	} else {
		atomic.AddInt64(&tt.writeCount, 1)
	}
}

// GetMetrics 获取吞吐量指标
func (tt *ThroughputTracker) GetMetrics(duration time.Duration) ThroughputMetrics {
	readCount := atomic.LoadInt64(&tt.readCount)
	writeCount := atomic.LoadInt64(&tt.writeCount)
	total := readCount + writeCount

	var rps, readRPS, writeRPS float64
	if duration > 0 {
		seconds := duration.Seconds()
		rps = float64(total) / seconds
		readRPS = float64(readCount) / seconds
		writeRPS = float64(writeCount) / seconds
	}

	return ThroughputMetrics{
		RPS:      rps,
		ReadRPS:  readRPS,
		WriteRPS: writeRPS,
	}
}

// Reset 重置吞吐量统计
func (tt *ThroughputTracker) Reset() {
	atomic.StoreInt64(&tt.readCount, 0)
	atomic.StoreInt64(&tt.writeCount, 0)
	tt.window.Reset()
}

// DefaultMetricsConfig 返回默认配置
func DefaultMetricsConfig() *MetricsConfig {
	return &MetricsConfig{
		Latency: LatencyConfig{
			HistorySize:     10000,
			Percentiles:     []float64{0.5, 0.9, 0.95, 0.99},
			SamplingRate:    1.0,
			ComputeInterval: time.Second,
		},
		Throughput: ThroughputConfig{
			WindowSize:     60 * time.Second,
			UpdateInterval: time.Second,
		},
		System: SystemConfig{
			MonitorInterval:   time.Second,
			SnapshotRetention: 100,
			Enabled:           true,
			HealthThresholds: HealthThresholds{
				MemoryUsage:    80.0,
				GCFrequency:    100,
				GoroutineCount: 1000,
				CPUUsage:       80.0,
			},
		},
		Storage: StorageConfig{
			MemoryLimit:    100 * 1024 * 1024, // 100MB
			UseCompression: false,
			FlushInterval:  5 * time.Second,
		},
		Export: ExportConfig{
			Format:   []string{"json"},
			Interval: 10 * time.Second,
			Enabled:  false,
		},
	}
}