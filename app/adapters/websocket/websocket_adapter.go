package websocket

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"abc-runner/app/adapters/websocket/config"
	"abc-runner/app/core/interfaces"
)

// WebSocketAdapter WebSocket协议适配器
type WebSocketAdapter struct {
	config           *config.WebSocketConfig
	connections      map[string]*WebSocketConnection
	metricsCollector interfaces.DefaultMetricsCollector
	mu               sync.RWMutex
	isConnected      bool
	
	// 统计信息
	sentMessages     int64
	receivedMessages int64
	reconnectCount   int64
	heartbeatCount   int64
}

// WebSocketConnection WebSocket连接封装
type WebSocketConnection struct {
	id              string
	url             string
	conn            interface{} // 模拟WebSocket连接，实际应使用真实WebSocket库
	isConnected     bool
	lastPingTime    time.Time
	lastPongTime    time.Time
	reconnectCount  int
	mu              sync.RWMutex
}

// NewWebSocketAdapter 创建WebSocket适配器
func NewWebSocketAdapter(metricsCollector interfaces.DefaultMetricsCollector) *WebSocketAdapter {
	return &WebSocketAdapter{
		metricsCollector: metricsCollector,
		connections:      make(map[string]*WebSocketConnection),
		isConnected:      false,
	}
}

// Connect 初始化连接
func (w *WebSocketAdapter) Connect(ctx context.Context, cfg interfaces.Config) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// 类型断言配置
	wsConfig, ok := cfg.(*config.WebSocketConfig)
	if !ok {
		return fmt.Errorf("invalid config type for WebSocket adapter")
	}

	w.config = wsConfig

	// 验证配置
	if err := wsConfig.Validate(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	// 创建连接池
	poolSize := wsConfig.Connection.Pool.PoolSize
	for i := 0; i < poolSize; i++ {
		connID := fmt.Sprintf("ws_conn_%d", i)
		conn, err := w.createConnection(connID, wsConfig.Connection.URL)
		if err != nil {
			// 如果无法创建连接，记录错误但继续
			fmt.Printf("Warning: failed to create WebSocket connection %s: %v\n", connID, err)
			continue
		}
		w.connections[connID] = conn
	}

	w.isConnected = true
	return nil
}

// createConnection 创建WebSocket连接
func (w *WebSocketAdapter) createConnection(id, url string) (*WebSocketConnection, error) {
	// 这里应该使用真实的WebSocket库，如gorilla/websocket
	// 目前使用模拟实现
	conn := &WebSocketConnection{
		id:          id,
		url:         url,
		conn:        nil, // 模拟连接对象
		isConnected: false,
		lastPingTime: time.Now(),
		lastPongTime: time.Now(),
	}

	// 模拟连接过程
	if err := w.simulateConnect(conn); err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", url, err)
	}

	conn.isConnected = true
	return conn, nil
}

// simulateConnect 模拟连接过程
func (w *WebSocketAdapter) simulateConnect(conn *WebSocketConnection) error {
	// 模拟WebSocket握手
	time.Sleep(time.Millisecond * 10) // 模拟连接延迟
	
	// 在实际实现中，这里会：
	// 1. 建立TCP连接
	// 2. 发送WebSocket握手请求
	// 3. 验证握手响应
	// 4. 配置连接选项（压缩、扩展等）
	
	return nil
}

// Execute 执行操作
func (w *WebSocketAdapter) Execute(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
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
	if !w.isConnected {
		result.Error = fmt.Errorf("adapter not connected")
		result.Duration = time.Since(startTime)
		return result, result.Error
	}

	// 根据操作类型执行不同的操作
	switch operation.Type {
	case "message_exchange":
		result, err := w.executeMessageExchange(ctx, operation)
		result.Duration = time.Since(startTime)
		if w.metricsCollector != nil {
			w.metricsCollector.Record(result)
		}
		return result, err
	case "ping_pong":
		result, err := w.executePingPong(ctx, operation)
		result.Duration = time.Since(startTime)
		if w.metricsCollector != nil {
			w.metricsCollector.Record(result)
		}
		return result, err
	case "broadcast":
		result, err := w.executeBroadcast(ctx, operation)
		result.Duration = time.Since(startTime)
		if w.metricsCollector != nil {
			w.metricsCollector.Record(result)
		}
		return result, err
	case "large_message":
		result, err := w.executeLargeMessage(ctx, operation)
		result.Duration = time.Since(startTime)
		if w.metricsCollector != nil {
			w.metricsCollector.Record(result)
		}
		return result, err
	default:
		result.Error = fmt.Errorf("unsupported operation type: %s", operation.Type)
		result.Duration = time.Since(startTime)
		return result, result.Error
	}
}

