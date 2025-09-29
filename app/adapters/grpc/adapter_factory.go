package grpc

import (
	"abc-runner/app/adapters/grpc/config"
	"abc-runner/app/core/interfaces"
	"fmt"
)

// AdapterFactory gRPC适配器工厂
type AdapterFactory struct {
	metricsCollector interfaces.DefaultMetricsCollector
	config           *config.GRPCConfig
}

// NewAdapterFactory 创建gRPC适配器工厂
func NewAdapterFactory(metricsCollector interfaces.DefaultMetricsCollector) *AdapterFactory {
	return &AdapterFactory{
		metricsCollector: metricsCollector,
		config:           config.NewDefaultGRPCConfig(),
	}
}

// NewAdapterFactoryWithConfig 创建带配置的gRPC适配器工厂
func NewAdapterFactoryWithConfig(metricsCollector interfaces.DefaultMetricsCollector, cfg *config.GRPCConfig) *AdapterFactory {
	return &AdapterFactory{
		metricsCollector: metricsCollector,
		config:           cfg,
	}
}

// CreateGRPCAdapter 创建gRPC适配器 (实现GRPCAdapterFactory接口)
func (f *AdapterFactory) CreateGRPCAdapter() interfaces.ProtocolAdapter {
	if f.metricsCollector == nil {
		return nil
	}

	adapter := NewGRPCAdapter(f.metricsCollector)
	return adapter
}

// GetProtocolName 获取支持的协议名称
func (f *AdapterFactory) GetProtocolName() string {
	return "grpc"
}

// GetConfig 获取配置
func (f *AdapterFactory) GetConfig() *config.GRPCConfig {
	return f.config
}

// SetConfig 设置配置
func (f *AdapterFactory) SetConfig(cfg *config.GRPCConfig) {
	f.config = cfg
}

// GetMetricsCollector 获取指标收集器
func (f *AdapterFactory) GetMetricsCollector() interfaces.DefaultMetricsCollector {
	return f.metricsCollector
}

// SetMetricsCollector 设置指标收集器
func (f *AdapterFactory) SetMetricsCollector(collector interfaces.DefaultMetricsCollector) {
	f.metricsCollector = collector
}

// ValidateConfig 验证配置
func (f *AdapterFactory) ValidateConfig() error {
	if f.config == nil {
		return fmt.Errorf("gRPC config is nil")
	}
	return f.config.Validate()
}

// 确保实现了interfaces.GRPCAdapterFactory接口
var _ interfaces.GRPCAdapterFactory = (*AdapterFactory)(nil)
