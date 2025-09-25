package redis

import (
	"math"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"abc-runner/app/core/interfaces"
)

// RedisOperationTracker Redis操作追踪器
type RedisOperationTracker struct {
	operations map[string]*operationData
	mutex      sync.RWMutex
}

// operationData 操作数据
type operationData struct {
	count        int64
	successCount int64
	failureCount int64
	totalLatency int64 // nanoseconds
	minLatency   int64 // nanoseconds
	maxLatency   int64 // nanoseconds
	latencies    []time.Duration // 延迟历史
	startTime    time.Time
	lastUpdate   time.Time
}

// NewRedisOperationTracker 创建操作追踪器
func NewRedisOperationTracker() *RedisOperationTracker {
	return &RedisOperationTracker{
		operations: make(map[string]*operationData),
	}
}

// Record 记录操作
func (rot *RedisOperationTracker) Record(operationType string, result *interfaces.OperationResult) {
	rot.mutex.Lock()
	defer rot.mutex.Unlock()

	data, exists := rot.operations[operationType]
	if !exists {
		data = &operationData{
			minLatency: math.MaxInt64,
			startTime:  time.Now(),
		}
		rot.operations[operationType] = data
	}

	// 更新计数
	atomic.AddInt64(&data.count, 1)
	if result.Success {
		atomic.AddInt64(&data.successCount, 1)
	} else {
		atomic.AddInt64(&data.failureCount, 1)
	}

	// 更新延迟统计
	latencyNanos := result.Duration.Nanoseconds()
	atomic.AddInt64(&data.totalLatency, latencyNanos)

	// 更新最小延迟
	for {
		current := atomic.LoadInt64(&data.minLatency)
		if latencyNanos >= current || atomic.CompareAndSwapInt64(&data.minLatency, current, latencyNanos) {
			break
		}
	}

	// 更新最大延迟
	for {
		current := atomic.LoadInt64(&data.maxLatency)
		if latencyNanos <= current || atomic.CompareAndSwapInt64(&data.maxLatency, current, latencyNanos) {
			break
		}
	}

	// 添加到延迟历史（限制大小）
	data.latencies = append(data.latencies, result.Duration)
	if len(data.latencies) > 1000 { // 保留最近1000个样本
		data.latencies = data.latencies[1:]
	}

	data.lastUpdate = time.Now()
}

// GetOperationStats 获取操作统计
func (rot *RedisOperationTracker) GetOperationStats() map[string]*RedisOperationStat {
	rot.mutex.RLock()
	defer rot.mutex.RUnlock()

	stats := make(map[string]*RedisOperationStat)

	for opType, data := range rot.operations {
		count := atomic.LoadInt64(&data.count)
		successCount := atomic.LoadInt64(&data.successCount)
		failureCount := atomic.LoadInt64(&data.failureCount)
		totalLatency := atomic.LoadInt64(&data.totalLatency)
		minLatency := atomic.LoadInt64(&data.minLatency)
		maxLatency := atomic.LoadInt64(&data.maxLatency)

		stat := &RedisOperationStat{
			Count:        count,
			SuccessCount: successCount,
			FailureCount: failureCount,
			TotalLatency: time.Duration(totalLatency),
			MinLatency:   time.Duration(minLatency),
			MaxLatency:   time.Duration(maxLatency),
		}

		// 计算平均延迟
		if count > 0 {
			stat.AvgLatency = time.Duration(totalLatency / count)
		}

		// 计算P95延迟
		if len(data.latencies) > 0 {
			sortedLatencies := make([]time.Duration, len(data.latencies))
			copy(sortedLatencies, data.latencies)
			sort.Slice(sortedLatencies, func(i, j int) bool {
				return sortedLatencies[i] < sortedLatencies[j]
			})

			p95Index := int(float64(len(sortedLatencies)) * 0.95)
			if p95Index >= len(sortedLatencies) {
				p95Index = len(sortedLatencies) - 1
			}
			stat.P95Latency = sortedLatencies[p95Index]
		}

		// 计算吞吐量
		duration := data.lastUpdate.Sub(data.startTime)
		if duration > 0 {
			stat.Throughput = float64(count) / duration.Seconds()
		}

		stats[opType] = stat
	}

	return stats
}

// Reset 重置操作统计
func (rot *RedisOperationTracker) Reset() {
	rot.mutex.Lock()
	defer rot.mutex.Unlock()

	rot.operations = make(map[string]*operationData)
}

// RedisConnectionTracker Redis连接追踪器
type RedisConnectionTracker struct {
	activeConnections    int32
	totalConnections     int64
	failedConnections    int64
	connectionTimeouts   int64
	totalConnectionTime  int64 // nanoseconds
	connectionCount      int64
	poolSize            int32
	mutex               sync.RWMutex
}

// NewRedisConnectionTracker 创建连接追踪器
func NewRedisConnectionTracker() *RedisConnectionTracker {
	return &RedisConnectionTracker{}
}

