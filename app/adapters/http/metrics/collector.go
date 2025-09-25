package metrics

import (
	"sort"
	"sync"
	"time"

	"abc-runner/app/core/interfaces"
)

// MetricsCollector HTTP指标收集器 (参考Redis架构)
type MetricsCollector struct {
	// 基础指标
	totalOperations   int64
	successOperations int64
	failedOperations  int64
	readOperations    int64
	writeOperations   int64

	// 延迟指标
	totalLatency   time.Duration
	minLatency     time.Duration
	maxLatency     time.Duration
	latencyHistory []time.Duration

	// HTTP特定指标
	statusCodeStats  map[int]*StatusCodeStat
	methodStats      map[string]*MethodStat
	urlStats         map[string]*URLStat
	contentTypeStats map[string]*ContentTypeStat
	networkStats     *NetworkStat
	connectionStats  *ConnectionStat
	windowStats      *WindowStat

	// 错误统计
	errorStats map[string]int64

	mutex     sync.RWMutex
	startTime time.Time
}

// StatusCodeStat 状态码统计
type StatusCodeStat struct {
	Count        int64         `json:"count"`
	SuccessCount int64         `json:"success_count"`
	TotalLatency time.Duration `json:"total_latency"`
	AvgLatency   time.Duration `json:"avg_latency"`
}

// MethodStat 请求方法统计
type MethodStat struct {
	Count        int64         `json:"count"`
	SuccessCount int64         `json:"success_count"`
	FailureCount int64         `json:"failure_count"`
	TotalLatency time.Duration `json:"total_latency"`
	AvgLatency   time.Duration `json:"avg_latency"`
	MinLatency   time.Duration `json:"min_latency"`
	MaxLatency   time.Duration `json:"max_latency"`
}

// URLStat URL统计
type URLStat struct {
	Count         int64         `json:"count"`
	SuccessCount  int64         `json:"success_count"`
	FailureCount  int64         `json:"failure_count"`
	TotalLatency  time.Duration `json:"total_latency"`
	AvgLatency    time.Duration `json:"avg_latency"`
	MinLatency    time.Duration `json:"min_latency"`
	MaxLatency    time.Duration `json:"max_latency"`
	ResponseSizes []int64       `json:"response_sizes"`
}

// ContentTypeStat Content-Type统计
type ContentTypeStat struct {
	Count int64 `json:"count"`
}

// NetworkStat 网络统计
type NetworkStat struct {
	DNSResolveTime   []time.Duration `json:"dns_resolve_time"`
	ConnectionTime   []time.Duration `json:"connection_time"`
	TLSHandshakeTime []time.Duration `json:"tls_handshake_time"`
	FirstByteTime    []time.Duration `json:"first_byte_time"`
}

// ConnectionStat 连接统计
type ConnectionStat struct {
	TotalConnections  int64         `json:"total_connections"`
	ActiveConnections int64         `json:"active_connections"`
	FailedConnections int64         `json:"failed_connections"`
	ConnectionLatency time.Duration `json:"connection_latency"`
	KeepAliveReused   int64         `json:"keep_alive_reused"`
}

// WindowStat 时间窗口统计
type WindowStat struct {
	WindowSize       time.Duration `json:"window_size"`
	CurrentWindow    int64         `json:"current_window"`
	WindowOperations []int64       `json:"window_operations"`
	RPS              float64       `json:"rps"`
	LastUpdate       time.Time     `json:"last_update"`
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
	MinLatency   time.Duration `json:"min_latency"`
	MaxLatency   time.Duration `json:"max_latency"`
	AvgLatency   time.Duration `json:"avg_latency"`
	P50Latency   time.Duration `json:"p50_latency"`
	P90Latency   time.Duration `json:"p90_latency"`
	P95Latency   time.Duration `json:"p95_latency"`
	P99Latency   time.Duration `json:"p99_latency"`
	TotalLatency time.Duration `json:"total_latency"`
}

// NewMetricsCollector 创建HTTP指标收集器
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		statusCodeStats:  make(map[int]*StatusCodeStat),
		methodStats:      make(map[string]*MethodStat),
		urlStats:         make(map[string]*URLStat),
		contentTypeStats: make(map[string]*ContentTypeStat),
		networkStats: &NetworkStat{
			DNSResolveTime:   make([]time.Duration, 0),
			ConnectionTime:   make([]time.Duration, 0),
			TLSHandshakeTime: make([]time.Duration, 0),
			FirstByteTime:    make([]time.Duration, 0),
		},
		connectionStats: &ConnectionStat{},
		windowStats: &WindowStat{
			WindowSize:       time.Second,
			WindowOperations: make([]int64, 60),
			LastUpdate:       time.Now(),
		},
		errorStats:     make(map[string]int64),
		latencyHistory: make([]time.Duration, 0),
		startTime:      time.Now(),
		minLatency:     time.Duration(^uint64(0) >> 1),
		maxLatency:     0,
	}
}

