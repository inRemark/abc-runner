package tcp

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"abc-runner/app/adapters/tcp/config"
	"abc-runner/app/adapters/tcp/connection"
	"abc-runner/app/adapters/tcp/operations"

	"abc-runner/app/core/interfaces"
)

// TCPAdapter TCP协议适配器 - 遵循统一架构模式
type TCPAdapter struct {
	config           *config.TCPConfig
	connectionPool   *connection.ConnectionPool
	tcpOperations    *operations.TCPOperations
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
	pool, err := connection.NewConnectionPool(tcpConfig)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}

	t.connectionPool = pool

	// 创建TCP操作执行器
	t.tcpOperations = operations.NewTCPOperations(pool, tcpConfig, t.metricsCollector)

	// 测试连接
	if err := t.testConnection(ctx); err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}

	t.isConnected = true
	return nil
}

// Execute 执行TCP操作 - 使用TCPOperations执行器
func (t *TCPAdapter) Execute(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	// 检查连接状态
	if !t.isConnected {
		return &interfaces.OperationResult{
			Success:  false,
			Duration: 0,
			Error:    fmt.Errorf("adapter not connected"),
		}, fmt.Errorf("adapter not connected")
	}

	// 使用TCPOperations执行器执行操作
	return t.tcpOperations.ExecuteOperation(ctx, operation)
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
	if len(testData) == 0 {
		testData = []byte(fmt.Sprintf("ECHO_TEST_%d_%d", time.Now().Unix(), len(testData)))
	}

	// 设置超时
	if err := conn.SetDeadline(time.Now().Add(t.config.Connection.Timeout)); err != nil {
		result.Error = fmt.Errorf("failed to set deadline: %w", err)
		return result, result.Error
	}

	// 发送数据
	sentBytes, err := conn.Write(testData)
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

	// 记录详细指标
	result.Metadata["sent_bytes"] = sentBytes
	result.Metadata["received_bytes"] = n
	result.Metadata["expected_bytes"] = len(testData)
	result.Metadata["data_match"] = string(testData) == string(receivedData)
	result.Metadata["echo_ratio"] = float64(n) / float64(sentBytes)

	// 检查数据完整性
	if n != len(testData) {
		result.Metadata["warning"] = "received data length mismatch"
	}

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
	if len(testData) == 0 {
		testData = []byte(fmt.Sprintf("SEND_ONLY_%d_%d", time.Now().Unix(), operation.Params["job_id"]))
	}

	// 设置写超时
	if err := conn.SetWriteDeadline(time.Now().Add(t.config.Connection.Timeout)); err != nil {
		result.Error = fmt.Errorf("failed to set write deadline: %w", err)
		return result, result.Error
	}

	// 发送数据（分块发送支持大数据包）
	totalSent := 0
	bufferSize := t.config.TCPSpecific.BufferSize
	if bufferSize <= 0 {
		bufferSize = 4096
	}

	for totalSent < len(testData) {
		// 检查上下文取消
		select {
		case <-ctx.Done():
			result.Error = ctx.Err()
			return result, result.Error
		default:
		}

		// 计算本次发送的数据大小
		remaining := len(testData) - totalSent
		chunkSize := remaining
		if chunkSize > bufferSize {
			chunkSize = bufferSize
		}

		// 发送数据块
		n, err := conn.Write(testData[totalSent : totalSent+chunkSize])
		if err != nil {
			result.Error = fmt.Errorf("failed to send data chunk (sent %d/%d bytes): %w", totalSent, len(testData), err)
			return result, result.Error
		}
		totalSent += n

		// 记录进度
		if totalSent%bufferSize == 0 || totalSent == len(testData) {
			result.Metadata["progress"] = fmt.Sprintf("%d/%d bytes", totalSent, len(testData))
		}
	}

	result.Success = true
	result.Value = totalSent
	result.Metadata["sent_bytes"] = totalSent
	result.Metadata["expected_bytes"] = len(testData)
	result.Metadata["send_complete"] = totalSent == len(testData)
	result.Metadata["throughput_bps"] = float64(totalSent) / result.Duration.Seconds()

	return result, nil
}

