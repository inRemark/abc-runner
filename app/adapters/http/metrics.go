package http

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"abc-runner/app/core/interfaces"
	"abc-runner/app/core/metrics"
)

// HttpMetrics HTTP协议特定指标
type HttpMetrics struct {
	// 状态码统计
	StatusCodes map[int]*HttpStatusCodeStat `json:"status_codes"`

	// 请求方法统计
	Methods map[string]*HttpMethodStat `json:"methods"`

	// URL路径统计
	URLs map[string]*HttpURLStat `json:"urls"`

	// 内容类型统计
	ContentTypes map[string]*HttpContentTypeStat `json:"content_types"`

	// 网络统计
	Network *HttpNetworkStat `json:"network"`

	// 连接统计
	Connection *HttpConnectionStat `json:"connection"`

	// 安全统计
	Security *HttpSecurityStat `json:"security"`

	// 性能统计
	Performance *HttpPerformanceStat `json:"performance"`
}

// HttpStatusCodeStat HTTP状态码统计
type HttpStatusCodeStat struct {
	Count        int64         `json:"count"`
	SuccessCount int64         `json:"success_count"`
	TotalLatency time.Duration `json:"total_latency"`
	AvgLatency   time.Duration `json:"avg_latency"`
	MinLatency   time.Duration `json:"min_latency"`
	MaxLatency   time.Duration `json:"max_latency"`
	Percentage   float64       `json:"percentage"`
}

// HttpMethodStat HTTP方法统计
type HttpMethodStat struct {
	Count        int64         `json:"count"`
	SuccessCount int64         `json:"success_count"`
	FailureCount int64         `json:"failure_count"`
	TotalLatency time.Duration `json:"total_latency"`
	AvgLatency   time.Duration `json:"avg_latency"`
	MinLatency   time.Duration `json:"min_latency"`
	MaxLatency   time.Duration `json:"max_latency"`
	Throughput   float64       `json:"throughput"` // requests/sec
	SuccessRate  float64       `json:"success_rate"`
}

// HttpURLStat HTTP URL统计
type HttpURLStat struct {
	Count         int64            `json:"count"`
	SuccessCount  int64            `json:"success_count"`
	FailureCount  int64            `json:"failure_count"`
	TotalLatency  time.Duration    `json:"total_latency"`
	AvgLatency    time.Duration    `json:"avg_latency"`
	MinLatency    time.Duration    `json:"min_latency"`
	MaxLatency    time.Duration    `json:"max_latency"`
	ResponseSizes []int64          `json:"response_sizes"`
	AvgSize       float64          `json:"avg_size"`
	ErrorTypes    map[string]int64 `json:"error_types"`
}

// HttpContentTypeStat HTTP内容类型统计
type HttpContentTypeStat struct {
	Count       int64   `json:"count"`
	TotalSize   int64   `json:"total_size"`
	AvgSize     float64 `json:"avg_size"`
	MinSize     int64   `json:"min_size"`
	MaxSize     int64   `json:"max_size"`
	Compression float64 `json:"compression_ratio"`
}

// HttpNetworkStat HTTP网络统计
type HttpNetworkStat struct {
	DNSLookupTime     time.Duration `json:"dns_lookup_time"`
	ConnectionTime    time.Duration `json:"connection_time"`
	TLSHandshakeTime  time.Duration `json:"tls_handshake_time"`
	FirstByteTime     time.Duration `json:"first_byte_time"`
	TransferTime      time.Duration `json:"transfer_time"`
	TotalBytesRead    int64         `json:"total_bytes_read"`
	TotalBytesWritten int64         `json:"total_bytes_written"`
	KeepAliveReused   int64         `json:"keep_alive_reused"`
	KeepAliveWaits    int64         `json:"keep_alive_waits"`
}

