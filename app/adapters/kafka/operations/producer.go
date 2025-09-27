package operations

import (
	"context"
	"fmt"
	"time"

	"abc-runner/app/adapters/kafka/connection"
	"abc-runner/app/core/interfaces"

	"github.com/segmentio/kafka-go"
)

// ProducerOperations 生产者操作实现
type ProducerOperations struct {
	pool             *connection.ConnectionPool
	metricsCollector interfaces.DefaultMetricsCollector
}

// NewProducerOperations 创建生产者操作实例
func NewProducerOperations(pool *connection.ConnectionPool, metricsCollector interfaces.DefaultMetricsCollector) *ProducerOperations {
	return &ProducerOperations{
		pool:             pool,
		metricsCollector: metricsCollector,
	}
}

// ExecuteProduceMessage 执行单条消息生产
func (p *ProducerOperations) ExecuteProduceMessage(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	startTime := time.Now()

	// 解析参数
	topic, ok := operation.Params["topic"].(string)
	if !ok || topic == "" {
		return &interfaces.OperationResult{
			Success:  false,
			Duration: time.Since(startTime),
			IsRead:   false,
			Error:    fmt.Errorf("topic parameter is required"),
		}, fmt.Errorf("topic parameter is required")
	}

	// 获取生产者
	producer, err := p.pool.GetProducer()
	if err != nil {
		duration := time.Since(startTime)
		// 使用核心接口记录指标
		operationResult := &interfaces.OperationResult{
			Success:  false,
			IsRead:   false,
			Duration: duration,
			Error:    err,
			Metadata: map[string]interface{}{
				"operation_type": "produce",
				"topic":          topic,
				"partition":      -1,
				"message_size":   0,
				"batch_size":     1,
			},
		}
		p.metricsCollector.Record(operationResult)
		return &interfaces.OperationResult{
			Success:  false,
			Duration: duration,
			IsRead:   false,
			Error:    fmt.Errorf("failed to get producer: %w", err),
		}, err
	}
	defer p.pool.ReturnProducer(producer)

	// 构建Kafka消息
	kafkaMessage := kafka.Message{
		Topic: topic,
		Key:   []byte(operation.Key),
		Value: []byte(fmt.Sprintf("%v", operation.Value)),
	}

	// 添加Headers
	if headers, ok := operation.Params["headers"].(map[string]string); ok {
		kafkaMessage.Headers = make([]kafka.Header, 0, len(headers))
		for k, v := range headers {
			kafkaMessage.Headers = append(kafkaMessage.Headers, kafka.Header{
				Key:   k,
				Value: []byte(v),
			})
		}
	}

	// 设置分区（如果指定）
	if partition, ok := operation.Params["partition"].(int32); ok {
		kafkaMessage.Partition = int(partition)
	}

	// 执行生产操作
	err = producer.WriteMessages(ctx, kafkaMessage)
	duration := time.Since(startTime)

	messageSize := len(kafkaMessage.Key) + len(kafkaMessage.Value)
	success := err == nil

	// 使用核心接口记录指标
	operationResult := &interfaces.OperationResult{
		Success:  success,
		IsRead:   false,
		Duration: duration,
		Error:    err,
		Metadata: map[string]interface{}{
			"operation_type": "produce",
			"topic":          topic,
			"partition":      int32(kafkaMessage.Partition),
			"message_size":   int64(messageSize),
			"batch_size":     1,
			"client_id":      "producer",
		},
	}
	p.metricsCollector.Record(operationResult)

	if err != nil {
		return &interfaces.OperationResult{
			Success:  false,
			Duration: duration,
			IsRead:   false,
			Error:    fmt.Errorf("failed to produce message: %w", err),
		}, err
	}

	// 构建结果
	result := &ProduceResult{
		Partition: int32(kafkaMessage.Partition),
		Offset:    -1, // kafka-go Writer不返回offset信息
		Timestamp: time.Now(),
		Duration:  duration,
	}

	return &interfaces.OperationResult{
		Success:  true,
		Duration: duration,
		IsRead:   false,
		Error:    nil,
		Value:    result,
		Metadata: map[string]interface{}{
			"topic":      topic,
			"partition":  kafkaMessage.Partition,
			"key":        operation.Key,
			"value_size": len(kafkaMessage.Value),
		},
	}, nil
}

