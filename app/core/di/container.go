package di

import (
	"go.uber.org/dig"

	"abc-runner/app/adapters/http"
	"abc-runner/app/adapters/kafka"
	"abc-runner/app/adapters/redis"
	"abc-runner/app/core/config"
	"abc-runner/app/core/interfaces"
	"abc-runner/app/core/monitoring"
	"abc-runner/app/core/utils"
)

// Container 依赖注入容器
type Container struct {
	container *dig.Container
}

// NewContainer 创建依赖注入容器
func NewContainer() *Container {
	container := dig.New()

	// 注册核心组件
	container.Provide(config.NewConfigManager)
	container.Provide(utils.NewOperationRegistry)
	container.Provide(utils.NewDefaultKeyGenerator)

	// 注册适配器工厂
	container.Provide(NewAdapterFactory)

	// 注册指标收集器
	container.Provide(NewMetricsCollector)

	// 注册配置源工厂
	container.Provide(NewConfigSourceFactory)

	return &Container{
		container: container,
	}
}

// NewAdapterFactory 创建适配器工厂
func NewAdapterFactory(metricsCollector interfaces.MetricsCollector) interfaces.AdapterFactory {
	return &SimpleAdapterFactory{
		metricsCollector: metricsCollector,
	}
}

// NewConfigSourceFactory 创建配置源工厂
func NewConfigSourceFactory() interfaces.ConfigSourceFactory {
	return &SimpleConfigSourceFactory{}
}

// NewMetricsCollector 创建指标收集器
func NewMetricsCollector() interfaces.MetricsCollector {
	// 返回增强的指标收集器实现
	return monitoring.NewEnhancedMetricsCollector()
}

// Provide 注册依赖
func (c *Container) Provide(constructor interface{}, opts ...dig.ProvideOption) error {
	return c.container.Provide(constructor, opts...)
}

// Invoke 调用函数并注入依赖
func (c *Container) Invoke(function interface{}, opts ...dig.InvokeOption) error {
	return c.container.Invoke(function, opts...)
}

// SimpleAdapterFactory 简单适配器工厂实现
type SimpleAdapterFactory struct {
	metricsCollector interfaces.MetricsCollector
}

// CreateRedisAdapter 创建Redis适配器
func (f *SimpleAdapterFactory) CreateRedisAdapter() interfaces.ProtocolAdapter {
	// 创建Redis适配器并注入指标收集器
	return redis.NewRedisAdapter(f.metricsCollector)
}

// CreateHttpAdapter 创建HTTP适配器
func (f *SimpleAdapterFactory) CreateHttpAdapter() interfaces.ProtocolAdapter {
	// 创建HTTP适配器并注入指标收集器
	return http.NewHttpAdapter(f.metricsCollector)
}

// CreateKafkaAdapter 创建Kafka适配器
func (f *SimpleAdapterFactory) CreateKafkaAdapter() interfaces.ProtocolAdapter {
	// 创建Kafka适配器并注入指标收集器
	return kafka.NewKafkaAdapter(f.metricsCollector)
}

// SimpleConfigSourceFactory 简单配置源工厂实现
type SimpleConfigSourceFactory struct{}

// CreateRedisConfigSource 创建Redis配置源
func (f *SimpleConfigSourceFactory) CreateRedisConfigSource() interfaces.ConfigSource {
	// TODO: 实现Redis配置源创建逻辑
	return nil
}

// CreateHttpConfigSource 创建HTTP配置源
func (f *SimpleConfigSourceFactory) CreateHttpConfigSource() interfaces.ConfigSource {
	// TODO: 实现HTTP配置源创建逻辑
	return nil
}

// CreateKafkaConfigSource 创建Kafka配置源
func (f *SimpleConfigSourceFactory) CreateKafkaConfigSource() interfaces.ConfigSource {
	// TODO: 实现Kafka配置源创建逻辑
	return nil
}