// HttpConnectionStat HTTP连接统计
type HttpConnectionStat struct {
	ActiveConnections  int32         `json:"active_connections"`
	TotalConnections   int64         `json:"total_connections"`
	FailedConnections  int64         `json:"failed_connections"`
	ConnectionTimeouts int64         `json:"connection_timeouts"`
	ConnectionReused   int64         `json:"connection_reused"`
	PoolUtilization    float64       `json:"pool_utilization"`
	AvgConnectionTime  time.Duration `json:"avg_connection_time"`
	MaxConcurrent      int32         `json:"max_concurrent"`
}

// HttpSecurityStat HTTP安全统计
type HttpSecurityStat struct {
	TLSConnections      int64            `json:"tls_connections"`
	TLSVersions         map[string]int64 `json:"tls_versions"`
	CipherSuites        map[string]int64 `json:"cipher_suites"`
	CertificateErrors   int64            `json:"certificate_errors"`
	SecurityHeaders     map[string]int64 `json:"security_headers"`
	RedirectCount       int64            `json:"redirect_count"`
	AuthenticationFails int64            `json:"authentication_fails"`
}

// HttpPerformanceStat HTTP性能统计
type HttpPerformanceStat struct {
	RequestsPerSecond  float64       `json:"requests_per_second"`
	AvgResponseTime    time.Duration `json:"avg_response_time"`
	MedianResponseTime time.Duration `json:"median_response_time"`
	P95ResponseTime    time.Duration `json:"p95_response_time"`
	P99ResponseTime    time.Duration `json:"p99_response_time"`
	SlowRequests       int64         `json:"slow_requests"`      // > 2s
	VerySlowRequests   int64         `json:"very_slow_requests"` // > 5s
	TimeoutRate        float64       `json:"timeout_rate"`
	ErrorRate          float64       `json:"error_rate"`
	SuccessRate        float64       `json:"success_rate"`
}

// HttpCollector HTTP指标收集器
type HttpCollector struct {
	*metrics.BaseCollector[HttpMetrics]

	// HTTP特定指标
	httpMetrics *HttpMetrics
	mutex       sync.RWMutex

	// 追踪器
	statusCodeTracker  *HttpStatusCodeTracker
	methodTracker      *HttpMethodTracker
	urlTracker         *HttpURLTracker
	contentTypeTracker *HttpContentTypeTracker
	networkTracker     *HttpNetworkTracker
	connectionTracker  *HttpConnectionTracker
	securityTracker    *HttpSecurityTracker
	performanceTracker *HttpPerformanceTracker

	// 配置
	config *metrics.MetricsConfig
}

// NewHttpCollector 创建HTTP指标收集器
func NewHttpCollector(config *metrics.MetricsConfig) *HttpCollector {
	if config == nil {
		config = metrics.DefaultMetricsConfig()
	}

	// 初始化HTTP指标
	httpMetrics := &HttpMetrics{
		StatusCodes:  make(map[int]*HttpStatusCodeStat),
		Methods:      make(map[string]*HttpMethodStat),
		URLs:         make(map[string]*HttpURLStat),
		ContentTypes: make(map[string]*HttpContentTypeStat),
		Network:      &HttpNetworkStat{},
		Connection:   &HttpConnectionStat{},
		Security: &HttpSecurityStat{
			TLSVersions:     make(map[string]int64),
			CipherSuites:    make(map[string]int64),
			SecurityHeaders: make(map[string]int64),
		},
		Performance: &HttpPerformanceStat{},
	}

	// 创建基础收集器
	baseCollector := metrics.NewBaseCollector(config, *httpMetrics)

	collector := &HttpCollector{
		BaseCollector:      baseCollector,
		httpMetrics:        httpMetrics,
		statusCodeTracker:  NewHttpStatusCodeTracker(),
		methodTracker:      NewHttpMethodTracker(),
		urlTracker:         NewHttpURLTracker(),
		contentTypeTracker: NewHttpContentTypeTracker(),
		networkTracker:     NewHttpNetworkTracker(),
		connectionTracker:  NewHttpConnectionTracker(),
		securityTracker:    NewHttpSecurityTracker(),
		performanceTracker: NewHttpPerformanceTracker(),
		config:             config,
	}

	return collector
}