// executeMessageExchange 执行消息交换
func (w *WebSocketAdapter) executeMessageExchange(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	result := &interfaces.OperationResult{
		Success:  false,
		IsRead:   true, // 消息交换既发送又接收
		Metadata: make(map[string]interface{}),
	}

	// 获取可用连接
	conn := w.getAvailableConnection()
	if conn == nil {
		result.Error = fmt.Errorf("no available connections")
		return result, result.Error
	}

	// 构造消息数据
	messageData := w.buildMessageData(operation)
	
	// 发送消息
	err := w.sendMessage(conn, messageData)
	if err != nil {
		result.Error = fmt.Errorf("failed to send message: %w", err)
		return result, result.Error
	}

	atomic.AddInt64(&w.sentMessages, 1)

	// 接收响应（模拟）
	responseData, err := w.receiveMessage(conn)
	if err != nil {
		result.Error = fmt.Errorf("failed to receive message: %w", err)
		return result, result.Error
	}

	atomic.AddInt64(&w.receivedMessages, 1)

	result.Success = true
	result.Value = responseData
	result.Metadata["sent_bytes"] = len(messageData)
	result.Metadata["received_bytes"] = len(responseData)
	result.Metadata["message_type"] = w.config.WebSocketSpecific.MessageType
	result.Metadata["connection_id"] = conn.id

	return result, nil
}

// executePingPong 执行心跳检测
func (w *WebSocketAdapter) executePingPong(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	result := &interfaces.OperationResult{
		Success:  false,
		IsRead:   true,
		Metadata: make(map[string]interface{}),
	}

	// 获取可用连接
	conn := w.getAvailableConnection()
	if conn == nil {
		result.Error = fmt.Errorf("no available connections")
		return result, result.Error
	}

	// 发送Ping
	pingData := []byte("ping")
	err := w.sendPing(conn, pingData)
	if err != nil {
		result.Error = fmt.Errorf("failed to send ping: %w", err)
		return result, result.Error
	}

	// 等待Pong
	pongData, err := w.receivePong(conn)
	if err != nil {
		result.Error = fmt.Errorf("failed to receive pong: %w", err)
		return result, result.Error
	}

	atomic.AddInt64(&w.heartbeatCount, 1)

	result.Success = true
	result.Value = pongData
	result.Metadata["ping_data"] = string(pingData)
	result.Metadata["pong_data"] = string(pongData)
	result.Metadata["connection_id"] = conn.id

	return result, nil
}

// executeBroadcast 执行广播测试
func (w *WebSocketAdapter) executeBroadcast(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	result := &interfaces.OperationResult{
		Success:  false,
		IsRead:   false,
		Metadata: make(map[string]interface{}),
	}

	// 构造广播数据
	broadcastData := w.buildMessageData(operation)
	
	// 向所有连接广播
	successCount := 0
	totalConnections := len(w.connections)
	
	for _, conn := range w.connections {
		if conn.isConnected {
			err := w.sendMessage(conn, broadcastData)
			if err == nil {
				successCount++
				atomic.AddInt64(&w.sentMessages, 1)
			}
		}
	}

	if successCount == 0 {
		result.Error = fmt.Errorf("failed to broadcast to any connections")
		return result, result.Error
	}

	result.Success = true
	result.Value = successCount
	result.Metadata["broadcast_size"] = len(broadcastData)
	result.Metadata["total_connections"] = totalConnections
	result.Metadata["successful_broadcasts"] = successCount

	return result, nil
}

