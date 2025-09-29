package grpc

import (
	"abc-runner/app/adapters/grpc/config"
	"abc-runner/app/adapters/grpc/connection"
	"abc-runner/app/core/interfaces"
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc/metadata"
)

// GRPCAdapter gRPC协议适配器
type GRPCAdapter struct {
	config           *config.GRPCConfig
	connectionPool   *connection.ConnectionPool
	metricsCollector interfaces.DefaultMetricsCollector
	mu               sync.RWMutex
	isConnected      bool

	// 统计信息
	totalCalls      int64
	successfulCalls int64
	failedCalls     int64
	totalLatency    time.Duration
	startTime       time.Time
}

// NewGRPCAdapter 创建新的gRPC适配器
func NewGRPCAdapter(metricsCollector interfaces.DefaultMetricsCollector) *GRPCAdapter {
	return &GRPCAdapter{
		metricsCollector: metricsCollector,
		isConnected:      false,
		startTime:        time.Now(),
	}
}

// Connect 连接到gRPC服务器
func (adapter *GRPCAdapter) Connect(ctx context.Context, cfg interfaces.Config) error {
	adapter.mu.Lock()
	defer adapter.mu.Unlock()

	// 转换配置
	grpcConfig, err := adapter.parseConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to parse gRPC config: %w", err)
	}

	adapter.config = grpcConfig

	// 验证配置
	if err := adapter.config.Validate(); err != nil {
		return fmt.Errorf("invalid gRPC config: %w", err)
	}

	// 创建连接池
	adapter.connectionPool = connection.NewConnectionPool(adapter.config)
	if err := adapter.connectionPool.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize connection pool: %w", err)
	}

	adapter.isConnected = true
	log.Printf("Successfully connected to gRPC server: %s:%d with %d connections",
		adapter.config.Connection.Address, adapter.config.Connection.Port,
		adapter.config.Connection.Pool.PoolSize)

	return nil
}

// Execute 执行gRPC操作
func (adapter *GRPCAdapter) Execute(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	if !adapter.isConnected {
		return nil, fmt.Errorf("adapter not connected")
	}

	startTime := time.Now()

	// 根据操作类型执行不同的测试
	var result *interfaces.OperationResult
	var err error

	switch operation.Type {
	case "unary_call":
		result, err = adapter.executeUnaryCall(ctx, operation)
	case "server_stream":
		result, err = adapter.executeServerStream(ctx, operation)
	case "client_stream":
		result, err = adapter.executeClientStream(ctx, operation)
	case "bidirectional_stream":
		result, err = adapter.executeBidirectionalStream(ctx, operation)
	default:
		err = fmt.Errorf("unsupported operation type: %s", operation.Type)
	}

	// 记录指标
	duration := time.Since(startTime)
	adapter.recordMetrics(operation.Type, duration, err == nil)

	if err != nil {
		return nil, fmt.Errorf("gRPC operation failed: %w", err)
	}

	return result, nil
}

// executeUnaryCall 执行一元调用
func (adapter *GRPCAdapter) executeUnaryCall(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	// 获取连接
	connWrapper, err := adapter.connectionPool.GetConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	conn := connWrapper.GetConn()
	if conn == nil {
		return nil, fmt.Errorf("connection is nil")
	}

	// 模拟一元调用执行
	log.Printf("Executing unary call: %s.%s with connection",
		adapter.config.GRPCSpecific.ServiceName,
		adapter.config.GRPCSpecific.MethodName)

	// 添加认证metadata
	ctx = adapter.addAuthMetadata(ctx)

	startTime := time.Now()
	time.Sleep(10 * time.Millisecond) // 模拟调用延迟
	duration := time.Since(startTime)

	return &interfaces.OperationResult{
		Success:  true,
		Duration: duration,
		Value:    fmt.Sprintf("Unary call result for key: %s", operation.Key),
		Metadata: map[string]interface{}{
			"operation_type": "unary_call",
			"service":        adapter.config.GRPCSpecific.ServiceName,
			"method":         adapter.config.GRPCSpecific.MethodName,
			"timestamp":      time.Now(),
		},
	}, nil
}

