package http

import (
	"context"
	"fmt"
	"sync"
	"time"

	httpConfig "abc-runner/app/adapters/http/config"
	"abc-runner/app/adapters/http/connection"
	"abc-runner/app/adapters/http/operations"
	"abc-runner/app/core/interfaces"
)

// HttpAdapter HTTP协议适配器
type HttpAdapter struct {
	// 核心组件
	connectionPool *connection.HTTPConnectionPool
	httpOperations *operations.HttpOperations
	config         *httpConfig.HttpAdapterConfig

	// 指标收集器
	metricsCollector interfaces.DefaultMetricsCollector

	// 状态管理
	isConnected bool
	mutex       sync.RWMutex

	// 统计信息
	totalOperations   int64
	successOperations int64
	failedOperations  int64
	startTime         time.Time
}

// NewHttpAdapter 创建HTTP适配器
func NewHttpAdapter(metricsCollector interfaces.DefaultMetricsCollector) *HttpAdapter {
	if metricsCollector == nil {
		panic("metricsCollector cannot be nil - dependency injection required")
	}

	return &HttpAdapter{
		metricsCollector: metricsCollector,
		startTime:        time.Now(),
	}
}

// Connect 初始化连接
func (h *HttpAdapter) Connect(ctx context.Context, config interfaces.Config) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// 类型断言配置
	httpConfig, ok := config.(*httpConfig.HttpAdapterConfig)
	if !ok {
		return fmt.Errorf("invalid config type for HTTP adapter: expected *httpConfig.HttpAdapterConfig, got %T", config)
	}

	// 验证配置
	if err := httpConfig.Validate(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	h.config = httpConfig

	// 创建连接池配置
	poolConfig := connection.PoolConfig{
		MaxConnections:      100, // 默认值，原配置中没有这个字段
		MaxIdleConns:        httpConfig.Connection.MaxIdleConns,
		MaxConnsPerHost:     httpConfig.Connection.MaxConnsPerHost,
		IdleConnTimeout:     httpConfig.Connection.IdleConnTimeout,
		ConnectionTimeout:   httpConfig.Connection.Timeout,
		RequestTimeout:      httpConfig.Connection.Timeout,
		TLSHandshakeTimeout: 10 * time.Second,
		DisableKeepAlives:   false,
		DisableCompression:  httpConfig.Connection.DisableCompression,
	}

	// 创建连接池
	pool, err := connection.NewHTTPConnectionPool(httpConfig, poolConfig)
	if err != nil {
		return fmt.Errorf("failed to create HTTP connection pool: %w", err)
	}

	h.connectionPool = pool

	// 创建HTTP操作执行器
	h.httpOperations = operations.NewHttpOperations(pool, httpConfig, h.metricsCollector)

	// 执行健康检查
	if err := h.HealthCheck(ctx); err != nil {
		return fmt.Errorf("initial health check failed: %w", err)
	}

	h.isConnected = true
	return nil
}

// Execute 执行操作
func (h *HttpAdapter) Execute(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	if !h.IsConnected() {
		return nil, fmt.Errorf("HTTP adapter is not connected")
	}

	// 统计操作数
	h.incrementTotalOperations()

	// 执行操作
	result, err := h.httpOperations.ExecuteOperation(ctx, operation)

	// 更新统计信息
	if err != nil || (result != nil && !result.Success) {
		h.incrementFailedOperations()
	} else {
		h.incrementSuccessOperations()
	}

	// 注意：不要在这里调用 h.metricsCollector.Record(result)
	// 因为执行引擎会负责记录指标，避免重复计数

	return result, err
}

// Close 关闭连接
func (h *HttpAdapter) Close() error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if h.connectionPool != nil {
		if err := h.connectionPool.Close(); err != nil {
			return fmt.Errorf("failed to close HTTP connection pool: %w", err)
		}
		h.connectionPool = nil
	}

	h.httpOperations = nil
	h.isConnected = false

	return nil
}

// HealthCheck 健康检查
func (h *HttpAdapter) HealthCheck(ctx context.Context) error {
	if h.connectionPool == nil {
		return fmt.Errorf("connection pool not initialized")
	}

	return h.connectionPool.HealthCheck()
}

// GetProtocolName 获取协议名称
func (h *HttpAdapter) GetProtocolName() string {
	return "http"
}

// GetMetricsCollector 获取指标收集器
func (h *HttpAdapter) GetMetricsCollector() interfaces.DefaultMetricsCollector {
	return h.metricsCollector
}

// GetProtocolMetrics 获取HTTP特定指标
func (h *HttpAdapter) GetProtocolMetrics() map[string]interface{} {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	metrics := map[string]interface{}{
		"protocol":           "http",
		"base_url":           h.getBaseURL(),
		"is_connected":       h.isConnected,
		"total_operations":   h.totalOperations,
		"success_operations": h.successOperations,
		"failed_operations":  h.failedOperations,
		"uptime_seconds":     time.Since(h.startTime).Seconds(),
	}

	// 添加连接池统计信息
	if h.connectionPool != nil {
		poolStats := h.connectionPool.GetStats()
		metrics["connection_pool"] = poolStats
	}

	// 添加配置信息
	if h.config != nil {
		metrics["config"] = map[string]interface{}{
			"base_url":           h.config.Connection.BaseURL,
			"timeout":            h.config.Connection.Timeout.String(),
			"max_idle_conns":     h.config.Connection.MaxIdleConns,
			"max_conns_per_host": h.config.Connection.MaxConnsPerHost,
		}
	}

	return metrics
}

// IsConnected 检查连接状态
func (h *HttpAdapter) IsConnected() bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return h.isConnected
}

// GetConfig 获取HTTP配置
func (h *HttpAdapter) GetConfig() *httpConfig.HttpAdapterConfig {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return h.config
}

// 私有辅助方法
func (h *HttpAdapter) incrementTotalOperations() {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.totalOperations++
}

func (h *HttpAdapter) incrementSuccessOperations() {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.successOperations++
}

func (h *HttpAdapter) incrementFailedOperations() {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.failedOperations++
}

func (h *HttpAdapter) getBaseURL() string {
	if h.config != nil {
		return h.config.Connection.BaseURL
	}
	return ""
}

// 确保实现了ProtocolAdapter接口
var _ interfaces.ProtocolAdapter = (*HttpAdapter)(nil)
