package config

import (
	"fmt"
	"time"

	"redis-runner/app/core/interfaces"
)

// UnifiedConfigManager 统一配置管理器
type UnifiedConfigManager interface {
	// LoadConfig 加载配置
	LoadConfig(source ConfigSource, protocol string) (interfaces.Config, error)
	
	// MergeConfigs 合并多个配置
	MergeConfigs(configs ...interfaces.Config) (interfaces.Config, error)
	
	// ValidateConfig 验证配置
	ValidateConfig(config interfaces.Config) error
	
	// GetDefaultConfig 获取默认配置
	GetDefaultConfig(protocol string) (interfaces.Config, error)
	
	// SaveConfig 保存配置
	SaveConfig(config interfaces.Config, destination string) error
	
	// RegisterConfigLoader 注册配置加载器
	RegisterConfigLoader(protocol string, loader ConfigLoader) error
	
	// GetProtocolConfig 获取特定协议的配置
	GetProtocolConfig(protocol string, source ConfigSource) (interfaces.Config, error)
}

// ConfigSource 配置源类型
type ConfigSource interface {
	// GetType 获取配置源类型
	GetType() ConfigSourceType
	
	// GetPath 获取配置路径
	GetPath() string
	
	// GetData 获取配置数据
	GetData() (map[string]interface{}, error)
	
	// Validate 验证配置源
	Validate() error
}

// ConfigSourceType 配置源类型
type ConfigSourceType string

const (
	FileSource      ConfigSourceType = "file"
	EnvironmentSource ConfigSourceType = "environment"
	CommandLineSource ConfigSourceType = "command_line"
	DefaultSource     ConfigSourceType = "default"
	DatabaseSource    ConfigSourceType = "database"
)

// ConfigLoader 配置加载器接口
type ConfigLoader interface {
	// LoadFromSource 从源加载配置
	LoadFromSource(source ConfigSource) (interfaces.Config, error)
	
	// GetSupportedSources 获取支持的配置源类型
	GetSupportedSources() []ConfigSourceType
	
	// ValidateConfig 验证配置
	ValidateConfig(config interfaces.Config) error
	
	// GetDefaultConfig 获取默认配置
	GetDefaultConfig() interfaces.Config
}

// UnifiedConfig 统一配置实现
type UnifiedConfig struct {
	Protocol     string                     `yaml:"protocol" json:"protocol"`
	Connection   *UnifiedConnectionConfig   `yaml:"connection" json:"connection"`
	Benchmark    *UnifiedBenchmarkConfig    `yaml:"benchmark" json:"benchmark"`
	Global       *GlobalConfig              `yaml:"global" json:"global"`
	Metadata     map[string]interface{}     `yaml:"metadata" json:"metadata"`
}

// UnifiedConnectionConfig 统一连接配置
type UnifiedConnectionConfig struct {
	Addresses      []string                `yaml:"addresses" json:"addresses"`
	Credentials    map[string]string       `yaml:"credentials" json:"credentials"`
	Pool           *UnifiedPoolConfig      `yaml:"pool" json:"pool"`
	Timeout        time.Duration           `yaml:"timeout" json:"timeout"`
	TLS            *TLSConfig              `yaml:"tls" json:"tls"`
	Proxy          *ProxyConfig            `yaml:"proxy" json:"proxy"`
	HealthCheck    *HealthCheckConfig      `yaml:"health_check" json:"health_check"`
}

// UnifiedBenchmarkConfig 统一基准测试配置
type UnifiedBenchmarkConfig struct {
	Total        int           `yaml:"total" json:"total"`
	Parallels    int           `yaml:"parallels" json:"parallels"`
	DataSize     int           `yaml:"data_size" json:"data_size"`
	TTL          time.Duration `yaml:"ttl" json:"ttl"`
	ReadPercent  int           `yaml:"read_percent" json:"read_percent"`
	RandomKeys   int           `yaml:"random_keys" json:"random_keys"`
	TestCase     string        `yaml:"test_case" json:"test_case"`
	Duration     time.Duration `yaml:"duration" json:"duration"`
	RateLimit    int           `yaml:"rate_limit" json:"rate_limit"`
	WarmupTime   time.Duration `yaml:"warmup_time" json:"warmup_time"`
}

// UnifiedPoolConfig 统一连接池配置
type UnifiedPoolConfig struct {
	PoolSize          int           `yaml:"pool_size" json:"pool_size"`
	MinIdle           int           `yaml:"min_idle" json:"min_idle"`
	MaxIdle           int           `yaml:"max_idle" json:"max_idle"`
	IdleTimeout       time.Duration `yaml:"idle_timeout" json:"idle_timeout"`
	ConnectionTimeout time.Duration `yaml:"connection_timeout" json:"connection_timeout"`
	MaxLifetime       time.Duration `yaml:"max_lifetime" json:"max_lifetime"`
	RetryInterval     time.Duration `yaml:"retry_interval" json:"retry_interval"`
	MaxRetries        int           `yaml:"max_retries" json:"max_retries"`
}

