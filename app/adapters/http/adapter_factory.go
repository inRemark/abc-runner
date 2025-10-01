package http

import (
	"abc-runner/app/core/interfaces"
)

// AdapterFactory HTTP适配器工厂
type AdapterFactory struct {
	metricsCollector interfaces.DefaultMetricsCollector
}

// NewAdapterFactory 创建HTTP适配器工厂
func NewAdapterFactory(metricsCollector interfaces.DefaultMetricsCollector) *AdapterFactory {
	return &AdapterFactory{
		metricsCollector: metricsCollector,
	}
}

// CreateHttpAdapter 创建HTTP适配器 (实现HttpAdapterFactory接口)
func (f *AdapterFactory) CreateHttpAdapter() interfaces.ProtocolAdapter {
	if f.metricsCollector == nil {
		panic("metricsCollector cannot be nil - dependency injection required")
	}

	adapter := NewHttpAdapter(f.metricsCollector)
	return adapter
}

// GetProtocolName 获取支持的协议名称
func (f *AdapterFactory) GetProtocolName() string {
	return "http"
}

// GetMetricsCollector 获取指标收集器
func (f *AdapterFactory) GetMetricsCollector() interfaces.DefaultMetricsCollector {
	return f.metricsCollector
}

// SetMetricsCollector 设置指标收集器
func (f *AdapterFactory) SetMetricsCollector(collector interfaces.DefaultMetricsCollector) {
	f.metricsCollector = collector
}

// 确保实现了interfaces.HttpAdapterFactory接口
var _ interfaces.HttpAdapterFactory = (*AdapterFactory)(nil)
