package monitoring

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"abc-runner/servers/pkg/interfaces"
)

// MetricsCollector 指标收集器实现
type MetricsCollector struct {
	// 请求指标
	totalRequests    int64
	successRequests  int64
	failedRequests   int64
	requestDurations []time.Duration
	durationMutex    sync.RWMutex

	// 连接指标
	totalConnections   int64
	activeConnections  int64
	closedConnections  int64

	// 错误指标
	errors     map[string]int64
	errorMutex sync.RWMutex

	// 协议指标
	protocolMetrics     map[string]map[string]int64
	protocolMetricMutex sync.RWMutex

	// 时间戳
	startTime time.Time
	mutex     sync.RWMutex
}

// NewMetricsCollector 创建指标收集器
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		requestDurations:    make([]time.Duration, 0, 10000), // 预分配空间
		errors:              make(map[string]int64),
		protocolMetrics:     make(map[string]map[string]int64),
		startTime:           time.Now(),
	}
}

// RecordRequest 记录请求
func (mc *MetricsCollector) RecordRequest(protocol string, operation string, duration time.Duration, success bool) {
	atomic.AddInt64(&mc.totalRequests, 1)
	
	if success {
		atomic.AddInt64(&mc.successRequests, 1)
	} else {
		atomic.AddInt64(&mc.failedRequests, 1)
	}
	
	// 记录持续时间
	mc.durationMutex.Lock()
	mc.requestDurations = append(mc.requestDurations, duration)
	
	// 限制存储的持续时间数量，避免内存泄漏
	if len(mc.requestDurations) > 10000 {
		// 移除最旧的一半记录
		copy(mc.requestDurations, mc.requestDurations[5000:])
		mc.requestDurations = mc.requestDurations[:5000]
	}
	mc.durationMutex.Unlock()
	
	// 记录协议特定指标
	mc.recordProtocolMetric(protocol, operation, 1)
}

// RecordConnection 记录连接
func (mc *MetricsCollector) RecordConnection(protocol string, action string) {
	atomic.AddInt64(&mc.totalConnections, 1)
	
	switch action {
	case "open":
		atomic.AddInt64(&mc.activeConnections, 1)
	case "close":
		atomic.AddInt64(&mc.activeConnections, -1)
		atomic.AddInt64(&mc.closedConnections, 1)
	}
	
	// 记录协议连接指标
	mc.recordProtocolMetric(protocol, "connections_"+action, 1)
}

// RecordError 记录错误
func (mc *MetricsCollector) RecordError(protocol string, operation string, errorType string) {
	key := protocol + ":" + operation + ":" + errorType
	
	mc.errorMutex.Lock()
	mc.errors[key]++
	mc.errorMutex.Unlock()
	
	// 记录协议错误指标
	mc.recordProtocolMetric(protocol, "errors_"+errorType, 1)
}

// recordProtocolMetric 记录协议特定指标
func (mc *MetricsCollector) recordProtocolMetric(protocol string, metric string, value int64) {
	mc.protocolMetricMutex.Lock()
	defer mc.protocolMetricMutex.Unlock()
	
	if mc.protocolMetrics[protocol] == nil {
		mc.protocolMetrics[protocol] = make(map[string]int64)
	}
	
	mc.protocolMetrics[protocol][metric] += value
}

