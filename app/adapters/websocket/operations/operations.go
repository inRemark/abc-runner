package operations

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"abc-runner/app/adapters/websocket/connection"
	"abc-runner/app/core/interfaces"
)

// WebSocketOperations WebSocket操作执行器 - 参考HTTP适配器架构
type WebSocketOperations struct {
	connectionPool *connection.WebSocketConnectionPool
}

// NewWebSocketOperations 创建WebSocket操作执行器
func NewWebSocketOperations(connectionPool *connection.WebSocketConnectionPool) *WebSocketOperations {
	return &WebSocketOperations{
		connectionPool: connectionPool,
	}
}

// ExecuteOperation 执行WebSocket操作 - 统一操作入口
func (ops *WebSocketOperations) ExecuteOperation(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	startTime := time.Now()
	result := &interfaces.OperationResult{
		IsRead:   ops.isReadOperation(operation.Type),
		Metadata: make(map[string]interface{}),
	}

	// 获取连接
	conn, err := ops.connectionPool.GetConnection()
	if err != nil {
		result.Error = fmt.Errorf("failed to get connection: %w", err)
		result.Duration = time.Since(startTime)
		return result, result.Error
	}
	defer ops.connectionPool.ReturnConnection(conn)

	var opErr error
	switch operation.Type {
	case "send_text":
		opErr = ops.executeSendText(ctx, operation, conn)
	case "send_binary":
		opErr = ops.executeSendBinary(ctx, operation, conn)
	case "echo_test":
		result.Value, opErr = ops.executeEchoTest(ctx, operation, conn)
	case "ping_pong":
		result.Value, opErr = ops.executePingPong(ctx, operation, conn)
	case "broadcast":
		result.Value, opErr = ops.executeBroadcast(ctx, operation, conn)
	case "subscribe":
		result.Value, opErr = ops.executeSubscribe(ctx, operation, conn)
	case "large_message":
		opErr = ops.executeLargeMessage(ctx, operation, conn)
		if opErr == nil {
			result.Value = len(operation.Value.([]byte))
		}
	case "stress_test":
		result.Value, opErr = ops.executeStressTest(ctx, operation, conn)
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
	result.Metadata["connection_id"] = conn.ID
	result.Metadata["operation_type"] = operation.Type

	return result, opErr
}

// executeSendText 发送文本消息操作
func (ops *WebSocketOperations) executeSendText(ctx context.Context, operation interfaces.Operation, conn *connection.WebSocketConnection) error {
	message, ok := operation.Value.(string)
	if !ok {
		return fmt.Errorf("invalid message type for send_text operation")
	}

	// WebSocket文本消息类型为1
	return conn.SendMessage(1, []byte(message))
}

// executeSendBinary 发送二进制消息操作
func (ops *WebSocketOperations) executeSendBinary(ctx context.Context, operation interfaces.Operation, conn *connection.WebSocketConnection) error {
	data, ok := operation.Value.([]byte)
	if !ok {
		return fmt.Errorf("invalid data type for send_binary operation")
	}

	// WebSocket二进制消息类型为2
	return conn.SendMessage(2, data)
}

// executeEchoTest 回显测试操作
func (ops *WebSocketOperations) executeEchoTest(ctx context.Context, operation interfaces.Operation, conn *connection.WebSocketConnection) (interface{}, error) {
	message, ok := operation.Value.(string)
	if !ok {
		return nil, fmt.Errorf("invalid message type for echo_test operation")
	}

	// 发送消息
	err := conn.SendMessage(1, []byte(message))
	if err != nil {
		return nil, fmt.Errorf("failed to send echo message: %w", err)
	}

	// 模拟接收回显响应
	// 在实际实现中，这里会等待服务器的回显响应
	time.Sleep(10 * time.Millisecond) // 模拟网络延迟

	return message + "_echo", nil
}

// executePingPong 心跳测试操作
func (ops *WebSocketOperations) executePingPong(ctx context.Context, operation interfaces.Operation, conn *connection.WebSocketConnection) (interface{}, error) {
	startTime := time.Now()

	// 发送Ping消息（WebSocket Ping消息类型为9）
	err := conn.SendMessage(9, []byte("ping"))
	if err != nil {
		return nil, fmt.Errorf("failed to send ping: %w", err)
	}

	// 模拟等待Pong响应
	time.Sleep(5 * time.Millisecond) // 模拟ping-pong延迟

	latency := time.Since(startTime)
	return latency, nil
}

// executeBroadcast 广播测试操作
func (ops *WebSocketOperations) executeBroadcast(ctx context.Context, operation interfaces.Operation, conn *connection.WebSocketConnection) (interface{}, error) {
	message, ok := operation.Value.(string)
	if !ok {
		return nil, fmt.Errorf("invalid message type for broadcast operation")
	}

	// 发送广播消息
	err := conn.SendMessage(1, []byte(message))
	if err != nil {
		return nil, fmt.Errorf("failed to send broadcast message: %w", err)
	}

	// 返回广播成功的连接数（这里简化为1）
	return 1, nil
}

// executeSubscribe 订阅操作
func (ops *WebSocketOperations) executeSubscribe(ctx context.Context, operation interfaces.Operation, conn *connection.WebSocketConnection) (interface{}, error) {
	channel := operation.Key
	timeout := operation.TTL
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	// 发送订阅请求
	subscribeMsg := fmt.Sprintf("SUBSCRIBE:%s", channel)
	err := conn.SendMessage(1, []byte(subscribeMsg))
	if err != nil {
		return nil, fmt.Errorf("failed to send subscribe message: %w", err)
	}

	// 模拟接收订阅消息
	messages := []string{"message1", "message2", "message3"}
	return messages, nil
}

// executeLargeMessage 大消息传输操作
func (ops *WebSocketOperations) executeLargeMessage(ctx context.Context, operation interfaces.Operation, conn *connection.WebSocketConnection) error {
	data, ok := operation.Value.([]byte)
	if !ok {
		return fmt.Errorf("invalid data type for large_message operation")
	}

	// 发送大消息（二进制类型）
	return conn.SendMessage(2, data)
}

// executeStressTest 压力测试操作
func (ops *WebSocketOperations) executeStressTest(ctx context.Context, operation interfaces.Operation, conn *connection.WebSocketConnection) (interface{}, error) {
	// 从元数据中获取压力测试参数
	connections := 1
	frequency := 10
	duration := 10 * time.Second

	if connStr, ok := operation.Metadata["connections"]; ok {
		if c, err := fmt.Sscanf(connStr, "%d", &connections); err != nil || c != 1 {
			connections = 1
		}
	}

	if freqStr, ok := operation.Metadata["frequency"]; ok {
		if f, err := fmt.Sscanf(freqStr, "%d", &frequency); err != nil || f != 1 {
			frequency = 10
		}
	}

	if durStr, ok := operation.Metadata["duration"]; ok {
		if d, err := time.ParseDuration(durStr); err == nil {
			duration = d
		}
	}

	// 执行压力测试
	startTime := time.Now()
	messageCount := 0
	successCount := 0

	for time.Since(startTime) < duration {
		err := conn.SendMessage(1, []byte("stress test message"))
		messageCount++
		if err == nil {
			successCount++
		}

		// 控制发送频率
		time.Sleep(time.Second / time.Duration(frequency))
	}

	stats := map[string]interface{}{
		"total_messages": messageCount,
		"success_count":  successCount,
		"success_rate":   float64(successCount) / float64(messageCount),
		"duration":       time.Since(startTime).String(),
	}

	return stats, nil
}

// isReadOperation 判断是否为读操作
func (ops *WebSocketOperations) isReadOperation(operationType string) bool {
	readOperations := map[string]bool{
		"echo_test":     true,
		"ping_pong":     true,
		"subscribe":     true,
		"send_text":     false,
		"send_binary":   false,
		"broadcast":     false,
		"large_message": false,
		"stress_test":   false,
	}

	return readOperations[operationType]
}

// GetSupportedOperations 获取支持的操作类型
func (ops *WebSocketOperations) GetSupportedOperations() []string {
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

// 统计辅助函数

var (
	totalMessagesSent     int64
	totalMessagesReceived int64
	totalHeartbeats       int64
)

// IncrementMessagesSent 增加发送消息计数
func IncrementMessagesSent() {
	atomic.AddInt64(&totalMessagesSent, 1)
}

// IncrementMessagesReceived 增加接收消息计数
func IncrementMessagesReceived() {
	atomic.AddInt64(&totalMessagesReceived, 1)
}

// IncrementHeartbeats 增加心跳计数
func IncrementHeartbeats() {
	atomic.AddInt64(&totalHeartbeats, 1)
}

// GetOperationStats 获取操作统计信息
func GetOperationStats() map[string]interface{} {
	return map[string]interface{}{
		"total_messages_sent":     atomic.LoadInt64(&totalMessagesSent),
		"total_messages_received": atomic.LoadInt64(&totalMessagesReceived),
		"total_heartbeats":        atomic.LoadInt64(&totalHeartbeats),
	}
}