// RecordConnection 记录连接事件
func (rct *RedisConnectionTracker) RecordConnection(success bool, duration time.Duration) {
	if success {
		atomic.AddInt32(&rct.activeConnections, 1)
		atomic.AddInt64(&rct.totalConnections, 1)
		atomic.AddInt64(&rct.totalConnectionTime, duration.Nanoseconds())
		atomic.AddInt64(&rct.connectionCount, 1)
	} else {
		atomic.AddInt64(&rct.failedConnections, 1)
	}
}

// RecordDisconnection 记录断开连接
func (rct *RedisConnectionTracker) RecordDisconnection() {
	atomic.AddInt32(&rct.activeConnections, -1)
}

// RecordTimeout 记录连接超时
func (rct *RedisConnectionTracker) RecordTimeout() {
	atomic.AddInt64(&rct.connectionTimeouts, 1)
}

// SetPoolSize 设置连接池大小
func (rct *RedisConnectionTracker) SetPoolSize(size int32) {
	atomic.StoreInt32(&rct.poolSize, size)
}

// GetConnectionStats 获取连接统计
func (rct *RedisConnectionTracker) GetConnectionStats() *RedisConnectionStat {
	activeConnections := atomic.LoadInt32(&rct.activeConnections)
	totalConnections := atomic.LoadInt64(&rct.totalConnections)
	failedConnections := atomic.LoadInt64(&rct.failedConnections)
	connectionTimeouts := atomic.LoadInt64(&rct.connectionTimeouts)
	totalConnectionTime := atomic.LoadInt64(&rct.totalConnectionTime)
	connectionCount := atomic.LoadInt64(&rct.connectionCount)
	poolSize := atomic.LoadInt32(&rct.poolSize)

	stat := &RedisConnectionStat{
		ActiveConnections:  activeConnections,
		TotalConnections:   totalConnections,
		FailedConnections:  failedConnections,
		ConnectionTimeouts: connectionTimeouts,
	}

	// 计算平均连接时间
	if connectionCount > 0 {
		stat.AvgConnectionTime = time.Duration(totalConnectionTime / connectionCount)
	}

	// 计算连接池利用率
	if poolSize > 0 {
		stat.PoolUtilization = float64(activeConnections) / float64(poolSize) * 100.0
	}

	return stat
}

// Reset 重置连接统计
func (rct *RedisConnectionTracker) Reset() {
	atomic.StoreInt32(&rct.activeConnections, 0)
	atomic.StoreInt64(&rct.totalConnections, 0)
	atomic.StoreInt64(&rct.failedConnections, 0)
	atomic.StoreInt64(&rct.connectionTimeouts, 0)
	atomic.StoreInt64(&rct.totalConnectionTime, 0)
	atomic.StoreInt64(&rct.connectionCount, 0)
}

// RedisPerformanceTracker Redis性能追踪器
type RedisPerformanceTracker struct {
	totalOps        int64
	readOps         int64
	writeOps        int64
	totalLatency    int64 // nanoseconds
	cacheHits       int64
	cacheMisses     int64
	bytesRead       int64
	bytesWritten    int64
	startTime       time.Time
	mutex           sync.RWMutex
}

// NewRedisPerformanceTracker 创建性能追踪器
func NewRedisPerformanceTracker() *RedisPerformanceTracker {
	return &RedisPerformanceTracker{
		startTime: time.Now(),
	}
}

// Record 记录操作性能
func (rpt *RedisPerformanceTracker) Record(result *interfaces.OperationResult) {
	atomic.AddInt64(&rpt.totalOps, 1)
	atomic.AddInt64(&rpt.totalLatency, result.Duration.Nanoseconds())

	if result.IsRead {
		atomic.AddInt64(&rpt.readOps, 1)
	} else {
		atomic.AddInt64(&rpt.writeOps, 1)
	}
}

// RecordCacheHit 记录缓存命中
func (rpt *RedisPerformanceTracker) RecordCacheHit(hit bool) {
	if hit {
		atomic.AddInt64(&rpt.cacheHits, 1)
	} else {
		atomic.AddInt64(&rpt.cacheMisses, 1)
	}
}

// RecordBytes 记录字节数
func (rpt *RedisPerformanceTracker) RecordBytes(read, written int64) {
	atomic.AddInt64(&rpt.bytesRead, read)
	atomic.AddInt64(&rpt.bytesWritten, written)
}

