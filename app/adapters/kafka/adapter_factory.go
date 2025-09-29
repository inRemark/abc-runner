package kafka

import (
	"abc-runner/app/core/interfaces"
)

// AdapterFactory Kafka适配器工厂
type AdapterFactory struct {
	metricsCollector interfaces.DefaultMetricsCollector
}

// NewAdapterFactory 创建Kafka适配器工厂
func NewAdapterFactory(metricsCollector interfaces.DefaultMetricsCollector) interfaces.KafkaAdapterFactory {
	return &AdapterFactory{
		metricsCollector: metricsCollector,
	}
}

// CreateKafkaAdapter 创建Kafka适配器1
func (f *AdapterFactory) CreateKafkaAdapter() interfaces.ProtocolAdapter {
	return NewKafkaAdapter(f.metricsCollector)
}