// Record 记录操作结果（覆盖基础实现以添加HTTP特定逻辑）
func (hc *HttpCollector) Record(result *interfaces.OperationResult) {
	// 调用基础记录方法
	hc.BaseCollector.Record(result)

	// HTTP特定记录
	hc.recordHttpOperation(result)
}

// recordHttpOperation 记录HTTP特定操作
func (hc *HttpCollector) recordHttpOperation(result *interfaces.OperationResult) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()

	// 从元数据中提取HTTP特定信息
	metadata := result.Metadata
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	// 记录状态码
	if statusCode, ok := metadata["status_code"].(int); ok {
		hc.statusCodeTracker.Record(statusCode, result)
	}

	// 记录HTTP方法
	if method, ok := metadata["method"].(string); ok {
		hc.methodTracker.Record(method, result)
	}

	// 记录URL
	if url, ok := metadata["url"].(string); ok {
		hc.urlTracker.Record(url, result)
	}

	// 记录内容类型
	if contentType, ok := metadata["content_type"].(string); ok {
		hc.contentTypeTracker.Record(contentType, result)
	}

	// 记录网络指标
	hc.networkTracker.Record(result)

	// 记录连接指标
	hc.connectionTracker.Record(result)

	// 记录安全指标
	hc.securityTracker.Record(result)

	// 记录性能指标
	hc.performanceTracker.Record(result)

	// 更新HTTP特定指标
	hc.updateHttpMetrics()
}

// updateHttpMetrics 更新HTTP指标
func (hc *HttpCollector) updateHttpMetrics() {
	// 更新各类统计
	hc.httpMetrics.StatusCodes = hc.statusCodeTracker.GetStats()
	hc.httpMetrics.Methods = hc.methodTracker.GetStats()
	hc.httpMetrics.URLs = hc.urlTracker.GetStats()
	hc.httpMetrics.ContentTypes = hc.contentTypeTracker.GetStats()
	hc.httpMetrics.Network = hc.networkTracker.GetStats()
	hc.httpMetrics.Connection = hc.connectionTracker.GetStats()
	hc.httpMetrics.Security = hc.securityTracker.GetStats()
	hc.httpMetrics.Performance = hc.performanceTracker.GetStats()

	// 更新基础收集器的协议数据
	hc.UpdateProtocolMetrics(*hc.httpMetrics)
}

