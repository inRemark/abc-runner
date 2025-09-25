package metrics

import (
	"sync"
	"time"

	"abc-runner/app/core/interfaces"
)

// HttpMetricsCollector HTTP指标收集器
type HttpMetricsCollector struct {
	mutex sync.RWMutex

	// 基础指标
	operations []interfaces.OperationResult
	totalOps   int64
	successOps int64
	failedOps  int64
	readOps    int64
	writeOps   int64
	durations  []time.Duration
	startTime  time.Time

	// HTTP特定指标
	responseCodeCount    map[int]int64              // 状态码分布
	methodCount          map[string]int64           // 请求方法统计
	urlLatencies         map[string][]time.Duration // URL延迟统计
	responseSizes        []int64                    // 响应大小分布
	tlsHandshakeTimes    []time.Duration            // TLS握手时间
	uploadSpeeds         map[string]float64         // 上传速度统计
	uploadFileSizes      []int64                    // 上传文件大小
	uploadSuccessCount   map[string]int64           // 上传成功统计
	contentTypeCount     map[string]int64           // Content-Type分布
	serverCount          map[string]int64           // 服务器分布
	errorTypeCount       map[string]int64           // 错误类型统计
	redirectCount        int64                      // 重定向次数
	timeoutCount         int64                      // 超时次数
	connectionErrorCount int64                      // 连接错误次数
	dnsResolveTime       []time.Duration            // DNS解析时间
	connectionTime       []time.Duration            // 连接建立时间
	firstByteTime        []time.Duration            // 首字节时间
}

// NewHttpMetricsCollector 创建HTTP指标收集器
func NewHttpMetricsCollector() *HttpMetricsCollector {
	return &HttpMetricsCollector{
		operations:         make([]interfaces.OperationResult, 0),
		durations:          make([]time.Duration, 0),
		startTime:          time.Now(),
		responseCodeCount:  make(map[int]int64),
		methodCount:        make(map[string]int64),
		urlLatencies:       make(map[string][]time.Duration),
		responseSizes:      make([]int64, 0),
		tlsHandshakeTimes:  make([]time.Duration, 0),
		uploadSpeeds:       make(map[string]float64),
		uploadFileSizes:    make([]int64, 0),
		uploadSuccessCount: make(map[string]int64),
		contentTypeCount:   make(map[string]int64),
		serverCount:        make(map[string]int64),
		errorTypeCount:     make(map[string]int64),
		dnsResolveTime:     make([]time.Duration, 0),
		connectionTime:     make([]time.Duration, 0),
		firstByteTime:      make([]time.Duration, 0),
	}
}

// RecordOperation 记录操作结果
func (c *HttpMetricsCollector) RecordOperation(result *interfaces.OperationResult) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

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

	// 记录错误类型
	if result.Error != nil {
		c.recordErrorType(result.Error.Error())
	}
}

// RecordHttpResponse 记录HTTP响应
func (c *HttpMetricsCollector) RecordHttpResponse(statusCode int, method, url string, duration time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 记录状态码分布
	c.responseCodeCount[statusCode]++

	// 记录请求方法统计
	c.methodCount[method]++

	// 记录URL延迟
	if c.urlLatencies[url] == nil {
		c.urlLatencies[url] = make([]time.Duration, 0)
	}
	c.urlLatencies[url] = append(c.urlLatencies[url], duration)

	// 分析状态码类型
	c.analyzeStatusCode(statusCode)
}

// RecordResponseSize 记录响应大小
func (c *HttpMetricsCollector) RecordResponseSize(size int64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.responseSizes = append(c.responseSizes, size)
}

// RecordTLSHandshakeTime 记录TLS握手时间
func (c *HttpMetricsCollector) RecordTLSHandshakeTime(duration time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.tlsHandshakeTimes = append(c.tlsHandshakeTimes, duration)
}

// RecordUploadSpeed 记录上传速度
func (c *HttpMetricsCollector) RecordUploadSpeed(url string, speed float64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.uploadSpeeds[url] = speed
}

// RecordUploadFileSize 记录上传文件大小
func (c *HttpMetricsCollector) RecordUploadFileSize(size int64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.uploadFileSizes = append(c.uploadFileSizes, size)
}

