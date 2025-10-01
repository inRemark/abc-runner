package operations

import (
	"fmt"

	"abc-runner/app/adapters/websocket/config"
	"abc-runner/app/core/execution"
	"abc-runner/app/core/interfaces"
)

// WebSocketEngineOperationFactory 适配ExecutionEngine的WebSocket操作工厂
type WebSocketEngineOperationFactory struct {
	config *config.WebSocketConfig
}

// NewWebSocketEngineOperationFactory 创建WebSocket引擎操作工厂
func NewWebSocketEngineOperationFactory(config *config.WebSocketConfig) execution.OperationFactory {
	return &WebSocketEngineOperationFactory{
		config: config,
	}
}

// CreateOperation 创建操作实例（适配ExecutionEngine接口）
func (f *WebSocketEngineOperationFactory) CreateOperation(jobID int, benchConfig execution.BenchmarkConfig) interfaces.Operation {
	testCase := f.config.BenchMark.TestCase

	// 根据测试用例创建相应的操作
	switch testCase {
	case "message_exchange":
		return f.createMessageExchangeOperation(jobID)
	case "ping_pong":
		return f.createPingPongOperation(jobID)
	case "broadcast":
		return f.createBroadcastOperation(jobID)
	case "large_message":
		return f.createLargeMessageOperation(jobID)
	default:
		// 默认使用消息交换测试
		return f.createMessageExchangeOperation(jobID)
	}
}

// createMessageExchangeOperation 创建消息交换操作
func (f *WebSocketEngineOperationFactory) createMessageExchangeOperation(jobID int) interfaces.Operation {
	message := f.generateTestMessage(f.config.BenchMark.DataSize)

	return interfaces.Operation{
		Type:  "send_text",
		Key:   fmt.Sprintf("msg_%d", jobID),
		Value: message,
		Metadata: map[string]string{
			"test_case":      "message_exchange",
			"job_id":         fmt.Sprintf("%d", jobID),
			"message_length": fmt.Sprintf("%d", len(message)),
		},
	}
}

// createPingPongOperation 创建心跳操作
func (f *WebSocketEngineOperationFactory) createPingPongOperation(jobID int) interfaces.Operation {
	return interfaces.Operation{
		Type:  "ping_pong",
		Key:   fmt.Sprintf("ping_%d", jobID),
		Value: "ping",
		Metadata: map[string]string{
			"test_case": "ping_pong",
			"job_id":    fmt.Sprintf("%d", jobID),
		},
	}
}

// createBroadcastOperation 创建广播操作
func (f *WebSocketEngineOperationFactory) createBroadcastOperation(jobID int) interfaces.Operation {
	message := f.generateTestMessage(f.config.BenchMark.DataSize)

	return interfaces.Operation{
		Type:  "broadcast",
		Key:   fmt.Sprintf("broadcast_%d", jobID),
		Value: message,
		Metadata: map[string]string{
			"test_case":      "broadcast",
			"job_id":         fmt.Sprintf("%d", jobID),
			"message_length": fmt.Sprintf("%d", len(message)),
		},
	}
}

// createLargeMessageOperation 创建大消息操作
func (f *WebSocketEngineOperationFactory) createLargeMessageOperation(jobID int) interfaces.Operation {
	dataSize := f.config.BenchMark.DataSize
	if dataSize < 1024*1024 { // 如果小于1MB，设为1MB
		dataSize = 1024 * 1024
	}

	data := make([]byte, dataSize)
	for i := range data {
		data[i] = byte('A' + (i % 26))
	}

	return interfaces.Operation{
		Type:  "large_message",
		Key:   fmt.Sprintf("large_%d", jobID),
		Value: data,
		Metadata: map[string]string{
			"test_case": "large_message",
			"job_id":    fmt.Sprintf("%d", jobID),
			"data_size": fmt.Sprintf("%d", dataSize),
		},
	}
}

// generateTestMessage 生成测试消息
func (f *WebSocketEngineOperationFactory) generateTestMessage(size int) string {
	if size <= 0 {
		size = 1024 // 默认1KB
	}

	message := make([]byte, size)
	for i := range message {
		message[i] = 'A' + byte(i%26)
	}
	return string(message)
}