// GetHttpMetrics 获取HTTP特定指标
func (hc *HttpCollector) GetHttpMetrics() *HttpMetrics {
	hc.mutex.RLock()
	defer hc.mutex.RUnlock()

	// 创建深拷贝
	metricsCopy := &HttpMetrics{
		StatusCodes:  make(map[int]*HttpStatusCodeStat),
		Methods:      make(map[string]*HttpMethodStat),
		URLs:         make(map[string]*HttpURLStat),
		ContentTypes: make(map[string]*HttpContentTypeStat),
		Network:      &HttpNetworkStat{},
		Connection:   &HttpConnectionStat{},
		Security: &HttpSecurityStat{
			TLSVersions:     make(map[string]int64),
			CipherSuites:    make(map[string]int64),
			SecurityHeaders: make(map[string]int64),
		},
		Performance: &HttpPerformanceStat{},
	}

	// 深拷贝状态码统计
	for code, stat := range hc.httpMetrics.StatusCodes {
		metricsCopy.StatusCodes[code] = &HttpStatusCodeStat{
			Count:        stat.Count,
			SuccessCount: stat.SuccessCount,
			TotalLatency: stat.TotalLatency,
			AvgLatency:   stat.AvgLatency,
			MinLatency:   stat.MinLatency,
			MaxLatency:   stat.MaxLatency,
			Percentage:   stat.Percentage,
		}
	}

	// 深拷贝方法统计
	for method, stat := range hc.httpMetrics.Methods {
		metricsCopy.Methods[method] = &HttpMethodStat{
			Count:        stat.Count,
			SuccessCount: stat.SuccessCount,
			FailureCount: stat.FailureCount,
			TotalLatency: stat.TotalLatency,
			AvgLatency:   stat.AvgLatency,
			MinLatency:   stat.MinLatency,
			MaxLatency:   stat.MaxLatency,
			Throughput:   stat.Throughput,
			SuccessRate:  stat.SuccessRate,
		}
	}

	// 深拷贝URL统计
	for url, stat := range hc.httpMetrics.URLs {
		sizeCopy := make([]int64, len(stat.ResponseSizes))
		copy(sizeCopy, stat.ResponseSizes)

		errorTypesCopy := make(map[string]int64)
		for errType, count := range stat.ErrorTypes {
			errorTypesCopy[errType] = count
		}

		metricsCopy.URLs[url] = &HttpURLStat{
			Count:         stat.Count,
			SuccessCount:  stat.SuccessCount,
			FailureCount:  stat.FailureCount,
			TotalLatency:  stat.TotalLatency,
			AvgLatency:    stat.AvgLatency,
			MinLatency:    stat.MinLatency,
			MaxLatency:    stat.MaxLatency,
			ResponseSizes: sizeCopy,
			AvgSize:       stat.AvgSize,
			ErrorTypes:    errorTypesCopy,
		}
	}

	// 深拷贝内容类型统计
	for contentType, stat := range hc.httpMetrics.ContentTypes {
		metricsCopy.ContentTypes[contentType] = &HttpContentTypeStat{
			Count:       stat.Count,
			TotalSize:   stat.TotalSize,
			AvgSize:     stat.AvgSize,
			MinSize:     stat.MinSize,
			MaxSize:     stat.MaxSize,
			Compression: stat.Compression,
		}
	}

	// 复制其他统计
	*metricsCopy.Network = *hc.httpMetrics.Network
	*metricsCopy.Connection = *hc.httpMetrics.Connection
	*metricsCopy.Performance = *hc.httpMetrics.Performance

	// 深拷贝安全统计
	for version, count := range hc.httpMetrics.Security.TLSVersions {
		metricsCopy.Security.TLSVersions[version] = count
	}
	for cipher, count := range hc.httpMetrics.Security.CipherSuites {
		metricsCopy.Security.CipherSuites[cipher] = count
	}
	for header, count := range hc.httpMetrics.Security.SecurityHeaders {
		metricsCopy.Security.SecurityHeaders[header] = count
	}

	metricsCopy.Security.TLSConnections = hc.httpMetrics.Security.TLSConnections
	metricsCopy.Security.CertificateErrors = hc.httpMetrics.Security.CertificateErrors
	metricsCopy.Security.RedirectCount = hc.httpMetrics.Security.RedirectCount
	metricsCopy.Security.AuthenticationFails = hc.httpMetrics.Security.AuthenticationFails

	return metricsCopy
}

// Export 导出指标（实现interfaces.MetricsCollector接口）
func (hc *HttpCollector) Export() map[string]interface{} {
	// 获取基础指标
	baseMetrics := hc.BaseCollector.Snapshot()

	// 获取HTTP特定指标
	httpMetrics := hc.GetHttpMetrics()

	// 合并指标
	result := make(map[string]interface{})

	// 添加基础指标
	result["core"] = baseMetrics.Core
	result["system"] = baseMetrics.System

	// 添加HTTP特定指标
	result["http"] = httpMetrics

	// 添加汇总信息
	result["protocol"] = "http"
	result["timestamp"] = baseMetrics.Timestamp
	result["duration"] = baseMetrics.Core.Duration

	return result
}

// GetSummary 获取指标摘要
func (hc *HttpCollector) GetSummary() map[string]interface{} {
	snapshot := hc.Snapshot()
	httpMetrics := hc.GetHttpMetrics()

	return map[string]interface{}{
		"protocol":         "http",
		"total_requests":   snapshot.Core.Operations.Total,
		"success_rate":     snapshot.Core.Operations.Rate,
		"avg_latency":      snapshot.Core.Latency.Average,
		"requests_per_sec": httpMetrics.Performance.RequestsPerSecond,
		"error_rate":       httpMetrics.Performance.ErrorRate,
		"timeout_rate":     httpMetrics.Performance.TimeoutRate,
		"status_codes":     len(httpMetrics.StatusCodes),
		"duration":         snapshot.Core.Duration,
	}
}

