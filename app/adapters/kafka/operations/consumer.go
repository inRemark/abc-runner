package operations

import (
	"context"
	"fmt"
	"time"

	"abc-runner/app/adapters/kafka/connection"
	"abc-runner/app/adapters/kafka/metrics"
	"abc-runner/app/core/interfaces"

	"github.com/segmentio/kafka-go"
)

// ConsumerOperations 消费者操作实现
type ConsumerOperations struct {
	pool             *connection.ConnectionPool
	metricsCollector *metrics.MetricsCollector
}

// NewConsumerOperations 创建消费者操作实例
func NewConsumerOperations(pool *connection.ConnectionPool, metricsCollector *metrics.MetricsCollector) *ConsumerOperations {
	return &ConsumerOperations{
		pool:             pool,
		metricsCollector: metricsCollector,
	}
}

// ExecuteConsumeMessage 执行单条消息消费
func (c *ConsumerOperations) ExecuteConsumeMessage(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	startTime := time.Now()

	// 解析参数
	topic, ok := operation.Params["topic"].(string)
	if !ok || topic == "" {
		return &interfaces.OperationResult{
			Success:  false,
			Duration: time.Since(startTime),
			IsRead:   true,
			Error:    fmt.Errorf("topic parameter is required"),
		}, fmt.Errorf("topic parameter is required")
	}

	// 获取消费者
	consumer, err := c.pool.GetConsumer()
	if err != nil {
		duration := time.Since(startTime)
		// 使用核心接口记录指标
		defaultOperationResult := &interfaces.OperationResult{
			Success:  false,
			IsRead:   true,
			Duration: duration,
			Error:    err,
			Metadata: map[string]interface{}{
				"operation_type": "consume",
				"topic":          topic,
				"partition":      -1,
				"message_size":   0,
				"offset":         -1,
			},
		}
		c.metricsCollector.RecordOperation(defaultOperationResult)
		return &interfaces.OperationResult{
			Success:  false,
			Duration: duration,
			IsRead:   true,
			Error:    fmt.Errorf("failed to get consumer: %w", err),
		}, err
	}
	defer c.pool.ReturnConsumer(consumer)

	// 设置读取超时
	timeoutCtx := ctx
	if timeout, ok := operation.Params["timeout"].(time.Duration); ok && timeout > 0 {
		var cancel context.CancelFunc
		timeoutCtx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	// 执行消费操作
	msg, err := consumer.ReadMessage(timeoutCtx)
	duration := time.Since(startTime)

	success := err == nil
	var messageSize int
	var offset int64 = -1
	var partition int32 = -1

	if success {
		messageSize = len(msg.Key) + len(msg.Value)
		offset = msg.Offset
		partition = int32(msg.Partition)
	}

	// 使用核心接口记录指标
	consumeResult := &interfaces.OperationResult{
		Success:  success,
		IsRead:   true,
		Duration: duration,
		Error:    err,
		Metadata: map[string]interface{}{
			"operation_type": "consume",
			"topic":          topic,
			"partition":      partition,
			"message_size":   int64(messageSize),
			"offset":         offset,
			"client_id":      "consumer",
		},
	}
	c.metricsCollector.RecordOperation(consumeResult)

	if err != nil {
		return &interfaces.OperationResult{
			Success:  false,
			Duration: duration,
			IsRead:   true,
			Error:    fmt.Errorf("failed to consume message: %w", err),
		}, err
	}

	// 构建消息结果
	message := &Message{
		Key:       string(msg.Key),
		Value:     string(msg.Value),
		Headers:   convertHeaders(msg.Headers),
		Timestamp: msg.Time,
		Partition: int32(msg.Partition),
		Offset:    msg.Offset,
		Topic:     msg.Topic,
	}

	return &interfaces.OperationResult{
		Success:  true,
		Duration: duration,
		IsRead:   true,
		Error:    nil,
		Value:    message,
		Metadata: map[string]interface{}{
			"topic":      msg.Topic,
			"partition":  msg.Partition,
			"offset":     msg.Offset,
			"key":        string(msg.Key),
			"value_size": len(msg.Value),
			"timestamp":  msg.Time,
		},
	}, nil
}

// ExecuteConsumeBatch 执行批量消息消费
func (c *ConsumerOperations) ExecuteConsumeBatch(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	startTime := time.Now()

	// 解析参数
	maxMessages, ok := operation.Params["max_messages"].(int)
	if !ok || maxMessages <= 0 {
		maxMessages = 100 // 默认批量大小
	}

	topic, ok := operation.Params["topic"].(string)
	if !ok || topic == "" {
		return &interfaces.OperationResult{
			Success:  false,
			Duration: time.Since(startTime),
			IsRead:   true,
			Error:    fmt.Errorf("topic parameter is required"),
		}, fmt.Errorf("topic parameter is required")
	}

	// 获取消费者
	consumer, err := c.pool.GetConsumer()
	if err != nil {
		duration := time.Since(startTime)
		// 使用核心接口记录指标
		defaultOperationResult := &interfaces.OperationResult{
			Success:  false,
			IsRead:   true,
			Duration: duration,
			Error:    err,
			Metadata: map[string]interface{}{
				"operation_type": "consume",
				"topic":          topic,
				"partition":      -1,
				"message_size":   0,
				"offset":         -1,
			},
		}
		c.metricsCollector.RecordOperation(defaultOperationResult)
		return &interfaces.OperationResult{
			Success:  false,
			Duration: duration,
			IsRead:   true,
			Error:    fmt.Errorf("failed to get consumer: %w", err),
		}, err
	}
	defer c.pool.ReturnConsumer(consumer)

	// 设置读取超时
	timeoutCtx := ctx
	if timeout, ok := operation.Params["timeout"].(time.Duration); ok && timeout > 0 {
		var cancel context.CancelFunc
		timeoutCtx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	// 批量消费消息
	messages := make([]*Message, 0, maxMessages)
	totalSize := 0
	successCount := 0
	var lastErr error

	for i := 0; i < maxMessages; i++ {
		msg, err := consumer.ReadMessage(timeoutCtx)
		if err != nil {
			lastErr = err
			// 如果是超时错误，退出循环
			if isTimeoutError(err) {
				break
			}
			// 其他错误也退出循环
			break
		}

		message := &Message{
			Key:       string(msg.Key),
			Value:     string(msg.Value),
			Headers:   convertHeaders(msg.Headers),
			Timestamp: msg.Time,
			Partition: int32(msg.Partition),
			Offset:    msg.Offset,
			Topic:     msg.Topic,
		}

		messages = append(messages, message)
		totalSize += len(msg.Key) + len(msg.Value)
		successCount++

		// 使用核心接口记录每条消息的指标
		msgResult := &interfaces.OperationResult{
			Success:  true,
			IsRead:   true,
			Duration: 0, // 单条消息不单独计时
			Error:    nil,
			Metadata: map[string]interface{}{
				"operation_type": "consume",
				"topic":          topic,
				"partition":      int32(msg.Partition),
				"message_size":   int64(len(msg.Key) + len(msg.Value)),
				"offset":         msg.Offset,
				"client_id":      "consumer",
			},
		}
		c.metricsCollector.RecordOperation(msgResult)
	}

	duration := time.Since(startTime)

	// 如果没有消费到任何消息且有错误，返回错误
	if len(messages) == 0 && lastErr != nil {
		// 使用核心接口记录指标
		batchFailResult := &interfaces.OperationResult{
			Success:  false,
			IsRead:   true,
			Duration: duration,
			Error:    lastErr,
			Metadata: map[string]interface{}{
				"operation_type": "consume",
				"topic":          topic,
				"partition":      -1,
				"message_size":   0,
				"offset":         -1,
			},
		}
		c.metricsCollector.RecordOperation(batchFailResult)
		return &interfaces.OperationResult{
			Success:  false,
			Duration: duration,
			IsRead:   true,
			Error:    fmt.Errorf("failed to consume batch messages: %w", lastErr),
		}, lastErr
	}

	// 构建批量结果
	batchResult := &ConsumeBatchResult{
		Messages:      messages,
		SuccessCount:  successCount,
		FailureCount:  maxMessages - successCount,
		TotalDuration: duration,
		TotalSize:     totalSize,
	}

	return &interfaces.OperationResult{
		Success:  true,
		Duration: duration,
		IsRead:   true,
		Error:    nil,
		Value:    batchResult,
		Metadata: map[string]interface{}{
			"topic":           topic,
			"requested_count": maxMessages,
			"actual_count":    len(messages),
			"success_count":   successCount,
			"total_size":      totalSize,
			"avg_message_size": func() int {
				if len(messages) > 0 {
					return totalSize / len(messages)
				}
				return 0
			}(),
		},
	}, nil
}

// convertHeaders 转换Kafka Headers
func convertHeaders(headers []kafka.Header) map[string]string {
	result := make(map[string]string, len(headers))
	for _, header := range headers {
		result[header.Key] = string(header.Value)
	}
	return result
}

// isTimeoutError 检查是否为超时错误
func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	return contains(errStr, "timeout") ||
		contains(errStr, "context deadline exceeded") ||
		contains(errStr, "i/o timeout")
}

// contains 检查字符串是否包含子字符串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && indexOf(s, substr) >= 0
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

// ConsumeBatchResult 批量消费结果
type ConsumeBatchResult struct {
	Messages      []*Message    `json:"messages"`
	SuccessCount  int           `json:"success_count"`
	FailureCount  int           `json:"failure_count"`
	TotalDuration time.Duration `json:"total_duration"`
	TotalSize     int           `json:"total_size"`
}
