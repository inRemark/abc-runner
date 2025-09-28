package tcp

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"abc-runner/app/adapters/tcp/config"
	"abc-runner/app/core/interfaces"
)

// TCPAdapter TCP协议适配器
type TCPAdapter struct {
	config           *config.TCPConfig
	connectionPool   *ConnectionPool
	metricsCollector interfaces.DefaultMetricsCollector
	mu               sync.RWMutex
	isConnected      bool
}

// NewTCPAdapter 创建TCP适配器
func NewTCPAdapter(metricsCollector interfaces.DefaultMetricsCollector) *TCPAdapter {
	return &TCPAdapter{
		metricsCollector: metricsCollector,
		isConnected:      false,
	}
}

// Connect 初始化连接
func (t *TCPAdapter) Connect(ctx context.Context, cfg interfaces.Config) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// 类型断言配置
	tcpConfig, ok := cfg.(*config.TCPConfig)
	if !ok {
		return fmt.Errorf("invalid config type for TCP adapter")
	}

	t.config = tcpConfig

	// 验证配置
	if err := tcpConfig.Validate(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	// 创建连接池
	pool, err := NewConnectionPool(tcpConfig)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}

	t.connectionPool = pool

	// 测试连接
	if err := t.testConnection(ctx); err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}

	t.isConnected = true
	return nil
}

// Execute 执行操作
func (t *TCPAdapter) Execute(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	startTime := time.Now()
	
	result := &interfaces.OperationResult{
		Success:  false,
		Duration: 0,
		IsRead:   false,
		Error:    nil,
		Value:    nil,
		Metadata: make(map[string]interface{}),
	}

	// 检查连接状态
	if !t.isConnected {
		result.Error = fmt.Errorf("adapter not connected")
		result.Duration = time.Since(startTime)
		return result, result.Error
	}

	// 获取连接
	conn, err := t.connectionPool.GetConnection()
	if err != nil {
		result.Error = fmt.Errorf("failed to get connection: %w", err)
		result.Duration = time.Since(startTime)
		return result, result.Error
	}
	defer t.connectionPool.ReturnConnection(conn)

	// 根据操作类型执行不同的操作
	switch operation.Type {
	case "echo_test":
		result, err = t.executeEchoTest(ctx, conn, operation)
	case "send_only":
		result, err = t.executeSendOnly(ctx, conn, operation)
	case "receive_only":
		result, err = t.executeReceiveOnly(ctx, conn, operation)
	case "bidirectional":
		result, err = t.executeBidirectional(ctx, conn, operation)
	default:
		result.Error = fmt.Errorf("unsupported operation type: %s", operation.Type)
	}

	result.Duration = time.Since(startTime)
	
	// 记录指标
	if t.metricsCollector != nil {
		t.metricsCollector.Record(result)
	}

	return result, err
}

// executeEchoTest 执行回显测试
func (t *TCPAdapter) executeEchoTest(ctx context.Context, conn net.Conn, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	result := &interfaces.OperationResult{
		Success:  false,
		IsRead:   true, // 回显测试既读又写，但主要是验证读取
		Metadata: make(map[string]interface{}),
	}

	// 构造测试数据
	testData := t.buildTestData(operation)
	
	// 设置超时
	if err := conn.SetDeadline(time.Now().Add(t.config.Connection.Timeout)); err != nil {
		result.Error = fmt.Errorf("failed to set deadline: %w", err)
		return result, result.Error
	}

	// 发送数据
	_, err := conn.Write(testData)
	if err != nil {
		result.Error = fmt.Errorf("failed to send data: %w", err)
		return result, result.Error
	}

	// 接收响应
	buffer := make([]byte, len(testData)*2) // 留出足够的缓冲区
	n, err := conn.Read(buffer)
	if err != nil {
		result.Error = fmt.Errorf("failed to receive data: %w", err)
		return result, result.Error
	}

	// 验证响应
	receivedData := buffer[:n]
	result.Value = receivedData
	result.Success = true
	result.Metadata["sent_bytes"] = len(testData)
	result.Metadata["received_bytes"] = len(receivedData)
	result.Metadata["data_match"] = string(testData) == string(receivedData)

	return result, nil
}

// executeSendOnly 执行仅发送操作
func (t *TCPAdapter) executeSendOnly(ctx context.Context, conn net.Conn, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	result := &interfaces.OperationResult{
		Success:  false,
		IsRead:   false,
		Metadata: make(map[string]interface{}),
	}

	// 构造测试数据
	testData := t.buildTestData(operation)
	
	// 设置超时
	if err := conn.SetWriteDeadline(time.Now().Add(t.config.Connection.Timeout)); err != nil {
		result.Error = fmt.Errorf("failed to set write deadline: %w", err)
		return result, result.Error
	}

	// 发送数据
	n, err := conn.Write(testData)
	if err != nil {
		result.Error = fmt.Errorf("failed to send data: %w", err)
		return result, result.Error
	}

	result.Success = true
	result.Value = n
	result.Metadata["sent_bytes"] = n
	result.Metadata["expected_bytes"] = len(testData)

	return result, nil
}