// Reset 重置所有指标
func (hc *HttpCollector) Reset() {
	hc.BaseCollector.Reset()

	hc.mutex.Lock()
	defer hc.mutex.Unlock()

	// 重置HTTP特定指标
	hc.httpMetrics = &HttpMetrics{
		StatusCodes:  make(map[int]*HttpStatusCodeStat),
		Methods:      make(map[string]*HttpMethodStat),
		URLs:         make(map[string]*HttpURLStat),
		ContentTypes: make(map[string]*HttpContentTypeStat),
		Network:      &HttpNetworkStat{},
		Connection:   &HttpConnectionStat{},
		Security: &HttpSecurityStat{
			TLSVersions:     make(map[string]int64),
			CipherSuites:    make(map[string]int64),
			SecurityHeaders: make(map[string]int64),
		},
		Performance: &HttpPerformanceStat{},
	}

	// 重置追踪器
	hc.statusCodeTracker.Reset()
	hc.methodTracker.Reset()
	hc.urlTracker.Reset()
	hc.contentTypeTracker.Reset()
	hc.networkTracker.Reset()
	hc.connectionTracker.Reset()
	hc.securityTracker.Reset()
	hc.performanceTracker.Reset()
}

// MarshalJSON 自定义JSON序列化
func (hm *HttpMetrics) MarshalJSON() ([]byte, error) {
	type Alias HttpMetrics
	return json.Marshal(&struct {
		*Alias
		Timestamp time.Time `json:"timestamp"`
	}{
		Alias:     (*Alias)(hm),
		Timestamp: time.Now(),
	})
}

// String 返回字符串表示
func (hm *HttpMetrics) String() string {
	data, _ := json.MarshalIndent(hm, "", "  ")
	return string(data)
}

// Validate 验证指标数据的有效性
func (hm *HttpMetrics) Validate() error {
	if hm.StatusCodes == nil {
		return fmt.Errorf("status codes map is nil")
	}
	if hm.Methods == nil {
		return fmt.Errorf("methods map is nil")
	}
	if hm.URLs == nil {
		return fmt.Errorf("urls map is nil")
	}
	if hm.ContentTypes == nil {
		return fmt.Errorf("content types map is nil")
	}
	if hm.Network == nil {
		return fmt.Errorf("network stats is nil")
	}
	if hm.Connection == nil {
		return fmt.Errorf("connection stats is nil")
	}
	if hm.Security == nil {
		return fmt.Errorf("security stats is nil")
	}
	if hm.Performance == nil {
		return fmt.Errorf("performance stats is nil")
	}

	return nil
}

// GetTopStatusCodes 获取最频繁的状态码
func (hm *HttpMetrics) GetTopStatusCodes(limit int) []*TopStatusCode {
	type statusPair struct {
		Code int
		Stat *HttpStatusCodeStat
	}

	var codes []statusPair
	for code, stat := range hm.StatusCodes {
		codes = append(codes, statusPair{Code: code, Stat: stat})
	}

	// 按出现次数排序
	for i := 0; i < len(codes)-1; i++ {
		for j := i + 1; j < len(codes); j++ {
			if codes[i].Stat.Count < codes[j].Stat.Count {
				codes[i], codes[j] = codes[j], codes[i]
			}
		}
	}

	// 限制结果数量
	if limit > 0 && limit < len(codes) {
		codes = codes[:limit]
	}

	result := make([]*TopStatusCode, len(codes))
	for i, code := range codes {
		result[i] = &TopStatusCode{
			Code:       code.Code,
			Count:      code.Stat.Count,
			Percentage: code.Stat.Percentage,
			AvgLatency: code.Stat.AvgLatency,
		}
	}

	return result
}

// TopStatusCode 顶级状态码统计
type TopStatusCode struct {
	Code       int           `json:"code"`
	Count      int64         `json:"count"`
	Percentage float64       `json:"percentage"`
	AvgLatency time.Duration `json:"avg_latency"`
}