// executeLargeMessage 执行大消息传输
func (w *WebSocketAdapter) executeLargeMessage(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	result := &interfaces.OperationResult{
		Success:  false,
		IsRead:   false,
		Metadata: make(map[string]interface{}),
	}

	// 获取可用连接
	conn := w.getAvailableConnection()
	if conn == nil {
		result.Error = fmt.Errorf("no available connections")
		return result, result.Error
	}

	// 构造大消息数据（比普通消息大10倍）
	largeSize := w.config.BenchMark.DataSize * 10
	largeData := make([]byte, largeSize)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	// 发送大消息
	err := w.sendMessage(conn, largeData)
	if err != nil {
		result.Error = fmt.Errorf("failed to send large message: %w", err)
		return result, result.Error
	}

	atomic.AddInt64(&w.sentMessages, 1)

	result.Success = true
	result.Value = len(largeData)
	result.Metadata["message_size"] = len(largeData)
	result.Metadata["connection_id"] = conn.id
	result.Metadata["message_type"] = "large"

	return result, nil
}

// 辅助方法

// getAvailableConnection 获取可用连接
func (w *WebSocketAdapter) getAvailableConnection() *WebSocketConnection {
	w.mu.RLock()
	defer w.mu.RUnlock()
	
	for _, conn := range w.connections {
		if conn.isConnected {
			return conn
		}
	}
	return nil
}

// sendMessage 发送消息（模拟）
func (w *WebSocketAdapter) sendMessage(conn *WebSocketConnection, data []byte) error {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	
	if !conn.isConnected {
		return fmt.Errorf("connection %s is not connected", conn.id)
	}
	
	// 模拟发送延迟
	time.Sleep(time.Microsecond * time.Duration(len(data)/100))
	
	return nil
}

// receiveMessage 接收消息（模拟）
func (w *WebSocketAdapter) receiveMessage(conn *WebSocketConnection) ([]byte, error) {
	conn.mu.RLock()
	defer conn.mu.RUnlock()
	
	if !conn.isConnected {
		return nil, fmt.Errorf("connection %s is not connected", conn.id)
	}
	
	// 模拟接收延迟和数据
	time.Sleep(time.Microsecond * 100)
	return []byte("echo response"), nil
}

// sendPing 发送Ping（模拟）
func (w *WebSocketAdapter) sendPing(conn *WebSocketConnection, data []byte) error {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	
	conn.lastPingTime = time.Now()
	return nil
}

// receivePong 接收Pong（模拟）
func (w *WebSocketAdapter) receivePong(conn *WebSocketConnection) ([]byte, error) {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	
	conn.lastPongTime = time.Now()
	return []byte("pong"), nil
}

// buildMessageData 构造消息数据
func (w *WebSocketAdapter) buildMessageData(operation interfaces.Operation) []byte {
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
	data := make([]byte, w.config.BenchMark.DataSize)
	for i := range data {
		data[i] = byte(i % 256)
	}
	return data
}

// Close 关闭连接
func (w *WebSocketAdapter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	for id, conn := range w.connections {
		if conn.isConnected {
			conn.mu.Lock()
			conn.isConnected = false
			// 这里应该关闭真实的WebSocket连接
			conn.mu.Unlock()
		}
		delete(w.connections, id)
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
	
	metrics["sent_messages"] = atomic.LoadInt64(&w.sentMessages)
	metrics["received_messages"] = atomic.LoadInt64(&w.receivedMessages)
	metrics["reconnect_count"] = atomic.LoadInt64(&w.reconnectCount)
	metrics["heartbeat_count"] = atomic.LoadInt64(&w.heartbeatCount)
	metrics["active_connections"] = w.getActiveConnectionCount()
	
	return metrics
}

// getActiveConnectionCount 获取活跃连接数
func (w *WebSocketAdapter) getActiveConnectionCount() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	
	count := 0
	for _, conn := range w.connections {
		if conn.isConnected {
			count++
		}
	}
	return count
}

// HealthCheck 健康检查
func (w *WebSocketAdapter) HealthCheck(ctx context.Context) error {
	if !w.isConnected {
		return fmt.Errorf("adapter not connected")
	}

	// 检查是否有可用连接
	if w.getAvailableConnection() == nil {
		return fmt.Errorf("no available WebSocket connections")
	}

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