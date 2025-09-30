package websocket

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"

	"abc-runner/servers/pkg/interfaces"
)

// Connection WebSocket连接封装
type Connection struct {
	id           string
	conn         *websocket.Conn
	remoteAddr   string
	userAgent    string
	connectedAt  time.Time
	lastActivity time.Time
	state        ConnectionState
	subprotocol  string
	headers      map[string]string

	// 统计信息
	bytesSent    int64
	bytesRecv    int64
	messagesSent int64
	messagesRecv int64
	pingsSent    int64
	pongsRecv    int64
	missedPongs  int

	// 消息队列
	sendQueue chan []byte
	closeOnce sync.Once
	done      chan struct{}
	
	// 心跳
	lastPingTime time.Time
	lastPongTime time.Time

	// 锁
	mutex sync.RWMutex

	// 配置和依赖
	config           *WebSocketServerConfig
	logger           interfaces.Logger
	metricsCollector interfaces.MetricsCollector
}

// ConnectionState 连接状态
type ConnectionState int

const (
	StateConnecting ConnectionState = iota
	StateConnected
	StateClosing
	StateClosed
)

// String 返回连接状态的字符串表示
func (s ConnectionState) String() string {
	switch s {
	case StateConnecting:
		return "connecting"
	case StateConnected:
		return "connected"
	case StateClosing:
		return "closing"
	case StateClosed:
		return "closed"
	default:
		return "unknown"
	}
}

// ConnectionManager WebSocket连接管理器
type ConnectionManager struct {
	connections      map[string]*Connection
	maxConnections   int
	cleanupInterval  time.Duration
	mutex            sync.RWMutex
	logger           interfaces.Logger
	metricsCollector interfaces.MetricsCollector
	config           *WebSocketServerConfig

	// 清理协程控制
	cleanupDone chan struct{}
	cleanupOnce sync.Once
}

// NewConnection 创建新的WebSocket连接
func NewConnection(conn *websocket.Conn, config *WebSocketServerConfig, logger interfaces.Logger, metricsCollector interfaces.MetricsCollector) *Connection {
	id := GenerateConnectionID()
	
	// 获取连接信息
	remoteAddr := conn.RemoteAddr().String()
	userAgent := ""
	headers := make(map[string]string)
	
	// 从升级请求中获取用户代理和头部信息
	// 注意：这里简化处理，实际实现中应该从HTTP升级请求中获取这些信息
	userAgent = "WebSocket Client"

	c := &Connection{
		id:               id,
		conn:             conn,
		remoteAddr:       remoteAddr,
		userAgent:        userAgent,
		connectedAt:      time.Now(),
		lastActivity:     time.Now(),
		state:            StateConnected,
		headers:          headers,
		sendQueue:        make(chan []byte, config.Message.QueueSize),
		done:             make(chan struct{}),
		config:           config,
		logger:           logger,
		metricsCollector: metricsCollector,
	}

	// 启动读写协程
	go c.readPump()
	go c.writePump()
	
	// 如果启用心跳，启动心跳协程
	if config.Heartbeat.Enabled {
		go c.heartbeatPump()
	}

	return c
}

// GetID 获取连接ID
func (c *Connection) GetID() string {
	return c.id
}

// GetState 获取连接状态
func (c *Connection) GetState() ConnectionState {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.state
}

// IsConnected 检查连接是否处于连接状态
func (c *Connection) IsConnected() bool {
	return c.GetState() == StateConnected
}

// SendMessage 发送消息
func (c *Connection) SendMessage(messageType int, data []byte) error {
	if !c.IsConnected() {
		return fmt.Errorf("connection %s is not connected", c.id)
	}

	// 检查消息大小
	if len(data) > c.config.Message.MaxMessageSize {
		return fmt.Errorf("message too large: %d bytes, maximum is %d", len(data), c.config.Message.MaxMessageSize)
	}

	select {
	case c.sendQueue <- data:
		return nil
	case <-c.done:
		return fmt.Errorf("connection %s is closed", c.id)
	default:
		return fmt.Errorf("send queue is full for connection %s", c.id)
	}
}

// SendText 发送文本消息
func (c *Connection) SendText(text string) error {
	return c.SendMessage(websocket.TextMessage, []byte(text))
}

// SendBinary 发送二进制消息
func (c *Connection) SendBinary(data []byte) error {
	return c.SendMessage(websocket.BinaryMessage, data)
}