// executeReceiveOnly 执行仅接收操作
func (t *TCPAdapter) executeReceiveOnly(ctx context.Context, conn net.Conn, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	result := &interfaces.OperationResult{
		Success:  false,
		IsRead:   true,
		Metadata: make(map[string]interface{}),
	}

	// 设置读超时
	if err := conn.SetReadDeadline(time.Now().Add(t.config.Connection.Timeout)); err != nil {
		result.Error = fmt.Errorf("failed to set read deadline: %w", err)
		return result, result.Error
	}

	// 准备接收缓冲区
	bufferSize := t.config.BenchMark.DataSize * 2
	if bufferSize < 4096 {
		bufferSize = 4096
	}
	buffer := make([]byte, bufferSize)

	// 接收数据（支持多次读取）
	totalReceived := 0
	allData := make([]byte, 0, bufferSize)
	maxReadAttempts := 10 // 防止无限等待

	for i := 0; i < maxReadAttempts; i++ {
		// 检查上下文取消
		select {
		case <-ctx.Done():
			result.Error = ctx.Err()
			return result, result.Error
		default:
		}

		// 读取数据
		n, err := conn.Read(buffer)
		if err != nil {
			// 如果已经读取了一些数据，这可能不是错误
			if totalReceived > 0 && (err.Error() == "EOF" || err.Error() == "connection reset by peer") {
				break
			}
			result.Error = fmt.Errorf("failed to receive data (attempt %d, received %d bytes): %w", i+1, totalReceived, err)
			return result, result.Error
		}

		if n > 0 {
			allData = append(allData, buffer[:n]...)
			totalReceived += n
			result.Metadata[fmt.Sprintf("read_%d_bytes", i+1)] = n
		}

		// 如果读取的数据小于缓冲区，可能已经读完
		if n < len(buffer) {
			break
		}
	}

	result.Success = totalReceived > 0
	result.Value = allData
	result.Metadata["received_bytes"] = totalReceived
	result.Metadata["read_attempts"] = maxReadAttempts
	result.Metadata["buffer_size"] = bufferSize

	// 计算接收速率
	if result.Duration > 0 {
		result.Metadata["receive_rate_bps"] = float64(totalReceived) / result.Duration.Seconds()
	}

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
	if len(testData) == 0 {
		testData = []byte(fmt.Sprintf("BIDI_TEST_%d_%d", time.Now().Unix(), operation.Params["job_id"]))
	}

	// 设置超时
	if err := conn.SetDeadline(time.Now().Add(t.config.Connection.Timeout)); err != nil {
		result.Error = fmt.Errorf("failed to set deadline: %w", err)
		return result, result.Error
	}

	// 使用带缓冲的通道进行异步发送和接收
	sendResult := make(chan error, 1)
	receiveResult := make(chan []byte, 1)
	receiveError := make(chan error, 1)

	// 发送协程
	go func() {
		defer func() {
			if r := recover(); r != nil {
				sendResult <- fmt.Errorf("send goroutine panic: %v", r)
			}
		}()

		totalSent := 0
		bufferSize := t.config.TCPSpecific.BufferSize
		if bufferSize <= 0 {
			bufferSize = 4096
		}

		// 分块发送
		for totalSent < len(testData) {
			remaining := len(testData) - totalSent
			chunkSize := remaining
			if chunkSize > bufferSize {
				chunkSize = bufferSize
			}

			n, err := conn.Write(testData[totalSent : totalSent+chunkSize])
			if err != nil {
				sendResult <- fmt.Errorf("failed to send data chunk: %w", err)
				return
			}
			totalSent += n

			// 等待一小段时间以允许接收操作
			time.Sleep(time.Millisecond)
		}
		sendResult <- nil
	}()

	// 接收协程
	go func() {
		defer func() {
			if r := recover(); r != nil {
				receiveError <- fmt.Errorf("receive goroutine panic: %v", r)
			}
		}()

		buffer := make([]byte, len(testData)*3) // 留出额外缓冲区
		n, err := conn.Read(buffer)
		if err != nil {
			receiveError <- fmt.Errorf("failed to receive data: %w", err)
			return
		}
		receiveResult <- buffer[:n]
	}()

	// 等待发送结果
	var sendErr error
	select {
	case sendErr = <-sendResult:
		if sendErr != nil {
			result.Error = fmt.Errorf("send operation failed: %w", sendErr)
			return result, result.Error
		}
	case <-ctx.Done():
		result.Error = fmt.Errorf("send operation cancelled: %w", ctx.Err())
		return result, result.Error
	case <-time.After(t.config.Connection.Timeout):
		result.Error = fmt.Errorf("send operation timeout")
		return result, result.Error
	}

	// 等待接收结果
	var receivedData []byte
	select {
	case receivedData = <-receiveResult:
		// 成功接收
	case err := <-receiveError:
		result.Error = fmt.Errorf("receive operation failed: %w", err)
		return result, result.Error
	case <-ctx.Done():
		result.Error = fmt.Errorf("receive operation cancelled: %w", ctx.Err())
		return result, result.Error
	case <-time.After(t.config.Connection.Timeout):
		result.Error = fmt.Errorf("receive operation timeout")
		return result, result.Error
	}

	result.Value = receivedData
	result.Success = true

	// 记录详细指标
	result.Metadata["sent_bytes"] = len(testData)
	result.Metadata["received_bytes"] = len(receivedData)
	result.Metadata["data_match"] = string(testData) == string(receivedData)
	result.Metadata["bidirectional_success"] = true
	result.Metadata["echo_ratio"] = float64(len(receivedData)) / float64(len(testData))

	// 计算双向吞吐量
	if result.Duration > 0 {
		totalBytes := len(testData) + len(receivedData)
		result.Metadata["bidirectional_throughput_bps"] = float64(totalBytes) / result.Duration.Seconds()
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
		// 连接池指标
		metrics["connection_pool_size"] = t.config.Connection.Pool.PoolSize
		metrics["connection_pool_active"] = t.connectionPool.ActiveConnections()
		metrics["connection_pool_available"] = t.connectionPool.AvailableConnections()

		// TCP特定配置指标
		metrics["connection_mode"] = t.config.TCPSpecific.ConnectionMode
		metrics["no_delay_enabled"] = t.config.TCPSpecific.NoDelay
		metrics["buffer_size"] = t.config.TCPSpecific.BufferSize
		metrics["linger_timeout"] = t.config.TCPSpecific.LingerTimeout
		metrics["reuse_address"] = t.config.TCPSpecific.ReuseAddress

		// 连接配置指标
		metrics["keep_alive"] = t.config.Connection.KeepAlive
		metrics["keep_alive_period"] = t.config.Connection.KeepAlivePeriod.String()
		metrics["connection_timeout"] = t.config.Connection.Timeout.String()

		// 连接池详细信息
		if poolStats := t.connectionPool.Stats(); poolStats != nil {
			for key, value := range poolStats {
				metrics["pool_"+key] = value
			}
		}
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
