package grpc

import (
	"fmt"
	"strconv"
	"time"

	"abc-runner/app/adapters/grpc/config"
	"abc-runner/app/core/execution"
	"abc-runner/app/core/interfaces"
)

// OperationFactory gRPC操作工厂
type OperationFactory struct {
	config     *config.GRPCConfig
	testCase   string
	dataSize   int
	serviceName string
	methodName  string
}

// NewOperationFactory 创建gRPC操作工厂
func NewOperationFactory(cfg *config.GRPCConfig) *OperationFactory {
	return &OperationFactory{
		config:      cfg,
		testCase:    cfg.BenchMark.TestCase,
		dataSize:    cfg.BenchMark.DataSize,
		serviceName: cfg.GRPCSpecific.ServiceName,
		methodName:  cfg.GRPCSpecific.MethodName,
	}
}

// CreateOperation 创建gRPC操作
func (f *OperationFactory) CreateOperation(jobID int, config execution.BenchmarkConfig) interfaces.Operation {
	// 生成操作键
	key := f.generateKey(jobID)
	
	// 生成请求数据
	requestData := f.generateRequestData(jobID)
	
	// 创建操作特定参数
	params := map[string]interface{}{
		"job_id":           jobID,
		"data_size":        f.dataSize,
		"test_case":        f.testCase,
		"service_name":     f.serviceName,
		"method_name":      f.methodName,
		"compression":      f.config.GRPCSpecific.Compression,
		"load_balancing":   f.config.GRPCSpecific.LoadBalancing,
		"max_message_size": f.config.GRPCSpecific.MaxMessageSize,
		"auth_enabled":     f.config.GRPCSpecific.Auth.Enabled,
	}
	
	// 添加认证信息
	if f.config.GRPCSpecific.Auth.Enabled {
		params["auth_method"] = f.config.GRPCSpecific.Auth.Method
		params["auth_token"] = f.config.GRPCSpecific.Auth.Token
		params["auth_metadata"] = f.config.GRPCSpecific.Auth.Metadata
	}
	
	// 创建操作元数据
	metadata := map[string]string{
		"operation_type":   f.testCase,
		"protocol":         "grpc",
		"service":          f.serviceName,
		"method":           f.methodName,
		"job_id":          strconv.Itoa(jobID),
		"timestamp":        time.Now().Format(time.RFC3339),
		"compression":      f.config.GRPCSpecific.Compression,
	}
	
	// 根据测试用例添加特定元数据
	switch f.testCase {
	case "server_stream":
		metadata["stream_type"] = "server"
		metadata["expected_messages"] = "5"
	case "client_stream":
		metadata["stream_type"] = "client"
		metadata["messages_to_send"] = "3"
	case "bidirectional_stream":
		metadata["stream_type"] = "bidirectional"
		metadata["message_pairs"] = "4"
	}
	
	return interfaces.Operation{
		Type:     f.testCase,
		Key:      key,
		Value:    requestData,
		Params:   params,
		TTL:      f.config.BenchMark.TTL,
		Metadata: metadata,
	}
}

// GetOperationType 获取操作类型
func (f *OperationFactory) GetOperationType() string {
	return f.testCase
}

// GetConfig 获取配置信息
func (f *OperationFactory) GetConfig() *config.GRPCConfig {
	return f.config
}

// GetSupportedOperations 获取支持的操作类型
func (f *OperationFactory) GetSupportedOperations() []string {
	return []string{"unary_call", "server_stream", "client_stream", "bidirectional_stream"}
}

// ValidateTestCase 验证测试用例是否支持
func (f *OperationFactory) ValidateTestCase(testCase string) error {
	supportedCases := f.GetSupportedOperations()
	for _, supported := range supportedCases {
		if testCase == supported {
			return nil
		}
	}
	return fmt.Errorf("unsupported test case: %s, supported: %v", testCase, supportedCases)
}

