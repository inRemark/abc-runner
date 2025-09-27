package kafka

import (
	"context"
	"fmt"
	"sync"
	"time"

	kafkaConfig "abc-runner/app/adapters/kafka/config"
	"abc-runner/app/adapters/kafka/connection"
	"abc-runner/app/adapters/kafka/operations"
	"abc-runner/app/core/base"
	"abc-runner/app/core/interfaces"

	"github.com/segmentio/kafka-go"
)

// KafkaAdapter Kafka协议适配器实现
type KafkaAdapter struct {
	*base.BaseAdapter

	// 连接管理
	connPool *connection.ConnectionPool
	config   *kafkaConfig.KafkaAdapterConfig

	// 不再需要协议特定的指标收集器，完全使用通用接口

	// 操作执行器
	producerOps      *operations.ProducerOperations
	consumerOps      *operations.ConsumerOperations
	operationFactory *operations.KafkaOperationFactory

	// 同步控制
	mutex sync.RWMutex
}

// NewKafkaAdapter 创建Kafka适配器
func NewKafkaAdapter(metricsCollector interfaces.MetricsCollector) *KafkaAdapter {
	adapter := &KafkaAdapter{
		BaseAdapter: base.NewBaseAdapter("kafka"),
	}

	// 使用新架构：只接受通用接口，不再有专用收集器后备
	if metricsCollector == nil {
		return nil // 新架构要求必须传入MetricsCollector
	}
	adapter.SetMetricsCollector(metricsCollector)

	return adapter
}

// Connect 初始化连接
func (k *KafkaAdapter) Connect(ctx context.Context, config interfaces.Config) error {
	k.mutex.Lock()
	defer k.mutex.Unlock()

	// 验证并转换配置
	kafkaConfig, ok := config.(*kafkaConfig.KafkaAdapterConfig)
	if !ok {
		return fmt.Errorf("invalid config type for Kafka adapter")
	}

	if err := k.ValidateConfig(config); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	k.config = kafkaConfig
	k.SetConfig(config)

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

	k.SetConnected(true)
	k.UpdateMetric("connected_at", time.Now())
	k.UpdateMetric("brokers", k.config.Brokers)

	// 初始化操作执行器
	// 使用新架构：直接传入通用指标收集器接口
	k.producerOps = operations.NewProducerOperations(k.connPool, k.GetMetricsCollector())
	k.consumerOps = operations.NewConsumerOperations(k.connPool, k.GetMetricsCollector())
	k.operationFactory = operations.NewKafkaOperationFactory(k.config)

	return nil
}

// Execute 执行操作并返回结果
func (k *KafkaAdapter) Execute(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	if !k.IsConnected() {
		return nil, fmt.Errorf("kafka adapter not connected")
	}

	startTime := time.Now()

	// 根据操作类型选择执行方法
	result, err := k.executeOperation(ctx, operation)
	if result != nil {
		result.Duration = time.Since(startTime)

		// 记录操作到指标收集器
		if metricsCollector := k.GetMetricsCollector(); metricsCollector != nil {
			metricsCollector.RecordOperation(result)
		}
	}

	return result, err
}

// executeOperation 执行具体操作
func (k *KafkaAdapter) executeOperation(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	switch operation.Type {
	case "produce", "produce_message":
		return k.executeProduceMessage(ctx, operation)
	case "produce_batch":
		return k.executeProduceBatch(ctx, operation)
	case "consume", "consume_message":
		return k.executeConsumeMessage(ctx, operation)
	case "consume_batch":
		return k.executeConsumeBatch(ctx, operation)
	case "create_topic":
		return k.executeCreateTopic(ctx, operation)
	case "delete_topic":
		return k.executeDeleteTopic(ctx, operation)
	case "list_topics":
		return k.executeListTopics(ctx, operation)
	case "describe_consumer_groups":
		return k.executeDescribeConsumerGroups(ctx, operation)
	default:
		return &interfaces.OperationResult{
			Success: false,
			IsRead:  false,
			Error:   fmt.Errorf("unsupported operation type: %s", operation.Type),
		}, fmt.Errorf("unsupported operation type: %s", operation.Type)
	}
}

// executeProduceMessage 执行单条消息生产
func (k *KafkaAdapter) executeProduceMessage(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	return k.producerOps.ExecuteProduceMessage(ctx, operation)
}

// executeProduceBatch 执行批量消息生产
func (k *KafkaAdapter) executeProduceBatch(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	return k.producerOps.ExecuteProduceBatch(ctx, operation)
}

