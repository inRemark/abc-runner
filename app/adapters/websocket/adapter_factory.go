package websocket

import (
	"abc-runner/app/core/interfaces"
)

// AdapterFactory WebSocket适配器工厂 - 统一接口实现
type AdapterFactory struct {
	metricsCollector interfaces.DefaultMetricsCollector
}

// NewAdapterFactory 创建WebSocket适配器工厂
func NewAdapterFactory(metricsCollector interfaces.DefaultMetricsCollector) *AdapterFactory {
	if metricsCollector == nil {
		panic("metricsCollector cannot be nil - dependency injection required")
	}

	return &AdapterFactory{
		metricsCollector: metricsCollector,
	}
}

// CreateWebSocketAdapter 创建WebSocket适配器 (实现WebSocketAdapterFactory接口)
func (f *AdapterFactory) CreateWebSocketAdapter() interfaces.ProtocolAdapter {
	return NewWebSocketAdapter(f.metricsCollector)
}

// GetProtocolName 获取支持的协议名称
func (f *AdapterFactory) GetProtocolName() string {
	return "websocket"
}

// GetMetricsCollector 获取指标收集器
func (f *AdapterFactory) GetMetricsCollector() interfaces.DefaultMetricsCollector {
	return f.metricsCollector
}

// SetMetricsCollector 设置指标收集器
func (f *AdapterFactory) SetMetricsCollector(collector interfaces.DefaultMetricsCollector) {
	f.metricsCollector = collector
}

// 确保实现了相关接口
var (
	_ interfaces.ProtocolAdapter = (*WebSocketAdapter)(nil)
	_                            = (*AdapterFactory)(nil) // WebSocket工厂接口检查
)