// RecordUploadSuccess 记录上传成功
func (c *HttpMetricsCollector) RecordUploadSuccess(url string, success bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if success {
		c.uploadSuccessCount[url+"_success"]++
	} else {
		c.uploadSuccessCount[url+"_failed"]++
	}
}

// RecordContentType 记录Content-Type
func (c *HttpMetricsCollector) RecordContentType(contentType string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.contentTypeCount[contentType]++
}

// RecordServer 记录服务器信息
func (c *HttpMetricsCollector) RecordServer(server string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.serverCount[server]++
}

// RecordDNSResolveTime 记录DNS解析时间
func (c *HttpMetricsCollector) RecordDNSResolveTime(duration time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.dnsResolveTime = append(c.dnsResolveTime, duration)
}

// RecordConnectionTime 记录连接建立时间
func (c *HttpMetricsCollector) RecordConnectionTime(duration time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.connectionTime = append(c.connectionTime, duration)
}

// RecordFirstByteTime 记录首字节时间
func (c *HttpMetricsCollector) RecordFirstByteTime(duration time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.firstByteTime = append(c.firstByteTime, duration)
}

// recordErrorType 记录错误类型
func (c *HttpMetricsCollector) recordErrorType(errorMsg string) {
	errorType := c.categorizeError(errorMsg)
	c.errorTypeCount[errorType]++
}

// categorizeError 分类错误类型
func (c *HttpMetricsCollector) categorizeError(errorMsg string) string {
	if contains(errorMsg, "timeout") {
		c.timeoutCount++
		return "timeout"
	} else if contains(errorMsg, "connection refused") || contains(errorMsg, "connection reset") {
		c.connectionErrorCount++
		return "connection"
	} else if contains(errorMsg, "redirect") {
		c.redirectCount++
		return "redirect"
	} else if contains(errorMsg, "dns") || contains(errorMsg, "resolve") {
		return "dns"
	} else if contains(errorMsg, "tls") || contains(errorMsg, "certificate") {
		return "tls"
	} else {
		return "other"
	}
}

// analyzeStatusCode 分析状态码
func (c *HttpMetricsCollector) analyzeStatusCode(statusCode int) {
	if statusCode >= 300 && statusCode < 400 {
		c.redirectCount++
	}
}

// GetMetrics 获取基础指标
func (c *HttpMetricsCollector) GetMetrics() *interfaces.Metrics {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if len(c.durations) == 0 {
		return &interfaces.Metrics{
			StartTime: c.startTime,
			EndTime:   time.Now(),
		}
	}

	// 复制并排序延迟数据
	durations := make([]time.Duration, len(c.durations))
	copy(durations, c.durations)
	c.sortDurations(durations)

	return c.calculateMetrics(durations)
}

// GetHttpSpecificMetrics 获取HTTP特定指标
func (c *HttpMetricsCollector) GetHttpSpecificMetrics() map[string]interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	metrics := make(map[string]interface{})

	// 状态码分布
	metrics["status_codes"] = c.copyStatusCodes()

	// 请求方法统计
	metrics["methods"] = c.copyMethodCount()

	// URL延迟统计
	metrics["url_latencies"] = c.calculateURLLatencies()

	// 响应大小统计
	metrics["response_sizes"] = c.calculateResponseSizeStats()

	// TLS握手时间统计
	metrics["tls_handshake_times"] = c.calculateTLSStats()

	// 上传统计
	metrics["upload_stats"] = c.calculateUploadStats()

	// Content-Type分布
	metrics["content_types"] = c.copyContentTypeCount()

	// 服务器分布
	metrics["servers"] = c.copyServerCount()

	// 错误类型统计
	metrics["error_types"] = c.copyErrorTypeCount()

	// 网络时间统计
	metrics["network_times"] = c.calculateNetworkTimeStats()

	// 计数器
	metrics["redirect_count"] = c.redirectCount
	metrics["timeout_count"] = c.timeoutCount
	metrics["connection_error_count"] = c.connectionErrorCount

	return metrics
}