// executeConsumeMessage 执行单条消息消费
func (k *KafkaAdapter) executeConsumeMessage(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	return k.consumerOps.ExecuteConsumeMessage(ctx, operation)
}

// executeConsumeBatch 执行批量消息消费
func (k *KafkaAdapter) executeConsumeBatch(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	return k.consumerOps.ExecuteConsumeBatch(ctx, operation)
}

// executeCreateTopic 执行创建主题
func (k *KafkaAdapter) executeCreateTopic(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	startTime := time.Now()

	// TODO: 实现主题创建逻辑
	// 这需要使用Kafka Admin API

	return &interfaces.OperationResult{
		Success:  false,
		Duration: time.Since(startTime),
		IsRead:   false,
		Error:    fmt.Errorf("create topic operation not implemented yet"),
	}, fmt.Errorf("create topic operation not implemented yet")
}

// executeDeleteTopic 执行删除主题
func (k *KafkaAdapter) executeDeleteTopic(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	startTime := time.Now()

	// TODO: 实现主题删除逻辑

	return &interfaces.OperationResult{
		Success:  false,
		Duration: time.Since(startTime),
		IsRead:   false,
		Error:    fmt.Errorf("delete topic operation not implemented yet"),
	}, fmt.Errorf("delete topic operation not implemented yet")
}

// executeListTopics 执行列出主题
func (k *KafkaAdapter) executeListTopics(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	startTime := time.Now()

	// TODO: 实现主题列表查询逻辑

	return &interfaces.OperationResult{
		Success:  false,
		Duration: time.Since(startTime),
		IsRead:   true,
		Error:    fmt.Errorf("list topics operation not implemented yet"),
	}, fmt.Errorf("list topics operation not implemented yet")
}

// executeDescribeConsumerGroups 执行描述消费者组
func (k *KafkaAdapter) executeDescribeConsumerGroups(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	startTime := time.Now()

	// TODO: 实现消费者组描述逻辑

	return &interfaces.OperationResult{
		Success:  false,
		Duration: time.Since(startTime),
		IsRead:   true,
		Error:    fmt.Errorf("describe consumer groups operation not implemented yet"),
	}, fmt.Errorf("describe consumer groups operation not implemented yet")
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

	k.SetConnected(false)
	k.UpdateMetric("disconnected_at", time.Now())

	return nil
}

// HealthCheck 健康检查
func (k *KafkaAdapter) HealthCheck(ctx context.Context) error {
	if !k.IsConnected() {
		return fmt.Errorf("adapter not connected")
	}

	return k.testConnection(ctx)
}

// GetProtocolMetrics 获取Kafka特定指标
func (k *KafkaAdapter) GetProtocolMetrics() map[string]interface{} {
	baseMetrics := k.BaseAdapter.GetProtocolMetrics()

	// 新架构：只使用通用指标收集器
	result := make(map[string]interface{})
	for key, value := range baseMetrics {
		result[key] = value
	}

	return result
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
func (k *KafkaAdapter) ProduceMessage(topic string, message *operations.Message) (*operations.ProduceResult, error) {
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
	if produceResult, ok := result.Value.(*operations.ProduceResult); ok {
		return produceResult, nil
	}

	return nil, fmt.Errorf("unexpected result type from produce operation")
}

// ProduceBatch 批量生产消息
func (k *KafkaAdapter) ProduceBatch(topic string, messages []*operations.Message) (*operations.BatchResult, error) {
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
	if batchResult, ok := result.Value.(*operations.BatchResult); ok {
		return batchResult, nil
	}

	return nil, fmt.Errorf("unexpected result type from batch produce operation")
}

// CreateOperation 创建操作（便捷方法）
func (k *KafkaAdapter) CreateOperation(params map[string]interface{}) (interfaces.Operation, error) {
	if k.operationFactory == nil {
		return interfaces.Operation{}, fmt.Errorf("operation factory not initialized")
	}

	return k.operationFactory.CreateOperation(params)
}

// GetOperationFactory 获取操作工厂
func (k *KafkaAdapter) GetOperationFactory() interfaces.OperationFactory {
	return k.operationFactory
}

// GetMetricsCollector 获取指标收集器
func (k *KafkaAdapter) GetMetricsCollector() interfaces.MetricsCollector {
	// 新架构：只返回BaseAdapter的通用指标收集器
	return k.BaseAdapter.GetMetricsCollector()
}
