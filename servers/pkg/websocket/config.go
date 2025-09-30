package websocket

import (
	"fmt"
	"time"

	"abc-runner/servers/internal/common"
	"abc-runner/servers/pkg/interfaces"
)

// WebSocketServerConfig WebSocket服务端配置
type WebSocketServerConfig struct {
	*common.BaseConfig `yaml:",inline"`

	// WebSocket升级配置
	Upgrader UpgraderConfig `yaml:"upgrader" json:"upgrader"`

	// 连接管理配置
	Connection ConnectionConfig `yaml:"connection" json:"connection"`

	// 心跳配置
	Heartbeat HeartbeatConfig `yaml:"heartbeat" json:"heartbeat"`

	// 消息配置
	Message MessageConfig `yaml:"message" json:"message"`

	// HTTP服务器配置（用于WebSocket升级）
	HTTPServer HTTPServerConfig `yaml:"http_server" json:"http_server"`

	// 日志配置
	Logging LoggingConfig `yaml:"logging" json:"logging"`
}

// UpgraderConfig WebSocket升级器配置
type UpgraderConfig struct {
	Path              string        `yaml:"path" json:"path"`                             // WebSocket端点路径，默认"/ws"
	ReadBufferSize    int           `yaml:"read_buffer_size" json:"read_buffer_size"`     // 读取缓冲区大小
	WriteBufferSize   int           `yaml:"write_buffer_size" json:"write_buffer_size"`   // 写入缓冲区大小
	HandshakeTimeout  time.Duration `yaml:"handshake_timeout" json:"handshake_timeout"`  // 握手超时时间
	EnableCompression bool          `yaml:"enable_compression" json:"enable_compression"` // 启用压缩支持
	CheckOrigin       bool          `yaml:"check_origin" json:"check_origin"`             // 检查来源
	AllowedOrigins    []string      `yaml:"allowed_origins" json:"allowed_origins"`       // 允许的来源列表
	Subprotocols      []string      `yaml:"subprotocols" json:"subprotocols"`             // 支持的子协议
}

// ConnectionConfig 连接管理配置
type ConnectionConfig struct {
	MaxConnections    int           `yaml:"max_connections" json:"max_connections"`       // 最大并发连接数
	ConnectionTimeout time.Duration `yaml:"connection_timeout" json:"connection_timeout"` // 连接超时时间
	ReadTimeout       time.Duration `yaml:"read_timeout" json:"read_timeout"`             // 读取超时时间
	WriteTimeout      time.Duration `yaml:"write_timeout" json:"write_timeout"`           // 写入超时时间
	IdleTimeout       time.Duration `yaml:"idle_timeout" json:"idle_timeout"`             // 空闲超时时间
	CleanupInterval   time.Duration `yaml:"cleanup_interval" json:"cleanup_interval"`     // 清理间隔
}

// HeartbeatConfig 心跳配置
type HeartbeatConfig struct {
	Enabled      bool          `yaml:"enabled" json:"enabled"`             // 启用心跳
	PingInterval time.Duration `yaml:"ping_interval" json:"ping_interval"` // Ping间隔
	PongTimeout  time.Duration `yaml:"pong_timeout" json:"pong_timeout"`   // Pong响应超时
	MaxMissed    int           `yaml:"max_missed" json:"max_missed"`       // 最大丢失次数
}

// MessageConfig 消息配置
type MessageConfig struct {
	MaxMessageSize int           `yaml:"max_message_size" json:"max_message_size"` // 最大消息大小
	MessageTimeout time.Duration `yaml:"message_timeout" json:"message_timeout"`  // 消息处理超时
	QueueSize      int           `yaml:"queue_size" json:"queue_size"`             // 消息队列大小
	EchoMode       bool          `yaml:"echo_mode" json:"echo_mode"`               // 回显模式
	ResponseDelay  time.Duration `yaml:"response_delay" json:"response_delay"`    // 响应延迟
}

