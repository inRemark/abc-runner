package config

import (
	"fmt"
	"strings"
	"time"

	"abc-runner/app/core/interfaces"
)

// WebSocketConfig WebSocket协议配置
type WebSocketConfig struct {
	Protocol          string                  `yaml:"protocol" json:"protocol"`
	Connection        ConnectionConfig        `yaml:"connection" json:"connection"`
	BenchMark         BenchmarkConfig         `yaml:"benchmark" json:"benchmark"`
	WebSocketSpecific WebSocketSpecificConfig `yaml:"websocket_specific" json:"websocket_specific"`
}

// ConnectionConfig WebSocket连接配置
type ConnectionConfig struct {
	URL          string        `yaml:"url" json:"url"`
	Timeout      time.Duration `yaml:"timeout" json:"timeout"`
	PingInterval time.Duration `yaml:"ping_interval" json:"ping_interval"`
	PongTimeout  time.Duration `yaml:"pong_timeout" json:"pong_timeout"`
	Pool         PoolConfig    `yaml:"pool" json:"pool"`
}

// PoolConfig WebSocket连接池配置
type PoolConfig struct {
	PoolSize          int           `yaml:"pool_size" json:"pool_size"`
	MinIdle           int           `yaml:"min_idle" json:"min_idle"`
	MaxIdle           int           `yaml:"max_idle" json:"max_idle"`
	IdleTimeout       time.Duration `yaml:"idle_timeout" json:"idle_timeout"`
	ConnectionTimeout time.Duration `yaml:"connection_timeout" json:"connection_timeout"`
}

// BenchmarkConfig WebSocket基准测试配置
type BenchmarkConfig struct {
	Total       int           `yaml:"total" json:"total"`
	Parallels   int           `yaml:"parallels" json:"parallels"`
	DataSize    int           `yaml:"data_size" json:"data_size"`
	TTL         time.Duration `yaml:"ttl" json:"ttl"`
	ReadPercent int           `yaml:"read_percent" json:"read_percent"`
	RandomKeys  int           `yaml:"random_keys" json:"random_keys"`
	TestCase    string        `yaml:"test_case" json:"test_case"`
	Duration    time.Duration `yaml:"duration" json:"duration"`
}

// WebSocketSpecificConfig WebSocket特定配置
type WebSocketSpecificConfig struct {
	Subprotocol          string          `yaml:"subprotocol" json:"subprotocol"`
	Compression          bool            `yaml:"compression" json:"compression"`
	MessageType          string          `yaml:"message_type" json:"message_type"` // "text", "binary"
	AutoReconnect        bool            `yaml:"auto_reconnect" json:"auto_reconnect"`
	ReconnectInterval    time.Duration   `yaml:"reconnect_interval" json:"reconnect_interval"`
	MaxReconnectAttempts int             `yaml:"max_reconnect_attempts" json:"max_reconnect_attempts"`
	BufferSize           int             `yaml:"buffer_size" json:"buffer_size"`
	WriteTimeout         time.Duration   `yaml:"write_timeout" json:"write_timeout"`
	ReadTimeout          time.Duration   `yaml:"read_timeout" json:"read_timeout"`
	Heartbeat            HeartbeatConfig `yaml:"heartbeat" json:"heartbeat"`
	Extensions           ExtensionConfig `yaml:"extensions" json:"extensions"`
}

// HeartbeatConfig 心跳配置
type HeartbeatConfig struct {
	Enabled   bool          `yaml:"enabled" json:"enabled"`
	Interval  time.Duration `yaml:"interval" json:"interval"`
	Timeout   time.Duration `yaml:"timeout" json:"timeout"`
	MaxMissed int           `yaml:"max_missed" json:"max_missed"`
}

// ExtensionConfig 扩展配置
type ExtensionConfig struct {
	PerMessageDeflate bool `yaml:"per_message_deflate" json:"per_message_deflate"`
	NoContextTakeover bool `yaml:"no_context_takeover" json:"no_context_takeover"`
	MaxWindowBits     int  `yaml:"max_window_bits" json:"max_window_bits"`
}

// NewDefaultWebSocketConfig 创建默认WebSocket配置
func NewDefaultWebSocketConfig() *WebSocketConfig {
	return &WebSocketConfig{
		Protocol: "websocket",
		Connection: ConnectionConfig{
			URL:          "ws://localhost:8080/ws",
			Timeout:      30 * time.Second,
			PingInterval: 30 * time.Second,
			PongTimeout:  10 * time.Second,
			Pool: PoolConfig{
				PoolSize:          10,
				MinIdle:           1,
				MaxIdle:           5,
				IdleTimeout:       300 * time.Second,
				ConnectionTimeout: 30 * time.Second,
			},
		},
		BenchMark: BenchmarkConfig{
			Total:       2000,
			Parallels:   30,
			DataSize:    1024,
			ReadPercent: 50,
			RandomKeys:  1000,
			TestCase:    "message_exchange",
			Duration:    60 * time.Second,
		},
		WebSocketSpecific: WebSocketSpecificConfig{
			Subprotocol:          "",
			Compression:          false,
			MessageType:          "text",
			AutoReconnect:        true,
			ReconnectInterval:    5 * time.Second,
			MaxReconnectAttempts: 3,
			BufferSize:           4096,
			WriteTimeout:         10 * time.Second,
			ReadTimeout:          10 * time.Second,
			Heartbeat: HeartbeatConfig{
				Enabled:   true,
				Interval:  30 * time.Second,
				Timeout:   10 * time.Second,
				MaxMissed: 3,
			},
			Extensions: ExtensionConfig{
				PerMessageDeflate: false,
				NoContextTakeover: false,
				MaxWindowBits:     15,
			},
		},
	}
}

