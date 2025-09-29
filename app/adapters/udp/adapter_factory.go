package udp

import (
	"abc-runner/app/core/interfaces"
)

// AdapterFactory UDP适配器工厂
type AdapterFactory struct {
	metricsCollector interfaces.DefaultMetricsCollector
}

// NewAdapterFactory 创建UDP适配器工厂
func NewAdapterFactory(metricsCollector interfaces.DefaultMetricsCollector) *AdapterFactory {
	return &AdapterFactory{
		metricsCollector: metricsCollector,
	}
}

// CreateUDPAdapter 创建UDP适配器 (实现UDPAdapterFactory接口)
func (f *AdapterFactory) CreateUDPAdapter() interfaces.ProtocolAdapter {
	if f.metricsCollector == nil {
		return nil
	}

	adapter := NewUDPAdapter(f.metricsCollector)
	return adapter
}

// GetProtocolName 获取支持的协议名称
func (f *AdapterFactory) GetProtocolName() string {
	return "udp"
}

// GetMetricsCollector 获取指标收集器
func (f *AdapterFactory) GetMetricsCollector() interfaces.DefaultMetricsCollector {
	return f.metricsCollector
}

// SetMetricsCollector 设置指标收集器
func (f *AdapterFactory) SetMetricsCollector(collector interfaces.DefaultMetricsCollector) {
	f.metricsCollector = collector
}

// 确保实现了interfaces.UDPAdapterFactory接口
var _ interfaces.UDPAdapterFactory = (*AdapterFactory)(nil)
