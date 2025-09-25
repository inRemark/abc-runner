package http

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	httpConfig "abc-runner/app/adapters/http/config"
	"abc-runner/app/adapters/http/connection"
	"abc-runner/app/adapters/http/metrics"
	"abc-runner/app/adapters/http/operations"
	"abc-runner/app/core/base"
	"abc-runner/app/core/interfaces"
)

// HttpAdapter HTTP协议适配器实现
type HttpAdapter struct {
	*base.BaseAdapter

	// 连接管理
	connPool *connection.ConnectionPool
	config   *httpConfig.HttpAdapterConfig

	// 指标收集
	metricsCollector *metrics.MetricsCollector
	metricsReporter  *metrics.HttpMetricsReporter

	// 操作执行器
	httpOps          *operations.HttpOperations
	operationFactory *operations.HttpOperationFactory

	// 同步控制
	mutex sync.RWMutex
}

// NewHttpAdapter 创建HTTP适配器
func NewHttpAdapter(metricsCollector interfaces.MetricsCollector) *HttpAdapter {
	adapter := &HttpAdapter{
		BaseAdapter: base.NewBaseAdapter("http"),
	}

	// 使用传入的增强指标收集器，而不是创建专用收集器
	if metricsCollector != nil {
		adapter.SetMetricsCollector(metricsCollector)
		// 不再使用专用的HTTP指标收集器，直接使用通用接口
		adapter.metricsCollector = nil
		adapter.metricsReporter = nil
	} else {
		// 如果没有传入收集器，使用默认的HTTP专用收集器作为后备
		httpCollector := metrics.NewMetricsCollector()
		adapter.metricsCollector = httpCollector
		adapter.metricsReporter = metrics.NewHttpMetricsReporter(httpCollector)
	}

	return adapter
}

// Connect 初始化连接
func (h *HttpAdapter) Connect(ctx context.Context, config interfaces.Config) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// 验证并转换配置
	httpConfig, ok := config.(*httpConfig.HttpAdapterConfig)
	if !ok {
		return fmt.Errorf("invalid config type for HTTP adapter")
	}

	if err := h.ValidateConfig(config); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	h.config = httpConfig
	h.SetConfig(config)

	// 初始化连接池
	poolConfig := connection.PoolConfig{
		MaxConnections:    httpConfig.Connection.MaxConnsPerHost,
		MaxIdleConns:      httpConfig.Connection.MaxIdleConns,
		MaxConnsPerHost:   httpConfig.Connection.MaxConnsPerHost,
		IdleConnTimeout:   httpConfig.Connection.IdleConnTimeout,
		ConnectionTimeout: httpConfig.Connection.Timeout,
		DisableKeepAlives: false,
	}

	var err error
	h.connPool, err = connection.NewConnectionPool(httpConfig, poolConfig)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}

	// 初始化操作执行器和工厂
	// 为操作执行器确保有一个有效的HTTP指标收集器
	var httpMetricsCollector *metrics.MetricsCollector
	if h.metricsCollector != nil {
		httpMetricsCollector = h.metricsCollector
	} else {
		// 如果没有专用收集器，创建一个用于操作执行器
		httpMetricsCollector = metrics.NewMetricsCollector()
		// 同时设置为适配器的专用收集器
		h.metricsCollector = httpMetricsCollector
	}

	h.httpOps = operations.NewHttpOperations(h.connPool, h.config, httpMetricsCollector)
	h.operationFactory = operations.NewHttpOperationFactory(h.config)

	// 测试连接
	if err := h.testConnection(ctx); err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}

	h.SetConnected(true)
	h.UpdateMetric("connected_at", time.Now())
	h.UpdateMetric("base_url", h.config.Connection.BaseURL)
	h.UpdateMetric("pool_size", poolConfig.MaxConnections)

	return nil
}

// Execute 执行操作并返回结果
func (h *HttpAdapter) Execute(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	if !h.IsConnected() {
		return nil, fmt.Errorf("HTTP adapter not connected")
	}

	startTime := time.Now()

	// 执行操作
	result, err := h.executeOperation(ctx, operation)
	if result != nil {
		result.Duration = time.Since(startTime)

		// 使用BaseAdapter的通用指标收集器，而不是专用的HTTP收集器
		if baseCollector := h.GetMetricsCollector(); baseCollector != nil {
			baseCollector.RecordOperation(result)
		}

		// 如果还有专用收集器，也同时记录（向后兼容）
		if h.metricsCollector != nil {
			// 转换为HTTP特定的结果类型
			h.metricsCollector.RecordOperation(result)
		}
	}

	return result, err
}