// GetProtocol 实现Config接口
func (c *WebSocketConfig) GetProtocol() string {
	return c.Protocol
}

// GetConnection 实现Config接口
func (c *WebSocketConfig) GetConnection() interfaces.ConnectionConfig {
	return &c.Connection
}

// GetBenchmark 实现Config接口
func (c *WebSocketConfig) GetBenchmark() interfaces.BenchmarkConfig {
	return &c.BenchMark
}

// Validate 实现Config接口
func (c *WebSocketConfig) Validate() error {
	if c.Connection.URL == "" {
		return fmt.Errorf("connection URL cannot be empty")
	}

	// 验证URL格式
	if !strings.HasPrefix(c.Connection.URL, "ws://") && !strings.HasPrefix(c.Connection.URL, "wss://") {
		return fmt.Errorf("invalid WebSocket URL: must start with ws:// or wss://")
	}

	if c.BenchMark.Total <= 0 {
		return fmt.Errorf("total operations must be greater than 0")
	}

	if c.BenchMark.Parallels <= 0 {
		return fmt.Errorf("parallel connections must be greater than 0")
	}

	if c.BenchMark.DataSize <= 0 {
		return fmt.Errorf("data size must be greater than 0")
	}

	// 验证测试用例
	validTestCases := []string{"message_exchange", "ping_pong", "broadcast", "large_message"}
	valid := false
	for _, testCase := range validTestCases {
		if c.BenchMark.TestCase == testCase {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid test case: %s, valid options: %s",
			c.BenchMark.TestCase, strings.Join(validTestCases, ", "))
	}

	// 验证消息类型
	if c.WebSocketSpecific.MessageType != "text" && c.WebSocketSpecific.MessageType != "binary" {
		return fmt.Errorf("invalid message type: %s, must be 'text' or 'binary'", c.WebSocketSpecific.MessageType)
	}

	return nil
}

// Clone 实现Config接口
func (c *WebSocketConfig) Clone() interfaces.Config {
	clone := *c
	return &clone
}

// ConnectionConfig接口实现

// GetAddresses 实现ConnectionConfig接口
func (c *ConnectionConfig) GetAddresses() []string {
	return []string{c.URL}
}

// GetCredentials 实现ConnectionConfig接口
func (c *ConnectionConfig) GetCredentials() map[string]string {
	return map[string]string{} // WebSocket认证通过其他机制处理
}

// GetPoolConfig 实现ConnectionConfig接口
func (c *ConnectionConfig) GetPoolConfig() interfaces.PoolConfig {
	return &c.Pool
}

// GetTimeout 实现ConnectionConfig接口
func (c *ConnectionConfig) GetTimeout() time.Duration {
	return c.Timeout
}

// PoolConfig接口实现

// GetPoolSize 实现PoolConfig接口
func (p *PoolConfig) GetPoolSize() int {
	return p.PoolSize
}

// GetMinIdle 实现PoolConfig接口
func (p *PoolConfig) GetMinIdle() int {
	return p.MinIdle
}

// GetMaxIdle 实现PoolConfig接口
func (p *PoolConfig) GetMaxIdle() int {
	return p.MaxIdle
}

// GetIdleTimeout 实现PoolConfig接口
func (p *PoolConfig) GetIdleTimeout() time.Duration {
	return p.IdleTimeout
}

// GetConnectionTimeout 实现PoolConfig接口
func (p *PoolConfig) GetConnectionTimeout() time.Duration {
	return p.ConnectionTimeout
}

// BenchmarkConfig接口实现

// GetTotal 实现BenchmarkConfig接口
func (b *BenchmarkConfig) GetTotal() int {
	return b.Total
}

// GetParallels 实现BenchmarkConfig接口
func (b *BenchmarkConfig) GetParallels() int {
	return b.Parallels
}

// GetDataSize 实现BenchmarkConfig接口
func (b *BenchmarkConfig) GetDataSize() int {
	return b.DataSize
}

// GetTTL 实现BenchmarkConfig接口
func (b *BenchmarkConfig) GetTTL() time.Duration {
	return b.TTL
}

// GetReadPercent 实现BenchmarkConfig接口
func (b *BenchmarkConfig) GetReadPercent() int {
	return b.ReadPercent
}

// GetRandomKeys 实现BenchmarkConfig接口
func (b *BenchmarkConfig) GetRandomKeys() int {
	return b.RandomKeys
}

// GetTestCase 实现BenchmarkConfig接口
func (b *BenchmarkConfig) GetTestCase() string {
	return b.TestCase
}

// 确保实现了所有接口
var (
	_ interfaces.Config           = (*WebSocketConfig)(nil)
	_ interfaces.ConnectionConfig = (*ConnectionConfig)(nil)
	_ interfaces.PoolConfig       = (*PoolConfig)(nil)
	_ interfaces.BenchmarkConfig  = (*BenchmarkConfig)(nil)
)
