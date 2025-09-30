package operations

import (
	"fmt"
	"strconv"
	"time"

	"abc-runner/app/core/interfaces"
	"abc-runner/app/core/utils"
)

// WebSocketOperationFactory WebSocket操作工厂
type WebSocketOperationFactory struct {
	operationType string
}

// NewWebSocketOperationFactory 创建WebSocket操作工厂
func NewWebSocketOperationFactory(operationType string) *WebSocketOperationFactory {
	return &WebSocketOperationFactory{
		operationType: operationType,
	}
}

// CreateOperation 创建操作实例
func (f *WebSocketOperationFactory) CreateOperation(params map[string]interface{}) (interfaces.Operation, error) {
	if err := f.ValidateParams(params); err != nil {
		return interfaces.Operation{}, err
	}

	operation := interfaces.Operation{
		Type:     f.operationType,
		Metadata: make(map[string]string),
	}

	switch f.operationType {
	case "send_text":
		return f.createSendTextOperation(params, operation)
	case "send_binary":
		return f.createSendBinaryOperation(params, operation)
	case "echo_test":
		return f.createEchoTestOperation(params, operation)
	case "ping_pong":
		return f.createPingPongOperation(params, operation)
	case "broadcast":
		return f.createBroadcastOperation(params, operation)
	case "subscribe":
		return f.createSubscribeOperation(params, operation)
	case "large_message":
		return f.createLargeMessageOperation(params, operation)
	case "stress_test":
		return f.createStressTestOperation(params, operation)
	default:
		return interfaces.Operation{}, fmt.Errorf("unsupported operation type: %s", f.operationType)
	}
}

// GetOperationType 获取操作类型
func (f *WebSocketOperationFactory) GetOperationType() string {
	return f.operationType
}

// ValidateParams 验证参数
func (f *WebSocketOperationFactory) ValidateParams(params map[string]interface{}) error {
	if params == nil {
		return fmt.Errorf("params cannot be nil")
	}

	switch f.operationType {
	case "send_text":
		return f.validateSendTextParams(params)
	case "send_binary":
		return f.validateSendBinaryParams(params)
	case "echo_test":
		return f.validateEchoTestParams(params)
	case "ping_pong":
		return f.validatePingPongParams(params)
	case "broadcast":
		return f.validateBroadcastParams(params)
	case "subscribe":
		return f.validateSubscribeParams(params)
	case "large_message":
		return f.validateLargeMessageParams(params)
	case "stress_test":
		return f.validateStressTestParams(params)
	default:
		return fmt.Errorf("unsupported operation type: %s", f.operationType)
	}
}

// 具体操作创建方法

// createSendTextOperation 创建发送文本消息操作
func (f *WebSocketOperationFactory) createSendTextOperation(params map[string]interface{}, operation interfaces.Operation) (interfaces.Operation, error) {
	message, ok := params["message"].(string)
	if !ok {
		message = "default text message"
	}

	operation.Key = "text_message"
	operation.Value = message
	operation.Metadata["message_type"] = "text"
	operation.Metadata["message_length"] = strconv.Itoa(len(message))

	if timeout, ok := params["timeout"].(time.Duration); ok {
		operation.TTL = timeout
		operation.Metadata["timeout"] = timeout.String()
	}

	return operation, nil
}

// createSendBinaryOperation 创建发送二进制消息操作
func (f *WebSocketOperationFactory) createSendBinaryOperation(params map[string]interface{}, operation interfaces.Operation) (interfaces.Operation, error) {
	var data []byte
	
	if rawData, ok := params["data"].([]byte); ok {
		data = rawData
	} else if size, ok := params["size"].(int); ok && size > 0 {
		data = make([]byte, size)
		for i := range data {
			data[i] = byte(i % 256)
		}
	} else {
		data = []byte("default binary data")
	}

	operation.Key = "binary_message"
	operation.Value = data
	operation.Metadata["message_type"] = "binary"
	operation.Metadata["data_size"] = strconv.Itoa(len(data))

	if timeout, ok := params["timeout"].(time.Duration); ok {
		operation.TTL = timeout
		operation.Metadata["timeout"] = timeout.String()
	}

	return operation, nil
}

// createEchoTestOperation 创建回显测试操作
func (f *WebSocketOperationFactory) createEchoTestOperation(params map[string]interface{}, operation interfaces.Operation) (interfaces.Operation, error) {
	message, ok := params["message"].(string)
	if !ok {
		message = "echo test message"
	}

	operation.Key = "echo_test"
	operation.Value = message
	operation.Metadata["message_type"] = "echo"
	operation.Metadata["expect_response"] = "true"
	operation.Metadata["message_length"] = strconv.Itoa(len(message))

	if timeout, ok := params["timeout"].(time.Duration); ok {
		operation.TTL = timeout
		operation.Metadata["timeout"] = timeout.String()
	} else {
		operation.TTL = 30 * time.Second
		operation.Metadata["timeout"] = "30s"
	}

	return operation, nil
}

