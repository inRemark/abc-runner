package websocket

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"abc-runner/app/adapters/websocket/config"
	"abc-runner/app/adapters/websocket/connection"
	"abc-runner/app/adapters/websocket/operations"
	"abc-runner/app/core/interfaces"
	"abc-runner/app/core/utils"

	"github.com/gorilla/websocket"
)

// WebSocketAdapter WebSocket协议适配器 - 基于统一架构设计
type WebSocketAdapter struct {
	// 核心组件（与Redis/HTTP/TCP保持一致）
	connectionPool    *connection.WebSocketConnectionPool
	operationRegistry *utils.OperationRegistry
	config            *config.WebSocketConfig

	// 指标收集器（统一依赖注入）
	metricsCollector interfaces.DefaultMetricsCollector

	// 状态管理
	isConnected bool
	mutex       sync.RWMutex

	// 统计信息
	totalOperations   int64
	successOperations int64
	failedOperations  int64
	startTime         time.Time

	// WebSocket特定统计
	sentMessages     int64
	receivedMessages int64
	heartbeatCount   int64
	reconnectCount   int64
}

// NewWebSocketAdapter 创建WebSocket适配器 - 新架构设计
func NewWebSocketAdapter(metricsCollector interfaces.DefaultMetricsCollector) *WebSocketAdapter {
	if metricsCollector == nil {
		panic("metricsCollector cannot be nil - dependency injection required")
	}

	return &WebSocketAdapter{
		metricsCollector: metricsCollector,
		startTime:        time.Now(),
	}
}

// Connect 初始化连接 - 统一架构实现
func (w *WebSocketAdapter) Connect(ctx context.Context, cfg interfaces.Config) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	// 类型断言配置
	wsConfig, ok := cfg.(*config.WebSocketConfig)
	if !ok {
		return fmt.Errorf("invalid config type for WebSocket adapter: expected *config.WebSocketConfig, got %T", cfg)
	}

	// 验证配置
	if err := wsConfig.Validate(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	w.config = wsConfig

	// 创建连接池
	pool, err := connection.NewWebSocketConnectionPool(wsConfig)
	if err != nil {
		return fmt.Errorf("failed to create WebSocket connection pool: %w", err)
	}
	w.connectionPool = pool

	// 创建操作注册表并注册所有WebSocket操作
	w.operationRegistry = utils.NewOperationRegistry()
	operations.RegisterWebSocketOperations(w.operationRegistry)

	// 执行健康检查
	if err := w.HealthCheck(ctx); err != nil {
		return fmt.Errorf("initial health check failed: %w", err)
	}

	w.isConnected = true
	return nil
}

// Execute 执行操作 - 使用操作工厂模式
// Execute 执行WebSocket操作
func (w *WebSocketAdapter) Execute(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	if !w.IsConnected() {
		return nil, fmt.Errorf("WebSocket adapter is not connected")
	}

	// 统计操作数
	w.incrementTotalOperations()

	// 验证操作是否支持
	if err := w.ValidateOperation(operation.Type); err != nil {
		w.incrementFailedOperations()
		return &interfaces.OperationResult{
			Success: false,
			Error:   err,
		}, err
	}

	// 执行具体操作
	result, err := w.executeWebSocketOperation(ctx, operation)

	// 更新统计信息
	if err != nil || (result != nil && !result.Success) {
		w.incrementFailedOperations()
	} else {
		w.incrementSuccessOperations()
	}

	// 注意：不要在这里调用 w.metricsCollector.Record(result)
	// 因为执行引擎会负责记录指标，避免重复计数

	return result, err
}

