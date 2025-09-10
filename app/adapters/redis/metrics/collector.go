package metrics

import (
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"abc-runner/app/adapters/redis/operations"
)

// MetricsCollector Redis指标收集器
type MetricsCollector struct {
	// 基础指标
	totalOperations    int64
	successOperations  int64
	failedOperations   int64
	readOperations     int64
	writeOperations    int64
	
	// 延迟指标
	totalLatency       time.Duration
	minLatency         time.Duration
	maxLatency         time.Duration
	latencyHistory     []time.Duration
	
	// 操作类型统计
	operationStats     map[operations.OperationType]*OperationStat
	
	// 连接指标
	connectionStats    *ConnectionStat
	
	// 时间窗口指标
	windowStats        *WindowStat
	
	// 错误统计
	errorStats         map[string]int64
	
	mutex              sync.RWMutex
	startTime          time.Time
}

// OperationStat 操作统计
type OperationStat struct {
	Count        int64         `json:"count"`
	SuccessCount int64         `json:"success_count"`
	FailureCount int64         `json:"failure_count"`
	TotalLatency time.Duration `json:"total_latency"`
	MinLatency   time.Duration `json:"min_latency"`
	MaxLatency   time.Duration `json:"max_latency"`
	AvgLatency   time.Duration `json:"avg_latency"`
}

// ConnectionStat 连接统计
type ConnectionStat struct {
	TotalConnections    int64         `json:"total_connections"`
	ActiveConnections   int64         `json:"active_connections"`
	FailedConnections   int64         `json:"failed_connections"`
	ConnectionLatency   time.Duration `json:"connection_latency"`
	ReconnectCount      int64         `json:"reconnect_count"`
}

// WindowStat 时间窗口统计
type WindowStat struct {
	WindowSize      time.Duration `json:"window_size"`
	CurrentWindow   int64         `json:"current_window"`
	WindowOperations []int64       `json:"window_operations"`
	RPS             float64       `json:"rps"`
	LastUpdate      time.Time     `json:"last_update"`
}

// MetricsSummary 指标摘要
type MetricsSummary struct {
	BasicMetrics      *BasicMetrics                             `json:"basic_metrics"`
	LatencyMetrics    *LatencyMetrics                           `json:"latency_metrics"`
	OperationMetrics  map[operations.OperationType]*OperationStat `json:"operation_metrics"`
	ConnectionMetrics *ConnectionStat                           `json:"connection_metrics"`
	WindowMetrics     *WindowStat                               `json:"window_metrics"`
	ErrorMetrics      map[string]int64                          `json:"error_metrics"`
	Duration          time.Duration                             `json:"duration"`
	Timestamp         time.Time                                 `json:"timestamp"`
}

// BasicMetrics 基础指标
type BasicMetrics struct {
	TotalOperations   int64   `json:"total_operations"`
	SuccessOperations int64   `json:"success_operations"`
	FailedOperations  int64   `json:"failed_operations"`
	ReadOperations    int64   `json:"read_operations"`
	WriteOperations   int64   `json:"write_operations"`
	SuccessRate       float64 `json:"success_rate"`
	ReadWriteRatio    float64 `json:"read_write_ratio"`
	RPS               float64 `json:"rps"`
}

// LatencyMetrics 延迟指标
type LatencyMetrics struct {
	MinLatency     time.Duration `json:"min_latency"`
	MaxLatency     time.Duration `json:"max_latency"`
	AvgLatency     time.Duration `json:"avg_latency"`
	P50Latency     time.Duration `json:"p50_latency"`
	P90Latency     time.Duration `json:"p90_latency"`
	P95Latency     time.Duration `json:"p95_latency"`
	P99Latency     time.Duration `json:"p99_latency"`
	TotalLatency   time.Duration `json:"total_latency"`
}

// NewMetricsCollector 创建指标收集器
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		operationStats:  make(map[operations.OperationType]*OperationStat),
		connectionStats: &ConnectionStat{},
		windowStats: &WindowStat{
			WindowSize:       time.Second,
			WindowOperations: make([]int64, 60), // 60秒窗口
			LastUpdate:       time.Now(),
		},
		errorStats:      make(map[string]int64),
		latencyHistory:  make([]time.Duration, 0),
		startTime:       time.Now(),
		minLatency:      time.Duration(^uint64(0) >> 1), // 最大值
		maxLatency:      0,
	}
}