// Reset 重置指标
func (c *HttpMetricsCollector) Reset() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.operations = make([]interfaces.OperationResult, 0)
	c.durations = make([]time.Duration, 0)
	c.startTime = time.Now()
	c.totalOps = 0
	c.successOps = 0
	c.failedOps = 0
	c.readOps = 0
	c.writeOps = 0

	// 重置HTTP特定指标
	c.responseCodeCount = make(map[int]int64)
	c.methodCount = make(map[string]int64)
	c.urlLatencies = make(map[string][]time.Duration)
	c.responseSizes = make([]int64, 0)
	c.tlsHandshakeTimes = make([]time.Duration, 0)
	c.uploadSpeeds = make(map[string]float64)
	c.uploadFileSizes = make([]int64, 0)
	c.uploadSuccessCount = make(map[string]int64)
	c.contentTypeCount = make(map[string]int64)
	c.serverCount = make(map[string]int64)
	c.errorTypeCount = make(map[string]int64)
	c.redirectCount = 0
	c.timeoutCount = 0
	c.connectionErrorCount = 0
	c.dnsResolveTime = make([]time.Duration, 0)
	c.connectionTime = make([]time.Duration, 0)
	c.firstByteTime = make([]time.Duration, 0)
}

// Export 导出指标
func (c *HttpMetricsCollector) Export() map[string]interface{} {
	baseMetrics := c.getBaseMetricsMap()
	httpMetrics := c.GetHttpSpecificMetrics()

	// 合并指标
	result := make(map[string]interface{})
	for k, v := range baseMetrics {
		result[k] = v
	}
	for k, v := range httpMetrics {
		result[k] = v
	}

	return result
}

// 辅助方法

