package redis

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"abc-runner/app/core/interfaces"
	"abc-runner/app/core/metrics"
)

// RedisMetrics Redis协议特定指标
type RedisMetrics struct {
	// 操作统计
	Operations map[string]*RedisOperationStat `json:"operations"`

	// 连接统计
	Connection *RedisConnectionStat `json:"connection"`

	// 性能指标
	Performance *RedisPerformanceStat `json:"performance"`

	// 错误统计
	Errors map[string]*RedisErrorStat `json:"errors"`

	// 时间统计
	Timing *RedisTimingStat `json:"timing"`
}

// RedisOperationStat Redis操作统计
type RedisOperationStat struct {
	Count        int64         `json:"count"`
	SuccessCount int64         `json:"success_count"`
	FailureCount int64         `json:"failure_count"`
	TotalLatency time.Duration `json:"total_latency"`
	MinLatency   time.Duration `json:"min_latency"`
	MaxLatency   time.Duration `json:"max_latency"`
	AvgLatency   time.Duration `json:"avg_latency"`
	P95Latency   time.Duration `json:"p95_latency"`
	Throughput   float64       `json:"throughput"` // ops/sec
}

// RedisConnectionStat Redis连接统计
type RedisConnectionStat struct {
	ActiveConnections  int32   `json:"active_connections"`
	TotalConnections   int64   `json:"total_connections"`
	FailedConnections  int64   `json:"failed_connections"`
	ConnectionTimeouts int64   `json:"connection_timeouts"`
	AvgConnectionTime  time.Duration `json:"avg_connection_time"`
	PoolUtilization    float64 `json:"pool_utilization"`
}

// RedisPerformanceStat Redis性能统计
type RedisPerformanceStat struct {
	QPS             float64       `json:"qps"`               // 每秒查询数
	ReadQPS         float64       `json:"read_qps"`          // 每秒读查询数
	WriteQPS        float64       `json:"write_qps"`         // 每秒写查询数
	AvgResponseTime time.Duration `json:"avg_response_time"` // 平均响应时间
	HitRate         float64       `json:"hit_rate"`          // 缓存命中率
	BytesRead       int64         `json:"bytes_read"`        // 读取字节数
	BytesWritten    int64         `json:"bytes_written"`     // 写入字节数
}

// RedisErrorStat Redis错误统计
type RedisErrorStat struct {
	Count       int64     `json:"count"`
	LastOccured time.Time `json:"last_occured"`
	ErrorType   string    `json:"error_type"`
	Impact      string    `json:"impact"` // "low", "medium", "high"
}

// RedisTimingStat Redis时间统计
type RedisTimingStat struct {
	FirstOperation time.Time     `json:"first_operation"`
	LastOperation  time.Time     `json:"last_operation"`
	TotalDuration  time.Duration `json:"total_duration"`
	ActiveTime     time.Duration `json:"active_time"`
	IdleTime       time.Duration `json:"idle_time"`
}

// RedisCollector Redis指标收集器
type RedisCollector struct {
	*metrics.BaseCollector[RedisMetrics]
	
	// Redis特定指标
	redisMetrics    *RedisMetrics
	mutex           sync.RWMutex
	
	// 操作追踪
	operationTracker *RedisOperationTracker
	
	// 连接追踪
	connectionTracker *RedisConnectionTracker
	
	// 性能追踪
	performanceTracker *RedisPerformanceTracker
	
	// 错误追踪
	errorTracker *RedisErrorTracker
	
	// 配置
	config *metrics.MetricsConfig
}

// NewRedisCollector 创建Redis指标收集器
func NewRedisCollector(config *metrics.MetricsConfig) *RedisCollector {
	if config == nil {
		config = metrics.DefaultMetricsConfig()
	}

	// 初始化Redis指标
	redisMetrics := &RedisMetrics{
		Operations: make(map[string]*RedisOperationStat),
		Connection: &RedisConnectionStat{},
		Performance: &RedisPerformanceStat{},
		Errors: make(map[string]*RedisErrorStat),
		Timing: &RedisTimingStat{
			FirstOperation: time.Now(),
		},
	}

	// 创建基础收集器
	baseCollector := metrics.NewBaseCollector(config, *redisMetrics)

	collector := &RedisCollector{
		BaseCollector:      baseCollector,
		redisMetrics:       redisMetrics,
		operationTracker:   NewRedisOperationTracker(),
		connectionTracker:  NewRedisConnectionTracker(),
		performanceTracker: NewRedisPerformanceTracker(),
		errorTracker:       NewRedisErrorTracker(),
		config:            config,
	}

	return collector
}

// Record 记录操作结果（覆盖基础实现以添加Redis特定逻辑）
func (rc *RedisCollector) Record(result *interfaces.OperationResult) {
	// 调用基础记录方法
	rc.BaseCollector.Record(result)

	// Redis特定记录
	rc.recordRedisOperation(result)
}

