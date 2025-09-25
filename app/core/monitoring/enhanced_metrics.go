package monitoring

import (
	"sort"
	"sync"
	"time"

	"abc-runner/app/core/interfaces"
)

// EnhancedMetricsCollector 增强的指标收集器（直接实现所有功能）
type EnhancedMetricsCollector struct {
	// 基础指标收集功能（原 DefaultMetricsCollector 功能）
	operations []interfaces.OperationResult
	startTime  time.Time
	totalOps   int64
	successOps int64
	failedOps  int64
	readOps    int64
	writeOps   int64
	durations  []time.Duration

	// 增强功能
	systemMonitor      *SystemMonitor
	performanceMonitor *PerformanceMonitor
	protocolMetrics    map[string]interface{}
	operationMetrics   map[string]*OperationTypeMetrics
	mutex              sync.RWMutex
}

// OperationTypeMetrics 操作类型指标
type OperationTypeMetrics struct {
	Count       int64         `json:"count"`
	SuccessRate float64       `json:"success_rate"`
	AvgLatency  time.Duration `json:"avg_latency"`
	MinLatency  time.Duration `json:"min_latency"`
	MaxLatency  time.Duration `json:"max_latency"`
	P95Latency  time.Duration `json:"p95_latency"`
	ErrorCount  int64         `json:"error_count"`
	TotalTime   time.Duration `json:"total_time"`
}

// NewEnhancedMetricsCollector 创建增强指标收集器
func NewEnhancedMetricsCollector() *EnhancedMetricsCollector {
	return &EnhancedMetricsCollector{
		// 初始化基础指标收集功能
		operations: make([]interfaces.OperationResult, 0),
		durations:  make([]time.Duration, 0),
		startTime:  time.Now(),

		// 初始化增强功能
		systemMonitor:      NewSystemMonitor(),
		performanceMonitor: NewPerformanceMonitor(200),
		protocolMetrics:    make(map[string]interface{}),
		operationMetrics:   make(map[string]*OperationTypeMetrics),
	}
}

// Start 启动监控
func (c *EnhancedMetricsCollector) Start() {
	c.systemMonitor.Start(time.Second)
	c.performanceMonitor.Start(time.Second)
}

// Stop 停止监控
func (c *EnhancedMetricsCollector) Stop() {
	c.systemMonitor.Stop()
	c.performanceMonitor.Stop()
}

// RecordOperation 记录操作结果（实现基础功能 + 增强功能）
func (c *EnhancedMetricsCollector) RecordOperation(result *interfaces.OperationResult) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 基础指标记录（原 DefaultMetricsCollector 功能）
	c.operations = append(c.operations, *result)
	c.durations = append(c.durations, result.Duration)
	c.totalOps++

	if result.Success {
		c.successOps++
	} else {
		c.failedOps++
	}

	if result.IsRead {
		c.readOps++
	} else {
		c.writeOps++
	}

	// 增强功能：记录操作类型指标
	c.recordOperationTypeMetrics(result)
}

// recordOperationTypeMetrics 记录操作类型指标
func (c *EnhancedMetricsCollector) recordOperationTypeMetrics(result *interfaces.OperationResult) {
	// 从结果元数据中获取操作类型
	operationType := "unknown"
	if result.Metadata != nil {
		if opType, ok := result.Metadata["operation_type"].(string); ok {
			operationType = opType
		}
	}

	// 获取或创建操作类型指标
	metrics, exists := c.operationMetrics[operationType]
	if !exists {
		metrics = &OperationTypeMetrics{
			MinLatency: result.Duration,
			MaxLatency: result.Duration,
		}
		c.operationMetrics[operationType] = metrics
	}

	// 更新指标
	metrics.Count++
	metrics.TotalTime += result.Duration

	if result.Success {
		metrics.SuccessRate = float64(metrics.Count-metrics.ErrorCount) / float64(metrics.Count) * 100
	} else {
		metrics.ErrorCount++
		metrics.SuccessRate = float64(metrics.Count-metrics.ErrorCount) / float64(metrics.Count) * 100
	}

	// 更新延迟统计
	if result.Duration < metrics.MinLatency {
		metrics.MinLatency = result.Duration
	}
	if result.Duration > metrics.MaxLatency {
		metrics.MaxLatency = result.Duration
	}

	metrics.AvgLatency = metrics.TotalTime / time.Duration(metrics.Count)
}