// SendPing 发送Ping消息
func (c *Connection) SendPing(data []byte) error {
	if !c.IsConnected() {
		return fmt.Errorf("connection %s is not connected", c.id)
	}

	c.mutex.Lock()
	c.lastPingTime = time.Now()
	c.pingsSent++
	c.mutex.Unlock()

	return c.conn.WriteMessage(websocket.PingMessage, data)
}

// Close 关闭连接
func (c *Connection) Close() error {
	var err error
	c.closeOnce.Do(func() {
		c.mutex.Lock()
		c.state = StateClosing
		c.mutex.Unlock()

		// 检查WebSocket连接是否存在（用于测试中的模拟连接）
		if c.conn != nil {
			// 发送关闭消息
			closeMessage := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "connection closed")
			c.conn.WriteMessage(websocket.CloseMessage, closeMessage)

			// 关闭连接
			err = c.conn.Close()
		}

		c.mutex.Lock()
		c.state = StateClosed
		c.mutex.Unlock()

		close(c.done)

		if c.logger != nil {
			c.logger.Debug("Connection closed", map[string]interface{}{
				"connection_id": c.id,
				"remote_addr":   c.remoteAddr,
				"duration":      time.Since(c.connectedAt).String(),
				"bytes_sent":    c.bytesSent,
				"bytes_recv":    c.bytesRecv,
			})
		}
	})

	return err
}

// GetInfo 获取连接信息
func (c *Connection) GetInfo() ConnectionInfo {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return ConnectionInfo{
		ID:           c.id,
		RemoteAddr:   c.remoteAddr,
		UserAgent:    c.userAgent,
		ConnectedAt:  c.connectedAt,
		LastActivity: c.lastActivity,
		BytesSent:    c.bytesSent,
		BytesRecv:    c.bytesRecv,
		MessagesSent: c.messagesSent,
		MessagesRecv: c.messagesRecv,
		PingsSent:    c.pingsSent,
		PongsRecv:    c.pongsRecv,
		MissedPongs:  c.missedPongs,
		State:        c.state.String(),
		Subprotocol:  c.subprotocol,
		Headers:      c.headers,
	}
}

// readPump 读取消息的协程
func (c *Connection) readPump() {
	defer c.Close()

	// 设置读取超时
	c.conn.SetReadDeadline(time.Now().Add(c.config.Connection.ReadTimeout))
	
	// 设置最大消息大小
	c.conn.SetReadLimit(int64(c.config.Message.MaxMessageSize))

	// 设置Pong处理器
	c.conn.SetPongHandler(func(data string) error {
		c.mutex.Lock()
		c.lastPongTime = time.Now()
		c.pongsRecv++
		c.lastActivity = time.Now()
		c.missedPongs = 0 // 重置丢失计数
		c.mutex.Unlock()

		c.conn.SetReadDeadline(time.Now().Add(c.config.Connection.ReadTimeout))
		
		if c.config.Logging.LogHeartbeat && c.logger != nil {
			c.logger.Debug("Pong received", map[string]interface{}{
				"connection_id": c.id,
				"data":          data,
			})
		}
		
		return nil
	})

	for {
		messageType, data, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				if c.logger != nil {
					c.logger.Error("WebSocket read error", err, map[string]interface{}{
						"connection_id": c.id,
					})
				}
			}
			break
		}

		// 更新统计信息
		c.mutex.Lock()
		c.lastActivity = time.Now()
		c.bytesRecv += int64(len(data))
		c.messagesRecv++
		c.mutex.Unlock()

		// 记录指标
		if c.metricsCollector != nil {
			c.metricsCollector.RecordRequest("websocket", "message_received", 0, true)
		}

		// 处理消息
		c.handleMessage(messageType, data)

		// 重置读取超时
		c.conn.SetReadDeadline(time.Now().Add(c.config.Connection.ReadTimeout))
	}
}