// createPingPongOperation 创建心跳测试操作
func (f *WebSocketOperationFactory) createPingPongOperation(params map[string]interface{}, operation interfaces.Operation) (interfaces.Operation, error) {
	operation.Key = "ping_pong"
	operation.Value = "ping"
	operation.Metadata["message_type"] = "ping"
	operation.Metadata["expect_response"] = "pong"

	if timeout, ok := params["timeout"].(time.Duration); ok {
		operation.TTL = timeout
		operation.Metadata["timeout"] = timeout.String()
	} else {
		operation.TTL = 10 * time.Second
		operation.Metadata["timeout"] = "10s"
	}

	if interval, ok := params["interval"].(time.Duration); ok {
		operation.Metadata["interval"] = interval.String()
	}

	return operation, nil
}

// createBroadcastOperation 创建广播测试操作
func (f *WebSocketOperationFactory) createBroadcastOperation(params map[string]interface{}, operation interfaces.Operation) (interfaces.Operation, error) {
	message, ok := params["message"].(string)
	if !ok {
		message = "broadcast message"
	}

	operation.Key = "broadcast"
	operation.Value = message
	operation.Metadata["message_type"] = "broadcast"
	operation.Metadata["message_length"] = strconv.Itoa(len(message))

	if targets, ok := params["targets"].(int); ok {
		operation.Metadata["target_connections"] = strconv.Itoa(targets)
	}

	if messageType, ok := params["broadcast_type"].(string); ok {
		operation.Metadata["broadcast_type"] = messageType
	} else {
		operation.Metadata["broadcast_type"] = "text"
	}

	return operation, nil
}

// createSubscribeOperation 创建订阅操作
func (f *WebSocketOperationFactory) createSubscribeOperation(params map[string]interface{}, operation interfaces.Operation) (interfaces.Operation, error) {
	channel, ok := params["channel"].(string)
	if !ok {
		channel = "default"
	}

	operation.Key = channel
	operation.Value = "subscribe"
	operation.Metadata["message_type"] = "subscribe"
	operation.Metadata["channel"] = channel

	if timeout, ok := params["timeout"].(time.Duration); ok {
		operation.TTL = timeout
		operation.Metadata["timeout"] = timeout.String()
	} else {
		operation.TTL = 60 * time.Second
		operation.Metadata["timeout"] = "60s"
	}

	return operation, nil
}

// createLargeMessageOperation 创建大消息传输操作
func (f *WebSocketOperationFactory) createLargeMessageOperation(params map[string]interface{}, operation interfaces.Operation) (interfaces.Operation, error) {
	size, ok := params["size"].(int)
	if !ok || size <= 0 {
		size = 1024 * 1024 // 默认1MB
	}

	// 生成大消息数据
	data := make([]byte, size)
	for i := range data {
		data[i] = byte('A' + (i % 26))
	}

	operation.Key = "large_message"
	operation.Value = data
	operation.Metadata["message_type"] = "large"
	operation.Metadata["data_size"] = strconv.Itoa(size)

	if timeout, ok := params["timeout"].(time.Duration); ok {
		operation.TTL = timeout
		operation.Metadata["timeout"] = timeout.String()
	} else {
		// 大消息需要更长的超时时间
		timeout := time.Duration(size/1024) * time.Second
		if timeout < 30*time.Second {
			timeout = 30 * time.Second
		}
		operation.TTL = timeout
		operation.Metadata["timeout"] = timeout.String()
	}

	return operation, nil
}

// createStressTestOperation 创建压力测试操作
func (f *WebSocketOperationFactory) createStressTestOperation(params map[string]interface{}, operation interfaces.Operation) (interfaces.Operation, error) {
	connections, ok := params["connections"].(int)
	if !ok || connections <= 0 {
		connections = 100
	}

	frequency, ok := params["frequency"].(int)
	if !ok || frequency <= 0 {
		frequency = 10 // 每秒10个消息
	}

	duration, ok := params["duration"].(time.Duration)
	if !ok {
		duration = 60 * time.Second
	}

	operation.Key = "stress_test"
	operation.Value = "stress test"
	operation.TTL = duration
	operation.Metadata["message_type"] = "stress"
	operation.Metadata["connections"] = strconv.Itoa(connections)
	operation.Metadata["frequency"] = strconv.Itoa(frequency)
	operation.Metadata["duration"] = duration.String()

	return operation, nil
}

// 参数验证方法

// validateSendTextParams 验证发送文本消息参数
func (f *WebSocketOperationFactory) validateSendTextParams(params map[string]interface{}) error {
	if message, ok := params["message"]; ok {
		if _, ok := message.(string); !ok {
			return fmt.Errorf("message must be a string")
		}
	}

	if timeout, ok := params["timeout"]; ok {
		if _, ok := timeout.(time.Duration); !ok {
			return fmt.Errorf("timeout must be a time.Duration")
		}
	}

	return nil
}

