package config

import (
	"fmt"
	"strings"
	"time"

	"abc-runner/app/core/interfaces"
)

// TCPConfig TCP协议配置
type TCPConfig struct {
	Protocol       string            `yaml:"protocol" json:"protocol"`
	Connection     ConnectionConfig  `yaml:"connection" json:"connection"`
	BenchMark      BenchmarkConfig   `yaml:"benchmark" json:"benchmark"`
	TCPSpecific    TCPSpecificConfig `yaml:"tcp_specific" json:"tcp_specific"`
}

// ConnectionConfig 连接配置
type ConnectionConfig struct {
	Address           string        `yaml:"address" json:"address"`
	Port              int           `yaml:"port" json:"port"`
	Timeout           time.Duration `yaml:"timeout" json:"timeout"`
	KeepAlive         bool          `yaml:"keep_alive" json:"keep_alive"`
	KeepAlivePeriod   time.Duration `yaml:"keep_alive_period" json:"keep_alive_period"`
	Pool              PoolConfig    `yaml:"pool" json:"pool"`
}

// PoolConfig 连接池配置
type PoolConfig struct {
	PoolSize          int           `yaml:"pool_size" json:"pool_size"`
	MinIdle           int           `yaml:"min_idle" json:"min_idle"`
	MaxIdle           int           `yaml:"max_idle" json:"max_idle"`
	IdleTimeout       time.Duration `yaml:"idle_timeout" json:"idle_timeout"`
	ConnectionTimeout time.Duration `yaml:"connection_timeout" json:"connection_timeout"`
}

// BenchmarkConfig 基准测试配置
type BenchmarkConfig struct {
	Total        int           `yaml:"total" json:"total"`
	Parallels    int           `yaml:"parallels" json:"parallels"`
	DataSize     int           `yaml:"data_size" json:"data_size"`
	TTL          time.Duration `yaml:"ttl" json:"ttl"`
	ReadPercent  int           `yaml:"read_percent" json:"read_percent"`
	RandomKeys   int           `yaml:"random_keys" json:"random_keys"`
	TestCase     string        `yaml:"test_case" json:"test_case"`
	Duration     time.Duration `yaml:"duration" json:"duration"`
}

// TCPSpecificConfig TCP特定配置
type TCPSpecificConfig struct {
	ConnectionMode string `yaml:"connection_mode" json:"connection_mode"` // "persistent", "transient"
	NoDelay        bool   `yaml:"no_delay" json:"no_delay"`               // 禁用Nagle算法
	BufferSize     int    `yaml:"buffer_size" json:"buffer_size"`         // 缓冲区大小
	LingerTimeout  int    `yaml:"linger_timeout" json:"linger_timeout"`   // SO_LINGER超时
	ReuseAddress   bool   `yaml:"reuse_address" json:"reuse_address"`     // SO_REUSEADDR
}

// NewDefaultTCPConfig 创建默认TCP配置
func NewDefaultTCPConfig() *TCPConfig {
	return &TCPConfig{
		Protocol: "tcp",
		Connection: ConnectionConfig{
			Address:         "localhost",
			Port:            9090, // 设计文档要求的默认端口
			Timeout:         30 * time.Second,
			KeepAlive:       true,
			KeepAlivePeriod: 30 * time.Second,
			Pool: PoolConfig{
				PoolSize:          10,
				MinIdle:           2, // 设计文档要求的最小空闲连接
				MaxIdle:           8, // 设计文档要求的最大空闲连接
				IdleTimeout:       300 * time.Second,
				ConnectionTimeout: 30 * time.Second,
			},
		},
		BenchMark: BenchmarkConfig{
			Total:       1000,
			Parallels:   10,
			DataSize:    1024,
			ReadPercent: 80,
			RandomKeys:  1000,
			TestCase:    "echo_test",
			Duration:    60 * time.Second,
		},
		TCPSpecific: TCPSpecificConfig{
			ConnectionMode: "persistent",
			NoDelay:        true,
			BufferSize:     4096,
			LingerTimeout:  -1,
			ReuseAddress:   true,
		},
	}
}

// GetProtocol 实现Config接口
func (c *TCPConfig) GetProtocol() string {
	return c.Protocol
}

// GetConnection 实现Config接口
func (c *TCPConfig) GetConnection() interfaces.ConnectionConfig {
	return &c.Connection
}

// GetBenchmark 实现Config接口
func (c *TCPConfig) GetBenchmark() interfaces.BenchmarkConfig {
	return &c.BenchMark
}

// Validate 实现Config接口
func (c *TCPConfig) Validate() error {
	if c.Connection.Address == "" {
		return fmt.Errorf("connection address cannot be empty")
	}
	
	if c.Connection.Port <= 0 || c.Connection.Port > 65535 {
		return fmt.Errorf("invalid port number: %d", c.Connection.Port)
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
	validTestCases := []string{"echo_test", "send_only", "receive_only", "bidirectional"}
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
	
	return nil
}

// Clone 实现Config接口
func (c *TCPConfig) Clone() interfaces.Config {
	clone := *c
	return &clone
}

// ConnectionConfig接口实现

// GetAddresses 实现ConnectionConfig接口
func (c *ConnectionConfig) GetAddresses() []string {
	return []string{fmt.Sprintf("%s:%d", c.Address, c.Port)}
}

// GetCredentials 实现ConnectionConfig接口
func (c *ConnectionConfig) GetCredentials() map[string]string {
	return map[string]string{} // TCP通常不需要认证凭据
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