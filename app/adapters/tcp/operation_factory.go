package tcp

import (
	"fmt"
	"strconv"
	
	"abc-runner/app/adapters/tcp/config"
	"abc-runner/app/core/execution"
	"abc-runner/app/core/interfaces"
)

// OperationFactory TCP操作工厂
type OperationFactory struct {
	config     *config.TCPConfig
	testCase   string
	dataSize   int
}

// NewOperationFactory 创建TCP操作工厂
func NewOperationFactory(cfg *config.TCPConfig) *OperationFactory {
	return &OperationFactory{
		config:   cfg,
		testCase: cfg.BenchMark.TestCase,
		dataSize: cfg.BenchMark.DataSize,
	}
}

// CreateOperation 创建TCP操作
func (f *OperationFactory) CreateOperation(jobID int, config execution.BenchmarkConfig) interfaces.Operation {
	// 生成键
	key := f.generateKey(jobID)
	
	// 生成测试数据
	testData := f.generateTestData(jobID)
	
	// 创建操作特定参数
	params := map[string]interface{}{
		"job_id":         jobID,
		"data_size":      f.dataSize,
		"test_case":      f.testCase,
		"connection_mode": f.config.TCPSpecific.ConnectionMode,
		"no_delay":       f.config.TCPSpecific.NoDelay,
		"buffer_size":    f.config.TCPSpecific.BufferSize,
	}
	
	// 创建操作元数据
	metadata := map[string]string{
		"operation_type":   f.testCase,
		"protocol":         "tcp",
		"connection_mode":  f.config.TCPSpecific.ConnectionMode,
		"job_id":          strconv.Itoa(jobID),
	}
	
	return interfaces.Operation{
		Type:     f.testCase,
		Key:      key,
		Value:    testData,
		Params:   params,
		Metadata: metadata,
	}
}

// GetOperationType 获取操作类型
func (f *OperationFactory) GetOperationType() string {
	return f.testCase
}

// GetConfig 获取配置信息
func (f *OperationFactory) GetConfig() *config.TCPConfig {
	return f.config
}

// GetSupportedOperations 获取支持的操作类型
func (f *OperationFactory) GetSupportedOperations() []string {
	return []string{"echo_test", "send_only", "receive_only", "bidirectional"}
}

// generateKey 生成操作键
func (f *OperationFactory) generateKey(jobID int) string {
	switch f.testCase {
	case "echo_test":
		return fmt.Sprintf("tcp_echo_%d", jobID)
	case "send_only":
		return fmt.Sprintf("tcp_send_%d", jobID)
	case "receive_only":
		return fmt.Sprintf("tcp_recv_%d", jobID)
	case "bidirectional":
		return fmt.Sprintf("tcp_bidi_%d", jobID)
	default:
		return fmt.Sprintf("tcp_op_%d", jobID)
	}
}

// generateTestData 生成测试数据
func (f *OperationFactory) generateTestData(jobID int) []byte {
	data := make([]byte, f.dataSize)
	
	// 根据测试用例类型生成不同的数据模式
	switch f.testCase {
	case "echo_test":
		// 回显测试使用可预测的模式
		pattern := fmt.Sprintf("ECHO_TEST_%d_", jobID)
		patternBytes := []byte(pattern)
		for i := range data {
			data[i] = patternBytes[i%len(patternBytes)]
		}
	case "send_only":
		// 发送测试使用递增模式
		for i := range data {
			data[i] = byte((jobID + i) % 256)
		}
	case "receive_only":
		// 接收测试使用空数据（不发送）
		return []byte{}
	case "bidirectional":
		// 双向测试使用混合模式
		for i := range data {
			if i%2 == 0 {
				data[i] = byte(jobID % 256)
			} else {
				data[i] = byte((jobID + i) % 256)
			}
		}
	default:
		// 默认模式
		for i := range data {
			data[i] = byte(i % 256)
		}
	}
	
	return data
}

// 确保实现了execution.OperationFactory接口
var _ execution.OperationFactory = (*OperationFactory)(nil)