// executeOperation 执行具体操作
func (h *HttpAdapter) executeOperation(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	switch operation.Type {
	case "http_get":
		return h.httpOps.ExecuteGetOperation(ctx, operation)
	case "http_post":
		return h.httpOps.ExecutePostOperation(ctx, operation)
	case "http_put":
		return h.httpOps.ExecutePutOperation(ctx, operation)
	case "http_patch":
		return h.httpOps.ExecutePatchOperation(ctx, operation)
	case "http_delete":
		return h.httpOps.ExecuteDeleteOperation(ctx, operation)
	case "http_head":
		return h.httpOps.ExecuteHeadOperation(ctx, operation)
	case "http_options":
		return h.httpOps.ExecuteOptionsOperation(ctx, operation)
	case "http_trace":
		return h.httpOps.ExecuteTraceOperation(ctx, operation)
	case "http_connect":
		return h.httpOps.ExecuteConnectOperation(ctx, operation)
	case "http_upload":
		return h.httpOps.ExecuteUploadOperation(ctx, operation)
	default:
		// 使用通用执行方法
		return h.httpOps.ExecuteOperation(ctx, operation)
	}
}

// Close 关闭连接
func (h *HttpAdapter) Close() error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if h.connPool != nil {
		if err := h.connPool.Close(); err != nil {
			return fmt.Errorf("failed to close connection pool: %w", err)
		}
	}

	h.SetConnected(false)
	h.UpdateMetric("disconnected_at", time.Now())

	return nil
}

// HealthCheck 健康检查
func (h *HttpAdapter) HealthCheck(ctx context.Context) error {
	if !h.IsConnected() {
		return fmt.Errorf("adapter not connected")
	}

	// 检查连接池健康状态
	if err := h.connPool.HealthCheck(); err != nil {
		return fmt.Errorf("connection pool health check failed: %w", err)
	}

	// 执行简单的HTTP请求测试连接
	return h.testConnection(ctx)
}

// GetProtocolMetrics 获取HTTP特定指标
func (h *HttpAdapter) GetProtocolMetrics() map[string]interface{} {
	baseMetrics := h.BaseAdapter.GetProtocolMetrics()

	// 使用通用指标收集器的Export方法
	result := make(map[string]interface{})
	for k, v := range baseMetrics {
		result[k] = v
	}

	// 如果有专用收集器，也添加其指标（向后兼容）
	if h.metricsCollector != nil {
		httpMetrics := h.metricsCollector.GetHttpSpecificMetrics()
		for k, v := range httpMetrics {
			result[k] = v
		}
	}

	// 添加连接池状态
	if h.connPool != nil {
		poolStats := h.connPool.GetStats()
		result["connection_pool"] = poolStats
	}

	return result
}

// testConnection 测试连接
func (h *HttpAdapter) testConnection(ctx context.Context) error {
	// 创建一个简单的GET请求来测试连接
	operation := interfaces.Operation{
		Type: "http_get",
		Key:  "connection_test",
		Params: map[string]interface{}{
			"method": "GET",
			"path":   "/",
		},
		Metadata: map[string]string{
			"test": "connection",
		},
	}

	// 使用短超时进行测试
	testCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	result, err := h.httpOps.ExecuteOperation(testCtx, operation)
	if err != nil {
		// 检查是否是可接受的错误
		if h.isAcceptableConnectionTestError(err) {
			return nil
		}
		return fmt.Errorf("connection test failed: %w", err)
	}

	// 如果结果成功或者是可接受的HTTP错误(如404)，说明连接正常
	if result.Success {
		return nil
	}

	// 检查HTTP状态码
	if result.Value != nil {
		if valueMap, ok := result.Value.(map[string]interface{}); ok {
			if statusCode, exists := valueMap["status_code"]; exists {
				if code, ok := statusCode.(int); ok {
					// 4xx和5xx错误说明连接正常，只是服务器端的问题
					if code >= 400 && code < 600 {
						return nil
					}
				}
			}
		}
	}

	return fmt.Errorf("connection test failed with result: %+v", result)
}

