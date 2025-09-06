package base

import (
	"context"
	"fmt"
	"sync"
	"time"

	"redis-runner/app/core/interfaces"
)

// BaseAdapter 协议适配器基础实现
type BaseAdapter struct {
	protocol     string
	config       interfaces.Config
	connected    bool
	connMutex    sync.RWMutex
	metrics      map[string]interface{}
	metricsMutex sync.RWMutex
}

// NewBaseAdapter 创建基础适配器
func NewBaseAdapter(protocol string) *BaseAdapter {
	return &BaseAdapter{
		protocol: protocol,
		metrics:  make(map[string]interface{}),
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

// ExecuteWithRetry 带重试的执行操作
func (b *BaseAdapter) ExecuteWithRetry(ctx context.Context, operation interfaces.Operation, executeFunc func(context.Context, interfaces.Operation) (*interfaces.OperationResult, error), maxRetries int) (*interfaces.OperationResult, error) {
	var lastErr error

	for i := 0; i <= maxRetries; i++ {
		result, err := executeFunc(ctx, operation)
		if err == nil {
			return result, nil
		}

		lastErr = err

		// 如果是最后一次重试，直接返回错误
		if i == maxRetries {
			break
		}

		// 检查是否应该重试
		if !b.shouldRetry(err) {
			break
		}

		// 等待重试间隔
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(b.getRetryInterval(i)):
			continue
		}
	}

	return &interfaces.OperationResult{
		Success:  false,
		Duration: 0,
		IsRead:   b.isReadOperation(operation.Type),
		Error:    fmt.Errorf("operation failed after %d retries: %w", maxRetries, lastErr),
	}, lastErr
}

// shouldRetry 判断是否应该重试
func (b *BaseAdapter) shouldRetry(err error) bool {
	// 默认实现：对于网络错误和超时错误进行重试
	if err == nil {
		return false
	}

	errStr := err.Error()
	return contains(errStr, "timeout") ||
		contains(errStr, "connection refused") ||
		contains(errStr, "network") ||
		contains(errStr, "broken pipe")
}

// getRetryInterval 获取重试间隔
func (b *BaseAdapter) getRetryInterval(retryCount int) time.Duration {
	// 指数退避：100ms, 200ms, 400ms, 800ms, 1600ms
	interval := time.Duration(100*(1<<retryCount)) * time.Millisecond
	if interval > 5*time.Second {
		interval = 5 * time.Second
	}
	return interval
}

// isReadOperation 判断是否为读操作
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

// DefaultMetricsCollector 默认指标收集器实现
type DefaultMetricsCollector struct {
	mutex      sync.RWMutex
	operations []interfaces.OperationResult
	startTime  time.Time
	totalOps   int64
	successOps int64
	failedOps  int64
	readOps    int64
	writeOps   int64
	durations  []time.Duration
}

// NewDefaultMetricsCollector 创建默认指标收集器
func NewDefaultMetricsCollector() *DefaultMetricsCollector {
	return &DefaultMetricsCollector{
		operations: make([]interfaces.OperationResult, 0),
		durations:  make([]time.Duration, 0),
		startTime:  time.Now(),
	}
}

// RecordOperation 记录操作结果
func (c *DefaultMetricsCollector) RecordOperation(result *interfaces.OperationResult) {
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
}

// GetMetrics 获取指标
func (c *DefaultMetricsCollector) GetMetrics() *interfaces.Metrics {
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
func (c *DefaultMetricsCollector) calculateMetrics(durations []time.Duration) *interfaces.Metrics {
	if len(durations) == 0 {
		return &interfaces.Metrics{}
	}

	// 排序计算分位数
	sortDurations(durations)

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

// Reset 重置指标
func (c *DefaultMetricsCollector) Reset() {
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
}

// Export 导出指标
func (c *DefaultMetricsCollector) Export() map[string]interface{} {
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

// sortDurations 排序时间切片
func sortDurations(durations []time.Duration) {
	// 简单的冒泡排序，对于小数据集足够
	n := len(durations)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if durations[j] > durations[j+1] {
				durations[j], durations[j+1] = durations[j+1], durations[j]
			}
		}
	}
}
