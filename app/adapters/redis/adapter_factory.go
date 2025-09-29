package redis

import (
	"abc-runner/app/core/interfaces"
)

// AdapterFactory Redis适配器工厂
type AdapterFactory struct {
	metricsCollector interfaces.DefaultMetricsCollector
}

// NewAdapterFactory 创建Redis适配器工厂
func NewAdapterFactory(metricsCollector interfaces.DefaultMetricsCollector) *AdapterFactory {
	return &AdapterFactory{
		metricsCollector: metricsCollector,
	}
}

// CreateRedisAdapter 创建Redis适配器 (实现RedisAdapterFactory接口)
func (f *AdapterFactory) CreateRedisAdapter() interfaces.ProtocolAdapter {
	if f.metricsCollector == nil {
		return nil
	}

	adapter := NewRedisAdapter(f.metricsCollector)
	return adapter
}

// GetProtocolName 获取支持的协议名称
func (f *AdapterFactory) GetProtocolName() string {
	return "redis"
}

// GetMetricsCollector 获取指标收集器
func (f *AdapterFactory) GetMetricsCollector() interfaces.DefaultMetricsCollector {
	return f.metricsCollector
}

// SetMetricsCollector 设置指标收集器
func (f *AdapterFactory) SetMetricsCollector(collector interfaces.DefaultMetricsCollector) {
	f.metricsCollector = collector
}

// 确保实现了interfaces.RedisAdapterFactory接口
var _ interfaces.RedisAdapterFactory = (*AdapterFactory)(nil)
