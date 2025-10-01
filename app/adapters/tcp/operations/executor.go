package operations

import (
	"context"
	"fmt"
	"net"
	"time"

	"abc-runner/app/adapters/tcp/config"
	"abc-runner/app/adapters/tcp/connection"
	"abc-runner/app/core/interfaces"
)

// TCPExecutor TCP操作执行器 - 遵循统一架构模式
type TCPExecutor struct {
	connectionPool   *connection.ConnectionPool
	config           *config.TCPConfig
	metricsCollector interfaces.DefaultMetricsCollector
}

// NewTCPExecutor 创建TCP操作执行器
func NewTCPExecutor(
	connectionPool *connection.ConnectionPool,
	config *config.TCPConfig,
	metricsCollector interfaces.DefaultMetricsCollector,
) *TCPExecutor {
	return &TCPExecutor{
		connectionPool:   connectionPool,
		config:           config,
		metricsCollector: metricsCollector,
	}
}

// ExecuteOperation 执行TCP操作 - 统一操作入口
func (t *TCPExecutor) ExecuteOperation(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	startTime := time.Now()
	result := &interfaces.OperationResult{
		IsRead:   t.isReadOperation(operation.Type),
		Metadata: make(map[string]interface{}),
	}

	// 获取连接
	conn, err := t.connectionPool.GetConnection()
	if err != nil {
		result.Error = fmt.Errorf("failed to get connection: %w", err)
		result.Duration = time.Since(startTime)
		return result, result.Error
	}
	defer t.connectionPool.ReturnConnection(conn)

	var opErr error
	switch operation.Type {
	case "echo_test":
		opErr = t.executeEchoTest(ctx, conn, operation, result)
	case "send_only":
		opErr = t.executeSendOnly(ctx, conn, operation, result)
	case "receive_only":
		opErr = t.executeReceiveOnly(ctx, conn, operation, result)
	case "bidirectional":
		opErr = t.executeBidirectional(ctx, conn, operation, result)
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
	result.Metadata["protocol"] = "tcp"
	result.Metadata["operation_type"] = operation.Type
	result.Metadata["connection_mode"] = t.config.TCPSpecific.ConnectionMode
	result.Metadata["no_delay"] = t.config.TCPSpecific.NoDelay
	result.Metadata["execution_time_ms"] = float64(result.Duration.Nanoseconds()) / 1e6

	return result, opErr
}

// executeEchoTest 执行回显测试
func (t *TCPExecutor) executeEchoTest(ctx context.Context, conn net.Conn, operation interfaces.Operation, result *interfaces.OperationResult) error {
	// 构造测试数据
	testData := t.buildTestData(operation)
	if len(testData) == 0 {
		testData = []byte(fmt.Sprintf("ECHO_TEST_%d_%d", time.Now().Unix(), len(testData)))
	}

	// 设置超时
	if err := conn.SetDeadline(time.Now().Add(t.config.Connection.Timeout)); err != nil {
		return fmt.Errorf("failed to set deadline: %w", err)
	}

	// 发送数据
	sentBytes, err := conn.Write(testData)
	if err != nil {
		return fmt.Errorf("failed to send data: %w", err)
	}

	// 接收响应
	buffer := make([]byte, len(testData)*2) // 留出足够的缓冲区
	n, err := conn.Read(buffer)
	if err != nil {
		return fmt.Errorf("failed to receive data: %w", err)
	}

	// 验证响应
	receivedData := buffer[:n]
	result.Value = receivedData

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

	return nil
}

// executeSendOnly 执行仅发送操作
func (t *TCPExecutor) executeSendOnly(ctx context.Context, conn net.Conn, operation interfaces.Operation, result *interfaces.OperationResult) error {
	// 构造测试数据
	testData := t.buildTestData(operation)
	if len(testData) == 0 {
		testData = []byte(fmt.Sprintf("SEND_ONLY_%d", time.Now().Unix()))
	}

	// 设置超时
	if err := conn.SetDeadline(time.Now().Add(t.config.Connection.Timeout)); err != nil {
		return fmt.Errorf("failed to set deadline: %w", err)
	}

	// 发送数据
	sentBytes, err := conn.Write(testData)
	if err != nil {
		return fmt.Errorf("failed to send data: %w", err)
	}

	result.Value = sentBytes
	result.Metadata["sent_bytes"] = sentBytes
	result.Metadata["data_size"] = len(testData)

	return nil
}

// executeReceiveOnly 执行仅接收操作
func (t *TCPExecutor) executeReceiveOnly(ctx context.Context, conn net.Conn, operation interfaces.Operation, result *interfaces.OperationResult) error {
	// 设置超时
	if err := conn.SetDeadline(time.Now().Add(t.config.Connection.Timeout)); err != nil {
		return fmt.Errorf("failed to set deadline: %w", err)
	}

	// 准备接收缓冲区
	bufferSize := 4096
	if size, ok := operation.Params["buffer_size"].(int); ok && size > 0 {
		bufferSize = size
	}

	buffer := make([]byte, bufferSize)
	n, err := conn.Read(buffer)
	if err != nil {
		return fmt.Errorf("failed to receive data: %w", err)
	}

	receivedData := buffer[:n]
	result.Value = receivedData
	result.Metadata["received_bytes"] = n
	result.Metadata["buffer_size"] = bufferSize

	return nil
}

// executeBidirectional 执行双向通信操作
func (t *TCPExecutor) executeBidirectional(ctx context.Context, conn net.Conn, operation interfaces.Operation, result *interfaces.OperationResult) error {
	// 构造测试数据
	testData := t.buildTestData(operation)
	if len(testData) == 0 {
		testData = []byte(fmt.Sprintf("BIDIRECTIONAL_%d", time.Now().Unix()))
	}

	// 设置超时
	if err := conn.SetDeadline(time.Now().Add(t.config.Connection.Timeout)); err != nil {
		return fmt.Errorf("failed to set deadline: %w", err)
	}

	// 发送数据
	sentBytes, err := conn.Write(testData)
	if err != nil {
		return fmt.Errorf("failed to send data: %w", err)
	}

	// 接收响应
	buffer := make([]byte, 4096)
	n, err := conn.Read(buffer)
	if err != nil {
		return fmt.Errorf("failed to receive data: %w", err)
	}

	receivedData := buffer[:n]
	result.Value = map[string]interface{}{
		"sent_data":     testData,
		"received_data": receivedData,
	}

	result.Metadata["sent_bytes"] = sentBytes
	result.Metadata["received_bytes"] = n
	result.Metadata["total_bytes"] = sentBytes + n

	return nil
}

// buildTestData 构造测试数据
func (t *TCPExecutor) buildTestData(operation interfaces.Operation) []byte {
	// 尝试从操作中获取数据
	if operation.Value != nil {
		switch v := operation.Value.(type) {
		case []byte:
			return v
		case string:
			return []byte(v)
		}
	}

	// 如果没有指定数据，根据配置生成
	if t.config != nil && t.config.BenchMark.DataSize > 0 {
		data := make([]byte, t.config.BenchMark.DataSize)
		for i := range data {
			data[i] = byte('A' + (i % 26))
		}
		return data
	}

	// 默认数据
	return []byte("TCP_TEST_DATA")
}

// isReadOperation 判断是否为读操作
func (t *TCPExecutor) isReadOperation(operationType string) bool {
	readOperations := map[string]bool{
		"echo_test":     true,  // 回显测试既读又写，但主要是验证读取
		"receive_only":  true,  // 仅接收
		"bidirectional": true,  // 双向通信包含读取
		"send_only":     false, // 仅发送
	}

	return readOperations[operationType]
}

// GetSupportedOperations 获取支持的操作类型
func (t *TCPExecutor) GetSupportedOperations() []string {
	return []string{
		"echo_test",
		"send_only",
		"receive_only",
		"bidirectional",
	}
}
