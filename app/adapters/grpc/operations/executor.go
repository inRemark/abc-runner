package operations

import (
	"context"
	"fmt"
	"log"
	"time"

	"abc-runner/app/adapters/grpc/config"
	"abc-runner/app/adapters/grpc/connection"
	"abc-runner/app/core/interfaces"

	"google.golang.org/grpc/metadata"
)

// GRPCOperations gRPC操作执行器 - 遵循统一架构模式
type GRPCOperations struct {
	connectionPool   *connection.ConnectionPool
	config           *config.GRPCConfig
	metricsCollector interfaces.DefaultMetricsCollector
}

// NewGRPCOperations 创建gRPC操作执行器
func NewGRPCOperations(
	connectionPool *connection.ConnectionPool,
	config *config.GRPCConfig,
	metricsCollector interfaces.DefaultMetricsCollector,
) *GRPCOperations {
	return &GRPCOperations{
		connectionPool:   connectionPool,
		config:           config,
		metricsCollector: metricsCollector,
	}
}

// ExecuteOperation 执行gRPC操作 - 统一操作入口
func (g *GRPCOperations) ExecuteOperation(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	startTime := time.Now()
	result := &interfaces.OperationResult{
		IsRead:   g.isReadOperation(operation.Type),
		Metadata: make(map[string]interface{}),
	}

	// 获取连接
	connWrapper, err := g.connectionPool.GetConnection()
	if err != nil {
		result.Success = false
		result.Error = fmt.Errorf("failed to get connection: %w", err)
		result.Duration = time.Since(startTime)
		return result, result.Error
	}

	conn := connWrapper.GetConn()
	if conn == nil {
		result.Success = false
		result.Error = fmt.Errorf("connection is nil")
		result.Duration = time.Since(startTime)
		return result, result.Error
	}

	// 添加认证metadata
	ctx = g.addAuthMetadata(ctx)

	var opErr error
	switch operation.Type {
	case "unary_call":
		opErr = g.executeUnaryCall(ctx, operation, result)
	case "server_stream":
		opErr = g.executeServerStream(ctx, operation, result)
	case "client_stream":
		opErr = g.executeClientStream(ctx, operation, result)
	case "bidirectional_stream":
		opErr = g.executeBidirectionalStream(ctx, operation, result)
	default:
		opErr = fmt.Errorf("unsupported operation type: %s", operation.Type)
	}

	result.Success = opErr == nil
	result.Error = opErr
	result.Duration = time.Since(startTime)

	// 添加操作特定元数据
	for k, v := range operation.Metadata {
		result.Metadata[k] = v
	}
	result.Metadata["protocol"] = "grpc"
	result.Metadata["operation_type"] = operation.Type
	result.Metadata["service"] = g.config.GRPCSpecific.ServiceName
	result.Metadata["method"] = g.config.GRPCSpecific.MethodName
	result.Metadata["execution_time_ms"] = float64(result.Duration.Nanoseconds()) / 1e6
	result.Metadata["timestamp"] = time.Now()

	return result, opErr
}

// executeUnaryCall 执行一元调用
func (g *GRPCOperations) executeUnaryCall(ctx context.Context, operation interfaces.Operation, result *interfaces.OperationResult) error {
	log.Printf("Executing unary call: %s.%s",
		g.config.GRPCSpecific.ServiceName,
		g.config.GRPCSpecific.MethodName)

	// 模拟一元调用执行
	operationStartTime := time.Now()
	time.Sleep(10 * time.Millisecond) // 模拟调用延迟
	operationDuration := time.Since(operationStartTime)

	result.Value = fmt.Sprintf("Unary call result for key: %s", operation.Key)
	result.Metadata["operation_duration_ms"] = float64(operationDuration.Nanoseconds()) / 1e6
	result.Metadata["call_type"] = "unary"

	return nil
}

