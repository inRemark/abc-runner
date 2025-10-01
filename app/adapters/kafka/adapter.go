package kafka

import (
	"context"
	"fmt"
	"sync"
	"time"

	kafkaConfig "abc-runner/app/adapters/kafka/config"
	"abc-runner/app/adapters/kafka/connection"
	"abc-runner/app/adapters/kafka/operations"

	"abc-runner/app/core/interfaces"

	"github.com/segmentio/kafka-go"
)

// KafkaAdapter Kafka协议适配器实现 - 遵循统一架构模式
// 职责：连接管理、状态维护、健康检查
type KafkaAdapter struct {
	// 连接管理
	connPool *connection.ConnectionPool
	config   *kafkaConfig.KafkaAdapterConfig

	// 操作执行器
	kafkaOperations *operations.KafkaOperations

	// 指标收集器
	metricsCollector interfaces.DefaultMetricsCollector

	// 状态管理
	isConnected bool
	mutex       sync.RWMutex

	// 统计信息
	startTime time.Time
}

// NewKafkaAdapter 创建Kafka适配器
func NewKafkaAdapter(metricsCollector interfaces.DefaultMetricsCollector) *KafkaAdapter {
	if metricsCollector == nil {
		panic("metricsCollector cannot be nil - dependency injection required")
	}

	return &KafkaAdapter{
		metricsCollector: metricsCollector,
		startTime:        time.Now(),
		isConnected:      false,
	}
}

// Connect 初始化连接
func (k *KafkaAdapter) Connect(ctx context.Context, config interfaces.Config) error {
	k.mutex.Lock()
	defer k.mutex.Unlock()

	// 验证并转换配置
	kafkaConfig, ok := config.(*kafkaConfig.KafkaAdapterConfig)
	if !ok {
		return fmt.Errorf("invalid config type for Kafka adapter: expected *kafkaConfig.KafkaAdapterConfig, got %T", config)
	}

	// 验证配置
	if err := kafkaConfig.Validate(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	k.config = kafkaConfig

	// 初始化连接池
	poolConfig := connection.PoolConfig{
		MaxConnections:    kafkaConfig.Performance.ConnectionPoolSize,
		ProducerPoolSize:  kafkaConfig.Performance.ProducerPoolSize,
		ConsumerPoolSize:  kafkaConfig.Performance.ConsumerPoolSize,
		ConnectionTimeout: kafkaConfig.Benchmark.GetTimeout(),
	}

	var err error
	k.connPool, err = connection.NewConnectionPool(kafkaConfig, poolConfig)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}

	// 测试连接
	if err := k.testConnection(ctx); err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}

	// 创建Kafka操作执行器
	k.kafkaOperations = operations.NewKafkaOperations(k.connPool, k.config, k.metricsCollector)

	k.isConnected = true

	return nil
}

// Execute 执行Kafka操作 - 使用执行器处理
func (k *KafkaAdapter) Execute(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	if !k.isConnected {
		return nil, fmt.Errorf("kafka adapter not connected")
	}

	// 委托给Kafka操作执行器处理
	return k.kafkaOperations.ExecuteOperation(ctx, operation)
}

// Close 关闭连接
func (k *KafkaAdapter) Close() error {
	k.mutex.Lock()
	defer k.mutex.Unlock()

	if k.connPool != nil {
		if err := k.connPool.Close(); err != nil {
			return fmt.Errorf("failed to close connection pool: %w", err)
		}
	}

	k.isConnected = false

	return nil
}

// HealthCheck 健康检查
func (k *KafkaAdapter) HealthCheck(ctx context.Context) error {
	if !k.isConnected {
		return fmt.Errorf("adapter not connected")
	}

	return k.testConnection(ctx)
}

// GetProtocolMetrics 获取Kafka特定指标
func (k *KafkaAdapter) GetProtocolMetrics() map[string]interface{} {
	k.mutex.RLock()
	defer k.mutex.RUnlock()

	metrics := map[string]interface{}{
		"protocol":       "kafka",
		"is_connected":   k.isConnected,
		"uptime_seconds": time.Since(k.startTime).Seconds(),
	}

	// 添加连接池统计信息
	if k.connPool != nil {
		poolStats := k.connPool.Stats()
		metrics["connection_pool"] = poolStats
	}

	// 添加配置信息
	if k.config != nil {
		metrics["config"] = map[string]interface{}{
			"brokers":              k.config.Brokers,
			"producer_pool_size":   k.config.Performance.ProducerPoolSize,
			"consumer_pool_size":   k.config.Performance.ConsumerPoolSize,
			"connection_pool_size": k.config.Performance.ConnectionPoolSize,
		}
	}

	return metrics
}