// recordRedisOperation 记录Redis特定操作
func (rc *RedisCollector) recordRedisOperation(result *interfaces.OperationResult) {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()

	// 更新时间统计
	rc.redisMetrics.Timing.LastOperation = time.Now()
	if rc.redisMetrics.Timing.FirstOperation.IsZero() {
		rc.redisMetrics.Timing.FirstOperation = time.Now()
	}

	// 记录操作类型统计
	operationType := rc.getOperationType(result)
	rc.operationTracker.Record(operationType, result)

	// 记录性能指标
	rc.performanceTracker.Record(result)

	// 记录错误
	if !result.Success && result.Error != nil {
		rc.errorTracker.Record(result.Error)
	}

	// 更新Redis特定指标
	rc.updateRedisMetrics()
}

// getOperationType 获取操作类型
func (rc *RedisCollector) getOperationType(result *interfaces.OperationResult) string {
	if result.Metadata != nil {
		if opType, exists := result.Metadata["operation_type"]; exists {
			if opTypeStr, ok := opType.(string); ok {
				return opTypeStr
			}
		}
	}
	
	// 根据操作基础信息推断类型
	if result.IsRead {
		return "read"
	}
	return "write"
}

// updateRedisMetrics 更新Redis指标
func (rc *RedisCollector) updateRedisMetrics() {
	// 更新操作统计
	rc.redisMetrics.Operations = rc.operationTracker.GetOperationStats()

	// 更新连接统计
	rc.redisMetrics.Connection = rc.connectionTracker.GetConnectionStats()

	// 更新性能统计
	rc.redisMetrics.Performance = rc.performanceTracker.GetPerformanceStats()

	// 更新错误统计
	rc.redisMetrics.Errors = rc.errorTracker.GetErrorStats()

	// 更新时间统计
	now := time.Now()
	rc.redisMetrics.Timing.TotalDuration = now.Sub(rc.redisMetrics.Timing.FirstOperation)

	// 更新基础收集器的协议数据
	rc.UpdateProtocolMetrics(*rc.redisMetrics)
}

// GetRedisMetrics 获取Redis特定指标
func (rc *RedisCollector) GetRedisMetrics() *RedisMetrics {
	rc.mutex.RLock()
	defer rc.mutex.RUnlock()

	// 创建副本以避免并发修改
	metricsCopy := &RedisMetrics{
		Operations:  make(map[string]*RedisOperationStat),
		Connection:  &RedisConnectionStat{},
		Performance: &RedisPerformanceStat{},
		Errors:      make(map[string]*RedisErrorStat),
		Timing:      &RedisTimingStat{},
	}

	// 深拷贝操作统计
	for opType, stat := range rc.redisMetrics.Operations {
		metricsCopy.Operations[opType] = &RedisOperationStat{
			Count:        stat.Count,
			SuccessCount: stat.SuccessCount,
			FailureCount: stat.FailureCount,
			TotalLatency: stat.TotalLatency,
			MinLatency:   stat.MinLatency,
			MaxLatency:   stat.MaxLatency,
			AvgLatency:   stat.AvgLatency,
			P95Latency:   stat.P95Latency,
			Throughput:   stat.Throughput,
		}
	}

	// 复制其他指标
	*metricsCopy.Connection = *rc.redisMetrics.Connection
	*metricsCopy.Performance = *rc.redisMetrics.Performance
	*metricsCopy.Timing = *rc.redisMetrics.Timing

	// 深拷贝错误统计
	for errType, errStat := range rc.redisMetrics.Errors {
		metricsCopy.Errors[errType] = &RedisErrorStat{
			Count:       errStat.Count,
			LastOccured: errStat.LastOccured,
			ErrorType:   errStat.ErrorType,
			Impact:      errStat.Impact,
		}
	}

	return metricsCopy
}

// RecordConnection 记录连接事件
func (rc *RedisCollector) RecordConnection(success bool, duration time.Duration) {
	rc.connectionTracker.RecordConnection(success, duration)
	rc.updateRedisMetrics()
}

// RecordCacheHit 记录缓存命中
func (rc *RedisCollector) RecordCacheHit(hit bool) {
	rc.performanceTracker.RecordCacheHit(hit)
	rc.updateRedisMetrics()
}

// RecordBytes 记录字节数
func (rc *RedisCollector) RecordBytes(read, written int64) {
	rc.performanceTracker.RecordBytes(read, written)
	rc.updateRedisMetrics()
}

