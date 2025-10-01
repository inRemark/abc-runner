package config

import (
	"fmt"
	"strings"
	"time"

	"abc-runner/app/core/execution"
	"abc-runner/app/core/interfaces"
)

// GRPCConfig gRPC协议配置
type GRPCConfig struct {
	Protocol     string             `yaml:"protocol" json:"protocol"`
	Connection   ConnectionConfig   `yaml:"connection" json:"connection"`
	BenchMark    BenchmarkConfig    `yaml:"benchmark" json:"benchmark"`
	GRPCSpecific GRPCSpecificConfig `yaml:"grpc_specific" json:"grpc_specific"`
}

// ConnectionConfig gRPC连接配置
type ConnectionConfig struct {
	Address         string        `yaml:"address" json:"address"`
	Port            int           `yaml:"port" json:"port"`
	Timeout         time.Duration `yaml:"timeout" json:"timeout"`
	KeepAlive       bool          `yaml:"keep_alive" json:"keep_alive"`
	KeepAlivePeriod time.Duration `yaml:"keep_alive_period" json:"keep_alive_period"`
	Pool            PoolConfig    `yaml:"pool" json:"pool"`
}

// PoolConfig gRPC连接池配置
type PoolConfig struct {
	PoolSize          int           `yaml:"pool_size" json:"pool_size"`
	MinIdle           int           `yaml:"min_idle" json:"min_idle"`
	MaxIdle           int           `yaml:"max_idle" json:"max_idle"`
	IdleTimeout       time.Duration `yaml:"idle_timeout" json:"idle_timeout"`
	ConnectionTimeout time.Duration `yaml:"connection_timeout" json:"connection_timeout"`
}

// BenchmarkConfig gRPC基准测试配置
type BenchmarkConfig struct {
	Total       int           `yaml:"total" json:"total"`
	Parallels   int           `yaml:"parallels" json:"parallels"`
	DataSize    int           `yaml:"data_size" json:"data_size"`
	TTL         time.Duration `yaml:"ttl" json:"ttl"`
	ReadPercent int           `yaml:"read_percent" json:"read_percent"`
	RandomKeys  int           `yaml:"random_keys" json:"random_keys"`
	TestCase    string        `yaml:"test_case" json:"test_case"`
	Duration    time.Duration `yaml:"duration" json:"duration"`
	Timeout     time.Duration `yaml:"timeout" json:"timeout"`
	RampUp      time.Duration `yaml:"ramp_up" json:"ramp_up"`
}

// GRPCSpecificConfig gRPC特定配置
type GRPCSpecificConfig struct {
	ServiceName    string            `yaml:"service_name" json:"service_name"`         // gRPC服务名
	MethodName     string            `yaml:"method_name" json:"method_name"`           // gRPC方法名
	LoadBalancing  string            `yaml:"load_balancing" json:"load_balancing"`     // 负载均衡策略
	TLS            TLSConfig         `yaml:"tls" json:"tls"`                           // TLS配置
	Auth           AuthConfig        `yaml:"auth" json:"auth"`                         // 认证配置
	Compression    string            `yaml:"compression" json:"compression"`           // 压缩算法
	MaxMessageSize int               `yaml:"max_message_size" json:"max_message_size"` // 最大消息大小
	Interceptors   InterceptorConfig `yaml:"interceptors" json:"interceptors"`         // 拦截器配置
}