// writePump 写入消息的协程
func (c *Connection) writePump() {
	ticker := time.NewTicker(54 * time.Second) // WebSocket ping间隔
	defer ticker.Stop()

	for {
		select {
		case data, ok := <-c.sendQueue:
			c.conn.SetWriteDeadline(time.Now().Add(c.config.Connection.WriteTimeout))
			
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// 发送消息
			messageType := websocket.TextMessage
			if !isTextMessage(data) {
				messageType = websocket.BinaryMessage
			}

			if err := c.conn.WriteMessage(messageType, data); err != nil {
				if c.logger != nil {
					c.logger.Error("WebSocket write error", err, map[string]interface{}{
						"connection_id": c.id,
					})
				}
				return
			}

			// 更新统计信息
			c.mutex.Lock()
			c.bytesSent += int64(len(data))
			c.messagesSent++
			c.mutex.Unlock()

			// 记录指标
			if c.metricsCollector != nil {
				c.metricsCollector.RecordRequest("websocket", "message_sent", 0, true)
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(c.config.Connection.WriteTimeout))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case <-c.done:
			return
		}
	}
}

// heartbeatPump 心跳协程
func (c *Connection) heartbeatPump() {
	ticker := time.NewTicker(c.config.Heartbeat.PingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !c.IsConnected() {
				return
			}

			// 检查是否超过最大丢失次数
			c.mutex.RLock()
			missedPongs := c.missedPongs
			c.mutex.RUnlock()

			if missedPongs >= c.config.Heartbeat.MaxMissed {
				if c.logger != nil {
					c.logger.Warn("Connection missed too many pongs, closing", map[string]interface{}{
						"connection_id": c.id,
						"missed_pongs":  missedPongs,
					})
				}
				c.Close()
				return
			}

			// 发送心跳
			if err := c.SendPing([]byte("ping")); err != nil {
				if c.logger != nil {
					c.logger.Error("Failed to send ping", err, map[string]interface{}{
						"connection_id": c.id,
					})
				}
				c.Close()
				return
			}

			// 增加丢失计数（将在收到pong时重置）
			c.mutex.Lock()
			c.missedPongs++
			c.mutex.Unlock()

		case <-c.done:
			return
		}
	}
}

// handleMessage 处理接收到的消息
func (c *Connection) handleMessage(messageType int, data []byte) {
	if c.config.Logging.LogMessages && c.logger != nil {
		c.logger.Debug("Message received", map[string]interface{}{
			"connection_id": c.id,
			"type":          messageType,
			"size":          len(data),
		})
	}

	// 如果启用回显模式
	if c.config.Message.EchoMode {
		// 添加响应延迟
		if c.config.Message.ResponseDelay > 0 {
			time.Sleep(c.config.Message.ResponseDelay)
		}

		// 回显消息
		if err := c.SendMessage(messageType, data); err != nil {
			if c.logger != nil {
				c.logger.Error("Failed to echo message", err, map[string]interface{}{
					"connection_id": c.id,
				})
			}
		}
	}
}

// NewConnectionManager 创建连接管理器
func NewConnectionManager(config *WebSocketServerConfig, logger interfaces.Logger, metricsCollector interfaces.MetricsCollector) *ConnectionManager {
	cm := &ConnectionManager{
		connections:      make(map[string]*Connection),
		maxConnections:   config.Connection.MaxConnections,
		cleanupInterval:  config.Connection.CleanupInterval,
		logger:           logger,
		metricsCollector: metricsCollector,
		config:           config,
		cleanupDone:      make(chan struct{}),
	}

	// 启动清理协程
	go cm.cleanupLoop()

	return cm
}

// AddConnection 添加连接
func (cm *ConnectionManager) AddConnection(conn *Connection) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if len(cm.connections) >= cm.maxConnections {
		return fmt.Errorf("maximum connections reached: %d", cm.maxConnections)
	}

	cm.connections[conn.GetID()] = conn

	if cm.metricsCollector != nil {
		cm.metricsCollector.RecordConnection("websocket", "open")
	}

	if cm.config.Logging.LogConnections && cm.logger != nil {
		cm.logger.Info("WebSocket connection added", map[string]interface{}{
			"connection_id":     conn.GetID(),
			"remote_addr":       conn.remoteAddr,
			"total_connections": len(cm.connections),
		})
	}

	return nil
}

// RemoveConnection 移除连接
func (cm *ConnectionManager) RemoveConnection(connectionID string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if conn, exists := cm.connections[connectionID]; exists {
		delete(cm.connections, connectionID)

		if cm.metricsCollector != nil {
			cm.metricsCollector.RecordConnection("websocket", "close")
		}

		if cm.config.Logging.LogConnections && cm.logger != nil {
			info := conn.GetInfo()
			cm.logger.Info("WebSocket connection removed", map[string]interface{}{
				"connection_id":         connectionID,
				"remote_addr":           info.RemoteAddr,
				"duration":              time.Since(info.ConnectedAt).String(),
				"bytes_sent":            info.BytesSent,
				"bytes_recv":            info.BytesRecv,
				"remaining_connections": len(cm.connections),
			})
		}
	}
}