// testConnection 测试连接
func (k *KafkaAdapter) testConnection(ctx context.Context) error {
	// 尝试获取一个生产者进行连接测试
	producer, err := k.connPool.GetProducer()
	if err != nil {
		return fmt.Errorf("failed to get producer for connection test: %w", err)
	}
	defer k.connPool.ReturnProducer(producer)

	// 简单的连接测试 - 尝试写入一个测试消息到不存在的topic（这会失败但能验证连接）
	testMsg := kafka.Message{
		Topic: "__connection_test__",
		Key:   []byte("test"),
		Value: []byte("connection_test"),
	}

	// 使用短超时进行测试
	testCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// 这个操作可能会失败，但如果是因为topic不存在等可预期的错误，说明连接是正常的
	err = producer.WriteMessages(testCtx, testMsg)
	if err != nil {
		// 检查是否是可接受的错误（如topic不存在）
		if !k.isAcceptableConnectionTestError(err) {
			return fmt.Errorf("connection test failed with unexpected error: %w", err)
		}
	}

	return nil
}

// isAcceptableConnectionTestError 检查是否为可接受的连接测试错误
func (k *KafkaAdapter) isAcceptableConnectionTestError(err error) bool {
	if err == nil {
		return true
	}

	errStr := err.Error()
	// 这些错误表明连接是正常的，只是操作本身有问题
	acceptableErrors := []string{
		"UNKNOWN_TOPIC_OR_PARTITION",
		"topic not found",
		"InvalidTopicException",
	}

	for _, acceptableErr := range acceptableErrors {
		if contains(errStr, acceptableErr) {
			return true
		}
	}

	return false
}

// contains 检查字符串是否包含子字符串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && indexOf(s, substr) >= 0))
}

// indexOf 查找子字符串位置
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// Kafka特定操作接口

// ProduceMessage 生产单条消息
func (k *KafkaAdapter) ProduceMessage(topic string, message *Message) (*ProduceResult, error) {
	ctx := context.Background()
	operation := interfaces.Operation{
		Type:  "produce_message",
		Key:   message.Key,
		Value: message.Value,
		Params: map[string]interface{}{
			"topic":     topic,
			"headers":   message.Headers,
			"partition": message.Partition,
		},
	}

	result, err := k.Execute(ctx, operation)
	if err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, result.Error
	}

	// 从结果中提取ProduceResult
	if produceResult, ok := result.Value.(*ProduceResult); ok {
		return produceResult, nil
	}

	return nil, fmt.Errorf("unexpected result type from produce operation")
}

// ProduceBatch 批量生产消息
func (k *KafkaAdapter) ProduceBatch(topic string, messages []*Message) (*BatchResult, error) {
	ctx := context.Background()
	operation := interfaces.Operation{
		Type: "produce_batch",
		Params: map[string]interface{}{
			"topic":    topic,
			"messages": messages,
		},
	}

	result, err := k.Execute(ctx, operation)
	if err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, result.Error
	}

	// 从结果中提取BatchResult
	if batchResult, ok := result.Value.(*BatchResult); ok {
		return batchResult, nil
	}

	return nil, fmt.Errorf("unexpected result type from batch produce operation")
}

// CreateOperation 创建操作（便捷方法）
func (k *KafkaAdapter) CreateOperation(params map[string]interface{}) (interfaces.Operation, error) {
	// 直接创建操作，不依赖外部工厂
	operationType := "produce"
	if opType, exists := params["type"]; exists {
		if opTypeStr, ok := opType.(string); ok {
			operationType = opTypeStr
		}
	}

	// 创建操作
	operation := interfaces.Operation{
		Type:     operationType,
		Key:      fmt.Sprintf("kafka_%s_%d", operationType, time.Now().UnixNano()),
		Value:    "default_kafka_message",
		Params:   params,
		Metadata: map[string]string{"protocol": "kafka"},
	}

	return operation, nil
}

// GetMetricsCollector 获取指标收集器
func (k *KafkaAdapter) GetMetricsCollector() interfaces.DefaultMetricsCollector {
	return k.metricsCollector
}

// GetProtocolName 获取协议名称
func (k *KafkaAdapter) GetProtocolName() string {
	return "kafka"
}