// GetPerformanceStats 获取性能统计
func (rpt *RedisPerformanceTracker) GetPerformanceStats() *RedisPerformanceStat {
	totalOps := atomic.LoadInt64(&rpt.totalOps)
	readOps := atomic.LoadInt64(&rpt.readOps)
	writeOps := atomic.LoadInt64(&rpt.writeOps)
	totalLatency := atomic.LoadInt64(&rpt.totalLatency)
	cacheHits := atomic.LoadInt64(&rpt.cacheHits)
	cacheMisses := atomic.LoadInt64(&rpt.cacheMisses)
	bytesRead := atomic.LoadInt64(&rpt.bytesRead)
	bytesWritten := atomic.LoadInt64(&rpt.bytesWritten)

	duration := time.Since(rpt.startTime)
	
	stat := &RedisPerformanceStat{
		BytesRead:    bytesRead,
		BytesWritten: bytesWritten,
	}

	// 计算QPS
	if duration > 0 {
		seconds := duration.Seconds()
		stat.QPS = float64(totalOps) / seconds
		stat.ReadQPS = float64(readOps) / seconds
		stat.WriteQPS = float64(writeOps) / seconds
	}

	// 计算平均响应时间
	if totalOps > 0 {
		stat.AvgResponseTime = time.Duration(totalLatency / totalOps)
	}

	// 计算缓存命中率
	totalCacheRequests := cacheHits + cacheMisses
	if totalCacheRequests > 0 {
		stat.HitRate = float64(cacheHits) / float64(totalCacheRequests) * 100.0
	}

	return stat
}

// Reset 重置性能统计
func (rpt *RedisPerformanceTracker) Reset() {
	atomic.StoreInt64(&rpt.totalOps, 0)
	atomic.StoreInt64(&rpt.readOps, 0)
	atomic.StoreInt64(&rpt.writeOps, 0)
	atomic.StoreInt64(&rpt.totalLatency, 0)
	atomic.StoreInt64(&rpt.cacheHits, 0)
	atomic.StoreInt64(&rpt.cacheMisses, 0)
	atomic.StoreInt64(&rpt.bytesRead, 0)
	atomic.StoreInt64(&rpt.bytesWritten, 0)
	rpt.startTime = time.Now()
}

// RedisErrorTracker Redis错误追踪器
type RedisErrorTracker struct {
	errors map[string]*errorData
	mutex  sync.RWMutex
}

// errorData 错误数据
type errorData struct {
	count       int64
	lastOccured time.Time
	errorType   string
	impact      string
}

// NewRedisErrorTracker 创建错误追踪器
func NewRedisErrorTracker() *RedisErrorTracker {
	return &RedisErrorTracker{
		errors: make(map[string]*errorData),
	}
}

// Record 记录错误
func (ret *RedisErrorTracker) Record(err error) {
	if err == nil {
		return
	}

	ret.mutex.Lock()
	defer ret.mutex.Unlock()

	errMsg := err.Error()
	data, exists := ret.errors[errMsg]
	if !exists {
		data = &errorData{
			errorType: ret.classifyError(err),
			impact:    ret.assessImpact(err),
		}
		ret.errors[errMsg] = data
	}

	atomic.AddInt64(&data.count, 1)
	data.lastOccured = time.Now()
}

// classifyError 分类错误类型
func (ret *RedisErrorTracker) classifyError(err error) string {
	errMsg := err.Error()
	
	switch {
	case contains(errMsg, "connection"):
		return "connection"
	case contains(errMsg, "timeout"):
		return "timeout"
	case contains(errMsg, "auth"):
		return "authentication"
	case contains(errMsg, "permission"):
		return "permission"
	case contains(errMsg, "protocol"):
		return "protocol"
	case contains(errMsg, "memory"):
		return "memory"
	default:
		return "unknown"
	}
}

// assessImpact 评估错误影响
func (ret *RedisErrorTracker) assessImpact(err error) string {
	errMsg := err.Error()
	
	switch {
	case contains(errMsg, "connection refused"), contains(errMsg, "network"):
		return "high"
	case contains(errMsg, "timeout"), contains(errMsg, "slow"):
		return "medium"
	case contains(errMsg, "auth"), contains(errMsg, "permission"):
		return "high"
	default:
		return "low"
	}
}

// contains 检查字符串是否包含子串（忽略大小写）
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		    (len(s) > len(substr) && 
		     func() bool {
		         for i := 0; i <= len(s)-len(substr); i++ {
		             if s[i:i+len(substr)] == substr {
		                 return true
		             }
		         }
		         return false
		     }()))
}

// GetErrorStats 获取错误统计
func (ret *RedisErrorTracker) GetErrorStats() map[string]*RedisErrorStat {
	ret.mutex.RLock()
	defer ret.mutex.RUnlock()

	stats := make(map[string]*RedisErrorStat)

	for errMsg, data := range ret.errors {
		count := atomic.LoadInt64(&data.count)
		
		stats[errMsg] = &RedisErrorStat{
			Count:       count,
			LastOccured: data.lastOccured,
			ErrorType:   data.errorType,
			Impact:      data.impact,
		}
	}

	return stats
}

// Reset 重置错误统计
func (ret *RedisErrorTracker) Reset() {
	ret.mutex.Lock()
	defer ret.mutex.Unlock()

	ret.errors = make(map[string]*errorData)
}