// GlobalConfig 全局配置
type GlobalConfig struct {
	LogLevel       string            `yaml:"log_level" json:"log_level"`
	OutputFormat   string            `yaml:"output_format" json:"output_format"`
	DefaultProtocol string           `yaml:"default_protocol" json:"default_protocol"`
	Aliases        map[string]string `yaml:"aliases" json:"aliases"`
	MetricsEnabled bool              `yaml:"metrics_enabled" json:"metrics_enabled"`
	Debug          bool              `yaml:"debug" json:"debug"`
}

// TLSConfig TLS配置
type TLSConfig struct {
	Enabled            bool   `yaml:"enabled" json:"enabled"`
	CertFile           string `yaml:"cert_file" json:"cert_file"`
	KeyFile            string `yaml:"key_file" json:"key_file"`
	CAFile             string `yaml:"ca_file" json:"ca_file"`
	InsecureSkipVerify bool   `yaml:"insecure_skip_verify" json:"insecure_skip_verify"`
}

// ProxyConfig 代理配置
type ProxyConfig struct {
	Enabled  bool   `yaml:"enabled" json:"enabled"`
	Type     string `yaml:"type" json:"type"`
	Address  string `yaml:"address" json:"address"`
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
}

// HealthCheckConfig 健康检查配置
type HealthCheckConfig struct {
	Enabled  bool          `yaml:"enabled" json:"enabled"`
	Interval time.Duration `yaml:"interval" json:"interval"`
	Timeout  time.Duration `yaml:"timeout" json:"timeout"`
	Retries  int           `yaml:"retries" json:"retries"`
}

// 实现 interfaces.Config 接口
func (c *UnifiedConfig) GetProtocol() string {
	return c.Protocol
}

func (c *UnifiedConfig) GetConnection() interfaces.ConnectionConfig {
	return c.Connection
}

func (c *UnifiedConfig) GetBenchmark() interfaces.BenchmarkConfig {
	return c.Benchmark
}

func (c *UnifiedConfig) Validate() error {
	if c.Protocol == "" {
		return fmt.Errorf("protocol cannot be empty")
	}
	
	if c.Connection == nil {
		return fmt.Errorf("connection config is required")
	}
	
	if err := c.Connection.Validate(); err != nil {
		return fmt.Errorf("connection config validation failed: %w", err)
	}
	
	if c.Benchmark == nil {
		return fmt.Errorf("benchmark config is required")
	}
	
	if err := c.Benchmark.Validate(); err != nil {
		return fmt.Errorf("benchmark config validation failed: %w", err)
	}
	
	return nil
}

func (c *UnifiedConfig) Clone() interfaces.Config {
	clone := &UnifiedConfig{
		Protocol: c.Protocol,
		Metadata: make(map[string]interface{}),
	}
	
	// 深拷贝连接配置
	if c.Connection != nil {
		clone.Connection = c.Connection.Clone()
	}
	
	// 深拷贝基准测试配置
	if c.Benchmark != nil {
		clone.Benchmark = c.Benchmark.Clone()
	}
	
	// 深拷贝全局配置
	if c.Global != nil {
		clone.Global = c.Global.Clone()
	}
	
	// 深拷贝元数据
	for k, v := range c.Metadata {
		clone.Metadata[k] = v
	}
	
	return clone
}

// 实现 interfaces.ConnectionConfig 接口
func (c *UnifiedConnectionConfig) GetAddresses() []string {
	return c.Addresses
}

func (c *UnifiedConnectionConfig) GetCredentials() map[string]string {
	return c.Credentials
}

func (c *UnifiedConnectionConfig) GetPoolConfig() interfaces.PoolConfig {
	return c.Pool
}

func (c *UnifiedConnectionConfig) GetTimeout() time.Duration {
	return c.Timeout
}

func (c *UnifiedConnectionConfig) Validate() error {
	if len(c.Addresses) == 0 {
		return fmt.Errorf("at least one address is required")
	}
	
	for _, addr := range c.Addresses {
		if addr == "" {
			return fmt.Errorf("empty address not allowed")
		}
	}
	
	if c.Pool != nil {
		if err := c.Pool.Validate(); err != nil {
			return fmt.Errorf("pool config validation failed: %w", err)
		}
	}
	
	return nil
}