// GetMetrics 获取基础指标（实现 interfaces.MetricsCollector 接口）
func (c *EnhancedMetricsCollector) GetMetrics() *interfaces.Metrics {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if len(c.durations) == 0 {
		return &interfaces.Metrics{
			StartTime: c.startTime,
			EndTime:   time.Now(),
		}
	}

	// 计算统计数据
	durations := make([]time.Duration, len(c.durations))
	copy(durations, c.durations)

	return c.calculateMetrics(durations)
}

// calculateMetrics 计算指标
func (c *EnhancedMetricsCollector) calculateMetrics(durations []time.Duration) *interfaces.Metrics {
	if len(durations) == 0 {
		return &interfaces.Metrics{}
	}

	// 排序计算分位数
	sort.Slice(durations, func(i, j int) bool {
		return durations[i] < durations[j]
	})

	endTime := time.Now()
	totalDuration := endTime.Sub(c.startTime)

	metrics := &interfaces.Metrics{
		TotalOps:   c.totalOps,
		SuccessOps: c.successOps,
		FailedOps:  c.failedOps,
		ReadOps:    c.readOps,
		WriteOps:   c.writeOps,
		MinLatency: durations[0],
		MaxLatency: durations[len(durations)-1],
		StartTime:  c.startTime,
		EndTime:    endTime,
		Duration:   totalDuration,
	}

	// 计算平均延迟
	var totalLatency time.Duration
	for _, d := range durations {
		totalLatency += d
	}
	metrics.AvgLatency = totalLatency / time.Duration(len(durations))

	// 计算分位数
	metrics.P90Latency = durations[int(float64(len(durations))*0.9)]
	metrics.P95Latency = durations[int(float64(len(durations))*0.95)]
	metrics.P99Latency = durations[int(float64(len(durations))*0.99)]

	// 计算RPS
	if totalDuration.Seconds() > 0 {
		metrics.RPS = int32(float64(c.totalOps) / totalDuration.Seconds())
	}

	// 计算错误率
	if c.totalOps > 0 {
		metrics.ErrorRate = float64(c.failedOps) / float64(c.totalOps) * 100
	}

	return metrics
}

// Reset 重置指标（实现 interfaces.MetricsCollector 接口）
func (c *EnhancedMetricsCollector) Reset() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 重置基础指标
	c.operations = make([]interfaces.OperationResult, 0)
	c.durations = make([]time.Duration, 0)
	c.startTime = time.Now()
	c.totalOps = 0
	c.successOps = 0
	c.failedOps = 0
	c.readOps = 0
	c.writeOps = 0

	// 重置增强指标
	c.operationMetrics = make(map[string]*OperationTypeMetrics)
	c.protocolMetrics = make(map[string]interface{})
}

// RecordProtocolMetric 记录协议特定指标
func (c *EnhancedMetricsCollector) RecordProtocolMetric(key string, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.protocolMetrics[key] = value
}

// GetEnhancedMetrics 获取增强指标
func (c *EnhancedMetricsCollector) GetEnhancedMetrics() *EnhancedMetrics {
	basicMetrics := c.GetMetrics()

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	// 复制操作指标
	operationMetrics := make(map[string]*OperationTypeMetrics)
	for k, v := range c.operationMetrics {
		operationMetrics[k] = &OperationTypeMetrics{
			Count:       v.Count,
			SuccessRate: v.SuccessRate,
			AvgLatency:  v.AvgLatency,
			MinLatency:  v.MinLatency,
			MaxLatency:  v.MaxLatency,
			P95Latency:  v.P95Latency,
			ErrorCount:  v.ErrorCount,
			TotalTime:   v.TotalTime,
		}
	}

	// 复制协议指标
	protocolMetrics := make(map[string]interface{})
	for k, v := range c.protocolMetrics {
		protocolMetrics[k] = v
	}

	return &EnhancedMetrics{
		BasicMetrics:       basicMetrics,
		OperationMetrics:   operationMetrics,
		ProtocolMetrics:    protocolMetrics,
		SystemSnapshot:     c.systemMonitor.GetSystemSnapshot(),
		PerformanceSummary: c.performanceMonitor.GetSummary(),
		CollectionDuration: time.Since(c.startTime),
	}
}