// CollectOperation 收集操作指标
func (mc *MetricsCollector) CollectOperation(result operations.OperationResult) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	// 更新基础指标
	mc.totalOperations++
	if result.Success {
		mc.successOperations++
	} else {
		mc.failedOperations++
		// 收集错误信息
		if result.Error != nil {
			mc.errorStats[result.Error.Error()]++
		}
	}

	if result.IsRead {
		mc.readOperations++
	} else {
		mc.writeOperations++
	}

	// 更新延迟指标
	mc.totalLatency += result.Duration
	if result.Duration < mc.minLatency {
		mc.minLatency = result.Duration
	}
	if result.Duration > mc.maxLatency {
		mc.maxLatency = result.Duration
	}

	// 保存延迟历史（限制大小）
	mc.latencyHistory = append(mc.latencyHistory, result.Duration)
	if len(mc.latencyHistory) > 10000 { // 保留最近10000个样本
		mc.latencyHistory = mc.latencyHistory[1:]
	}

	// 更新操作类型统计
	opType := operations.OperationType("unknown")
	if result.ExtraData != nil {
		if ot, exists := result.ExtraData["operation_type"]; exists {
			if otStr, ok := ot.(string); ok {
				opType = operations.OperationType(otStr)
			}
		}
	}

	if _, exists := mc.operationStats[opType]; !exists {
		mc.operationStats[opType] = &OperationStat{
			MinLatency: time.Duration(^uint64(0) >> 1),
		}
	}

	stat := mc.operationStats[opType]
	stat.Count++
	stat.TotalLatency += result.Duration
	if result.Success {
		stat.SuccessCount++
	} else {
		stat.FailureCount++
	}

	if result.Duration < stat.MinLatency {
		stat.MinLatency = result.Duration
	}
	if result.Duration > stat.MaxLatency {
		stat.MaxLatency = result.Duration
	}

	// 计算平均延迟
	if stat.Count > 0 {
		stat.AvgLatency = stat.TotalLatency / time.Duration(stat.Count)
	}

	// 更新时间窗口统计
	mc.updateWindowStats()
}

// CollectConnection 收集连接指标
func (mc *MetricsCollector) CollectConnection(connectionType string, success bool, duration time.Duration) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	mc.connectionStats.TotalConnections++
	if success {
		mc.connectionStats.ActiveConnections++
	} else {
		mc.connectionStats.FailedConnections++
	}

	mc.connectionStats.ConnectionLatency += duration
}

// updateWindowStats 更新时间窗口统计
func (mc *MetricsCollector) updateWindowStats() {
	now := time.Now()
	timeSinceLastUpdate := now.Sub(mc.windowStats.LastUpdate)
	
	if timeSinceLastUpdate >= mc.windowStats.WindowSize {
		// 移动窗口
		windowsToMove := int(timeSinceLastUpdate / mc.windowStats.WindowSize)
		for i := 0; i < windowsToMove && i < len(mc.windowStats.WindowOperations); i++ {
			// 左移窗口
			copy(mc.windowStats.WindowOperations, mc.windowStats.WindowOperations[1:])
			mc.windowStats.WindowOperations[len(mc.windowStats.WindowOperations)-1] = 0
		}
		mc.windowStats.LastUpdate = now
	}

	// 更新当前窗口
	currentIndex := len(mc.windowStats.WindowOperations) - 1
	mc.windowStats.WindowOperations[currentIndex]++

	// 计算RPS
	totalOpsInWindow := int64(0)
	for _, ops := range mc.windowStats.WindowOperations {
		totalOpsInWindow += ops
	}
	windowDuration := time.Duration(len(mc.windowStats.WindowOperations)) * mc.windowStats.WindowSize
	mc.windowStats.RPS = float64(totalOpsInWindow) / windowDuration.Seconds()
}