// GetConnection 获取连接
func (cm *ConnectionManager) GetConnection(connectionID string) (*Connection, bool) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	conn, exists := cm.connections[connectionID]
	return conn, exists
}

// GetConnectionCount 获取连接数量
func (cm *ConnectionManager) GetConnectionCount() int {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	return len(cm.connections)
}

// GetAllConnections 获取所有连接信息
func (cm *ConnectionManager) GetAllConnections() []ConnectionInfo {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	connections := make([]ConnectionInfo, 0, len(cm.connections))
	for _, conn := range cm.connections {
		connections = append(connections, conn.GetInfo())
	}

	return connections
}

// BroadcastMessage 广播消息到所有连接
func (cm *ConnectionManager) BroadcastMessage(messageType int, data []byte) int {
	cm.mutex.RLock()
	connections := make([]*Connection, 0, len(cm.connections))
	for _, conn := range cm.connections {
		if conn.IsConnected() {
			connections = append(connections, conn)
		}
	}
	cm.mutex.RUnlock()

	successCount := 0
	for _, conn := range connections {
		if err := conn.SendMessage(messageType, data); err == nil {
			successCount++
		}
	}

	return successCount
}

// CloseAll 关闭所有连接
func (cm *ConnectionManager) CloseAll() {
	cm.mutex.Lock()
	connections := make([]*Connection, 0, len(cm.connections))
	for _, conn := range cm.connections {
		connections = append(connections, conn)
	}
	cm.connections = make(map[string]*Connection)
	cm.mutex.Unlock()

	for _, conn := range connections {
		conn.Close()
	}

	if cm.logger != nil {
		cm.logger.Info("All WebSocket connections closed")
	}
}

// Shutdown 关闭连接管理器
func (cm *ConnectionManager) Shutdown() {
	cm.cleanupOnce.Do(func() {
		close(cm.cleanupDone)
	})
	cm.CloseAll()
}

// cleanupLoop 清理循环
func (cm *ConnectionManager) cleanupLoop() {
	ticker := time.NewTicker(cm.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cm.cleanup()
		case <-cm.cleanupDone:
			return
		}
	}
}

// cleanup 清理无效连接
func (cm *ConnectionManager) cleanup() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	now := time.Now()
	toRemove := make([]string, 0)

	for id, conn := range cm.connections {
		// 检查连接状态
		if conn.GetState() == StateClosed {
			toRemove = append(toRemove, id)
			continue
		}

		// 检查空闲超时
		conn.mutex.RLock()
		lastActivity := conn.lastActivity
		conn.mutex.RUnlock()

		if now.Sub(lastActivity) > cm.config.Connection.IdleTimeout {
			toRemove = append(toRemove, id)
			conn.Close()
		}
	}

	// 移除无效连接
	for _, id := range toRemove {
		delete(cm.connections, id)
	}

	if len(toRemove) > 0 && cm.logger != nil {
		cm.logger.Debug("Cleaned up inactive connections", map[string]interface{}{
			"removed_count":        len(toRemove),
			"remaining_connections": len(cm.connections),
		})
	}
}

// 工具函数

// GenerateConnectionID 生成连接ID
func GenerateConnectionID() string {
	return fmt.Sprintf("ws-%d-%d", time.Now().UnixNano(), atomic.AddInt64(&connectionCounter, 1))
}

// isTextMessage 检查是否为文本消息
func isTextMessage(data []byte) bool {
	// 简单检查是否为UTF-8文本
	for _, b := range data {
		// 检查是否包含控制字符（除了tab、换行、回车）
		if b < 32 && b != 9 && b != 10 && b != 13 {
			return false
		}
		// 检查高位字节
		if b > 127 {
			// 对于高位字节，进行更严格的UTF-8检查
			// 这里简化处理，假设高位字节可能是二进制
			if b >= 240 { // 0xF0及以上通常是二进制数据
				return false
			}
		}
	}
	return true
}

var connectionCounter int64