// validateSendBinaryParams 验证发送二进制消息参数
func (f *WebSocketOperationFactory) validateSendBinaryParams(params map[string]interface{}) error {
	if data, ok := params["data"]; ok {
		if _, ok := data.([]byte); !ok {
			return fmt.Errorf("data must be []byte")
		}
	}

	if size, ok := params["size"]; ok {
		if s, ok := size.(int); !ok || s <= 0 {
			return fmt.Errorf("size must be a positive integer")
		}
	}

	if timeout, ok := params["timeout"]; ok {
		if _, ok := timeout.(time.Duration); !ok {
			return fmt.Errorf("timeout must be a time.Duration")
		}
	}

	return nil
}

// validateEchoTestParams 验证回显测试参数
func (f *WebSocketOperationFactory) validateEchoTestParams(params map[string]interface{}) error {
	if message, ok := params["message"]; ok {
		if _, ok := message.(string); !ok {
			return fmt.Errorf("message must be a string")
		}
	}

	if timeout, ok := params["timeout"]; ok {
		if t, ok := timeout.(time.Duration); !ok || t <= 0 {
			return fmt.Errorf("timeout must be a positive time.Duration")
		}
	}

	return nil
}

// validatePingPongParams 验证心跳测试参数
func (f *WebSocketOperationFactory) validatePingPongParams(params map[string]interface{}) error {
	if timeout, ok := params["timeout"]; ok {
		if t, ok := timeout.(time.Duration); !ok || t <= 0 {
			return fmt.Errorf("timeout must be a positive time.Duration")
		}
	}

	if interval, ok := params["interval"]; ok {
		if i, ok := interval.(time.Duration); !ok || i <= 0 {
			return fmt.Errorf("interval must be a positive time.Duration")
		}
	}

	return nil
}

// validateBroadcastParams 验证广播测试参数
func (f *WebSocketOperationFactory) validateBroadcastParams(params map[string]interface{}) error {
	if message, ok := params["message"]; ok {
		if _, ok := message.(string); !ok {
			return fmt.Errorf("message must be a string")
		}
	}

	if targets, ok := params["targets"]; ok {
		if t, ok := targets.(int); !ok || t <= 0 {
			return fmt.Errorf("targets must be a positive integer")
		}
	}

	if broadcastType, ok := params["broadcast_type"]; ok {
		if bt, ok := broadcastType.(string); !ok {
			return fmt.Errorf("broadcast_type must be a string")
		} else if bt != "text" && bt != "binary" {
			return fmt.Errorf("broadcast_type must be 'text' or 'binary'")
		}
	}

	return nil
}

// validateSubscribeParams 验证订阅参数
func (f *WebSocketOperationFactory) validateSubscribeParams(params map[string]interface{}) error {
	if channel, ok := params["channel"]; ok {
		if _, ok := channel.(string); !ok {
			return fmt.Errorf("channel must be a string")
		}
	}

	if timeout, ok := params["timeout"]; ok {
		if t, ok := timeout.(time.Duration); !ok || t <= 0 {
			return fmt.Errorf("timeout must be a positive time.Duration")
		}
	}

	return nil
}

// validateLargeMessageParams 验证大消息参数
func (f *WebSocketOperationFactory) validateLargeMessageParams(params map[string]interface{}) error {
	if size, ok := params["size"]; ok {
		if s, ok := size.(int); !ok || s <= 0 {
			return fmt.Errorf("size must be a positive integer")
		} else if s > 100*1024*1024 { // 100MB限制
			return fmt.Errorf("size too large, maximum is 100MB")
		}
	}

	if timeout, ok := params["timeout"]; ok {
		if t, ok := timeout.(time.Duration); !ok || t <= 0 {
			return fmt.Errorf("timeout must be a positive time.Duration")
		}
	}

	return nil
}

// validateStressTestParams 验证压力测试参数
func (f *WebSocketOperationFactory) validateStressTestParams(params map[string]interface{}) error {
	if connections, ok := params["connections"]; ok {
		if c, ok := connections.(int); !ok || c <= 0 {
			return fmt.Errorf("connections must be a positive integer")
		}
	}

	if frequency, ok := params["frequency"]; ok {
		if f, ok := frequency.(int); !ok || f <= 0 {
			return fmt.Errorf("frequency must be a positive integer")
		}
	}

	if duration, ok := params["duration"]; ok {
		if d, ok := duration.(time.Duration); !ok || d <= 0 {
			return fmt.Errorf("duration must be a positive time.Duration")
		}
	}

	return nil
}

// RegisterWebSocketOperations 注册所有WebSocket操作到操作注册表
func RegisterWebSocketOperations(registry *utils.OperationRegistry) {
	operations := []string{
		"send_text",
		"send_binary", 
		"echo_test",
		"ping_pong",
		"broadcast",
		"subscribe",
		"large_message",
		"stress_test",
	}

	for _, opType := range operations {
		factory := NewWebSocketOperationFactory(opType)
		registry.Register(opType, factory)
	}
}

// GetSupportedOperations 获取支持的操作类型
func GetSupportedOperations() []string {
	return []string{
		"send_text",
		"send_binary",
		"echo_test",
		"ping_pong",
		"broadcast",
		"subscribe",
		"large_message",
		"stress_test",
	}
}