// GetMetrics 获取指标
func (mc *MetricsCollector) GetMetrics() map[string]interface{} {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	duration := time.Since(mc.startTime)
	
	// 基础指标
	basicMetrics := &BasicMetrics{
		TotalOperations:   mc.totalOperations,
		SuccessOperations: mc.successOperations,
		FailedOperations:  mc.failedOperations,
		ReadOperations:    mc.readOperations,
		WriteOperations:   mc.writeOperations,
	}

	if mc.totalOperations > 0 {
		basicMetrics.SuccessRate = float64(mc.successOperations) / float64(mc.totalOperations) * 100
		basicMetrics.RPS = float64(mc.totalOperations) / duration.Seconds()
	}

	if mc.writeOperations > 0 {
		basicMetrics.ReadWriteRatio = float64(mc.readOperations) / float64(mc.writeOperations)
	}

	// 延迟指标
	latencyMetrics := mc.calculateLatencyMetrics()

	// 组装结果
	result := map[string]interface{}{
		"basic_metrics":      basicMetrics,
		"latency_metrics":    latencyMetrics,
		"operation_metrics":  mc.operationStats,
		"connection_metrics": mc.connectionStats,
		"window_metrics":     mc.windowStats,
		"error_metrics":      mc.errorStats,
		"duration":          duration,
		"timestamp":         time.Now(),
	}

	return result
}

// calculateLatencyMetrics 计算延迟指标
func (mc *MetricsCollector) calculateLatencyMetrics() *LatencyMetrics {
	latencyMetrics := &LatencyMetrics{
		MinLatency:   mc.minLatency,
		MaxLatency:   mc.maxLatency,
		TotalLatency: mc.totalLatency,
	}

	if mc.totalOperations > 0 {
		latencyMetrics.AvgLatency = mc.totalLatency / time.Duration(mc.totalOperations)
	}

	// 计算百分位数
	if len(mc.latencyHistory) > 0 {
		sortedLatencies := make([]time.Duration, len(mc.latencyHistory))
		copy(sortedLatencies, mc.latencyHistory)
		sort.Slice(sortedLatencies, func(i, j int) bool {
			return sortedLatencies[i] < sortedLatencies[j]
		})

		latencyMetrics.P50Latency = mc.getPercentile(sortedLatencies, 50)
		latencyMetrics.P90Latency = mc.getPercentile(sortedLatencies, 90)
		latencyMetrics.P95Latency = mc.getPercentile(sortedLatencies, 95)
		latencyMetrics.P99Latency = mc.getPercentile(sortedLatencies, 99)
	}

	return latencyMetrics
}

// getPercentile 获取百分位数
func (mc *MetricsCollector) getPercentile(sortedLatencies []time.Duration, percentile int) time.Duration {
	if len(sortedLatencies) == 0 {
		return 0
	}

	index := int(float64(len(sortedLatencies)) * float64(percentile) / 100.0)
	if index >= len(sortedLatencies) {
		index = len(sortedLatencies) - 1
	}
	if index < 0 {
		index = 0
	}

	return sortedLatencies[index]
}

// Reset 重置指标
func (mc *MetricsCollector) Reset() {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	mc.totalOperations = 0
	mc.successOperations = 0
	mc.failedOperations = 0
	mc.readOperations = 0
	mc.writeOperations = 0
	mc.totalLatency = 0
	mc.minLatency = time.Duration(^uint64(0) >> 1)
	mc.maxLatency = 0
	mc.latencyHistory = mc.latencyHistory[:0]
	
	// 重置操作统计
	for opType := range mc.operationStats {
		delete(mc.operationStats, opType)
	}
	
	// 重置错误统计
	for errorMsg := range mc.errorStats {
		delete(mc.errorStats, errorMsg)
	}
	
	// 重置连接统计
	mc.connectionStats = &ConnectionStat{}
	
	// 重置时间窗口
	mc.windowStats = &WindowStat{
		WindowSize:       time.Second,
		WindowOperations: make([]int64, 60),
		LastUpdate:       time.Now(),
	}
	
	mc.startTime = time.Now()
}