// RecordOperation 实现核心接口（适配 interfaces.MetricsCollector）
func (hc *MetricsCollector) RecordOperation(result *interfaces.OperationResult) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()

	// 更新基础指标
	hc.totalOperations++
	if result.Success {
		hc.successOperations++
	} else {
		hc.failedOperations++
		if result.Error != nil {
			hc.errorStats[result.Error.Error()]++
		}
	}

	if result.IsRead {
		hc.readOperations++
	} else {
		hc.writeOperations++
	}

	// 更新延迟指标
	hc.totalLatency += result.Duration
	if result.Duration < hc.minLatency {
		hc.minLatency = result.Duration
	}
	if result.Duration > hc.maxLatency {
		hc.maxLatency = result.Duration
	}

	// 保存延迟历史（限制大小）
	hc.latencyHistory = append(hc.latencyHistory, result.Duration)
	if len(hc.latencyHistory) > 10000 { // 保留最近10000个样本
		hc.latencyHistory = hc.latencyHistory[1:]
	}

	// 更新时间窗口统计
	hc.updateWindowStats()

	// 处理HTTP特定元数据
	if result.Metadata != nil {
		if statusCode, exists := result.Metadata["status_code"]; exists {
			if sc, ok := statusCode.(int); ok {
				hc.recordStatusCode(sc, result.Duration, result.Success)
			}
		}

		if method, exists := result.Metadata["method"]; exists {
			if m, ok := method.(string); ok {
				hc.recordMethod(m, result.Duration, result.Success)
			}
		}

		if url, exists := result.Metadata["url"]; exists {
			if u, ok := url.(string); ok {
				hc.recordURL(u, result.Duration, result.Success)
			}
		}
	}
}

// GetMetrics 实现核心接口
func (hc *MetricsCollector) GetMetrics() *interfaces.Metrics {
	return hc.GetMetricsForCore()
}

// GetHttpMetrics 获取HTTP特定指标
func (hc *MetricsCollector) GetHttpMetrics() map[string]interface{} {
	hc.mutex.RLock()
	defer hc.mutex.RUnlock()

	duration := time.Since(hc.startTime)

	// 基础指标
	basicMetrics := &BasicMetrics{
		TotalOperations:   hc.totalOperations,
		SuccessOperations: hc.successOperations,
		FailedOperations:  hc.failedOperations,
		ReadOperations:    hc.readOperations,
		WriteOperations:   hc.writeOperations,
	}

	if hc.totalOperations > 0 {
		basicMetrics.SuccessRate = float64(hc.successOperations) / float64(hc.totalOperations) * 100
		basicMetrics.RPS = float64(hc.totalOperations) / duration.Seconds()
	}

	if hc.writeOperations > 0 {
		basicMetrics.ReadWriteRatio = float64(hc.readOperations) / float64(hc.writeOperations)
	}

	// 延迟指标
	latencyMetrics := hc.calculateLatencyMetrics()

	// 组装结果
	result := map[string]interface{}{
		"basic_metrics":      basicMetrics,
		"latency_metrics":    latencyMetrics,
		"status_code_stats":  hc.statusCodeStats,
		"method_stats":       hc.methodStats,
		"url_stats":          hc.urlStats,
		"content_type_stats": hc.contentTypeStats,
		"network_stats":      hc.networkStats,
		"connection_stats":   hc.connectionStats,
		"window_stats":       hc.windowStats,
		"error_stats":        hc.errorStats,
		"duration":           duration,
		"timestamp":          time.Now(),
	}

	return result
}
func (hc *MetricsCollector) GetMetricsForCore() *interfaces.Metrics {
	hc.mutex.RLock()
	defer hc.mutex.RUnlock()

	duration := time.Since(hc.startTime)

	metrics := &interfaces.Metrics{
		TotalOps:   hc.totalOperations,
		SuccessOps: hc.successOperations,
		FailedOps:  hc.failedOperations,
		ReadOps:    hc.readOperations,
		WriteOps:   hc.writeOperations,
		StartTime:  hc.startTime,
		EndTime:    time.Now(),
		Duration:   duration,
	}

	// 计算延迟指标
	if hc.totalOperations > 0 {
		metrics.AvgLatency = hc.totalLatency / time.Duration(hc.totalOperations)
		metrics.MinLatency = hc.minLatency
		metrics.MaxLatency = hc.maxLatency

		// 计算百分位数
		if len(hc.latencyHistory) > 0 {
			sortedLatencies := make([]time.Duration, len(hc.latencyHistory))
			copy(sortedLatencies, hc.latencyHistory)
			sort.Slice(sortedLatencies, func(i, j int) bool {
				return sortedLatencies[i] < sortedLatencies[j]
			})

			if len(sortedLatencies) > 0 {
				metrics.P90Latency = sortedLatencies[len(sortedLatencies)*90/100]
				metrics.P95Latency = sortedLatencies[len(sortedLatencies)*95/100]
				metrics.P99Latency = sortedLatencies[len(sortedLatencies)*99/100]
			}
		}
	}

	return metrics
}