// executeServerStream 执行服务器流调用
func (g *GRPCOperations) executeServerStream(ctx context.Context, operation interfaces.Operation, result *interfaces.OperationResult) error {
	log.Printf("Executing server stream call: %s.%s",
		g.config.GRPCSpecific.ServiceName,
		g.config.GRPCSpecific.MethodName)

	operationStartTime := time.Now()

	// 模拟服务器流
	messageCount := 5
	for i := 0; i < messageCount; i++ {
		time.Sleep(5 * time.Millisecond)
		log.Printf("Received stream message %d", i+1)
	}

	operationDuration := time.Since(operationStartTime)

	result.Value = fmt.Sprintf("Server stream completed, received %d messages", messageCount)
	result.Metadata["operation_duration_ms"] = float64(operationDuration.Nanoseconds()) / 1e6
	result.Metadata["call_type"] = "server_stream"
	result.Metadata["message_count"] = messageCount

	return nil
}

// executeClientStream 执行客户端流调用
func (g *GRPCOperations) executeClientStream(ctx context.Context, operation interfaces.Operation, result *interfaces.OperationResult) error {
	log.Printf("Executing client stream call: %s.%s",
		g.config.GRPCSpecific.ServiceName,
		g.config.GRPCSpecific.MethodName)

	operationStartTime := time.Now()

	// 模拟客户端流
	messageCount := 3
	for i := 0; i < messageCount; i++ {
		time.Sleep(5 * time.Millisecond)
		log.Printf("Sent stream message %d", i+1)
	}

	operationDuration := time.Since(operationStartTime)

	result.Value = fmt.Sprintf("Client stream completed, sent %d messages", messageCount)
	result.Metadata["operation_duration_ms"] = float64(operationDuration.Nanoseconds()) / 1e6
	result.Metadata["call_type"] = "client_stream"
	result.Metadata["message_count"] = messageCount

	return nil
}

// executeBidirectionalStream 执行双向流调用
func (g *GRPCOperations) executeBidirectionalStream(ctx context.Context, operation interfaces.Operation, result *interfaces.OperationResult) error {
	log.Printf("Executing bidirectional stream call: %s.%s",
		g.config.GRPCSpecific.ServiceName,
		g.config.GRPCSpecific.MethodName)

	operationStartTime := time.Now()

	// 模拟双向流
	messageCount := 4
	for i := 0; i < messageCount; i++ {
		time.Sleep(8 * time.Millisecond)
		log.Printf("Bidirectional stream message %d", i+1)
	}

	operationDuration := time.Since(operationStartTime)

	result.Value = fmt.Sprintf("Bidirectional stream completed, exchanged %d messages", messageCount)
	result.Metadata["operation_duration_ms"] = float64(operationDuration.Nanoseconds()) / 1e6
	result.Metadata["call_type"] = "bidirectional_stream"
	result.Metadata["message_count"] = messageCount

	return nil
}

// addAuthMetadata 添加认证metadata
func (g *GRPCOperations) addAuthMetadata(ctx context.Context) context.Context {
	if !g.config.GRPCSpecific.Auth.Enabled {
		return ctx
	}

	md := metadata.New(map[string]string{})

	switch g.config.GRPCSpecific.Auth.Method {
	case "token":
		if g.config.GRPCSpecific.Auth.Token != "" {
			md.Set("authorization", "Bearer "+g.config.GRPCSpecific.Auth.Token)
		}
	case "basic":
		if g.config.GRPCSpecific.Auth.Username != "" && g.config.GRPCSpecific.Auth.Password != "" {
			md.Set("authorization", fmt.Sprintf("Basic %s:%s",
				g.config.GRPCSpecific.Auth.Username,
				g.config.GRPCSpecific.Auth.Password))
		}
	}

	// 添加自定义metadata
	for key, value := range g.config.GRPCSpecific.Auth.Metadata {
		md.Set(key, value)
	}

	return metadata.NewOutgoingContext(ctx, md)
}

// isReadOperation 判断是否为读操作
func (g *GRPCOperations) isReadOperation(operationType string) bool {
	readOperations := map[string]bool{
		"unary_call":           true,  // 一元调用通常是读取数据
		"server_stream":        true,  // 服务器流是从服务器读取数据
		"client_stream":        false, // 客户端流是向服务器发送数据
		"bidirectional_stream": true,  // 双向流主要是交互，设为读操作
	}
	return readOperations[operationType]
}

// GetSupportedOperations 获取支持的操作类型
func (g *GRPCOperations) GetSupportedOperations() []string {
	return []string{
		"unary_call",
		"server_stream",
		"client_stream",
		"bidirectional_stream",
	}
}
