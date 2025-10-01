package operations

import (
	"context"
	"fmt"
	"time"

	kafkaConfig "abc-runner/app/adapters/kafka/config"
	"abc-runner/app/adapters/kafka/connection"
	"abc-runner/app/core/interfaces"
)

// KafkaExecutor Kafka操作执行器 - 遵循统一架构模式
type KafkaExecutor struct {
	connPool         *connection.ConnectionPool
	config           *kafkaConfig.KafkaAdapterConfig
	metricsCollector interfaces.DefaultMetricsCollector
	producer         *ProducerExecutor
	consumer         *ConsumerExecutor
}

// NewKafkaExecutor 创建Kafka操作执行器
func NewKafkaExecutor(
	connPool *connection.ConnectionPool,
	config *kafkaConfig.KafkaAdapterConfig,
	metricsCollector interfaces.DefaultMetricsCollector,
) *KafkaExecutor {
	return &KafkaExecutor{
		connPool:         connPool,
		config:           config,
		metricsCollector: metricsCollector,
		producer:         NewProducerExecutor(connPool, metricsCollector),
		consumer:         NewConsumerExecutor(connPool, metricsCollector),
	}
}

// ExecuteOperation 执行Kafka操作 - 统一操作入口
func (k *KafkaExecutor) ExecuteOperation(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	startTime := time.Now()
	result := &interfaces.OperationResult{
		IsRead:   k.isReadOperation(operation.Type),
		Metadata: make(map[string]interface{}),
	}

	var opErr error
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
		opErr = k.executeCreateTopic(ctx, operation, result)
	case "delete_topic":
		opErr = k.executeDeleteTopic(ctx, operation, result)
	case "list_topics":
		opErr = k.executeListTopics(ctx, operation, result)
	case "describe_consumer_groups":
		opErr = k.executeDescribeConsumerGroups(ctx, operation, result)
	default:
		opErr = fmt.Errorf("unsupported operation type: %s", operation.Type)
	}

	if result.Duration == 0 {
		result.Success = opErr == nil
		result.Error = opErr
		result.Duration = time.Since(startTime)
	}

	// 添加操作特定元数据
	for k, v := range operation.Metadata {
		result.Metadata[k] = v
	}
	result.Metadata["protocol"] = "kafka"
	result.Metadata["operation_type"] = operation.Type
	result.Metadata["execution_time_ms"] = float64(result.Duration.Nanoseconds()) / 1e6
	result.Metadata["timestamp"] = time.Now()

	// 添加Kafka特定配置信息
	if k.config != nil {
		result.Metadata["default_topic"] = k.config.Benchmark.DefaultTopic
		result.Metadata["consumer_group"] = k.config.Consumer.GroupID
		result.Metadata["bootstrap_servers"] = k.config.Brokers
	}

	return result, opErr
}

// executeProduceMessage 执行单条消息生产
func (k *KafkaExecutor) executeProduceMessage(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	if k.producer == nil {
		return &interfaces.OperationResult{
			Success: false,
			Error:   fmt.Errorf("producer not initialized"),
		}, fmt.Errorf("producer not initialized")
	}
	return k.producer.ExecuteProduceMessage(ctx, operation)
}

// executeProduceBatch 执行批量消息生产
func (k *KafkaExecutor) executeProduceBatch(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	if k.producer == nil {
		return &interfaces.OperationResult{
			Success: false,
			Error:   fmt.Errorf("producer not initialized"),
		}, fmt.Errorf("producer not initialized")
	}
	return k.producer.ExecuteProduceBatch(ctx, operation)
}

// executeConsumeMessage 执行单条消息消费
func (k *KafkaExecutor) executeConsumeMessage(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	if k.consumer == nil {
		return &interfaces.OperationResult{
			Success: false,
			Error:   fmt.Errorf("consumer not initialized"),
		}, fmt.Errorf("consumer not initialized")
	}
	return k.consumer.ExecuteConsumeMessage(ctx, operation)
}

// executeConsumeBatch 执行批量消息消费
func (k *KafkaExecutor) executeConsumeBatch(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	if k.consumer == nil {
		return &interfaces.OperationResult{
			Success: false,
			Error:   fmt.Errorf("consumer not initialized"),
		}, fmt.Errorf("consumer not initialized")
	}
	return k.consumer.ExecuteConsumeBatch(ctx, operation)
}

// executeCreateTopic 执行创建主题
func (k *KafkaExecutor) executeCreateTopic(ctx context.Context, operation interfaces.Operation, result *interfaces.OperationResult) error {
	// TODO: 实现主题创建逻辑
	// 这需要使用Kafka Admin API
	result.Value = fmt.Sprintf("Topic creation for operation: %s", operation.Key)
	result.Metadata["admin_operation"] = "create_topic"
	return fmt.Errorf("create topic operation not implemented yet")
}

// executeDeleteTopic 执行删除主题
func (k *KafkaExecutor) executeDeleteTopic(ctx context.Context, operation interfaces.Operation, result *interfaces.OperationResult) error {
	// TODO: 实现主题删除逻辑
	result.Value = fmt.Sprintf("Topic deletion for operation: %s", operation.Key)
	result.Metadata["admin_operation"] = "delete_topic"
	return fmt.Errorf("delete topic operation not implemented yet")
}

// executeListTopics 执行列出主题
func (k *KafkaExecutor) executeListTopics(ctx context.Context, operation interfaces.Operation, result *interfaces.OperationResult) error {
	// TODO: 实现主题列表查询逻辑
	result.Value = []string{} // 空列表
	result.Metadata["admin_operation"] = "list_topics"
	return fmt.Errorf("list topics operation not implemented yet")
}

// executeDescribeConsumerGroups 执行描述消费者组
func (k *KafkaExecutor) executeDescribeConsumerGroups(ctx context.Context, operation interfaces.Operation, result *interfaces.OperationResult) error {
	// TODO: 实现消费者组描述逻辑
	result.Value = map[string]interface{}{
		"consumer_groups": []string{},
	}
	result.Metadata["admin_operation"] = "describe_consumer_groups"
	return fmt.Errorf("describe consumer groups operation not implemented yet")
}

// isReadOperation 判断是否为读操作
func (k *KafkaExecutor) isReadOperation(operationType string) bool {
	readOperations := map[string]bool{
		"produce":                  false,
		"produce_message":          false,
		"produce_batch":            false,
		"consume":                  true,
		"consume_message":          true,
		"consume_batch":            true,
		"create_topic":             false,
		"delete_topic":             false,
		"list_topics":              true,
		"describe_consumer_groups": true,
	}
	return readOperations[operationType]
}

// GetSupportedOperations 获取支持的操作类型
func (k *KafkaExecutor) GetSupportedOperations() []string {
	return []string{
		"produce",
		"produce_message",
		"produce_batch",
		"consume",
		"consume_message",
		"consume_batch",
		"create_topic",
		"delete_topic",
		"list_topics",
		"describe_consumer_groups",
	}
}
