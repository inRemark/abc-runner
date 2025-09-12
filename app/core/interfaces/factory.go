package interfaces

import (
	"context"
)

// AdapterFactory 适配器工厂接口
type AdapterFactory interface {
	// CreateRedisAdapter 创建Redis适配器
	CreateRedisAdapter() ProtocolAdapter

	// CreateHttpAdapter 创建HTTP适配器
	CreateHttpAdapter() ProtocolAdapter

	// CreateKafkaAdapter 创建Kafka适配器
	CreateKafkaAdapter() ProtocolAdapter
}

// ConfigSourceFactory 配置源工厂接口
type ConfigSourceFactory interface {
	// CreateRedisConfigSource 创建Redis配置源
	CreateRedisConfigSource() ConfigSource

	// CreateHttpConfigSource 创建HTTP配置源
	CreateHttpConfigSource() ConfigSource

	// CreateKafkaConfigSource 创建Kafka配置源
	CreateKafkaConfigSource() ConfigSource
}

// ConfigSource 配置源接口
type ConfigSource interface {
	// Load 加载配置
	Load(ctx context.Context) (Config, error)

	// GetProtocol 获取协议名称
	GetProtocol() string
}

// SimpleAdapterFactory 简单适配器工厂实现
type SimpleAdapterFactory struct {
	metricsCollector MetricsCollector
}

// NewSimpleAdapterFactory 创建简单适配器工厂
func NewSimpleAdapterFactory(metricsCollector MetricsCollector) *SimpleAdapterFactory {
	return &SimpleAdapterFactory{
		metricsCollector: metricsCollector,
	}
}

// CreateRedisAdapter 创建Redis适配器
func (f *SimpleAdapterFactory) CreateRedisAdapter() ProtocolAdapter {
	return nil // TODO: 实现
}

// CreateHttpAdapter 创建HTTP适配器
func (f *SimpleAdapterFactory) CreateHttpAdapter() ProtocolAdapter {
	return nil // TODO: 实现
}

// CreateKafkaAdapter 创建Kafka适配器
func (f *SimpleAdapterFactory) CreateKafkaAdapter() ProtocolAdapter {
	return nil // TODO: 实现
}