// GetMetrics 获取指标快照
func (mc *MetricsCollector) GetMetrics() map[string]interface{} {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()
	
	// 基础指标
	metrics := map[string]interface{}{
		"total_requests":     atomic.LoadInt64(&mc.totalRequests),
		"success_requests":   atomic.LoadInt64(&mc.successRequests),
		"failed_requests":    atomic.LoadInt64(&mc.failedRequests),
		"total_connections":  atomic.LoadInt64(&mc.totalConnections),
		"active_connections": atomic.LoadInt64(&mc.activeConnections),
		"closed_connections": atomic.LoadInt64(&mc.closedConnections),
		"start_time":         mc.startTime,
		"uptime":             time.Since(mc.startTime).String(),
	}
	
	// 计算请求统计
	if requestStats := mc.calculateRequestStats(); requestStats != nil {
		for k, v := range requestStats {
			metrics[k] = v
		}
	}
	
	// 错误统计
	mc.errorMutex.RLock()
	if len(mc.errors) > 0 {
		errorMetrics := make(map[string]int64)
		for k, v := range mc.errors {
			errorMetrics[k] = v
		}
		metrics["errors"] = errorMetrics
	}
	mc.errorMutex.RUnlock()
	
	// 协议指标
	mc.protocolMetricMutex.RLock()
	if len(mc.protocolMetrics) > 0 {
		protocolMetrics := make(map[string]map[string]int64)
		for protocol, protocolData := range mc.protocolMetrics {
			protocolMetrics[protocol] = make(map[string]int64)
			for metric, value := range protocolData {
				protocolMetrics[protocol][metric] = value
			}
		}
		metrics["protocols"] = protocolMetrics
	}
	mc.protocolMetricMutex.RUnlock()
	
	return metrics
}

// calculateRequestStats 计算请求统计
func (mc *MetricsCollector) calculateRequestStats() map[string]interface{} {
	mc.durationMutex.RLock()
	defer mc.durationMutex.RUnlock()
	
	if len(mc.requestDurations) == 0 {
		return nil
	}
	
	// 复制切片以避免在计算过程中被修改
	durations := make([]time.Duration, len(mc.requestDurations))
	copy(durations, mc.requestDurations)
	
	// 计算基础统计
	var total, min, max time.Duration
	min = durations[0]
	max = durations[0]
	
	for _, d := range durations {
		total += d
		if d < min {
			min = d
		}
		if d > max {
			max = d
		}
	}
	
	avg := total / time.Duration(len(durations))
	
	// 计算百分位数（需要先排序）
	sortedDurations := make([]time.Duration, len(durations))
	copy(sortedDurations, durations)
	quickSort(sortedDurations, 0, len(sortedDurations)-1)
	
	p50 := percentile(sortedDurations, 50)
	p90 := percentile(sortedDurations, 90)
	p95 := percentile(sortedDurations, 95)
	p99 := percentile(sortedDurations, 99)
	
	return map[string]interface{}{
		"request_stats": map[string]interface{}{
			"count":         len(durations),
			"avg_duration":  avg.String(),
			"min_duration":  min.String(),
			"max_duration":  max.String(),
			"p50_duration":  p50.String(),
			"p90_duration":  p90.String(),
			"p95_duration":  p95.String(),
			"p99_duration":  p99.String(),
			"total_duration": total.String(),
		},
	}
}

// Reset 重置指标
func (mc *MetricsCollector) Reset() {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	
	atomic.StoreInt64(&mc.totalRequests, 0)
	atomic.StoreInt64(&mc.successRequests, 0)
	atomic.StoreInt64(&mc.failedRequests, 0)
	atomic.StoreInt64(&mc.totalConnections, 0)
	atomic.StoreInt64(&mc.activeConnections, 0)
	atomic.StoreInt64(&mc.closedConnections, 0)
	
	mc.durationMutex.Lock()
	mc.requestDurations = mc.requestDurations[:0]
	mc.durationMutex.Unlock()
	
	mc.errorMutex.Lock()
	mc.errors = make(map[string]int64)
	mc.errorMutex.Unlock()
	
	mc.protocolMetricMutex.Lock()
	mc.protocolMetrics = make(map[string]map[string]int64)
	mc.protocolMetricMutex.Unlock()
	
	mc.startTime = time.Now()
}

// GetRequestRate 获取请求速率（每秒请求数）
func (mc *MetricsCollector) GetRequestRate() float64 {
	totalRequests := atomic.LoadInt64(&mc.totalRequests)
	duration := time.Since(mc.startTime).Seconds()
	
	if duration > 0 {
		return float64(totalRequests) / duration
	}
	
	return 0
}