// executeServerStream 执行服务器流调用
func (adapter *GRPCAdapter) executeServerStream(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	// 获取连接
	connWrapper, err := adapter.connectionPool.GetConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	conn := connWrapper.GetConn()
	if conn == nil {
		return nil, fmt.Errorf("connection is nil")
	}

	log.Printf("Executing server stream call: %s.%s",
		adapter.config.GRPCSpecific.ServiceName,
		adapter.config.GRPCSpecific.MethodName)

	ctx = adapter.addAuthMetadata(ctx)
	startTime := time.Now()

	// 模拟服务器流
	messageCount := 5
	for i := 0; i < messageCount; i++ {
		time.Sleep(5 * time.Millisecond)
		log.Printf("Received stream message %d", i+1)
	}

	duration := time.Since(startTime)

	return &interfaces.OperationResult{
		Success:  true,
		Duration: duration,
		Value:    fmt.Sprintf("Server stream completed, received %d messages", messageCount),
		Metadata: map[string]interface{}{
			"operation_type": "server_stream",
			"message_count":  messageCount,
			"timestamp":      time.Now(),
		},
	}, nil
}

// executeClientStream 执行客户端流调用
func (adapter *GRPCAdapter) executeClientStream(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	// 获取连接
	connWrapper, err := adapter.connectionPool.GetConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	conn := connWrapper.GetConn()
	if conn == nil {
		return nil, fmt.Errorf("connection is nil")
	}

	log.Printf("Executing client stream call: %s.%s",
		adapter.config.GRPCSpecific.ServiceName,
		adapter.config.GRPCSpecific.MethodName)

	ctx = adapter.addAuthMetadata(ctx)
	startTime := time.Now()

	// 模拟客户端流
	messageCount := 3
	for i := 0; i < messageCount; i++ {
		time.Sleep(5 * time.Millisecond)
		log.Printf("Sent stream message %d", i+1)
	}

	duration := time.Since(startTime)

	return &interfaces.OperationResult{
		Success:  true,
		Duration: duration,
		Value:    fmt.Sprintf("Client stream completed, sent %d messages", messageCount),
		Metadata: map[string]interface{}{
			"operation_type": "client_stream",
			"message_count":  messageCount,
			"timestamp":      time.Now(),
		},
	}, nil
}

// executeBidirectionalStream 执行双向流调用
func (adapter *GRPCAdapter) executeBidirectionalStream(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	// 获取连接
	connWrapper, err := adapter.connectionPool.GetConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	conn := connWrapper.GetConn()
	if conn == nil {
		return nil, fmt.Errorf("connection is nil")
	}

	log.Printf("Executing bidirectional stream call: %s.%s",
		adapter.config.GRPCSpecific.ServiceName,
		adapter.config.GRPCSpecific.MethodName)

	ctx = adapter.addAuthMetadata(ctx)
	startTime := time.Now()

	// 模拟双向流
	messageCount := 4
	for i := 0; i < messageCount; i++ {
		time.Sleep(8 * time.Millisecond)
		log.Printf("Bidirectional stream message %d", i+1)
	}

	duration := time.Since(startTime)

	return &interfaces.OperationResult{
		Success:  true,
		Duration: duration,
		Value:    fmt.Sprintf("Bidirectional stream completed, exchanged %d messages", messageCount),
		Metadata: map[string]interface{}{
			"operation_type": "bidirectional_stream",
			"message_count":  messageCount,
			"timestamp":      time.Now(),
		},
	}, nil
}

// Close 关闭适配器
func (adapter *GRPCAdapter) Close() error {
	adapter.mu.Lock()
	defer adapter.mu.Unlock()

	if !adapter.isConnected {
		return nil
	}

	// 关闭连接池
	if adapter.connectionPool != nil {
		if err := adapter.connectionPool.Close(); err != nil {
			log.Printf("Error closing connection pool: %v", err)
		}
	}

	adapter.isConnected = false
	log.Println("gRPC adapter closed")
	return nil
}

