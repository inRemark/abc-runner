package tcp

import (
	"abc-runner/app/core/interfaces"
)

// AdapterFactory TCP适配器工厂
type AdapterFactory struct {
	metricsCollector interfaces.DefaultMetricsCollector
}

// NewAdapterFactory 创建TCP适配器工厂
func NewAdapterFactory(metricsCollector interfaces.DefaultMetricsCollector) *AdapterFactory {
	return &AdapterFactory{
		metricsCollector: metricsCollector,
	}
}

// CreateTCPAdapter 创建TCP适配器 (实现TCPAdapterFactory接口)
func (f *AdapterFactory) CreateTCPAdapter() interfaces.ProtocolAdapter {
	if f.metricsCollector == nil {
		return nil
	}

	adapter := NewTCPAdapter(f.metricsCollector)
	return adapter
}

// GetProtocolName 获取支持的协议名称
func (f *AdapterFactory) GetProtocolName() string {
	return "tcp"
}

// GetMetricsCollector 获取指标收集器
func (f *AdapterFactory) GetMetricsCollector() interfaces.DefaultMetricsCollector {
	return f.metricsCollector
}

// SetMetricsCollector 设置指标收集器
func (f *AdapterFactory) SetMetricsCollector(collector interfaces.DefaultMetricsCollector) {
	f.metricsCollector = collector
}

// 确保实现了interfaces.TCPAdapterFactory接口
var _ interfaces.TCPAdapterFactory = (*AdapterFactory)(nil)