func (c *HttpMetricsCollector) calculateMetrics(durations []time.Duration) *interfaces.Metrics {
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

func (c *HttpMetricsCollector) getBaseMetricsMap() map[string]interface{} {
	metrics := c.GetMetrics()
	return map[string]interface{}{
		"rps":         metrics.RPS,
		"total_ops":   metrics.TotalOps,
		"success_ops": metrics.SuccessOps,
		"failed_ops":  metrics.FailedOps,
		"read_ops":    metrics.ReadOps,
		"write_ops":   metrics.WriteOps,
		"avg_latency": metrics.AvgLatency.Nanoseconds(),
		"min_latency": metrics.MinLatency.Nanoseconds(),
		"max_latency": metrics.MaxLatency.Nanoseconds(),
		"p90_latency": metrics.P90Latency.Nanoseconds(),
		"p95_latency": metrics.P95Latency.Nanoseconds(),
		"p99_latency": metrics.P99Latency.Nanoseconds(),
		"error_rate":  metrics.ErrorRate,
		"duration":    metrics.Duration.Nanoseconds(),
	}
}

func (c *HttpMetricsCollector) copyStatusCodes() map[int]int64 {
	result := make(map[int]int64)
	for k, v := range c.responseCodeCount {
		result[k] = v
	}
	return result
}

func (c *HttpMetricsCollector) copyMethodCount() map[string]int64 {
	result := make(map[string]int64)
	for k, v := range c.methodCount {
		result[k] = v
	}
	return result
}

func (c *HttpMetricsCollector) copyContentTypeCount() map[string]int64 {
	result := make(map[string]int64)
	for k, v := range c.contentTypeCount {
		result[k] = v
	}
	return result
}

func (c *HttpMetricsCollector) copyServerCount() map[string]int64 {
	result := make(map[string]int64)
	for k, v := range c.serverCount {
		result[k] = v
	}
	return result
}

func (c *HttpMetricsCollector) copyErrorTypeCount() map[string]int64 {
	result := make(map[string]int64)
	for k, v := range c.errorTypeCount {
		result[k] = v
	}
	return result
}

func (c *HttpMetricsCollector) calculateURLLatencies() map[string]interface{} {
	result := make(map[string]interface{})
	for url, latencies := range c.urlLatencies {
		if len(latencies) > 0 {
			result[url] = c.calculateLatencyStats(latencies)
		}
	}
	return result
}

func (c *HttpMetricsCollector) calculateResponseSizeStats() map[string]interface{} {
	if len(c.responseSizes) == 0 {
		return nil
	}

	sizes := make([]int64, len(c.responseSizes))
	copy(sizes, c.responseSizes)
	c.sortInt64Slice(sizes)

	var total int64
	for _, size := range sizes {
		total += size
	}

	return map[string]interface{}{
		"min":     sizes[0],
		"max":     sizes[len(sizes)-1],
		"avg":     total / int64(len(sizes)),
		"p90":     sizes[int(float64(len(sizes))*0.9)],
		"p95":     sizes[int(float64(len(sizes))*0.95)],
		"p99":     sizes[int(float64(len(sizes))*0.99)],
		"total":   total,
		"samples": len(sizes),
	}
}

func (c *HttpMetricsCollector) calculateTLSStats() map[string]interface{} {
	if len(c.tlsHandshakeTimes) == 0 {
		return nil
	}

	times := make([]time.Duration, len(c.tlsHandshakeTimes))
	copy(times, c.tlsHandshakeTimes)
	c.sortDurations(times)

	return c.calculateLatencyStats(times)
}

func (c *HttpMetricsCollector) calculateUploadStats() map[string]interface{} {
	if len(c.uploadSpeeds) == 0 && len(c.uploadFileSizes) == 0 {
		return nil
	}

	stats := make(map[string]interface{})

	// 上传速度统计
	if len(c.uploadSpeeds) > 0 {
		var totalSpeed float64
		for _, speed := range c.uploadSpeeds {
			totalSpeed += speed
		}
		stats["avg_upload_speed"] = totalSpeed / float64(len(c.uploadSpeeds))
	}

	// 上传文件大小统计
	if len(c.uploadFileSizes) > 0 {
		sizes := make([]int64, len(c.uploadFileSizes))
		copy(sizes, c.uploadFileSizes)
		c.sortInt64Slice(sizes)

		var total int64
		for _, size := range sizes {
			total += size
		}

		stats["file_sizes"] = map[string]interface{}{
			"min":     sizes[0],
			"max":     sizes[len(sizes)-1],
			"avg":     total / int64(len(sizes)),
			"total":   total,
			"samples": len(sizes),
		}
	}

	// 上传成功率
	stats["success_rates"] = c.uploadSuccessCount

	return stats
}

func (c *HttpMetricsCollector) calculateNetworkTimeStats() map[string]interface{} {
	stats := make(map[string]interface{})

	if len(c.dnsResolveTime) > 0 {
		stats["dns_resolve"] = c.calculateLatencyStats(c.dnsResolveTime)
	}

	if len(c.connectionTime) > 0 {
		stats["connection"] = c.calculateLatencyStats(c.connectionTime)
	}

	if len(c.firstByteTime) > 0 {
		stats["first_byte"] = c.calculateLatencyStats(c.firstByteTime)
	}

	return stats
}

func (c *HttpMetricsCollector) calculateLatencyStats(latencies []time.Duration) map[string]interface{} {
	if len(latencies) == 0 {
		return nil
	}

	times := make([]time.Duration, len(latencies))
	copy(times, latencies)
	c.sortDurations(times)

	var total time.Duration
	for _, t := range times {
		total += t
	}

	return map[string]interface{}{
		"min":     times[0].Nanoseconds(),
		"max":     times[len(times)-1].Nanoseconds(),
		"avg":     (total / time.Duration(len(times))).Nanoseconds(),
		"p90":     times[int(float64(len(times))*0.9)].Nanoseconds(),
		"p95":     times[int(float64(len(times))*0.95)].Nanoseconds(),
		"p99":     times[int(float64(len(times))*0.99)].Nanoseconds(),
		"samples": len(times),
	}
}

func (c *HttpMetricsCollector) sortDurations(durations []time.Duration) {
	n := len(durations)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if durations[j] > durations[j+1] {
				durations[j], durations[j+1] = durations[j+1], durations[j]
			}
		}
	}
}

func (c *HttpMetricsCollector) sortInt64Slice(slice []int64) {
	n := len(slice)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if slice[j] > slice[j+1] {
				slice[j], slice[j+1] = slice[j+1], slice[j]
			}
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && indexOf(s, substr) >= 0
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