// GetProtocolMetrics 获取协议指标
func (adapter *GRPCAdapter) GetProtocolMetrics() map[string]interface{} {
	adapter.mu.RLock()
	defer adapter.mu.RUnlock()

	var avgLatency float64
	if adapter.totalCalls > 0 {
		avgLatency = float64(adapter.totalLatency.Nanoseconds()) / float64(adapter.totalCalls) / 1e6
	}

	successRate := float64(0)
	if adapter.totalCalls > 0 {
		successRate = float64(adapter.successfulCalls) / float64(adapter.totalCalls) * 100
	}

	metrics := map[string]interface{}{
		"protocol":         "grpc",
		"total_calls":      adapter.totalCalls,
		"successful_calls": adapter.successfulCalls,
		"failed_calls":     adapter.failedCalls,
		"success_rate":     successRate,
		"avg_latency_ms":   avgLatency,
		"uptime":           time.Since(adapter.startTime),
		"service_name":     adapter.config.GRPCSpecific.ServiceName,
		"method_name":      adapter.config.GRPCSpecific.MethodName,
	}

	// 添加连接池指标
	if adapter.connectionPool != nil {
		poolStats := adapter.connectionPool.GetStats()
		metrics["connection_pool"] = poolStats
	}

	return metrics
}

// HealthCheck 健康检查
func (adapter *GRPCAdapter) HealthCheck(ctx context.Context) error {
	if !adapter.isConnected {
		return fmt.Errorf("adapter not connected")
	}

	if adapter.connectionPool == nil {
		return fmt.Errorf("connection pool not initialized")
	}

	// 尝试获取一个连接来验证健康状态
	connWrapper, err := adapter.connectionPool.GetConnection()
	if err != nil {
		return fmt.Errorf("failed to get healthy connection: %w", err)
	}

	if !connWrapper.IsHealthy() {
		return fmt.Errorf("connection is not healthy")
	}

	return nil
}

// GetProtocolName 获取协议名称
func (adapter *GRPCAdapter) GetProtocolName() string {
	return "grpc"
}

// GetMetricsCollector 获取指标收集器
func (adapter *GRPCAdapter) GetMetricsCollector() interfaces.DefaultMetricsCollector {
	return adapter.metricsCollector
}

// 内部辅助方法

// parseConfig 解析配置
func (adapter *GRPCAdapter) parseConfig(cfg interfaces.Config) (*config.GRPCConfig, error) {
	// 如果传入的是gRPC配置类型，直接使用
	if gCfg, ok := cfg.(*config.GRPCConfig); ok {
		return gCfg, nil
	}

	// 使用默认配置
	grpcConfig := config.NewDefaultGRPCConfig()
	return grpcConfig, nil
}

// addAuthMetadata 添加认证metadata
func (adapter *GRPCAdapter) addAuthMetadata(ctx context.Context) context.Context {
	if !adapter.config.GRPCSpecific.Auth.Enabled {
		return ctx
	}

	md := metadata.New(map[string]string{})

	switch adapter.config.GRPCSpecific.Auth.Method {
	case "token":
		if adapter.config.GRPCSpecific.Auth.Token != "" {
			md.Set("authorization", "Bearer "+adapter.config.GRPCSpecific.Auth.Token)
		}
	case "basic":
		if adapter.config.GRPCSpecific.Auth.Username != "" && adapter.config.GRPCSpecific.Auth.Password != "" {
			md.Set("authorization", fmt.Sprintf("Basic %s:%s",
				adapter.config.GRPCSpecific.Auth.Username,
				adapter.config.GRPCSpecific.Auth.Password))
		}
	}

	// 添加自定义metadata
	for key, value := range adapter.config.GRPCSpecific.Auth.Metadata {
		md.Set(key, value)
	}

	return metadata.NewOutgoingContext(ctx, md)
}

// recordMetrics 记录指标
func (adapter *GRPCAdapter) recordMetrics(operationType string, duration time.Duration, success bool) {
	adapter.mu.Lock()
	defer adapter.mu.Unlock()

	adapter.totalCalls++
	adapter.totalLatency += duration

	if success {
		adapter.successfulCalls++
	} else {
		adapter.failedCalls++
	}

	// 记录到指标收集器
	if adapter.metricsCollector != nil {
		result := &interfaces.OperationResult{
			Success:  success,
			Duration: duration,
			IsRead:   false, // gRPC调用一般被视为写操作
		}
		adapter.metricsCollector.Record(result)
	}
}
