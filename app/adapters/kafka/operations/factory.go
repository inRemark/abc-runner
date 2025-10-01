package operations

import (
	"fmt"

	"abc-runner/app/adapters/kafka/config"
	"abc-runner/app/core/execution"
	"abc-runner/app/core/interfaces"
)

// OperationFactory Kafka操作工厂
type OperationFactory struct {
	config *config.KafkaAdapterConfig
}

// NewOperationFactory 创建Kafka操作工厂
func NewOperationFactory(config *config.KafkaAdapterConfig) execution.OperationFactory {
	return &OperationFactory{config: config}
}

func (k *OperationFactory) CreateOperation(jobID int, benchmarkConfig execution.BenchmarkConfig) interfaces.Operation {
	benchmark := k.config.Benchmark

	// 根据测试类型确定操作
	var opType string
	switch benchmark.TestCase {
	case "consumer", "consume":
		opType = "consume_message"
	case "producer", "produce":
		opType = "produce_message"
	default:
		// 根据读写比例决定
		if (jobID % 100) < benchmark.ReadPercent {
			opType = "consume_message"
		} else {
			opType = "produce_message"
		}
	}

	// 生成消息
	key := fmt.Sprintf("key_%d", jobID)
	value := fmt.Sprintf("message_%d", jobID)

	// 如果有指定数据大小，生成相应大小的值
	if benchmark.DataSize > 0 {
		value = generateRandomValue(benchmark.DataSize)
	}

	return interfaces.Operation{
		Type:  opType,
		Key:   key,
		Value: value,
		Params: map[string]interface{}{
			"topic":        benchmark.DefaultTopic,
			"partition":    jobID % 3, // 简单分区策略
			"job_id":       jobID,
			"test_type":    benchmark.TestCase,
			"message_size": benchmark.DataSize,
		},
	}
}

// generateRandomValue 生成指定大小的随机值
func generateRandomValue(size int) string {
	if size <= 0 {
		return ""
	}

	// 简单的值生成逻辑
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, size)

	for i := 0; i < size; i++ {
		result[i] = charset[i%len(charset)]
	}

	return string(result)
}
