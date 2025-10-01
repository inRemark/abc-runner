package kafka

import (
	"context"
	"fmt"
	"sync"
	"time"

	kafkaConfig "abc-runner/app/adapters/kafka/config"
	"abc-runner/app/adapters/kafka/connection"

	"abc-runner/app/core/interfaces"

	"github.com/segmentio/kafka-go"
)

// KafkaAdapter Kafka协议适配器实现
type KafkaAdapter struct {
	// 连接管理
	connPool *connection.ConnectionPool
	config   *kafkaConfig.KafkaAdapterConfig

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

	k.isConnected = true

	return nil
}

// Execute 执行操作并返回结果
// Execute 执行Kafka操作
func (k *KafkaAdapter) Execute(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	if !k.isConnected {
		return nil, fmt.Errorf("kafka adapter not connected")
	}

	startTime := time.Now()

	// 根据操作类型选择执行方法
	result, err := k.executeOperation(ctx, operation)
	if result != nil {
		result.Duration = time.Since(startTime)

		// 注意：不要在这里调用 k.metricsCollector.Record(result)
		// 因为执行引擎会负责记录指标，避免重复计数
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
	return k.executeKafkaOperation(ctx, operation)
}

// executeProduceBatch 执行批量消息生产
func (k *KafkaAdapter) executeProduceBatch(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	return k.executeKafkaOperation(ctx, operation)
}

// executeConsumeMessage 执行单条消息消费
func (k *KafkaAdapter) executeConsumeMessage(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	return k.executeKafkaOperation(ctx, operation)
}

// executeConsumeBatch 执行批量消息消费
func (k *KafkaAdapter) executeConsumeBatch(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	return k.executeKafkaOperation(ctx, operation)
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

// === 架构兼容性方法，与 operations 系统集成 ===

// GetSupportedOperations 获取支持的操作类型（架构兼容性）
func (k *KafkaAdapter) GetSupportedOperations() []string {
	return []string{
		"produce", "produce_message", "produce_batch",
		"consume", "consume_message", "consume_batch",
		"create_topic", "delete_topic", "list_topics",
		"describe_consumer_groups", "get_metadata",
	}
}

// ValidateOperation 验证操作是否受支持（架构兼容性）
func (k *KafkaAdapter) ValidateOperation(operationType string) error {
	supportedOps := k.GetSupportedOperations()
	for _, op := range supportedOps {
		if op == operationType {
			return nil
		}
	}
	return fmt.Errorf("unsupported operation type: %s", operationType)
}

// GetOperationMetadata 获取操作元数据（架构兼容性）
func (k *KafkaAdapter) GetOperationMetadata(operationType string) map[string]interface{} {
	metadata := map[string]interface{}{
		"protocol":       "kafka",
		"adapter_type":   "kafka_adapter",
		"operation_type": operationType,
		"is_read":        k.isReadOperation(operationType),
	}

	if k.config != nil {
		metadata["brokers"] = k.config.Brokers
		metadata["producer_pool_size"] = k.config.Performance.ProducerPoolSize
		metadata["consumer_pool_size"] = k.config.Performance.ConsumerPoolSize
	}

	return metadata
}

// isReadOperation 判断是否为读操作
func (k *KafkaAdapter) isReadOperation(operationType string) bool {
	readOps := []string{"consume", "consume_message", "consume_batch", "list_topics", "describe_consumer_groups", "get_metadata"}
	for _, readOp := range readOps {
		if readOp == operationType {
			return true
		}
	}
	return false
}

// executeKafkaOperation 执行具体的Kafka操作
func (k *KafkaAdapter) executeKafkaOperation(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	startTime := time.Now()
	result := &interfaces.OperationResult{
		IsRead: k.isReadOperation(operation.Type),
	}

	var err error
	switch operation.Type {
	case "produce", "produce_message":
		result.Value, err = k.executeProduceOperation(ctx, operation)
	case "produce_batch":
		result.Value, err = k.executeProduceBatchOperation(ctx, operation)
	case "consume", "consume_message":
		result.Value, err = k.executeConsumeOperation(ctx, operation)
	case "consume_batch":
		result.Value, err = k.executeConsumeBatchOperation(ctx, operation)
	default:
		err = fmt.Errorf("unsupported operation type: %s", operation.Type)
	}

	result.Duration = time.Since(startTime)
	result.Success = err == nil
	result.Error = err

	// 设置结果元数据
	result.Metadata = map[string]interface{}{
		"operation_type": operation.Type,
		"topic":          operation.Params["topic"],
		"duration_ms":    result.Duration.Milliseconds(),
	}

	return result, nil
}

// executeProduceOperation 执行生产操作
func (k *KafkaAdapter) executeProduceOperation(ctx context.Context, operation interfaces.Operation) (interface{}, error) {
	// 获取生产者
	producer, err := k.connPool.GetProducer()
	if err != nil {
		return nil, fmt.Errorf("failed to get producer: %w", err)
	}
	defer k.connPool.ReturnProducer(producer)

	// 构造消息
	topic := ""
	if topicParam, exists := operation.Params["topic"]; exists {
		if topicStr, ok := topicParam.(string); ok {
			topic = topicStr
		}
	}
	if topic == "" {
		topic = k.config.Benchmark.DefaultTopic
	}

	msg := kafka.Message{
		Topic: topic,
		Key:   []byte(operation.Key),
		Value: []byte(fmt.Sprintf("%v", operation.Value)),
	}

	// 设置分区
	if partitionParam, exists := operation.Params["partition"]; exists {
		if partition, ok := partitionParam.(int); ok {
			msg.Partition = partition
		}
	}

	// 发送消息
	err = producer.WriteMessages(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to produce message: %w", err)
	}

	return &ProduceResult{
		Partition: int32(msg.Partition),
		Offset:    0, // kafka-go不直接返回offset
		Timestamp: time.Now(),
		Duration:  0, // 将在上层设置
	}, nil
}

// executeProduceBatchOperation 执行批量生产操作
func (k *KafkaAdapter) executeProduceBatchOperation(ctx context.Context, operation interfaces.Operation) (interface{}, error) {
	// TODO: 实现批量生产逻辑
	return k.executeProduceOperation(ctx, operation)
}

// executeConsumeOperation 执行消费操作
func (k *KafkaAdapter) executeConsumeOperation(ctx context.Context, operation interfaces.Operation) (interface{}, error) {
	// 获取消费者
	consumer, err := k.connPool.GetConsumer()
	if err != nil {
		return nil, fmt.Errorf("failed to get consumer: %w", err)
	}
	defer k.connPool.ReturnConsumer(consumer)

	// 读取消息
	msg, err := consumer.ReadMessage(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to consume message: %w", err)
	}

	return map[string]interface{}{
		"topic":     msg.Topic,
		"key":       string(msg.Key),
		"value":     string(msg.Value),
		"partition": msg.Partition,
		"offset":    msg.Offset,
		"timestamp": msg.Time,
	}, nil
}

// executeConsumeBatchOperation 执行批量消费操作
func (k *KafkaAdapter) executeConsumeBatchOperation(ctx context.Context, operation interfaces.Operation) (interface{}, error) {
	// TODO: 实现批量消费逻辑
	return k.executeConsumeOperation(ctx, operation)
}