// HTTPServerConfig HTTP服务器配置
type HTTPServerConfig struct {
	ReadTimeout     time.Duration `yaml:"read_timeout" json:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout" json:"write_timeout"`
	IdleTimeout     time.Duration `yaml:"idle_timeout" json:"idle_timeout"`
	MaxHeaderBytes  int           `yaml:"max_header_bytes" json:"max_header_bytes"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" json:"shutdown_timeout"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	LogConnections bool `yaml:"log_connections" json:"log_connections"` // 记录连接日志
	LogMessages    bool `yaml:"log_messages" json:"log_messages"`       // 记录消息日志
	LogHeartbeat   bool `yaml:"log_heartbeat" json:"log_heartbeat"`     // 记录心跳日志
}

// NewWebSocketServerConfig 创建WebSocket服务端配置
func NewWebSocketServerConfig() *WebSocketServerConfig {
	return &WebSocketServerConfig{
		BaseConfig: &common.BaseConfig{
			Protocol: "websocket",
			Host:     "localhost",
			Port:     7070,
		},
		Upgrader: UpgraderConfig{
			Path:              "/ws",
			ReadBufferSize:    4096,
			WriteBufferSize:   4096,
			HandshakeTimeout:  10 * time.Second,
			EnableCompression: false,
			CheckOrigin:       false,
			AllowedOrigins:    []string{"*"},
			Subprotocols:      []string{},
		},
		Connection: ConnectionConfig{
			MaxConnections:    1000,
			ConnectionTimeout: 30 * time.Second,
			ReadTimeout:       60 * time.Second,
			WriteTimeout:      10 * time.Second,
			IdleTimeout:       300 * time.Second, // 5分钟
			CleanupInterval:   60 * time.Second,  // 1分钟
		},
		Heartbeat: HeartbeatConfig{
			Enabled:      true,
			PingInterval: 30 * time.Second,
			PongTimeout:  10 * time.Second,
			MaxMissed:    3,
		},
		Message: MessageConfig{
			MaxMessageSize: 1024 * 1024, // 1MB
			MessageTimeout: 30 * time.Second,
			QueueSize:      100,
			EchoMode:       true,
			ResponseDelay:  0,
		},
		HTTPServer: HTTPServerConfig{
			ReadTimeout:     30 * time.Second,
			WriteTimeout:    30 * time.Second,
			IdleTimeout:     120 * time.Second,
			MaxHeaderBytes:  1048576, // 1MB
			ShutdownTimeout: 30 * time.Second,
		},
		Logging: LoggingConfig{
			LogConnections: true,
			LogMessages:    false,
			LogHeartbeat:   false,
		},
	}
}

// Validate 验证WebSocket配置
func (c *WebSocketServerConfig) Validate() error {
	if err := c.BaseConfig.Validate(); err != nil {
		return fmt.Errorf("base config validation failed: %w", err)
	}

	// 验证升级器配置
	if c.Upgrader.Path == "" {
		return fmt.Errorf("upgrader path cannot be empty")
	}
	if c.Upgrader.ReadBufferSize <= 0 {
		return fmt.Errorf("read_buffer_size must be positive")
	}
	if c.Upgrader.WriteBufferSize <= 0 {
		return fmt.Errorf("write_buffer_size must be positive")
	}
	if c.Upgrader.HandshakeTimeout <= 0 {
		return fmt.Errorf("handshake_timeout must be positive")
	}

	// 验证连接配置
	if c.Connection.MaxConnections <= 0 {
		return fmt.Errorf("max_connections must be positive")
	}
	if c.Connection.ConnectionTimeout <= 0 {
		return fmt.Errorf("connection_timeout must be positive")
	}
	if c.Connection.ReadTimeout <= 0 {
		return fmt.Errorf("read_timeout must be positive")
	}
	if c.Connection.WriteTimeout <= 0 {
		return fmt.Errorf("write_timeout must be positive")
	}
	if c.Connection.IdleTimeout <= 0 {
		return fmt.Errorf("idle_timeout must be positive")
	}
	if c.Connection.CleanupInterval <= 0 {
		return fmt.Errorf("cleanup_interval must be positive")
	}

	// 验证心跳配置
	if c.Heartbeat.Enabled {
		if c.Heartbeat.PingInterval <= 0 {
			return fmt.Errorf("ping_interval must be positive when heartbeat is enabled")
		}
		if c.Heartbeat.PongTimeout <= 0 {
			return fmt.Errorf("pong_timeout must be positive when heartbeat is enabled")
		}
		if c.Heartbeat.MaxMissed <= 0 {
			return fmt.Errorf("max_missed must be positive when heartbeat is enabled")
		}
	}

	// 验证消息配置
	if c.Message.MaxMessageSize <= 0 {
		return fmt.Errorf("max_message_size must be positive")
	}
	if c.Message.MaxMessageSize > 100*1024*1024 { // 100MB limit
		return fmt.Errorf("max_message_size too large, maximum is 100MB")
	}
	if c.Message.MessageTimeout <= 0 {
		return fmt.Errorf("message_timeout must be positive")
	}
	if c.Message.QueueSize <= 0 {
		return fmt.Errorf("queue_size must be positive")
	}

	// 验证HTTP服务器配置
	if c.HTTPServer.ReadTimeout <= 0 {
		return fmt.Errorf("http_server read_timeout must be positive")
	}
	if c.HTTPServer.WriteTimeout <= 0 {
		return fmt.Errorf("http_server write_timeout must be positive")
	}
	if c.HTTPServer.IdleTimeout <= 0 {
		return fmt.Errorf("http_server idle_timeout must be positive")
	}
	if c.HTTPServer.MaxHeaderBytes <= 0 {
		return fmt.Errorf("http_server max_header_bytes must be positive")
	}
	if c.HTTPServer.ShutdownTimeout <= 0 {
		return fmt.Errorf("http_server shutdown_timeout must be positive")
	}

	return nil
}

// Clone 克隆WebSocket配置
func (c *WebSocketServerConfig) Clone() interfaces.ServerConfig {
	clone := *c
	clone.BaseConfig = c.BaseConfig.Clone().(*common.BaseConfig)

	// 深拷贝切片
	clone.Upgrader.AllowedOrigins = make([]string, len(c.Upgrader.AllowedOrigins))
	copy(clone.Upgrader.AllowedOrigins, c.Upgrader.AllowedOrigins)

	clone.Upgrader.Subprotocols = make([]string, len(c.Upgrader.Subprotocols))
	copy(clone.Upgrader.Subprotocols, c.Upgrader.Subprotocols)

	return &clone
}

// ConnectionInfo WebSocket连接信息
type ConnectionInfo struct {
	ID             string            `json:"id"`
	RemoteAddr     string            `json:"remote_addr"`
	UserAgent      string            `json:"user_agent"`
	ConnectedAt    time.Time         `json:"connected_at"`
	LastActivity   time.Time         `json:"last_activity"`
	BytesSent      int64             `json:"bytes_sent"`
	BytesRecv      int64             `json:"bytes_recv"`
	MessagesSent   int64             `json:"messages_sent"`
	MessagesRecv   int64             `json:"messages_recv"`
	PingsSent      int64             `json:"pings_sent"`
	PongsRecv      int64             `json:"pongs_recv"`
	MissedPongs    int               `json:"missed_pongs"`
	State          string            `json:"state"` // "connecting", "connected", "closing", "closed"
	Subprotocol    string            `json:"subprotocol"`
	Headers        map[string]string `json:"headers"`
}

// MessageInfo WebSocket消息信息
type MessageInfo struct {
	ConnectionID string    `json:"connection_id"`
	Type         int       `json:"type"`       // WebSocket消息类型 (1:text, 2:binary, 8:close, 9:ping, 10:pong)
	Direction    string    `json:"direction"`  // "in" or "out"
	Size         int       `json:"size"`
	Timestamp    time.Time `json:"timestamp"`
	Data         []byte    `json:"data,omitempty"`
}

// HeartbeatInfo 心跳信息
type HeartbeatInfo struct {
	ConnectionID string    `json:"connection_id"`
	Type         string    `json:"type"`      // "ping" or "pong"
	Timestamp    time.Time `json:"timestamp"`
	Latency      time.Duration `json:"latency,omitempty"` // 仅对pong有效
}