// isAcceptableConnectionTestError 检查是否为可接受的连接测试错误
func (h *HttpAdapter) isAcceptableConnectionTestError(err error) bool {
	if err == nil {
		return true
	}

	errStr := err.Error()
	// 这些错误表明连接是正常的，只是服务器端的问题
	acceptableErrors := []string{
		"404",
		"403",
		"405", // Method Not Allowed
		"500",
		"502",
		"503",
	}

	for _, acceptableErr := range acceptableErrors {
		if contains(errStr, acceptableErr) {
			return true
		}
	}

	return false
}

// CreateOperation 创建操作（便捷方法）
func (h *HttpAdapter) CreateOperation(params map[string]interface{}) (interfaces.Operation, error) {
	if h.operationFactory == nil {
		return interfaces.Operation{}, fmt.Errorf("operation factory not initialized")
	}

	return h.operationFactory.CreateOperation(params)
}

// GetOperationFactory 获取操作工厂
func (h *HttpAdapter) GetOperationFactory() interfaces.OperationFactory {
	return h.operationFactory
}

// GetMetricsCollector 获取指标收集器
func (h *HttpAdapter) GetMetricsCollector() interfaces.MetricsCollector {
	// 优先返回BaseAdapter的通用指标收集器（包含系统指标）
	if baseCollector := h.BaseAdapter.GetMetricsCollector(); baseCollector != nil {
		return baseCollector
	}

	// 如果没有通用收集器，返回HTTP专用收集器（向后兼容）
	return h.metricsCollector
}

// GetMetricsReporter 获取指标报告器
func (h *HttpAdapter) GetMetricsReporter() *metrics.HttpMetricsReporter {
	return h.metricsReporter
}

// GenerateReport 生成报告
func (h *HttpAdapter) GenerateReport() string {
	return h.metricsReporter.GenerateReport()
}

// GenerateSimpleReport 生成简单报告
func (h *HttpAdapter) GenerateSimpleReport() string {
	return h.metricsReporter.GenerateSimpleReport()
}

// ExportMetrics 导出指标
func (h *HttpAdapter) ExportMetrics() map[string]interface{} {
	return h.metricsCollector.Export()
}

// ResetMetrics 重置指标
func (h *HttpAdapter) ResetMetrics() {
	h.metricsCollector.Reset()
}

// HTTP特定操作接口

// ExecuteHttpRequest 执行HTTP请求
func (h *HttpAdapter) ExecuteHttpRequest(ctx context.Context, method, path string, body interface{}, headers map[string]string) (*interfaces.OperationResult, error) {
	operation := interfaces.Operation{
		Type:  fmt.Sprintf("http_%s", strings.ToLower(method)),
		Key:   fmt.Sprintf("%s:%s", method, path),
		Value: body,
		Params: map[string]interface{}{
			"method":  method,
			"path":    path,
			"headers": headers,
		},
	}

	return h.Execute(ctx, operation)
}

// ExecuteGetRequest 执行GET请求
func (h *HttpAdapter) ExecuteGetRequest(ctx context.Context, path string, headers map[string]string) (*interfaces.OperationResult, error) {
	return h.ExecuteHttpRequest(ctx, "GET", path, nil, headers)
}

// ExecutePostRequest 执行POST请求
func (h *HttpAdapter) ExecutePostRequest(ctx context.Context, path string, body interface{}, headers map[string]string) (*interfaces.OperationResult, error) {
	return h.ExecuteHttpRequest(ctx, "POST", path, body, headers)
}

// ExecutePutRequest 执行PUT请求
func (h *HttpAdapter) ExecutePutRequest(ctx context.Context, path string, body interface{}, headers map[string]string) (*interfaces.OperationResult, error) {
	return h.ExecuteHttpRequest(ctx, "PUT", path, body, headers)
}

// ExecuteDeleteRequest 执行DELETE请求
func (h *HttpAdapter) ExecuteDeleteRequest(ctx context.Context, path string, headers map[string]string) (*interfaces.OperationResult, error) {
	return h.ExecuteHttpRequest(ctx, "DELETE", path, nil, headers)
}

// GetConnectionPoolStats 获取连接池统计
func (h *HttpAdapter) GetConnectionPoolStats() map[string]interface{} {
	if h.connPool == nil {
		return nil
	}
	return h.connPool.GetStats()
}

// GetConfig 获取HTTP配置
func (h *HttpAdapter) GetHttpConfig() *httpConfig.HttpAdapterConfig {
	return h.config
}

// 辅助函数

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
