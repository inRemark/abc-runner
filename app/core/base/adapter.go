package base

import (
	"context"
	"fmt"
	"sync"
	"time"

	"abc-runner/app/core/interfaces"
	errorhandler "abc-runner/app/core/error"
)

// BaseAdapter 协议适配器基础实现
type BaseAdapter struct {
	protocol         string
	config           interfaces.Config
	connected        bool
	connMutex        sync.RWMutex
	metrics          map[string]interface{}
	metricsMutex     sync.RWMutex
	metricsCollector interfaces.MetricsCollector        // 外部注入的指标收集器
	errorHandler     *errorhandler.ErrorHandler // 错误处理器
}

// NewBaseAdapter 创建基础适配器（兼容性重载）
func NewBaseAdapter(protocol string, metricsCollector ...interfaces.MetricsCollector) *BaseAdapter {
	var collector interfaces.MetricsCollector
	if len(metricsCollector) > 0 {
		collector = metricsCollector[0]
	} else {
		// 使用默认的增强指标收集器
		collector = createDefaultMetricsCollector()
	}
	return &BaseAdapter{
		protocol:         protocol,
		metrics:          make(map[string]interface{}),
		metricsCollector: collector,
		errorHandler:     errorhandler.NewErrorHandler(),
	}
}

// GetProtocolName 获取协议名称
func (b *BaseAdapter) GetProtocolName() string {
	return b.protocol
}

// IsConnected 检查连接状态
func (b *BaseAdapter) IsConnected() bool {
	b.connMutex.RLock()
	defer b.connMutex.RUnlock()
	return b.connected
}

// SetConnected 设置连接状态
func (b *BaseAdapter) SetConnected(connected bool) {
	b.connMutex.Lock()
	defer b.connMutex.Unlock()
	b.connected = connected
}

// GetConfig 获取配置
func (b *BaseAdapter) GetConfig() interfaces.Config {
	return b.config
}

// SetConfig 设置配置
func (b *BaseAdapter) SetConfig(config interfaces.Config) {
	b.config = config
}

// GetMetricsCollector 获取指标收集器
func (b *BaseAdapter) GetMetricsCollector() interfaces.MetricsCollector {
	return b.metricsCollector
}

// SetMetricsCollector 设置指标收集器
func (b *BaseAdapter) SetMetricsCollector(collector interfaces.MetricsCollector) {
	b.metricsCollector = collector
}

// UpdateMetric 更新指标
func (b *BaseAdapter) UpdateMetric(key string, value interface{}) {
	b.metricsMutex.Lock()
	defer b.metricsMutex.Unlock()
	b.metrics[key] = value
}

// GetProtocolMetrics 获取协议指标
func (b *BaseAdapter) GetProtocolMetrics() map[string]interface{} {
	b.metricsMutex.RLock()
	defer b.metricsMutex.RUnlock()

	result := make(map[string]interface{})
	for k, v := range b.metrics {
		result[k] = v
	}
	return result
}

// ValidateConfig 验证配置
func (b *BaseAdapter) ValidateConfig(config interfaces.Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if config.GetProtocol() != b.protocol {
		return fmt.Errorf("protocol mismatch: expected %s, got %s", b.protocol, config.GetProtocol())
	}

	return config.Validate()
}

// ExecuteWithRetry 带重试的执行操作（委托给ErrorHandler）
func (b *BaseAdapter) ExecuteWithRetry(ctx context.Context, operation interfaces.Operation, executeFunc func(context.Context, interfaces.Operation) (*interfaces.OperationResult, error)) (*interfaces.OperationResult, error) {
	return b.errorHandler.ExecuteWithRetry(ctx, operation, executeFunc)
}

// createDefaultMetricsCollector 创建默认指标收集器（避免循环引用）
func createDefaultMetricsCollector() interfaces.MetricsCollector {
	// 返回一个简单的实现，避免引入 monitoring 包导致的循环引用
	return &simpleMetricsCollector{}
}

// simpleMetricsCollector 简单指标收集器实现
type simpleMetricsCollector struct {
	operations []interfaces.OperationResult
	startTime  time.Time
}

func (s *simpleMetricsCollector) RecordOperation(result *interfaces.OperationResult) {
	s.operations = append(s.operations, *result)
}

func (s *simpleMetricsCollector) GetMetrics() *interfaces.Metrics {
	return &interfaces.Metrics{
		TotalOps:  int64(len(s.operations)),
		StartTime: s.startTime,
		EndTime:   time.Now(),
	}
}

func (s *simpleMetricsCollector) Reset() {
	s.operations = make([]interfaces.OperationResult, 0)
	s.startTime = time.Now()
}

func (s *simpleMetricsCollector) Export() map[string]interface{} {
	return map[string]interface{}{
		"total_ops": len(s.operations),
	}
}
func (b *BaseAdapter) isReadOperation(operationType string) bool {
	readOps := []string{"get", "hget", "hgetall", "lrange", "smembers", "zrange", "exists", "ttl"}
	return contains(readOps, operationType)
}

// contains 检查切片是否包含指定元素
func contains(slice interface{}, item string) bool {
	switch s := slice.(type) {
	case []string:
		for _, v := range s {
			if v == item {
				return true
			}
		}
	case string:
		return s == item || s == ""
	}
	return false
}