// Export 实现核心接口（适配 interfaces.MetricsCollector）
func (hc *MetricsCollector) Export() map[string]interface{} {
	metrics := hc.GetMetricsForCore()
	duration := time.Since(hc.startTime)

	// 计算基本指标
	errorRate := float64(0)
	rps := float64(0)
	successRate := float64(0)

	if hc.totalOperations > 0 {
		errorRate = float64(hc.failedOperations) / float64(hc.totalOperations) * 100
		successRate = float64(hc.successOperations) / float64(hc.totalOperations) * 100
		rps = float64(hc.totalOperations) / duration.Seconds()
	}

	return map[string]interface{}{
		"total_ops":    hc.totalOperations,
		"success_ops":  hc.successOperations,
		"failed_ops":   hc.failedOperations,
		"read_ops":     hc.readOperations,
		"write_ops":    hc.writeOperations,
		"error_rate":   errorRate,
		"success_rate": successRate,
		"rps":          rps,
		"avg_latency":  metrics.AvgLatency.Nanoseconds(),
		"min_latency":  metrics.MinLatency.Nanoseconds(),
		"max_latency":  metrics.MaxLatency.Nanoseconds(),
		"p90_latency":  metrics.P90Latency.Nanoseconds(),
		"p95_latency":  metrics.P95Latency.Nanoseconds(),
		"p99_latency":  metrics.P99Latency.Nanoseconds(),
		"duration":     duration.Nanoseconds(),
		"start_time":   hc.startTime.Unix(),
		"end_time":     time.Now().Unix(),
	}
}

// Reset 重置指标
func (hc *MetricsCollector) Reset() {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()

	hc.totalOperations = 0
	hc.successOperations = 0
	hc.failedOperations = 0
	hc.readOperations = 0
	hc.writeOperations = 0
	hc.totalLatency = 0
	hc.minLatency = time.Duration(^uint64(0) >> 1)
	hc.maxLatency = 0
	hc.latencyHistory = hc.latencyHistory[:0]

	// 重置HTTP特定指标
	for statusCode := range hc.statusCodeStats {
		delete(hc.statusCodeStats, statusCode)
	}
	for method := range hc.methodStats {
		delete(hc.methodStats, method)
	}
	for url := range hc.urlStats {
		delete(hc.urlStats, url)
	}
	for contentType := range hc.contentTypeStats {
		delete(hc.contentTypeStats, contentType)
	}

	// 重置错误统计
	for errorMsg := range hc.errorStats {
		delete(hc.errorStats, errorMsg)
	}

	// 重置网络统计
	hc.networkStats = &NetworkStat{
		DNSResolveTime:   make([]time.Duration, 0),
		ConnectionTime:   make([]time.Duration, 0),
		TLSHandshakeTime: make([]time.Duration, 0),
		FirstByteTime:    make([]time.Duration, 0),
	}

	// 重置连接统计
	hc.connectionStats = &ConnectionStat{}

	// 重置时间窗口
	hc.windowStats = &WindowStat{
		WindowSize:       time.Second,
		WindowOperations: make([]int64, 60),
		LastUpdate:       time.Now(),
	}

	hc.startTime = time.Now()
}

// recordStatusCode 记录状态码统计
func (hc *MetricsCollector) recordStatusCode(statusCode int, duration time.Duration, success bool) {
	if _, exists := hc.statusCodeStats[statusCode]; !exists {
		hc.statusCodeStats[statusCode] = &StatusCodeStat{}
	}
	stat := hc.statusCodeStats[statusCode]
	stat.Count++
	stat.TotalLatency += duration
	if success {
		stat.SuccessCount++
	}
	if stat.Count > 0 {
		stat.AvgLatency = stat.TotalLatency / time.Duration(stat.Count)
	}
}

