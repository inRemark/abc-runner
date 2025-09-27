package interfaces

// AdapterFactory 适配器工厂接口
type AdapterFactory interface {
	// CreateRedisAdapter 创建Redis适配器
	CreateRedisAdapter() ProtocolAdapter

	// CreateHttpAdapter 创建HTTP适配器
	CreateHttpAdapter() ProtocolAdapter

	// CreateKafkaAdapter 创建Kafka适配器
	CreateKafkaAdapter() ProtocolAdapter
}