// ExecuteProduceBatch 执行批量消息生产
func (p *ProducerOperations) ExecuteProduceBatch(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	startTime := time.Now()

	// 解析参数
	topic, ok := operation.Params["topic"].(string)
	if !ok || topic == "" {
		return &interfaces.OperationResult{
			Success:  false,
			Duration: time.Since(startTime),
			IsRead:   false,
			Error:    fmt.Errorf("topic parameter is required"),
		}, fmt.Errorf("topic parameter is required")
	}

	messages, ok := operation.Params["messages"].([]*Message)
	if !ok || len(messages) == 0 {
		return &interfaces.OperationResult{
			Success:  false,
			Duration: time.Since(startTime),
			IsRead:   false,
			Error:    fmt.Errorf("messages parameter is required"),
		}, fmt.Errorf("messages parameter is required")
	}

	// 获取生产者
	producer, err := p.pool.GetProducer()
	if err != nil {
		duration := time.Since(startTime)
		// 使用核心接口记录指标
		operationResult := &interfaces.OperationResult{
			Success:  false,
			IsRead:   false,
			Duration: duration,
			Error:    err,
			Metadata: map[string]interface{}{
				"operation_type": "produce",
				"topic":          topic,
				"partition":      -1,
				"message_size":   0,
				"batch_size":     len(messages),
			},
		}
		p.metricsCollector.Record(operationResult)
		return &interfaces.OperationResult{
			Success:  false,
			Duration: duration,
			IsRead:   false,
			Error:    fmt.Errorf("failed to get producer: %w", err),
		}, err
	}
	defer p.pool.ReturnProducer(producer)

	// 转换为Kafka消息
	kafkaMessages := make([]kafka.Message, 0, len(messages))
	totalSize := 0

	for _, msg := range messages {
		kafkaMessage := kafka.Message{
			Topic:     topic,
			Key:       []byte(msg.Key),
			Value:     []byte(msg.Value),
			Partition: int(msg.Partition),
		}

		// 添加Headers
		if len(msg.Headers) > 0 {
			kafkaMessage.Headers = make([]kafka.Header, 0, len(msg.Headers))
			for k, v := range msg.Headers {
				kafkaMessage.Headers = append(kafkaMessage.Headers, kafka.Header{
					Key:   k,
					Value: []byte(v),
				})
			}
		}

		kafkaMessages = append(kafkaMessages, kafkaMessage)
		totalSize += len(kafkaMessage.Key) + len(kafkaMessage.Value)
	}

	// 执行批量生产操作
	err = producer.WriteMessages(ctx, kafkaMessages...)
	duration := time.Since(startTime)

	success := err == nil
	batchSize := len(messages)

	// 使用核心接口记录指标
	batchOperationResult := &interfaces.OperationResult{
		Success:  success,
		IsRead:   false,
		Duration: duration,
		Error:    err,
		Metadata: map[string]interface{}{
			"operation_type": "produce",
			"topic":          topic,
			"partition":      -1,
			"message_size":   int64(totalSize),
			"batch_size":     batchSize,
			"client_id":      "producer",
		},
	}
	p.metricsCollector.Record(batchOperationResult)

	if err != nil {
		return &interfaces.OperationResult{
			Success:  false,
			Duration: duration,
			IsRead:   false,
			Error:    fmt.Errorf("failed to produce batch messages: %w", err),
		}, err
	}

	// 构建批量结果
	results := make([]ProduceResult, len(messages))
	for i := range results {
		results[i] = ProduceResult{
			Partition: messages[i].Partition,
			Offset:    -1, // kafka-go Writer不返回offset信息
			Timestamp: time.Now(),
			Duration:  duration / time.Duration(len(messages)), // 平均时间
		}
	}

	batchResult := &BatchResult{
		Results:       results,
		SuccessCount:  len(results),
		FailureCount:  0,
		TotalDuration: duration,
	}

	return &interfaces.OperationResult{
		Success:  true,
		Duration: duration,
		IsRead:   false,
		Error:    nil,
		Value:    batchResult,
		Metadata: map[string]interface{}{
			"topic":        topic,
			"batch_size":   batchSize,
			"total_size":   totalSize,
			"avg_msg_size": totalSize / batchSize,
		},
	}, nil
}

// Message Kafka消息结构（重新定义以避免循环导入）
type Message struct {
	Key       string            `json:"key"`
	Value     string            `json:"value"`
	Headers   map[string]string `json:"headers"`
	Timestamp time.Time         `json:"timestamp"`
	Partition int32             `json:"partition"`
	Offset    int64             `json:"offset"`
	Topic     string            `json:"topic"`
}

// ProduceResult 生产结果（重新定义以避免循环导入）
type ProduceResult struct {
	Partition int32         `json:"partition"`
	Offset    int64         `json:"offset"`
	Timestamp time.Time     `json:"timestamp"`
	Duration  time.Duration `json:"duration"`
}

// BatchResult 批量操作结果（重新定义以避免循环导入）
type BatchResult struct {
	Results       []ProduceResult `json:"results"`
	SuccessCount  int             `json:"success_count"`
	FailureCount  int             `json:"failure_count"`
	TotalDuration time.Duration   `json:"total_duration"`
}