// GetSummary 获取指标摘要
func (mc *MetricsCollector) GetSummary() *MetricsSummary {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	duration := time.Since(mc.startTime)
	
	// 基础指标
	basicMetrics := &BasicMetrics{
		TotalOperations:   mc.totalOperations,
		SuccessOperations: mc.successOperations,
		FailedOperations:  mc.failedOperations,
		ReadOperations:    mc.readOperations,
		WriteOperations:   mc.writeOperations,
	}

	if mc.totalOperations > 0 {
		basicMetrics.SuccessRate = float64(mc.successOperations) / float64(mc.totalOperations) * 100
		basicMetrics.RPS = float64(mc.totalOperations) / duration.Seconds()
	}

	if mc.writeOperations > 0 {
		basicMetrics.ReadWriteRatio = float64(mc.readOperations) / float64(mc.writeOperations)
	}

	// 延迟指标
	latencyMetrics := mc.calculateLatencyMetrics()

	return &MetricsSummary{
		BasicMetrics:      basicMetrics,
		LatencyMetrics:    latencyMetrics,
		OperationMetrics:  mc.operationStats,
		ConnectionMetrics: mc.connectionStats,
		WindowMetrics:     mc.windowStats,
		ErrorMetrics:      mc.errorStats,
		Duration:          duration,
		Timestamp:         time.Now(),
	}
}

// ToJSON 转换为JSON
func (mc *MetricsCollector) ToJSON() ([]byte, error) {
	metrics := mc.GetMetrics()
	return json.MarshalIndent(metrics, "", "  ")
}

// PrintSummary 打印摘要
func (mc *MetricsCollector) PrintSummary() {
	summary := mc.GetSummary()
	
	fmt.Printf("\n=== Redis Performance Metrics Summary ===\n")
	fmt.Printf("Duration: %v\n", summary.Duration)
	fmt.Printf("Timestamp: %v\n\n", summary.Timestamp.Format("2006-01-02 15:04:05"))
	
	// 基础指标
	fmt.Printf("Basic Metrics:\n")
	fmt.Printf("  Total Operations: %d\n", summary.BasicMetrics.TotalOperations)
	fmt.Printf("  Success Operations: %d\n", summary.BasicMetrics.SuccessOperations)
	fmt.Printf("  Failed Operations: %d\n", summary.BasicMetrics.FailedOperations)
	fmt.Printf("  Read Operations: %d\n", summary.BasicMetrics.ReadOperations)
	fmt.Printf("  Write Operations: %d\n", summary.BasicMetrics.WriteOperations)
	fmt.Printf("  Success Rate: %.2f%%\n", summary.BasicMetrics.SuccessRate)
	fmt.Printf("  RPS: %.2f\n", summary.BasicMetrics.RPS)
	fmt.Printf("  Read/Write Ratio: %.2f\n\n", summary.BasicMetrics.ReadWriteRatio)
	
	// 延迟指标
	fmt.Printf("Latency Metrics:\n")
	fmt.Printf("  Min Latency: %v\n", summary.LatencyMetrics.MinLatency)
	fmt.Printf("  Max Latency: %v\n", summary.LatencyMetrics.MaxLatency)
	fmt.Printf("  Avg Latency: %v\n", summary.LatencyMetrics.AvgLatency)
	fmt.Printf("  P50 Latency: %v\n", summary.LatencyMetrics.P50Latency)
	fmt.Printf("  P90 Latency: %v\n", summary.LatencyMetrics.P90Latency)
	fmt.Printf("  P95 Latency: %v\n", summary.LatencyMetrics.P95Latency)
	fmt.Printf("  P99 Latency: %v\n\n", summary.LatencyMetrics.P99Latency)
	
	// 操作类型统计
	if len(summary.OperationMetrics) > 0 {
		fmt.Printf("Operation Metrics:\n")
		for opType, stat := range summary.OperationMetrics {
			fmt.Printf("  %s:\n", opType)
			fmt.Printf("    Count: %d\n", stat.Count)
			fmt.Printf("    Success: %d\n", stat.SuccessCount)
			fmt.Printf("    Failure: %d\n", stat.FailureCount)
			fmt.Printf("    Avg Latency: %v\n", stat.AvgLatency)
		}
		fmt.Printf("\n")
	}
	
	// 错误统计
	if len(summary.ErrorMetrics) > 0 {
		fmt.Printf("Error Metrics:\n")
		for errorMsg, count := range summary.ErrorMetrics {
			fmt.Printf("  %s: %d\n", errorMsg, count)
		}
		fmt.Printf("\n")
	}
	
	fmt.Printf("==========================================\n")
}