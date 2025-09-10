package operations

import (
	"fmt"
	"math/rand"
	"time"

	kafkaConfig "abc-runner/app/adapters/kafka/config"
	"abc-runner/app/core/interfaces"
)

// KafkaOperationFactory Kafka操作工厂
type KafkaOperationFactory struct {
	config     *kafkaConfig.KafkaAdapterConfig
	rand       *rand.Rand
	operations []operationType
}

// operationType 操作类型定义
type operationType struct {
	name   string
	weight int
	isRead bool
}

// NewKafkaOperationFactory 创建Kafka操作工厂
func NewKafkaOperationFactory(config *kafkaConfig.KafkaAdapterConfig) *KafkaOperationFactory {
	factory := &KafkaOperationFactory{
		config: config,
		rand:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	factory.initializeOperationTypes()
	return factory
}

// initializeOperationTypes 初始化操作类型
func (f *KafkaOperationFactory) initializeOperationTypes() {
	f.operations = []operationType{
		{name: "produce_message", weight: 100 - f.config.Benchmark.ReadPercent, isRead: false},
		{name: "consume_message", weight: f.config.Benchmark.ReadPercent, isRead: true},
		{name: "produce_batch", weight: (100 - f.config.Benchmark.ReadPercent) / 4, isRead: false},
		{name: "consume_batch", weight: f.config.Benchmark.ReadPercent / 4, isRead: true},
	}
}

// CreateOperation 创建操作
func (f *KafkaOperationFactory) CreateOperation(params map[string]interface{}) (interfaces.Operation, error) {
	// 选择操作类型
	operationType := f.selectOperationType(params)

	// 生成操作键
	operationKey := f.generateOperationKey(operationType, params)

	// 根据操作类型创建特定的操作参数
	operationParams, err := f.createOperationParams(operationType, params)
	if err != nil {
		return interfaces.Operation{}, fmt.Errorf("failed to create operation params: %w", err)
	}

	// 生成操作值
	operationValue := f.generateOperationValue(operationType, params)

	// 创建操作
	operation := interfaces.Operation{
		Type:   operationType,
		Key:    operationKey,
		Value:  operationValue,
		Params: operationParams,
		Metadata: map[string]string{
			"operation_factory": "kafka",
			"operation_type":    operationType,
		},
	}

	return operation, nil
}

// selectOperationType 选择操作类型
func (f *KafkaOperationFactory) selectOperationType(params map[string]interface{}) string {
	// 如果明确指定了操作类型
	if opType, exists := params["operation_type"]; exists {
		if opTypeStr, ok := opType.(string); ok {
			if f.isValidOperationType(opTypeStr) {
				return opTypeStr
			}
		}
	}

	// 按权重随机选择
	totalWeight := 0
	for _, op := range f.operations {
		totalWeight += op.weight
	}

	if totalWeight == 0 {
		return "produce_message" // 默认操作
	}

	randomValue := f.rand.Intn(totalWeight)
	currentWeight := 0

	for _, op := range f.operations {
		currentWeight += op.weight
		if randomValue < currentWeight {
			return op.name
		}
	}

	return "produce_message" // 默认操作
}

// isValidOperationType 检查是否为有效的操作类型
func (f *KafkaOperationFactory) isValidOperationType(opType string) bool {
	validTypes := []string{
		"produce_message", "produce_batch",
		"consume_message", "consume_batch",
		"create_topic", "delete_topic", "list_topics",
		"describe_consumer_groups",
	}

	for _, validType := range validTypes {
		if opType == validType {
			return true
		}
	}

	return false
}

// generateOperationKey 生成操作键
func (f *KafkaOperationFactory) generateOperationKey(operationType string, params map[string]interface{}) string {
	// 基于操作类型和参数生成唯一键
	key := operationType

	// 如果有索引参数，加入键中
	if index, exists := params["index"]; exists {
		key = fmt.Sprintf("%s:%v", key, index)
	}

	// 如果是随机键模式
	if f.config.Benchmark.RandomKeys > 0 {
		randomSuffix := f.rand.Intn(f.config.Benchmark.RandomKeys)
		key = fmt.Sprintf("%s:random_%d", key, randomSuffix)
	}

	return key
}

// createOperationParams 创建操作参数
func (f *KafkaOperationFactory) createOperationParams(operationType string, params map[string]interface{}) (map[string]interface{}, error) {
	operationParams := make(map[string]interface{})

	// 复制基础参数
	for k, v := range params {
		operationParams[k] = v
	}

	// 设置默认topic
	if _, exists := operationParams["topic"]; !exists {
		operationParams["topic"] = f.config.Benchmark.DefaultTopic
	}

	// 根据操作类型设置特定参数
	switch operationType {
	case "produce_message":
		return f.createProduceMessageParams(operationParams)

	case "produce_batch":
		return f.createProduceBatchParams(operationParams)

	case "consume_message":
		return f.createConsumeMessageParams(operationParams)

	case "consume_batch":
		return f.createConsumeBatchParams(operationParams)

	default:
		// 对于其他操作类型，返回基础参数
		return operationParams, nil
	}
}

// createProduceMessageParams 创建单条消息生产参数
func (f *KafkaOperationFactory) createProduceMessageParams(params map[string]interface{}) (map[string]interface{}, error) {
	// 设置默认Headers
	if _, exists := params["headers"]; !exists {
		params["headers"] = map[string]string{
			"producer":  "abc-runner",
			"timestamp": fmt.Sprintf("%d", time.Now().Unix()),
		}
	}

	// 设置分区策略
	if _, exists := params["partition"]; !exists && f.config.Benchmark.PartitionStrategy == "manual" {
		// 手动分区分配（简单轮询）
		params["partition"] = int32(f.rand.Intn(8)) // 假设8个分区
	}

	return params, nil
}

// createProduceBatchParams 创建批量消息生产参数
func (f *KafkaOperationFactory) createProduceBatchParams(params map[string]interface{}) (map[string]interface{}, error) {
	// 如果没有指定消息，生成一批消息
	if _, exists := params["messages"]; !exists {
		batchSize := f.selectBatchSize()
		messages := make([]*Message, batchSize)

		for i := 0; i < batchSize; i++ {
			messages[i] = &Message{
				Key:   fmt.Sprintf("batch_key_%d_%d", time.Now().UnixNano(), i),
				Value: f.generateMessageValue(),
				Headers: map[string]string{
					"batch_index": fmt.Sprintf("%d", i),
					"batch_size":  fmt.Sprintf("%d", batchSize),
				},
				Partition: int32(f.rand.Intn(8)), // 假设8个分区
			}
		}

		params["messages"] = messages
	}

	return params, nil
}

// createConsumeMessageParams 创建单条消息消费参数
func (f *KafkaOperationFactory) createConsumeMessageParams(params map[string]interface{}) (map[string]interface{}, error) {
	// 设置读取超时
	if _, exists := params["timeout"]; !exists {
		params["timeout"] = f.config.Benchmark.Timeout
	}

	return params, nil
}

// createConsumeBatchParams 创建批量消息消费参数
func (f *KafkaOperationFactory) createConsumeBatchParams(params map[string]interface{}) (map[string]interface{}, error) {
	// 设置批量大小
	if _, exists := params["max_messages"]; !exists {
		params["max_messages"] = f.selectBatchSize()
	}

	// 设置读取超时
	if _, exists := params["timeout"]; !exists {
		params["timeout"] = f.config.Benchmark.Timeout
	}

	return params, nil
}

// selectBatchSize 选择批量大小
func (f *KafkaOperationFactory) selectBatchSize() int {
	if len(f.config.Benchmark.BatchSizes) == 0 {
		return 10 // 默认批量大小
	}

	// 从配置的批量大小中随机选择
	return f.config.Benchmark.BatchSizes[f.rand.Intn(len(f.config.Benchmark.BatchSizes))]
}

// generateOperationValue 生成操作值
func (f *KafkaOperationFactory) generateOperationValue(operationType string, params map[string]interface{}) interface{} {
	switch operationType {
	case "produce_message":
		return f.generateMessageValue()
	case "produce_batch":
		// 批量操作的值在参数中处理
		return nil
	default:
		return nil
	}
}

// generateMessageValue 生成消息值
func (f *KafkaOperationFactory) generateMessageValue() string {
	// 根据配置的消息大小范围生成消息
	minSize := f.config.Benchmark.MessageSizeRange.Min
	maxSize := f.config.Benchmark.MessageSizeRange.Max

	if minSize <= 0 {
		minSize = f.config.Benchmark.DataSize
	}
	if maxSize <= minSize {
		maxSize = minSize + 100
	}

	messageSize := minSize + f.rand.Intn(maxSize-minSize+1)

	// 生成指定长度的随机字符串
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	message := make([]byte, messageSize)
	for i := range message {
		message[i] = charset[f.rand.Intn(len(charset))]
	}

	return string(message)
}

// GetOperationType 获取操作类型（实现OperationFactory接口）
func (f *KafkaOperationFactory) GetOperationType() string {
	return "kafka"
}

// ValidateParams 验证参数（实现OperationFactory接口）
func (f *KafkaOperationFactory) ValidateParams(params map[string]interface{}) error {
	// 验证基本参数
	if f.config.Benchmark.DefaultTopic == "" {
		if _, exists := params["topic"]; !exists {
			return fmt.Errorf("topic is required when default_topic is not configured")
		}
	}

	// 如果指定了operation_type，验证是否有效
	if opType, exists := params["operation_type"]; exists {
		if opTypeStr, ok := opType.(string); ok {
			if !f.isValidOperationType(opTypeStr) {
				return fmt.Errorf("invalid operation type: %s", opTypeStr)
			}
		}
	}

	return nil
}
