package connection

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"abc-runner/app/adapters/websocket/config"
)

// WebSocketConnectionPool WebSocket连接池
type WebSocketConnectionPool struct {
	config       *config.WebSocketConfig
	connections  map[string]*WebSocketConnection
	activeCount  int64
	totalCreated int64

	// 连接管理
	maxConnections       int
	currentConnections   int
	availableConnections chan *WebSocketConnection

	// 并发控制
	mutex  sync.RWMutex
	closed bool

	// 健康检查
	healthCheckTicker *time.Ticker
	stopHealthCheck   chan struct{}
}

// WebSocketConnection WebSocket连接封装
type WebSocketConnection struct {
	ID        string
	URL       string
	conn      interface{} // 模拟连接对象
	isActive  bool
	lastUsed  time.Time
	createdAt time.Time

	// 统计信息
	messagesSent int64
	messagesRecv int64
	bytesSent    int64
	bytesRecv    int64

	// 心跳信息
	lastPingTime time.Time
	lastPongTime time.Time
	missedPongs  int

	// 并发控制
	mutex sync.RWMutex

	// 连接控制
	sendChan     chan []byte
	done         chan struct{}
	writeTimeout time.Duration
	readTimeout  time.Duration
}

// NewWebSocketConnectionPool 创建WebSocket连接池
func NewWebSocketConnectionPool(cfg *config.WebSocketConfig) (*WebSocketConnectionPool, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	pool := &WebSocketConnectionPool{
		config:               cfg,
		connections:          make(map[string]*WebSocketConnection),
		maxConnections:       cfg.Connection.Pool.PoolSize,
		availableConnections: make(chan *WebSocketConnection, cfg.Connection.Pool.PoolSize),
		stopHealthCheck:      make(chan struct{}),
	}

	return pool, nil
}

// GetConnection 获取可用连接
func (p *WebSocketConnectionPool) GetConnection() (*WebSocketConnection, error) {
	if p.closed {
		return nil, fmt.Errorf("connection pool is closed")
	}

	// 创建新连接
	conn, err := p.createConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to create new connection: %w", err)
	}

	return conn, nil
}

// ReturnConnection 归还连接到池
func (p *WebSocketConnectionPool) ReturnConnection(conn *WebSocketConnection) error {
	return nil // 简化实现
}

// createConnection 创建新的WebSocket连接
func (p *WebSocketConnectionPool) createConnection() (*WebSocketConnection, error) {
	connID := fmt.Sprintf("ws_%d", atomic.AddInt64(&p.totalCreated, 1))

	// 创建连接对象
	wsConn := &WebSocketConnection{
		ID:        connID,
		URL:       p.config.Connection.URL,
		conn:      "mock_websocket_connection",
		isActive:  true,
		lastUsed:  time.Now(),
		createdAt: time.Now(),
		sendChan:  make(chan []byte, 100),
		done:      make(chan struct{}),
		// 初始化心跳相关字段
		lastPingTime: time.Now(),
		lastPongTime: time.Now(),
		missedPongs:  0,
	}

	return wsConn, nil
}

// Close 关闭连接池
func (p *WebSocketConnectionPool) Close() error {
	p.closed = true
	return nil
}

// GetStats 获取连接池统计信息
func (p *WebSocketConnectionPool) GetStats() map[string]interface{} {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	return map[string]interface{}{
		"max_connections":       p.maxConnections,
		"current_connections":   p.currentConnections,
		"active_connections":    int(atomic.LoadInt64(&p.activeCount)),
		"available_connections": p.maxConnections - int(atomic.LoadInt64(&p.activeCount)),
		"total_created":         atomic.LoadInt64(&p.totalCreated),
		"closed":                p.closed,
	}
}

// SendMessage 发送消息
func (c *WebSocketConnection) SendMessage(messageType int, data []byte) error {
	if !c.isActive {
		return fmt.Errorf("connection is not active")
	}

	atomic.AddInt64(&c.messagesSent, 1)
	atomic.AddInt64(&c.bytesSent, int64(len(data)))
	return nil
}

// GetStats 获取连接统计信息
func (c *WebSocketConnection) GetStats() map[string]interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return map[string]interface{}{
		"id":             c.ID,
		"url":            c.URL,
		"is_active":      c.isActive,
		"created_at":     c.createdAt,
		"last_used":      c.lastUsed,
		"messages_sent":  atomic.LoadInt64(&c.messagesSent),
		"messages_recv":  atomic.LoadInt64(&c.messagesRecv),
		"bytes_sent":     atomic.LoadInt64(&c.bytesSent),
		"bytes_recv":     atomic.LoadInt64(&c.bytesRecv),
		"last_ping_time": c.lastPingTime,
		"last_pong_time": c.lastPongTime,
		"missed_pongs":   c.missedPongs,
	}
}