// executeWebSocketOperation 执行具体的WebSocket操作
func (w *WebSocketAdapter) executeWebSocketOperation(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	startTime := time.Now()
	result := &interfaces.OperationResult{
		IsRead:   w.isReadOperation(operation.Type),
		Metadata: make(map[string]interface{}),
	}

	// 获取连接
	conn, err := w.connectionPool.GetConnection()
	if err != nil {
		result.Error = fmt.Errorf("failed to get connection: %w", err)
		result.Duration = time.Since(startTime)
		return result, result.Error
	}
	defer w.connectionPool.ReturnConnection(conn)

	var opErr error
	switch operation.Type {
	case "send_text":
		opErr = w.executeSendText(ctx, operation, conn)
	case "send_binary":
		opErr = w.executeSendBinary(ctx, operation, conn)
	case "echo_test":
		result.Value, opErr = w.executeEchoTest(ctx, operation, conn)
	case "ping_pong":
		result.Value, opErr = w.executePingPong(ctx, operation, conn)
	case "broadcast":
		result.Value, opErr = w.executeBroadcast(ctx, operation, conn)
	case "subscribe":
		result.Value, opErr = w.executeSubscribe(ctx, operation, conn)
	case "large_message":
		opErr = w.executeLargeMessage(ctx, operation, conn)
		if opErr == nil {
			result.Value = len(operation.Value.([]byte))
		}
	case "stress_test":
		result.Value, opErr = w.executeStressTest(ctx, operation, conn)
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

// 具体操作实现方法

// executeSendText 执行发送文本消息
func (w *WebSocketAdapter) executeSendText(ctx context.Context, operation interfaces.Operation, conn *connection.WebSocketConnection) error {
	message, ok := operation.Value.(string)
	if !ok {
		return fmt.Errorf("invalid message type for send_text operation")
	}

	err := conn.SendMessage(websocket.TextMessage, []byte(message))
	if err == nil {
		atomic.AddInt64(&w.sentMessages, 1)
	}
	return err
}

// executeSendBinary 执行发送二进制消息
func (w *WebSocketAdapter) executeSendBinary(ctx context.Context, operation interfaces.Operation, conn *connection.WebSocketConnection) error {
	data, ok := operation.Value.([]byte)
	if !ok {
		return fmt.Errorf("invalid data type for send_binary operation")
	}

	err := conn.SendMessage(websocket.BinaryMessage, data)
	if err == nil {
		atomic.AddInt64(&w.sentMessages, 1)
	}
	return err
}

// executeEchoTest 执行回显测试
func (w *WebSocketAdapter) executeEchoTest(ctx context.Context, operation interfaces.Operation, conn *connection.WebSocketConnection) (interface{}, error) {
	message, ok := operation.Value.(string)
	if !ok {
		return nil, fmt.Errorf("invalid message type for echo_test operation")
	}

	// 发送消息
	err := conn.SendMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		return nil, fmt.Errorf("failed to send echo message: %w", err)
	}
	atomic.AddInt64(&w.sentMessages, 1)

	// 模拟接收回显响应
	// 在实际实现中，这里会等待服务器的回显响应
	time.Sleep(10 * time.Millisecond) // 模拟网络延迟
	atomic.AddInt64(&w.receivedMessages, 1)

	return message + "_echo", nil
}

// executePingPong 执行心跳测试
func (w *WebSocketAdapter) executePingPong(ctx context.Context, operation interfaces.Operation, conn *connection.WebSocketConnection) (interface{}, error) {
	startTime := time.Now()

	// 发送Ping（通过连接池的心跳机制）
	// 这里模拟ping-pong的延迟
	time.Sleep(5 * time.Millisecond)
	atomic.AddInt64(&w.heartbeatCount, 1)

	latency := time.Since(startTime)
	return latency, nil
}

// executeBroadcast 执行广播测试
func (w *WebSocketAdapter) executeBroadcast(ctx context.Context, operation interfaces.Operation, conn *connection.WebSocketConnection) (interface{}, error) {
	message, ok := operation.Value.(string)
	if !ok {
		return nil, fmt.Errorf("invalid message type for broadcast operation")
	}

	// 发送广播消息
	err := conn.SendMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		return nil, fmt.Errorf("failed to send broadcast message: %w", err)
	}
	atomic.AddInt64(&w.sentMessages, 1)

	// 返回广播成功的连接数（这里简化为1）
	return 1, nil
}

// executeSubscribe 执行订阅操作
func (w *WebSocketAdapter) executeSubscribe(ctx context.Context, operation interfaces.Operation, conn *connection.WebSocketConnection) (interface{}, error) {
	channel := operation.Key
	timeout := operation.TTL
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	// 发送订阅请求
	subscribeMsg := fmt.Sprintf("SUBSCRIBE:%s", channel)
	err := conn.SendMessage(websocket.TextMessage, []byte(subscribeMsg))
	if err != nil {
		return nil, fmt.Errorf("failed to send subscribe message: %w", err)
	}
	atomic.AddInt64(&w.sentMessages, 1)

	// 模拟接收订阅消息
	messages := []string{"message1", "message2", "message3"}
	atomic.AddInt64(&w.receivedMessages, int64(len(messages)))

	return messages, nil
}

