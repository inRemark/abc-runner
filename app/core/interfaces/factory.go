package interfaces

// 协议特定适配器工厂接口 - 接口分离设计
// 符合SOLID原则
// 单一职责：每个工厂只负责一种协议
// 开闭原则：扩展新协议时不需修改现有代码
// 接口隔离：客户端只依赖需要的接口方法
// 依赖倒置：依赖抽象而非具体实现
//
// RedisAdapterFactory Redis适配器工厂接口
type RedisAdapterFactory interface {
	CreateRedisAdapter() ProtocolAdapter
}

// HttpAdapterFactory HTTP适配器工厂接口
type HttpAdapterFactory interface {
	CreateHttpAdapter() ProtocolAdapter
}

// KafkaAdapterFactory Kafka适配器工厂接口
type KafkaAdapterFactory interface {
	CreateKafkaAdapter() ProtocolAdapter
}

// TCPAdapterFactory TCP适配器工厂接口
type TCPAdapterFactory interface {
	CreateTCPAdapter() ProtocolAdapter
}

// UDPAdapterFactory UDP适配器工厂接口
type UDPAdapterFactory interface {
	CreateUDPAdapter() ProtocolAdapter
}

// GRPCAdapterFactory gRPC适配器工厂接口
type GRPCAdapterFactory interface {
	CreateGRPCAdapter() ProtocolAdapter
}

// WebSocketAdapterFactory WebSocket适配器工厂接口
type WebSocketAdapterFactory interface {
	CreateWebSocketAdapter() ProtocolAdapter
}
