package grpc

import (
	"abc-runner/app/adapters/grpc/config"
	"abc-runner/app/core/interfaces"
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	// 暂时注释gRPC依赖，避免编译错误
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
)

// GRPCAdapter gRPC协议适配器
type GRPCAdapter struct {
	config           *config.GRPCConfig
	connections      map[string]*GRPCConnection
	metricsCollector interfaces.DefaultMetricsCollector
	mu               sync.RWMutex
	isConnected      bool
	totalCalls       int64
	successfulCalls  int64
	failedCalls      int64
	totalLatency     time.Duration
}

// GRPCConnection gRPC连接包装器 (模拟版本)
type GRPCConnection struct {
	// conn     *grpc.ClientConn // 暂时注释
	address  string
	lastUsed time.Time
	mu       sync.RWMutex
}

// NewGRPCAdapter 创建新的gRPC适配器
func NewGRPCAdapter(metricsCollector interfaces.DefaultMetricsCollector) *GRPCAdapter {
	return &GRPCAdapter{
		connections:      make(map[string]*GRPCConnection),
		metricsCollector: metricsCollector,
		isConnected:      false,
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
	if err := adapter.createConnectionPool(ctx); err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}

	adapter.isConnected = true
	log.Printf("Successfully connected to gRPC server: %s:%d",
		adapter.config.Connection.Address, adapter.config.Connection.Port)

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
	conn, err := adapter.getConnection()
	if err != nil {
		return nil, err
	}

	// 模拟一元调用（实际使用时需要根据具体的proto定义）
	log.Printf("Executing unary call: %s.%s",
		adapter.config.GRPCSpecific.ServiceName,
		adapter.config.GRPCSpecific.MethodName)

	// 添加认证metadata
	ctx = adapter.addAuthMetadata(ctx)

	// 这里应该调用实际的gRPC方法
	// 由于没有具体的proto定义，我们模拟一个调用
	startTime := time.Now()

	// 模拟调用延迟
	time.Sleep(10 * time.Millisecond)

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
	conn, err := adapter.getConnection()
	if err != nil {
		return nil, err
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
	conn, err := adapter.getConnection()
	if err != nil {
		return nil, err
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
	conn, err := adapter.getConnection()
	if err != nil {
		return nil, err
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

	// 关闭所有连接
	for address, conn := range adapter.connections {
		if conn.conn != nil {
			conn.conn.Close()
			log.Printf("Closed gRPC connection to %s", address)
		}
	}

	adapter.connections = make(map[string]*GRPCConnection)
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
		avgLatency = float64(adapter.totalLatency.Nanoseconds()) / float64(adapter.totalCalls) / 1e6 // Convert to milliseconds
	}

	successRate := float64(0)
	if adapter.totalCalls > 0 {
		successRate = float64(adapter.successfulCalls) / float64(adapter.totalCalls) * 100
	}

	return map[string]interface{}{
		"protocol":         "grpc",
		"total_calls":      adapter.totalCalls,
		"successful_calls": adapter.successfulCalls,
		"failed_calls":     adapter.failedCalls,
		"success_rate":     successRate,
		"avg_latency_ms":   avgLatency,
		"connections":      len(adapter.connections),
		"service_name":     adapter.config.GRPCSpecific.ServiceName,
		"method_name":      adapter.config.GRPCSpecific.MethodName,
	}
}

// HealthCheck 健康检查
func (adapter *GRPCAdapter) HealthCheck(ctx context.Context) error {
	if !adapter.isConnected {
		return fmt.Errorf("adapter not connected")
	}

	// 检查至少一个连接是否可用
	for address, conn := range adapter.connections {
		if conn.conn != nil {
			// 简单的状态检查
			state := conn.conn.GetState()
			log.Printf("Connection %s state: %v", address, state)
		}
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
	grpcConfig := config.NewDefaultGRPCConfig()

	if address, ok := cfg["address"].(string); ok {
		grpcConfig.Connection.Address = address
	}
	if port, ok := cfg["port"].(int); ok {
		grpcConfig.Connection.Port = port
	}
	if timeout, ok := cfg["timeout"].(time.Duration); ok {
		grpcConfig.Connection.Timeout = timeout
	}
	if serviceName, ok := cfg["service_name"].(string); ok {
		grpcConfig.GRPCSpecific.ServiceName = serviceName
	}
	if methodName, ok := cfg["method_name"].(string); ok {
		grpcConfig.GRPCSpecific.MethodName = methodName
	}
	if testCase, ok := cfg["test_case"].(string); ok {
		grpcConfig.BenchMark.TestCase = testCase
	}
	if parallels, ok := cfg["parallels"].(int); ok {
		grpcConfig.BenchMark.Parallels = parallels
		grpcConfig.Connection.Pool.PoolSize = parallels
	}
	if total, ok := cfg["total"].(int); ok {
		grpcConfig.BenchMark.Total = total
	}

	return grpcConfig, nil
}

// createConnectionPool 创建连接池
func (adapter *GRPCAdapter) createConnectionPool(ctx context.Context) error {
	poolSize := adapter.config.Connection.Pool.PoolSize
	address := fmt.Sprintf("%s:%d", adapter.config.Connection.Address, adapter.config.Connection.Port)

	for i := 0; i < poolSize; i++ {
		conn, err := adapter.createConnection(ctx, address)
		if err != nil {
			return fmt.Errorf("failed to create connection %d: %w", i, err)
		}

		connectionKey := fmt.Sprintf("%s_%d", address, i)
		adapter.connections[connectionKey] = &GRPCConnection{
			conn:     conn,
			address:  address,
			lastUsed: time.Now(),
		}
	}

	log.Printf("Created gRPC connection pool with %d connections", poolSize)
	return nil
}

// createConnection 创建gRPC连接
func (adapter *GRPCAdapter) createConnection(ctx context.Context, address string) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption

	// TLS配置
	if adapter.config.GRPCSpecific.TLS.Enabled {
		var creds credentials.TransportCredentials
		if adapter.config.GRPCSpecific.TLS.InsecureSkipVerify {
			creds = credentials.NewTLS(nil)
		} else {
			creds = insecure.NewCredentials()
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Keep-alive配置
	if adapter.config.Connection.KeepAlive {
		kacp := keepalive.ClientParameters{
			Time:                adapter.config.Connection.KeepAlivePeriod,
			Timeout:             10 * time.Second,
			PermitWithoutStream: true,
		}
		opts = append(opts, grpc.WithKeepaliveParams(kacp))
	}

	// 负载均衡
	opts = append(opts, grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"loadBalancingConfig": [{"%s":{}}]}`,
		adapter.config.GRPCSpecific.LoadBalancing)))

	// 创建连接
	conn, err := grpc.DialContext(ctx, address, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial gRPC server: %w", err)
	}

	return conn, nil
}

// getConnection 获取可用连接
func (adapter *GRPCAdapter) getConnection() (*grpc.ClientConn, error) {
	adapter.mu.RLock()
	defer adapter.mu.RUnlock()

	// 简单的轮询策略
	for _, conn := range adapter.connections {
		conn.mu.Lock()
		conn.lastUsed = time.Now()
		result := conn.conn
		conn.mu.Unlock()
		return result, nil
	}

	return nil, fmt.Errorf("no available connections")
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
			// 实际应用中应该进行Base64编码
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
		adapter.metricsCollector.RecordLatency(operationType, duration)
		if success {
			adapter.metricsCollector.IncrementCounter("grpc_successful_calls")
		} else {
			adapter.metricsCollector.IncrementCounter("grpc_failed_calls")
		}
	}
}