// TLSConfig TLS配置
type TLSConfig struct {
	Enabled            bool   `yaml:"enabled" json:"enabled"`
	CertFile           string `yaml:"cert_file" json:"cert_file"`
	KeyFile            string `yaml:"key_file" json:"key_file"`
	CAFile             string `yaml:"ca_file" json:"ca_file"`
	InsecureSkipVerify bool   `yaml:"insecure_skip_verify" json:"insecure_skip_verify"`
	ServerName         string `yaml:"server_name" json:"server_name"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	Enabled  bool              `yaml:"enabled" json:"enabled"`
	Method   string            `yaml:"method" json:"method"` // "token", "oauth", "basic"
	Token    string            `yaml:"token" json:"token"`
	Username string            `yaml:"username" json:"username"`
	Password string            `yaml:"password" json:"password"`
	Metadata map[string]string `yaml:"metadata" json:"metadata"`
}

// InterceptorConfig 拦截器配置
type InterceptorConfig struct {
	Logging bool `yaml:"logging" json:"logging"`
	Tracing bool `yaml:"tracing" json:"tracing"`
	Metrics bool `yaml:"metrics" json:"metrics"`
	Retry   bool `yaml:"retry" json:"retry"`
	Timeout bool `yaml:"timeout" json:"timeout"`
}

// NewDefaultGRPCConfig 创建默认gRPC配置
func NewDefaultGRPCConfig() *GRPCConfig {
	return &GRPCConfig{
		Protocol: "grpc",
		Connection: ConnectionConfig{
			Address:         "localhost",
			Port:            50051,
			Timeout:         30 * time.Second,
			KeepAlive:       true,
			KeepAlivePeriod: 30 * time.Second,
			Pool: PoolConfig{
				PoolSize:          10,
				MinIdle:           1,
				MaxIdle:           5,
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
			TestCase:    "unary_call",
			Duration:    60 * time.Second,
			Timeout:     30 * time.Second,
			RampUp:      5 * time.Second,
		},
		GRPCSpecific: GRPCSpecificConfig{
			ServiceName:    "TestService",
			MethodName:     "Echo",
			LoadBalancing:  "round_robin",
			Compression:    "gzip",
			MaxMessageSize: 4 * 1024 * 1024, // 4MB
			TLS: TLSConfig{
				Enabled:            false,
				InsecureSkipVerify: true,
			},
			Auth: AuthConfig{
				Enabled: false,
				Method:  "token",
			},
			Interceptors: InterceptorConfig{
				Logging: true,
				Metrics: true,
				Retry:   false,
				Timeout: true,
			},
		},
	}
}

// GetProtocol 实现Config接口
func (c *GRPCConfig) GetProtocol() string {
	return c.Protocol
}

// GetConnection 实现Config接口
func (c *GRPCConfig) GetConnection() interfaces.ConnectionConfig {
	return &c.Connection
}

// GetBenchmark 实现Config接口
func (c *GRPCConfig) GetBenchmark() interfaces.BenchmarkConfig {
	return &c.BenchMark
}

// Validate 实现Config接口
func (c *GRPCConfig) Validate() error {
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
	validTestCases := []string{"unary_call", "server_stream", "client_stream", "bidirectional_stream"}
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

	// 验证服务和方法名
	if c.GRPCSpecific.ServiceName == "" {
		return fmt.Errorf("service name cannot be empty")
	}

	if c.GRPCSpecific.MethodName == "" {
		return fmt.Errorf("method name cannot be empty")
	}

	// 验证ExecutionEngine相关配置
	if c.BenchMark.Timeout <= 0 {
		return fmt.Errorf("timeout must be greater than 0")
	}

	if c.BenchMark.RampUp < 0 {
		return fmt.Errorf("ramp up duration cannot be negative")
	}

	// 验证负载均衡策略
	validStrategies := []string{"round_robin", "pick_first", "random"}
	valid = false
	for _, strategy := range validStrategies {
		if c.GRPCSpecific.LoadBalancing == strategy {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid load balancing strategy: %s, valid options: %s",
			c.GRPCSpecific.LoadBalancing, strings.Join(validStrategies, ", "))
	}

	return nil
}

// Clone 实现Config接口
func (c *GRPCConfig) Clone() interfaces.Config {
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
	return map[string]string{} // gRPC认证通过Auth配置处理
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

// GetDuration 实现execution.BenchmarkConfig接口
func (b *BenchmarkConfig) GetDuration() time.Duration {
	return b.Duration
}

// GetTimeout 实现execution.BenchmarkConfig接口
func (b *BenchmarkConfig) GetTimeout() time.Duration {
	return b.Timeout
}

// GetRampUp 实现execution.BenchmarkConfig接口
func (b *BenchmarkConfig) GetRampUp() time.Duration {
	return b.RampUp
}

// 确保BenchmarkConfig实现了execution.BenchmarkConfig接口
var _ execution.BenchmarkConfig = (*BenchmarkConfig)(nil)