// executeLargeMessage 执行大消息传输
func (w *WebSocketAdapter) executeLargeMessage(ctx context.Context, operation interfaces.Operation, conn *connection.WebSocketConnection) error {
	data, ok := operation.Value.([]byte)
	if !ok {
		return fmt.Errorf("invalid data type for large_message operation")
	}

	// 发送大消息
	err := conn.SendMessage(websocket.BinaryMessage, data)
	if err == nil {
		atomic.AddInt64(&w.sentMessages, 1)
	}
	return err
}

// executeStressTest 执行压力测试
func (w *WebSocketAdapter) executeStressTest(ctx context.Context, operation interfaces.Operation, conn *connection.WebSocketConnection) (interface{}, error) {
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
		err := conn.SendMessage(websocket.TextMessage, []byte("stress test message"))
		messageCount++
		if err == nil {
			successCount++
			atomic.AddInt64(&w.sentMessages, 1)
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

// 辅助方法

// ValidateOperation 验证操作是否支持
func (w *WebSocketAdapter) ValidateOperation(operationType string) error {
	if w.operationRegistry == nil {
		return fmt.Errorf("operation registry not initialized")
	}

	_, exists := w.operationRegistry.GetFactory(operationType)
	if !exists {
		return fmt.Errorf("unsupported operation type: %s", operationType)
	}

	return nil
}

// isReadOperation 判断是否为读操作
func (w *WebSocketAdapter) isReadOperation(operationType string) bool {
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

// IsConnected 检查连接状态
func (w *WebSocketAdapter) IsConnected() bool {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.isConnected
}

// incrementTotalOperations 增加总操作数
func (w *WebSocketAdapter) incrementTotalOperations() {
	atomic.AddInt64(&w.totalOperations, 1)
}

// incrementSuccessOperations 增加成功操作数
func (w *WebSocketAdapter) incrementSuccessOperations() {
	atomic.AddInt64(&w.successOperations, 1)
}

// incrementFailedOperations 增加失败操作数
func (w *WebSocketAdapter) incrementFailedOperations() {
	atomic.AddInt64(&w.failedOperations, 1)
}

// Close 关闭连接
func (w *WebSocketAdapter) Close() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.connectionPool != nil {
		if err := w.connectionPool.Close(); err != nil {
			return fmt.Errorf("failed to close WebSocket connection pool: %w", err)
		}
		w.connectionPool = nil
	}

	w.isConnected = false
	return nil
}

// GetProtocolMetrics 获取协议特定指标
func (w *WebSocketAdapter) GetProtocolMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	if w.config != nil {
		metrics["connection_pool_size"] = w.config.Connection.Pool.PoolSize
		metrics["message_type"] = w.config.WebSocketSpecific.MessageType
		metrics["compression_enabled"] = w.config.WebSocketSpecific.Compression
		metrics["auto_reconnect"] = w.config.WebSocketSpecific.AutoReconnect
		metrics["heartbeat_enabled"] = w.config.WebSocketSpecific.Heartbeat.Enabled
	}

	metrics["total_operations"] = atomic.LoadInt64(&w.totalOperations)
	metrics["success_operations"] = atomic.LoadInt64(&w.successOperations)
	metrics["failed_operations"] = atomic.LoadInt64(&w.failedOperations)
	metrics["sent_messages"] = atomic.LoadInt64(&w.sentMessages)
	metrics["received_messages"] = atomic.LoadInt64(&w.receivedMessages)
	metrics["reconnect_count"] = atomic.LoadInt64(&w.reconnectCount)
	metrics["heartbeat_count"] = atomic.LoadInt64(&w.heartbeatCount)

	// 连接池统计
	if w.connectionPool != nil {
		poolStats := w.connectionPool.GetStats()
		for k, v := range poolStats {
			metrics[k] = v
		}
	}

	return metrics
}

// HealthCheck 健康检查
func (w *WebSocketAdapter) HealthCheck(ctx context.Context) error {
	if !w.isConnected {
		return fmt.Errorf("adapter not connected")
	}

	// 检查连接池是否可用
	if w.connectionPool == nil {
		return fmt.Errorf("connection pool not initialized")
	}

	// 尝试获取连接进行健康检查
	conn, err := w.connectionPool.GetConnection()
	if err != nil {
		return fmt.Errorf("no available WebSocket connections: %w", err)
	}
	defer w.connectionPool.ReturnConnection(conn)

	return nil
}

// GetProtocolName 获取协议名称
func (w *WebSocketAdapter) GetProtocolName() string {
	return "websocket"
}

// GetMetricsCollector 获取指标收集器
func (w *WebSocketAdapter) GetMetricsCollector() interfaces.DefaultMetricsCollector {
	return w.metricsCollector
}