// generateKey 生成操作键
func (f *OperationFactory) generateKey(jobID int) string {
	switch f.testCase {
	case "unary_call":
		return fmt.Sprintf("grpc_unary_%s_%s_%d", f.serviceName, f.methodName, jobID)
	case "server_stream":
		return fmt.Sprintf("grpc_server_stream_%s_%s_%d", f.serviceName, f.methodName, jobID)
	case "client_stream":
		return fmt.Sprintf("grpc_client_stream_%s_%s_%d", f.serviceName, f.methodName, jobID)
	case "bidirectional_stream":
		return fmt.Sprintf("grpc_bidi_stream_%s_%s_%d", f.serviceName, f.methodName, jobID)
	default:
		return fmt.Sprintf("grpc_op_%s_%s_%d", f.serviceName, f.methodName, jobID)
	}
}

// generateRequestData 生成请求数据
func (f *OperationFactory) generateRequestData(jobID int) interface{} {
	switch f.testCase {
	case "unary_call":
		return f.generateUnaryCallData(jobID)
	case "server_stream":
		return f.generateServerStreamData(jobID)
	case "client_stream":
		return f.generateClientStreamData(jobID)
	case "bidirectional_stream":
		return f.generateBidirectionalStreamData(jobID)
	default:
		return f.generateDefaultData(jobID)
	}
}

// generateUnaryCallData 生成一元调用数据
func (f *OperationFactory) generateUnaryCallData(jobID int) map[string]interface{} {
	// 创建具有特定大小的测试数据
	payload := make([]byte, f.dataSize)
	pattern := fmt.Sprintf("UNARY_CALL_%d_", jobID)
	patternBytes := []byte(pattern)
	
	for i := range payload {
		payload[i] = patternBytes[i%len(patternBytes)]
	}
	
	return map[string]interface{}{
		"message_type": "unary_request",
		"job_id":       jobID,
		"payload":      payload,
		"service":      f.serviceName,
		"method":       f.methodName,
		"timestamp":    time.Now().Unix(),
		"data_size":    f.dataSize,
	}
}

// generateServerStreamData 生成服务器流数据
func (f *OperationFactory) generateServerStreamData(jobID int) map[string]interface{} {
	return map[string]interface{}{
		"message_type":     "server_stream_request",
		"job_id":           jobID,
		"service":          f.serviceName,
		"method":           f.methodName,
		"stream_count":     5, // 期望接收5条消息
		"timestamp":        time.Now().Unix(),
		"request_data":     fmt.Sprintf("SERVER_STREAM_REQUEST_%d", jobID),
	}
}

// generateClientStreamData 生成客户端流数据
func (f *OperationFactory) generateClientStreamData(jobID int) map[string]interface{} {
	messages := make([]string, 3) // 发送3条消息
	for i := range messages {
		messages[i] = fmt.Sprintf("CLIENT_STREAM_MSG_%d_%d", jobID, i)
	}
	
	return map[string]interface{}{
		"message_type": "client_stream_request",
		"job_id":       jobID,
		"service":      f.serviceName,
		"method":       f.methodName,
		"messages":     messages,
		"timestamp":    time.Now().Unix(),
	}
}

// generateBidirectionalStreamData 生成双向流数据
func (f *OperationFactory) generateBidirectionalStreamData(jobID int) map[string]interface{} {
	messagePairs := make([]map[string]string, 4) // 4对消息交换
	for i := range messagePairs {
		messagePairs[i] = map[string]string{
			"send":    fmt.Sprintf("BIDI_SEND_%d_%d", jobID, i),
			"expect":  fmt.Sprintf("BIDI_RESP_%d_%d", jobID, i),
		}
	}
	
	return map[string]interface{}{
		"message_type":   "bidirectional_stream_request",
		"job_id":         jobID,
		"service":        f.serviceName,
		"method":         f.methodName,
		"message_pairs":  messagePairs,
		"timestamp":      time.Now().Unix(),
	}
}

// generateDefaultData 生成默认数据
func (f *OperationFactory) generateDefaultData(jobID int) map[string]interface{} {
	return map[string]interface{}{
		"message_type": "default_request",
		"job_id":       jobID,
		"service":      f.serviceName,
		"method":       f.methodName,
		"timestamp":    time.Now().Unix(),
		"data":         fmt.Sprintf("DEFAULT_DATA_%d", jobID),
	}
}

// 确保实现了execution.OperationFactory接口
var _ execution.OperationFactory = (*OperationFactory)(nil)