// Export 导出指标（实现interfaces.MetricsCollector接口）
func (rc *RedisCollector) Export() map[string]interface{} {
	// 获取基础指标
	baseMetrics := rc.BaseCollector.Snapshot()

	// 获取Redis特定指标
	redisMetrics := rc.GetRedisMetrics()

	// 合并指标
	result := make(map[string]interface{})

	// 添加基础指标
	result["core"] = baseMetrics.Core
	result["system"] = baseMetrics.System

	// 添加Redis特定指标
	result["redis"] = redisMetrics

	// 添加汇总信息
	result["protocol"] = "redis"
	result["timestamp"] = baseMetrics.Timestamp
	result["duration"] = baseMetrics.Core.Duration

	return result
}

// GetSummary 获取指标摘要
func (rc *RedisCollector) GetSummary() map[string]interface{} {
	snapshot := rc.Snapshot()
	redisMetrics := rc.GetRedisMetrics()

	return map[string]interface{}{
		"protocol":       "redis",
		"total_ops":      snapshot.Core.Operations.Total,
		"success_rate":   snapshot.Core.Operations.Rate,
		"avg_latency":    snapshot.Core.Latency.Average,
		"qps":           redisMetrics.Performance.QPS,
		"hit_rate":      redisMetrics.Performance.HitRate,
		"error_count":   len(redisMetrics.Errors),
		"duration":      snapshot.Core.Duration,
	}
}

// Reset 重置所有指标
func (rc *RedisCollector) Reset() {
	rc.BaseCollector.Reset()

	rc.mutex.Lock()
	defer rc.mutex.Unlock()

	// 重置Redis特定指标
	rc.redisMetrics = &RedisMetrics{
		Operations: make(map[string]*RedisOperationStat),
		Connection: &RedisConnectionStat{},
		Performance: &RedisPerformanceStat{},
		Errors: make(map[string]*RedisErrorStat),
		Timing: &RedisTimingStat{
			FirstOperation: time.Now(),
		},
	}

	// 重置追踪器
	rc.operationTracker.Reset()
	rc.connectionTracker.Reset()
	rc.performanceTracker.Reset()
	rc.errorTracker.Reset()
}

// MarshalJSON 自定义JSON序列化
func (rm *RedisMetrics) MarshalJSON() ([]byte, error) {
	type Alias RedisMetrics
	return json.Marshal(&struct {
		*Alias
		Timestamp time.Time `json:"timestamp"`
	}{
		Alias:     (*Alias)(rm),
		Timestamp: time.Now(),
	})
}

// String 返回字符串表示
func (rm *RedisMetrics) String() string {
	data, _ := json.MarshalIndent(rm, "", "  ")
	return string(data)
}

// Validate 验证指标数据的有效性
func (rm *RedisMetrics) Validate() error {
	if rm.Operations == nil {
		return fmt.Errorf("operations map is nil")
	}
	if rm.Connection == nil {
		return fmt.Errorf("connection stats is nil")
	}
	if rm.Performance == nil {
		return fmt.Errorf("performance stats is nil")
	}
	if rm.Errors == nil {
		return fmt.Errorf("errors map is nil")
	}
	if rm.Timing == nil {
		return fmt.Errorf("timing stats is nil")
	}

	// 验证操作统计的一致性
	for opType, stat := range rm.Operations {
		if stat.Count != stat.SuccessCount+stat.FailureCount {
			return fmt.Errorf("operation %s: count mismatch", opType)
		}
		if stat.Count > 0 && stat.TotalLatency > 0 {
			expectedAvg := stat.TotalLatency / time.Duration(stat.Count)
			if stat.AvgLatency != expectedAvg {
				return fmt.Errorf("operation %s: average latency mismatch", opType)
			}
		}
	}

	return nil
}

// GetTopOperations 获取最活跃的操作类型
func (rm *RedisMetrics) GetTopOperations(limit int) []*TopOperation {
	type opPair struct {
		Type string
		Stat *RedisOperationStat
	}

	var ops []opPair
	for opType, stat := range rm.Operations {
		ops = append(ops, opPair{Type: opType, Stat: stat})
	}

	// 按操作数量排序
	for i := 0; i < len(ops)-1; i++ {
		for j := i + 1; j < len(ops); j++ {
			if ops[i].Stat.Count < ops[j].Stat.Count {
				ops[i], ops[j] = ops[j], ops[i]
			}
		}
	}

	// 限制结果数量
	if limit > 0 && limit < len(ops) {
		ops = ops[:limit]
	}

	result := make([]*TopOperation, len(ops))
	for i, op := range ops {
		result[i] = &TopOperation{
			Type:       op.Type,
			Count:      op.Stat.Count,
			Throughput: op.Stat.Throughput,
			AvgLatency: op.Stat.AvgLatency,
		}
	}

	return result
}

// TopOperation 顶级操作统计
type TopOperation struct {
	Type       string        `json:"type"`
	Count      int64         `json:"count"`
	Throughput float64       `json:"throughput"`
	AvgLatency time.Duration `json:"avg_latency"`
}