package operation

import (
	"fmt"

	"abc-runner/app/core/execution"
	"abc-runner/app/core/interfaces"
)

// OperationFactory Redis操作工厂
type OperationFactory struct {
	config interfaces.Config
}

// NewOperationFactory 创建Redis操作工厂
func NewOperationFactory(config interfaces.Config) execution.OperationFactory {
	return &OperationFactory{config: config}
}

func (r *OperationFactory) CreateOperation(jobID int, benchmarkConfig execution.BenchmarkConfig) interfaces.Operation {
	benchmark := r.config.GetBenchmark()

	// 根据读写比例决定操作类型
	isRead := (jobID % 100) < benchmark.GetReadPercent()

	var opType string
	var key, value string

	// 生成键
	if benchmark.GetRandomKeys() > 0 {
		key = fmt.Sprintf("key_%d", jobID%benchmark.GetRandomKeys())
	} else {
		key = fmt.Sprintf("key_%d", jobID)
	}

	if isRead {
		opType = "get"
	} else {
		opType = "set"
		// 生成指定大小的值
		dataSize := benchmark.GetDataSize()
		if dataSize <= 0 {
			dataSize = 64
		}
		value = generateRandomValue(dataSize)
	}

	operation := interfaces.Operation{
		Type:  opType,
		Key:   key,
		Value: value,
		TTL:   benchmark.GetTTL(),
		Params: map[string]interface{}{
			"operation_type": opType,
			"job_id":         jobID,
			"is_read":        isRead,
		},
	}

	return operation
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
