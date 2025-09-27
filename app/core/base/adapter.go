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
	metricsCollector interfaces.DefaultMetricsCollector        // 外部注入的指标收集器
	errorHandler     *errorhandler.ErrorHandler // 错误处理器
}

// NewBaseAdapter 创建基础适配器（完全新架构）
func NewBaseAdapter(protocol string, metricsCollector ...interfaces.DefaultMetricsCollector) *BaseAdapter {
	var collector interfaces.DefaultMetricsCollector
	if len(metricsCollector) > 0 {
		collector = metricsCollector[0]
	} else {
		// 使用默认的泛型指标收集器
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
func (b *BaseAdapter) GetMetricsCollector() interfaces.DefaultMetricsCollector {
	return b.metricsCollector
}

// SetMetricsCollector 设置指标收集器
func (b *BaseAdapter) SetMetricsCollector(collector interfaces.DefaultMetricsCollector) {
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

// createDefaultMetricsCollector 创建默认指标收集器
func createDefaultMetricsCollector() interfaces.DefaultMetricsCollector {
	// 使用新架构的默认实现
	return &simpleMetricsCollector{}
}

// simpleMetricsCollector 简单指标收集器实现
type simpleMetricsCollector struct {
	operations []interfaces.OperationResult
	startTime  time.Time
}

func (s *simpleMetricsCollector) Record(result *interfaces.OperationResult) {
	s.operations = append(s.operations, *result)
}

func (s *simpleMetricsCollector) Snapshot() *interfaces.DefaultMetricsSnapshot {
	total := int64(len(s.operations))
	var success, failed, read, write int64
	
	for _, op := range s.operations {
		if op.Success {
			success++
		} else {
			failed++
		}
		if op.IsRead {
			read++
		} else {
			write++
		}
	}
	
	var rate float64
	if total > 0 {
		rate = float64(success) / float64(total) * 100.0
	}
	
	return &interfaces.DefaultMetricsSnapshot{
		Core: interfaces.CoreMetrics{
			Operations: interfaces.OperationMetrics{
				Total:   total,
				Success: success,
				Failed:  failed,
				Read:    read,
				Write:   write,
				Rate:    rate,
			},
			Latency: interfaces.LatencyMetrics{},
			Throughput: interfaces.ThroughputMetrics{},
			Duration: time.Since(s.startTime),
		},
		Protocol: map[string]interface{}{
			"simple": "default_collector",
		},
		System: interfaces.SystemMetrics{},
		Timestamp: time.Now(),
	}
}

func (s *simpleMetricsCollector) Reset() {
	s.operations = make([]interfaces.OperationResult, 0)
	s.startTime = time.Now()
}

func (s *simpleMetricsCollector) Stop() {
	// 简单实现无需停止操作
}

// 以下是兼容性方法，用于支持旧代码调用
func (s *simpleMetricsCollector) RecordOperation(result *interfaces.OperationResult) {
	s.Record(result)
}

func (s *simpleMetricsCollector) GetMetrics() *interfaces.Metrics {
	snapshot := s.Snapshot()
	return &interfaces.Metrics{
		TotalOps:  snapshot.Core.Operations.Total,
		SuccessOps: snapshot.Core.Operations.Success,
		FailedOps: snapshot.Core.Operations.Failed,
		StartTime: s.startTime,
		EndTime:   time.Now(),
		Duration:  snapshot.Core.Duration,
	}
}

func (s *simpleMetricsCollector) Export() map[string]interface{} {
	snapshot := s.Snapshot()
	return map[string]interface{}{
		"total_ops": snapshot.Core.Operations.Total,
		"success_ops": snapshot.Core.Operations.Success,
		"failed_ops": snapshot.Core.Operations.Failed,
		"success_rate": snapshot.Core.Operations.Rate,
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