func (c *UnifiedConnectionConfig) Clone() *UnifiedConnectionConfig {
	clone := &UnifiedConnectionConfig{
		Addresses:   append([]string{}, c.Addresses...),
		Credentials: make(map[string]string),
		Timeout:     c.Timeout,
	}
	
	for k, v := range c.Credentials {
		clone.Credentials[k] = v
	}
	
	if c.Pool != nil {
		clone.Pool = c.Pool.Clone()
	}
	
	if c.TLS != nil {
		clone.TLS = &TLSConfig{
			Enabled:            c.TLS.Enabled,
			CertFile:           c.TLS.CertFile,
			KeyFile:            c.TLS.KeyFile,
			CAFile:             c.TLS.CAFile,
			InsecureSkipVerify: c.TLS.InsecureSkipVerify,
		}
	}
	
	if c.Proxy != nil {
		clone.Proxy = &ProxyConfig{
			Enabled:  c.Proxy.Enabled,
			Type:     c.Proxy.Type,
			Address:  c.Proxy.Address,
			Username: c.Proxy.Username,
			Password: c.Proxy.Password,
		}
	}
	
	if c.HealthCheck != nil {
		clone.HealthCheck = &HealthCheckConfig{
			Enabled:  c.HealthCheck.Enabled,
			Interval: c.HealthCheck.Interval,
			Timeout:  c.HealthCheck.Timeout,
			Retries:  c.HealthCheck.Retries,
		}
	}
	
	return clone
}

// 实现 interfaces.BenchmarkConfig 接口
func (c *UnifiedBenchmarkConfig) GetTotal() int {
	return c.Total
}

func (c *UnifiedBenchmarkConfig) GetParallels() int {
	return c.Parallels
}

func (c *UnifiedBenchmarkConfig) GetDataSize() int {
	return c.DataSize
}

func (c *UnifiedBenchmarkConfig) GetTTL() time.Duration {
	return c.TTL
}

func (c *UnifiedBenchmarkConfig) GetReadPercent() int {
	return c.ReadPercent
}

func (c *UnifiedBenchmarkConfig) GetRandomKeys() int {
	return c.RandomKeys
}

func (c *UnifiedBenchmarkConfig) GetTestCase() string {
	return c.TestCase
}

func (c *UnifiedBenchmarkConfig) Validate() error {
	if c.Total <= 0 {
		return fmt.Errorf("total must be greater than 0")
	}
	
	if c.Parallels <= 0 {
		return fmt.Errorf("parallels must be greater than 0")
	}
	
	if c.DataSize < 0 {
		return fmt.Errorf("data_size cannot be negative")
	}
	
	if c.ReadPercent < 0 || c.ReadPercent > 100 {
		return fmt.Errorf("read_percent must be between 0 and 100")
	}
	
	if c.RandomKeys < 0 {
		return fmt.Errorf("random_keys cannot be negative")
	}
	
	return nil
}

func (c *UnifiedBenchmarkConfig) Clone() *UnifiedBenchmarkConfig {
	return &UnifiedBenchmarkConfig{
		Total:       c.Total,
		Parallels:   c.Parallels,
		DataSize:    c.DataSize,
		TTL:         c.TTL,
		ReadPercent: c.ReadPercent,
		RandomKeys:  c.RandomKeys,
		TestCase:    c.TestCase,
		Duration:    c.Duration,
		RateLimit:   c.RateLimit,
		WarmupTime:  c.WarmupTime,
	}
}

// 实现 interfaces.PoolConfig 接口
func (c *UnifiedPoolConfig) GetPoolSize() int {
	return c.PoolSize
}

func (c *UnifiedPoolConfig) GetMinIdle() int {
	return c.MinIdle
}

func (c *UnifiedPoolConfig) GetMaxIdle() int {
	return c.MaxIdle
}

func (c *UnifiedPoolConfig) GetIdleTimeout() time.Duration {
	return c.IdleTimeout
}

func (c *UnifiedPoolConfig) GetConnectionTimeout() time.Duration {
	return c.ConnectionTimeout
}

func (c *UnifiedPoolConfig) Validate() error {
	if c.PoolSize <= 0 {
		return fmt.Errorf("pool_size must be greater than 0")
	}
	
	if c.MinIdle < 0 {
		return fmt.Errorf("min_idle cannot be negative")
	}
	
	if c.MaxIdle < c.MinIdle {
		return fmt.Errorf("max_idle cannot be less than min_idle")
	}
	
	if c.MaxIdle > c.PoolSize {
		return fmt.Errorf("max_idle cannot be greater than pool_size")
	}
	
	return nil
}

func (c *UnifiedPoolConfig) Clone() *UnifiedPoolConfig {
	return &UnifiedPoolConfig{
		PoolSize:          c.PoolSize,
		MinIdle:           c.MinIdle,
		MaxIdle:           c.MaxIdle,
		IdleTimeout:       c.IdleTimeout,
		ConnectionTimeout: c.ConnectionTimeout,
		MaxLifetime:       c.MaxLifetime,
		RetryInterval:     c.RetryInterval,
		MaxRetries:        c.MaxRetries,
	}
}

// Clone 克隆全局配置
func (c *GlobalConfig) Clone() *GlobalConfig {
	clone := &GlobalConfig{
		LogLevel:        c.LogLevel,
		OutputFormat:    c.OutputFormat,
		DefaultProtocol: c.DefaultProtocol,
		MetricsEnabled:  c.MetricsEnabled,
		Debug:           c.Debug,
		Aliases:         make(map[string]string),
	}
	
	for k, v := range c.Aliases {
		clone.Aliases[k] = v
	}
	
	return clone
}