// executeReceiveOnly 执行仅接收操作
func (t *TCPAdapter) executeReceiveOnly(ctx context.Context, conn net.Conn, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	result := &interfaces.OperationResult{
		Success:  false,
		IsRead:   true,
		Metadata: make(map[string]interface{}),
	}

	// 设置超时
	if err := conn.SetReadDeadline(time.Now().Add(t.config.Connection.Timeout)); err != nil {
		result.Error = fmt.Errorf("failed to set read deadline: %w", err)
		return result, result.Error
	}

	// 接收数据
	buffer := make([]byte, t.config.BenchMark.DataSize*2)
	n, err := conn.Read(buffer)
	if err != nil {
		result.Error = fmt.Errorf("failed to receive data: %w", err)
		return result, result.Error
	}

	result.Success = true
	result.Value = buffer[:n]
	result.Metadata["received_bytes"] = n

	return result, nil
}

// executeBidirectional 执行双向数据传输
func (t *TCPAdapter) executeBidirectional(ctx context.Context, conn net.Conn, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	result := &interfaces.OperationResult{
		Success:  false,
		IsRead:   true,
		Metadata: make(map[string]interface{}),
	}

	// 构造测试数据
	testData := t.buildTestData(operation)
	
	// 设置超时
	if err := conn.SetDeadline(time.Now().Add(t.config.Connection.Timeout)); err != nil {
		result.Error = fmt.Errorf("failed to set deadline: %w", err)
		return result, result.Error
	}

	// 异步发送和接收
	sendCh := make(chan error, 1)
	receiveCh := make(chan []byte, 1)
	
	// 发送协程
	go func() {
		_, err := conn.Write(testData)
		sendCh <- err
	}()
	
	// 接收协程
	go func() {
		buffer := make([]byte, len(testData)*2)
		n, err := conn.Read(buffer)
		if err != nil {
			receiveCh <- nil
		} else {
			receiveCh <- buffer[:n]
		}
	}()

	// 等待结果
	select {
	case err := <-sendCh:
		if err != nil {
			result.Error = fmt.Errorf("failed to send data: %w", err)
			return result, result.Error
		}
	case <-ctx.Done():
		result.Error = ctx.Err()
		return result, result.Error
	}

	select {
	case receivedData := <-receiveCh:
		if receivedData == nil {
			result.Error = fmt.Errorf("failed to receive data")
			return result, result.Error
		}
		result.Value = receivedData
		result.Success = true
		result.Metadata["sent_bytes"] = len(testData)
		result.Metadata["received_bytes"] = len(receivedData)
	case <-ctx.Done():
		result.Error = ctx.Err()
		return result, result.Error
	}

	return result, nil
}

// buildTestData 构造测试数据
func (t *TCPAdapter) buildTestData(operation interfaces.Operation) []byte {
	// 如果操作中包含自定义数据，使用它
	if operation.Value != nil {
		if data, ok := operation.Value.([]byte); ok {
			return data
		}
		if str, ok := operation.Value.(string); ok {
			return []byte(str)
		}
	}

	// 否则生成默认大小的测试数据
	data := make([]byte, t.config.BenchMark.DataSize)
	for i := range data {
		data[i] = byte(i % 256)
	}
	return data
}

// testConnection 测试连接
func (t *TCPAdapter) testConnection(ctx context.Context) error {
	address := fmt.Sprintf("%s:%d", t.config.Connection.Address, t.config.Connection.Port)
	
	conn, err := net.DialTimeout("tcp", address, t.config.Connection.Timeout)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", address, err)
	}
	defer conn.Close()

	return nil
}

// Close 关闭连接
func (t *TCPAdapter) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.connectionPool != nil {
		if err := t.connectionPool.Close(); err != nil {
			return fmt.Errorf("failed to close connection pool: %w", err)
		}
	}

	t.isConnected = false
	return nil
}

// GetProtocolMetrics 获取协议特定指标
func (t *TCPAdapter) GetProtocolMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})
	
	if t.connectionPool != nil {
		metrics["connection_pool_size"] = t.config.Connection.Pool.PoolSize
		metrics["connection_pool_active"] = t.connectionPool.ActiveConnections()
		metrics["connection_mode"] = t.config.TCPSpecific.ConnectionMode
		metrics["no_delay_enabled"] = t.config.TCPSpecific.NoDelay
		metrics["buffer_size"] = t.config.TCPSpecific.BufferSize
	}
	
	return metrics
}

// HealthCheck 健康检查
func (t *TCPAdapter) HealthCheck(ctx context.Context) error {
	if !t.isConnected {
		return fmt.Errorf("adapter not connected")
	}

	// 测试连接池中的连接
	conn, err := t.connectionPool.GetConnection()
	if err != nil {
		return fmt.Errorf("failed to get connection from pool: %w", err)
	}
	defer t.connectionPool.ReturnConnection(conn)

	// 简单的连接测试
	if err := conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return fmt.Errorf("failed to set deadline for health check: %w", err)
	}

	return nil
}

// GetProtocolName 获取协议名称
func (t *TCPAdapter) GetProtocolName() string {
	return "tcp"
}

// GetMetricsCollector 获取指标收集器
func (t *TCPAdapter) GetMetricsCollector() interfaces.DefaultMetricsCollector {
	return t.metricsCollector
}