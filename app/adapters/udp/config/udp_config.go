package config

import (
	"fmt"
	"net"
	"strings"
	"time"

	"abc-runner/app/core/interfaces"
)

// UDPConfig UDP协议配置
type UDPConfig struct {
	Protocol       string            `yaml:"protocol" json:"protocol"`
	Connection     ConnectionConfig  `yaml:"connection" json:"connection"`
	BenchMark      BenchmarkConfig   `yaml:"benchmark" json:"benchmark"`
	UDPSpecific    UDPSpecificConfig `yaml:"udp_specific" json:"udp_specific"`
}

// ConnectionConfig UDP连接配置
type ConnectionConfig struct {
	Address         string        `yaml:"address" json:"address"`
	Port            int           `yaml:"port" json:"port"`
	Timeout         time.Duration `yaml:"timeout" json:"timeout"`
	BufferSize      int           `yaml:"buffer_size" json:"buffer_size"`
	ReadBufferSize  int           `yaml:"read_buffer_size" json:"read_buffer_size"`
	WriteBufferSize int           `yaml:"write_buffer_size" json:"write_buffer_size"`
}

// BenchmarkConfig UDP基准测试配置
type BenchmarkConfig struct {
	Total        int           `yaml:"total" json:"total"`
	Parallels    int           `yaml:"parallels" json:"parallels"`
	DataSize     int           `yaml:"data_size" json:"data_size"`
	TTL          time.Duration `yaml:"ttl" json:"ttl"`
	ReadPercent  int           `yaml:"read_percent" json:"read_percent"`
	RandomKeys   int           `yaml:"random_keys" json:"random_keys"`
	TestCase     string        `yaml:"test_case" json:"test_case"`
	Duration     time.Duration `yaml:"duration" json:"duration"`
	PacketRate   int           `yaml:"packet_rate" json:"packet_rate"` // 每秒数据包发送率
}

// UDPSpecificConfig UDP特定配置
type UDPSpecificConfig struct {
	PacketMode      string `yaml:"packet_mode" json:"packet_mode"`         // "unicast", "broadcast", "multicast"
	Broadcast       bool   `yaml:"broadcast" json:"broadcast"`             // 启用广播
	MulticastGroup  string `yaml:"multicast_group" json:"multicast_group"` // 组播地址
	TTL             int    `yaml:"ttl" json:"ttl"`                         // 数据包TTL
	LocalAddress    string `yaml:"local_address" json:"local_address"`     // 本地绑定地址
	ReuseAddress    bool   `yaml:"reuse_address" json:"reuse_address"`     // SO_REUSEADDR
	ChecksumEnabled bool   `yaml:"checksum_enabled" json:"checksum_enabled"` // 启用校验和验证
}

// NewDefaultUDPConfig 创建默认UDP配置
func NewDefaultUDPConfig() *UDPConfig {
	return &UDPConfig{
		Protocol: "udp",
		Connection: ConnectionConfig{
			Address:         "localhost",
			Port:            9090,
			Timeout:         10 * time.Second,
			BufferSize:      64 * 1024, // 64KB
			ReadBufferSize:  64 * 1024,
			WriteBufferSize: 64 * 1024,
		},
		BenchMark: BenchmarkConfig{
			Total:       1000,
			Parallels:   20,
			DataSize:    1024,
			ReadPercent: 50,
			RandomKeys:  1000,
			TestCase:    "packet_send",
			Duration:    60 * time.Second,
			PacketRate:  1000, // 1000 packets per second
		},
		UDPSpecific: UDPSpecificConfig{
			PacketMode:      "unicast",
			Broadcast:       false,
			MulticastGroup:  "",
			TTL:             64,
			LocalAddress:    "",
			ReuseAddress:    true,
			ChecksumEnabled: true,
		},
	}
}

// GetProtocol 实现Config接口
func (c *UDPConfig) GetProtocol() string {
	return c.Protocol
}

// GetConnection 实现Config接口
func (c *UDPConfig) GetConnection() interfaces.ConnectionConfig {
	return &c.Connection
}

// GetBenchmark 实现Config接口
func (c *UDPConfig) GetBenchmark() interfaces.BenchmarkConfig {
	return &c.BenchMark
}

// Validate 实现Config接口
func (c *UDPConfig) Validate() error {
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
	
	if c.BenchMark.DataSize <= 0 || c.BenchMark.DataSize > 65507 { // UDP最大数据长度
		return fmt.Errorf("data size must be between 1 and 65507 bytes")
	}
	
	// 验证测试用例
	validTestCases := []string{"packet_send", "packet_receive", "echo_udp", "multicast"}
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
	
	// 验证数据包模式
	validModes := []string{"unicast", "broadcast", "multicast"}
	valid = false
	for _, mode := range validModes {
		if c.UDPSpecific.PacketMode == mode {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid packet mode: %s, valid options: %s", 
			c.UDPSpecific.PacketMode, strings.Join(validModes, ", "))
	}
	
	// 验证组播地址
	if c.UDPSpecific.PacketMode == "multicast" {
		if c.UDPSpecific.MulticastGroup == "" {
			return fmt.Errorf("multicast group address required for multicast mode")
		}
		if ip := net.ParseIP(c.UDPSpecific.MulticastGroup); ip == nil || !ip.IsMulticast() {
			return fmt.Errorf("invalid multicast address: %s", c.UDPSpecific.MulticastGroup)
		}
	}
	
	// 验证TTL
	if c.UDPSpecific.TTL < 1 || c.UDPSpecific.TTL > 255 {
		return fmt.Errorf("TTL must be between 1 and 255")
	}
	
	return nil
}

// Clone 实现Config接口
func (c *UDPConfig) Clone() interfaces.Config {
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
	return map[string]string{} // UDP通常不需要认证凭据
}

// GetPoolConfig 实现ConnectionConfig接口
func (c *ConnectionConfig) GetPoolConfig() interfaces.PoolConfig {
	// UDP不使用连接池，返回nil实现
	return &EmptyPoolConfig{}
}

// GetTimeout 实现ConnectionConfig接口
func (c *ConnectionConfig) GetTimeout() time.Duration {
	return c.Timeout
}

// EmptyPoolConfig 空的连接池配置（UDP不需要连接池）
type EmptyPoolConfig struct{}

func (p *EmptyPoolConfig) GetPoolSize() int           { return 0 }
func (p *EmptyPoolConfig) GetMinIdle() int            { return 0 }
func (p *EmptyPoolConfig) GetMaxIdle() int            { return 0 }
func (p *EmptyPoolConfig) GetIdleTimeout() time.Duration { return 0 }
func (p *EmptyPoolConfig) GetConnectionTimeout() time.Duration { return 0 }

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