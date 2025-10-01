package operations

import (
	"abc-runner/app/core/execution"
	"abc-runner/app/core/interfaces"
)

// SimpleOperationFactory 简单操作工厂
type SimpleOperationFactory struct {
	operationType string
	dataSize      int
}

// NewSimpleOperationFactory 创建简单操作工厂
func NewSimpleOperationFactory(operationType string, dataSize int) *SimpleOperationFactory {
	return &SimpleOperationFactory{
		operationType: operationType,
		dataSize:      dataSize,
	}
}

// CreateOperation 创建操作
func (f *SimpleOperationFactory) CreateOperation(jobID int, config execution.BenchmarkConfig) interfaces.Operation {
	return interfaces.Operation{
		Type:  f.operationType,
		Key:   "key_" + string(rune(jobID)),
		Value: generateTestData(f.dataSize),
		Params: map[string]interface{}{
			"job_id":    jobID,
			"data_size": f.dataSize,
		},
		Metadata: map[string]string{
			"operation_type": f.operationType,
		},
	}
}

// generateTestData 生成测试数据
func generateTestData(size int) []byte {
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(i % 256)
	}
	return data
}