// recordMethod 记录方法统计
func (hc *MetricsCollector) recordMethod(method string, duration time.Duration, success bool) {
	if _, exists := hc.methodStats[method]; !exists {
		hc.methodStats[method] = &MethodStat{
			MinLatency: time.Duration(^uint64(0) >> 1),
		}
	}
	methodStat := hc.methodStats[method]
	methodStat.Count++
	methodStat.TotalLatency += duration
	if success {
		methodStat.SuccessCount++
	} else {
		methodStat.FailureCount++
	}
	if duration < methodStat.MinLatency {
		methodStat.MinLatency = duration
	}
	if duration > methodStat.MaxLatency {
		methodStat.MaxLatency = duration
	}
	if methodStat.Count > 0 {
		methodStat.AvgLatency = methodStat.TotalLatency / time.Duration(methodStat.Count)
	}
}

// recordURL 记录URL统计
func (hc *MetricsCollector) recordURL(url string, duration time.Duration, success bool) {
	if _, exists := hc.urlStats[url]; !exists {
		hc.urlStats[url] = &URLStat{
			MinLatency:    time.Duration(^uint64(0) >> 1),
			ResponseSizes: make([]int64, 0),
		}
	}
	urlStat := hc.urlStats[url]
	urlStat.Count++
	urlStat.TotalLatency += duration
	if success {
		urlStat.SuccessCount++
	} else {
		urlStat.FailureCount++
	}
	if duration < urlStat.MinLatency {
		urlStat.MinLatency = duration
	}
	if duration > urlStat.MaxLatency {
		urlStat.MaxLatency = duration
	}
	if urlStat.Count > 0 {
		urlStat.AvgLatency = urlStat.TotalLatency / time.Duration(urlStat.Count)
	}
}

// calculateLatencyMetrics 计算延迟指标
func (hc *MetricsCollector) calculateLatencyMetrics() *LatencyMetrics {
	latencyMetrics := &LatencyMetrics{
		MinLatency:   hc.minLatency,
		MaxLatency:   hc.maxLatency,
		TotalLatency: hc.totalLatency,
	}

	if hc.totalOperations > 0 {
		latencyMetrics.AvgLatency = hc.totalLatency / time.Duration(hc.totalOperations)
	}

	// 计算百分位数
	if len(hc.latencyHistory) > 0 {
		sortedLatencies := make([]time.Duration, len(hc.latencyHistory))
		copy(sortedLatencies, hc.latencyHistory)
		sort.Slice(sortedLatencies, func(i, j int) bool {
			return sortedLatencies[i] < sortedLatencies[j]
		})

		latencyMetrics.P50Latency = hc.getPercentile(sortedLatencies, 50)
		latencyMetrics.P90Latency = hc.getPercentile(sortedLatencies, 90)
		latencyMetrics.P95Latency = hc.getPercentile(sortedLatencies, 95)
		latencyMetrics.P99Latency = hc.getPercentile(sortedLatencies, 99)
	}

	return latencyMetrics
}

// getPercentile 获取百分位数
func (hc *MetricsCollector) getPercentile(sortedLatencies []time.Duration, percentile int) time.Duration {
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

// updateWindowStats 更新时间窗口统计
func (hc *MetricsCollector) updateWindowStats() {
	now := time.Now()
	timeSinceLastUpdate := now.Sub(hc.windowStats.LastUpdate)

	if timeSinceLastUpdate >= hc.windowStats.WindowSize {
		windowsToMove := int(timeSinceLastUpdate / hc.windowStats.WindowSize)
		for i := 0; i < windowsToMove && i < len(hc.windowStats.WindowOperations); i++ {
			copy(hc.windowStats.WindowOperations, hc.windowStats.WindowOperations[1:])
			hc.windowStats.WindowOperations[len(hc.windowStats.WindowOperations)-1] = 0
		}
		hc.windowStats.LastUpdate = now
	}

	currentIndex := len(hc.windowStats.WindowOperations) - 1
	hc.windowStats.WindowOperations[currentIndex]++

	totalOpsInWindow := int64(0)
	for _, ops := range hc.windowStats.WindowOperations {
		totalOpsInWindow += ops
	}
	windowDuration := time.Duration(len(hc.windowStats.WindowOperations)) * hc.windowStats.WindowSize
	hc.windowStats.RPS = float64(totalOpsInWindow) / windowDuration.Seconds()
}

// GetHttpSpecificMetrics 获取HTTP特定指标
func (hc *MetricsCollector) GetHttpSpecificMetrics() map[string]interface{} {
	hc.mutex.RLock()
	defer hc.mutex.RUnlock()

	return map[string]interface{}{
		"status_codes":     hc.statusCodeStats,
		"methods":          hc.methodStats,
		"urls":             hc.urlStats,
		"content_types":    hc.contentTypeStats,
		"network_stats":    hc.networkStats,
		"connection_stats": hc.connectionStats,
		"window_stats":     hc.windowStats,
		"error_stats":      hc.errorStats,
	}
}
