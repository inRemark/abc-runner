package tcp

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"abc-runner/servers/internal/common"
	"abc-runner/servers/pkg/interfaces"
)

// TCPServerConfig TCP服务端配置
type TCPServerConfig struct {
	*common.BaseConfig `yaml:",inline"`
	
	// TCP特定配置
	MaxConnections    int           `yaml:"max_connections" json:"max_connections"`
	ConnectionTimeout time.Duration `yaml:"connection_timeout" json:"connection_timeout"`
	ReadTimeout       time.Duration `yaml:"read_timeout" json:"read_timeout"`
	WriteTimeout      time.Duration `yaml:"write_timeout" json:"write_timeout"`
	KeepAlive         bool          `yaml:"keep_alive" json:"keep_alive"`
	NoDelay           bool          `yaml:"no_delay" json:"no_delay"`
	
	// 缓冲区配置
	BufferSize     int `yaml:"buffer_size" json:"buffer_size"`
	ReadBufferSize int `yaml:"read_buffer_size" json:"read_buffer_size"`
	WriteBufferSize int `yaml:"write_buffer_size" json:"write_buffer_size"`
	
	// 行为配置
	EchoMode        bool          `yaml:"echo_mode" json:"echo_mode"`
	ResponseDelay   time.Duration `yaml:"response_delay" json:"response_delay"`
	MaxMessageSize  int           `yaml:"max_message_size" json:"max_message_size"`
	
	// 日志配置
	LogConnections bool `yaml:"log_connections" json:"log_connections"`
	LogMessages    bool `yaml:"log_messages" json:"log_messages"`
}

// NewTCPServerConfig 创建TCP服务端配置
func NewTCPServerConfig() *TCPServerConfig {
	return &TCPServerConfig{
		BaseConfig: &common.BaseConfig{
			Protocol: "tcp",
			Host:     "localhost",
			Port:     9090,
		},
		MaxConnections:    1000,
		ConnectionTimeout: 30 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		KeepAlive:         true,
		NoDelay:           true,
		BufferSize:        4096,
		ReadBufferSize:    4096,
		WriteBufferSize:   4096,
		EchoMode:          true,
		ResponseDelay:     0,
		MaxMessageSize:    65536, // 64KB
		LogConnections:    true,
		LogMessages:       false,
	}
}

// Validate 验证TCP配置
func (c *TCPServerConfig) Validate() error {
	if err := c.BaseConfig.Validate(); err != nil {
		return fmt.Errorf("base config validation failed: %w", err)
	}
	
	if c.MaxConnections <= 0 {
		return fmt.Errorf("max_connections must be positive")
	}
	
	if c.ConnectionTimeout <= 0 {
		return fmt.Errorf("connection_timeout must be positive")
	}
	
	if c.ReadTimeout <= 0 {
		return fmt.Errorf("read_timeout must be positive")
	}
	
	if c.WriteTimeout <= 0 {
		return fmt.Errorf("write_timeout must be positive")
	}
	
	if c.BufferSize <= 0 {
		return fmt.Errorf("buffer_size must be positive")
	}
	
	if c.ReadBufferSize <= 0 {
		return fmt.Errorf("read_buffer_size must be positive")
	}
	
	if c.WriteBufferSize <= 0 {
		return fmt.Errorf("write_buffer_size must be positive")
	}
	
	if c.MaxMessageSize <= 0 {
		return fmt.Errorf("max_message_size must be positive")
	}
	
	if c.MaxMessageSize > 10*1024*1024 { // 10MB limit
		return fmt.Errorf("max_message_size too large, maximum is 10MB")
	}
	
	return nil
}

// Clone 克隆TCP配置
func (c *TCPServerConfig) Clone() interfaces.ServerConfig {
	clone := *c
	clone.BaseConfig = c.BaseConfig.Clone().(*common.BaseConfig)
	return &clone
}

// ConnectionInfo 连接信息
type ConnectionInfo struct {
	ID         string    `json:"id"`
	RemoteAddr string    `json:"remote_addr"`
	LocalAddr  string    `json:"local_addr"`
	ConnectedAt time.Time `json:"connected_at"`
	BytesSent   int64     `json:"bytes_sent"`
	BytesRecv   int64     `json:"bytes_recv"`
	MessagesSent int64    `json:"messages_sent"`
	MessagesRecv int64    `json:"messages_recv"`
}

// MessageInfo 消息信息
type MessageInfo struct {
	ConnectionID string    `json:"connection_id"`
	Direction    string    `json:"direction"` // "in" or "out"
	Size         int       `json:"size"`
	Timestamp    time.Time `json:"timestamp"`
	Data         []byte    `json:"data,omitempty"`
}

// ConnectionManager 连接管理器
type ConnectionManager struct {
	connections map[string]*Connection
	maxConnections int
	mutex       sync.RWMutex
	logger      interfaces.Logger
	metrics     interfaces.MetricsCollector
}

// NewConnectionManager 创建连接管理器
func NewConnectionManager(maxConnections int, logger interfaces.Logger, metrics interfaces.MetricsCollector) *ConnectionManager {
	return &ConnectionManager{
		connections:    make(map[string]*Connection),
		maxConnections: maxConnections,
		logger:         logger,
		metrics:        metrics,
	}
}

// AddConnection 添加连接
func (cm *ConnectionManager) AddConnection(conn *Connection) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	if len(cm.connections) >= cm.maxConnections {
		return fmt.Errorf("maximum connections reached: %d", cm.maxConnections)
	}
	
	cm.connections[conn.ID] = conn
	
	if cm.metrics != nil {
		cm.metrics.RecordConnection("tcp", "open")
	}
	
	if cm.logger != nil {
		cm.logger.Debug("Connection added", map[string]interface{}{
			"connection_id": conn.ID,
			"remote_addr":   conn.RemoteAddr,
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
		
		if cm.metrics != nil {
			cm.metrics.RecordConnection("tcp", "close")
		}
		
		if cm.logger != nil {
			cm.logger.Debug("Connection removed", map[string]interface{}{
				"connection_id": connectionID,
				"remote_addr":   conn.RemoteAddr,
				"duration":      time.Since(conn.ConnectedAt).String(),
				"bytes_sent":    conn.BytesSent,
				"bytes_recv":    conn.BytesRecv,
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

// CloseAll 关闭所有连接
func (cm *ConnectionManager) CloseAll() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	for _, conn := range cm.connections {
		conn.Close()
	}
	
	cm.connections = make(map[string]*Connection)
	
	if cm.logger != nil {
		cm.logger.Info("All connections closed")
	}
}

// Utils 工具函数

// GenerateConnectionID 生成连接ID
func GenerateConnectionID() string {
	return fmt.Sprintf("tcp-%d-%d", time.Now().UnixNano(), rand.Int63())
}

// ValidateMessage 验证消息
func ValidateMessage(data []byte, maxSize int) error {
	if len(data) == 0 {
		return fmt.Errorf("empty message")
	}
	
	if len(data) > maxSize {
		return fmt.Errorf("message too large: %d bytes, maximum is %d", len(data), maxSize)
	}
	
	return nil
}