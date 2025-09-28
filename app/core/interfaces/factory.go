package interfaces

// AdapterFactory 适配器工厂接口
type AdapterFactory interface {
	// CreateRedisAdapter 创建Redis适配器
	CreateRedisAdapter() ProtocolAdapter

	// CreateHttpAdapter 创建HTTP适配器
	CreateHttpAdapter() ProtocolAdapter

	// CreateKafkaAdapter 创建Kafka适配器
	CreateKafkaAdapter() ProtocolAdapter

	// CreateTCPAdapter 创建TCP适配器
	CreateTCPAdapter() ProtocolAdapter

	// CreateUDPAdapter 创建UDP适配器
	CreateUDPAdapter() ProtocolAdapter

	// CreateGRPCAdapter 创建gRPC适配器
	CreateGRPCAdapter() ProtocolAdapter

	// CreateWebSocketAdapter 创建WebSocket适配器
	CreateWebSocketAdapter() ProtocolAdapter
}