// GetSuccessRate 获取成功率
func (mc *MetricsCollector) GetSuccessRate() float64 {
	totalRequests := atomic.LoadInt64(&mc.totalRequests)
	successRequests := atomic.LoadInt64(&mc.successRequests)
	
	if totalRequests > 0 {
		return float64(successRequests) / float64(totalRequests) * 100
	}
	
	return 0
}

// GetErrorRate 获取错误率
func (mc *MetricsCollector) GetErrorRate() float64 {
	totalRequests := atomic.LoadInt64(&mc.totalRequests)
	failedRequests := atomic.LoadInt64(&mc.failedRequests)
	
	if totalRequests > 0 {
		return float64(failedRequests) / float64(totalRequests) * 100
	}
	
	return 0
}

// HealthChecker 健康检查器实现
type HealthChecker struct {
	server interfaces.Server
	checks map[string]HealthCheckFunc
	mutex  sync.RWMutex
}

// HealthCheckFunc 健康检查函数类型
type HealthCheckFunc func() error

// NewHealthChecker 创建健康检查器
func NewHealthChecker(server interfaces.Server) *HealthChecker {
	return &HealthChecker{
		server: server,
		checks: make(map[string]HealthCheckFunc),
	}
}

// AddCheck 添加健康检查
func (hc *HealthChecker) AddCheck(name string, checkFunc HealthCheckFunc) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()
	hc.checks[name] = checkFunc
}

// RemoveCheck 移除健康检查
func (hc *HealthChecker) RemoveCheck(name string) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()
	delete(hc.checks, name)
}

// Check 执行健康检查
func (hc *HealthChecker) Check(ctx context.Context) error {
	hc.mutex.RLock()
	defer hc.mutex.RUnlock()
	
	// 检查服务端是否运行
	if !hc.server.IsRunning() {
		return fmt.Errorf("server is not running")
	}
	
	// 执行所有健康检查
	for name, checkFunc := range hc.checks {
		if err := checkFunc(); err != nil {
			return fmt.Errorf("health check '%s' failed: %w", name, err)
		}
	}
	
	return nil
}

// GetStatus 获取健康状态
func (hc *HealthChecker) GetStatus() interfaces.HealthStatus {
	start := time.Now()
	err := hc.Check(context.Background())
	duration := time.Since(start)
	
	status := "healthy"
	details := make(map[string]string)
	
	if err != nil {
		status = "unhealthy"
		details["error"] = err.Error()
	}
	
	details["protocol"] = hc.server.GetProtocol()
	details["address"] = hc.server.GetAddress()
	
	return interfaces.HealthStatus{
		Status:    status,
		Timestamp: time.Now(),
		Details:   details,
		Duration:  duration,
	}
}

// 工具函数

// quickSort 快速排序
func quickSort(arr []time.Duration, low, high int) {
	if low < high {
		pi := partition(arr, low, high)
		quickSort(arr, low, pi-1)
		quickSort(arr, pi+1, high)
	}
}

// partition 分区函数
func partition(arr []time.Duration, low, high int) int {
	pivot := arr[high]
	i := low - 1
	
	for j := low; j < high; j++ {
		if arr[j] < pivot {
			i++
			arr[i], arr[j] = arr[j], arr[i]
		}
	}
	
	arr[i+1], arr[high] = arr[high], arr[i+1]
	return i + 1
}

// percentile 计算百分位数
func percentile(sortedData []time.Duration, p int) time.Duration {
	if len(sortedData) == 0 {
		return 0
	}
	
	if p >= 100 {
		return sortedData[len(sortedData)-1]
	}
	
	if p <= 0 {
		return sortedData[0]
	}
	
	index := float64(p) / 100.0 * float64(len(sortedData)-1)
	lower := int(index)
	upper := lower + 1
	
	if upper >= len(sortedData) {
		return sortedData[len(sortedData)-1]
	}
	
	weight := index - float64(lower)
	return time.Duration(float64(sortedData[lower]) + weight*float64(sortedData[upper]-sortedData[lower]))
}