// Export 导出指标（实现 interfaces.MetricsCollector 接口）
func (c *EnhancedMetricsCollector) Export() map[string]interface{} {
	basicMetrics := c.GetMetrics()
	enhancedMetrics := c.GetEnhancedMetrics()

	// 合并指标
	result := make(map[string]interface{})

	// 基础指标
	result["rps"] = basicMetrics.RPS
	result["total_ops"] = basicMetrics.TotalOps
	result["success_ops"] = basicMetrics.SuccessOps
	result["failed_ops"] = basicMetrics.FailedOps
	result["read_ops"] = basicMetrics.ReadOps
	result["write_ops"] = basicMetrics.WriteOps
	result["avg_latency"] = basicMetrics.AvgLatency.Nanoseconds()
	result["min_latency"] = basicMetrics.MinLatency.Nanoseconds()
	result["max_latency"] = basicMetrics.MaxLatency.Nanoseconds()
	result["p90_latency"] = basicMetrics.P90Latency.Nanoseconds()
	result["p95_latency"] = basicMetrics.P95Latency.Nanoseconds()
	result["p99_latency"] = basicMetrics.P99Latency.Nanoseconds()
	result["error_rate"] = basicMetrics.ErrorRate
	result["duration"] = basicMetrics.Duration.Nanoseconds()

	// 操作类型指标
	result["operation_metrics"] = enhancedMetrics.OperationMetrics

	// 协议指标
	result["protocol_metrics"] = enhancedMetrics.ProtocolMetrics

	// 系统指标
	result["system_snapshot"] = enhancedMetrics.SystemSnapshot

	// 性能摘要
	result["performance_summary"] = enhancedMetrics.PerformanceSummary

	// 收集时长
	result["collection_duration"] = enhancedMetrics.CollectionDuration.Nanoseconds()

	return result
}

// EnhancedMetrics 增强指标结构
type EnhancedMetrics struct {
	BasicMetrics       *interfaces.Metrics              `json:"basic_metrics"`
	OperationMetrics   map[string]*OperationTypeMetrics `json:"operation_metrics"`
	ProtocolMetrics    map[string]interface{}           `json:"protocol_metrics"`
	SystemSnapshot     *SystemSnapshot                  `json:"system_snapshot"`
	PerformanceSummary *PerformanceSummary              `json:"performance_summary"`
	CollectionDuration time.Duration                    `json:"collection_duration"`
}

// GetOperationSummary 获取操作摘要
func (c *EnhancedMetricsCollector) GetOperationSummary() map[string]interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	summary := make(map[string]interface{})

	var totalOps int64
	var totalErrors int64

	for opType, metrics := range c.operationMetrics {
		totalOps += metrics.Count
		totalErrors += metrics.ErrorCount

		summary[opType] = map[string]interface{}{
			"count":        metrics.Count,
			"success_rate": metrics.SuccessRate,
			"avg_latency":  metrics.AvgLatency.Nanoseconds(),
			"error_count":  metrics.ErrorCount,
		}
	}

	summary["total_operations"] = totalOps
	summary["total_errors"] = totalErrors
	if totalOps > 0 {
		summary["overall_success_rate"] = float64(totalOps-totalErrors) / float64(totalOps) * 100
	}

	return summary
}

// GetSystemHealth 获取系统健康状况
func (c *EnhancedMetricsCollector) GetSystemHealth() *SystemHealth {
	snapshot := c.systemMonitor.GetSystemSnapshot()
	summary := c.performanceMonitor.GetSummary()

	health := &SystemHealth{
		Status:    "healthy",
		Timestamp: time.Now(),
	}

	// 检查内存使用率
	memUsagePercent := float64(snapshot.Memory.Alloc) / float64(snapshot.Memory.Sys) * 100
	if memUsagePercent > 80 {
		health.Status = "warning"
		health.Issues = append(health.Issues, "High memory usage")
	}

	// 检查GC频率
	if summary.TotalGC > 100 {
		health.Status = "warning"
		health.Issues = append(health.Issues, "High GC frequency")
	}

	// 检查goroutine数量
	if snapshot.GoRoutines > 1000 {
		health.Status = "warning"
		health.Issues = append(health.Issues, "High goroutine count")
	}

	health.MemoryUsagePercent = memUsagePercent
	health.GoRoutineCount = snapshot.GoRoutines
	health.GCCount = snapshot.GC.NumGC

	return health
}

// SystemHealth 系统健康状况
type SystemHealth struct {
	Status             string    `json:"status"`
	Timestamp          time.Time `json:"timestamp"`
	MemoryUsagePercent float64   `json:"memory_usage_percent"`
	GoRoutineCount     int       `json:"goroutine_count"`
	GCCount            uint32    `json:"gc_count"`
	Issues             []string  `json:"issues"`
}
