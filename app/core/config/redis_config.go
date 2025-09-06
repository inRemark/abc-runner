package config

import (
	"fmt"
	"os"
	"time"

	"redis-runner/app/core/interfaces"
	"gopkg.in/yaml.v2"
)

// RedisConfig Redis配置实现
type RedisConfig struct {
	Protocol   string              `yaml:"protocol"`
	Mode       string              `yaml:"mode"`
	BenchMark  BenchmarkConfigImpl `yaml:"benchmark"`
	Pool       PoolConfigImpl      `yaml:"pool"`
	Standalone StandAloneInfo      `yaml:"standalone"`
	Sentinel   SentinelInfo        `yaml:"sentinel"`
	Cluster    ClusterInfo         `yaml:"cluster"`
}

// StandAloneInfo 单机配置
type StandAloneInfo struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	Db       int    `yaml:"db"`
}

// SentinelInfo 哨兵配置
type SentinelInfo struct {
	MasterName string   `yaml:"master_name"`
	Addrs      []string `yaml:"addrs"`
	Password   string   `yaml:"password"`
	Db         int      `yaml:"db"`
}

// ClusterInfo 集群配置
type ClusterInfo struct {
	Addrs    []string `yaml:"addrs"`
	Password string   `yaml:"password"`
}

// PoolConfigImpl 连接池配置实现
type PoolConfigImpl struct {
	PoolSize          int           `yaml:"pool_size"`
	MinIdle           int           `yaml:"min_idle"`
	MaxIdle           int           `yaml:"max_idle"`
	IdleTimeout       time.Duration `yaml:"idle_timeout"`
	ConnectionTimeout time.Duration `yaml:"connection_timeout"`
}

// BenchmarkConfigImpl 基准测试配置实现
type BenchmarkConfigImpl struct {
	DataSize    int    `yaml:"data_size"`
	Parallels   int    `yaml:"parallels"`
	Total       int    `yaml:"total"`
	TTL         int    `yaml:"ttl"`
	ReadPercent int    `yaml:"read_percent"`
	RandomKeys  int    `yaml:"random_keys"`
	Case        string `yaml:"case"`
}

// ConnectionConfigImpl 连接配置实现
type ConnectionConfigImpl struct {
	Addresses   []string          `json:"addresses"`
	Credentials map[string]string `json:"credentials"`
	Pool        PoolConfigImpl    `json:"pool"`
	Timeout     time.Duration     `json:"timeout"`
}

// GetProtocol 获取协议类型
func (c *RedisConfig) GetProtocol() string {
	if c.Protocol == "" {
		return "redis"
	}
	return c.Protocol
}

// GetConnection 获取连接配置
func (c *RedisConfig) GetConnection() interfaces.ConnectionConfig {
	conn := &ConnectionConfigImpl{
		Pool:    c.Pool,
		Timeout: 30 * time.Second,
	}
	
	credentials := make(map[string]string)
	
	switch c.Mode {
	case "cluster":
		conn.Addresses = c.Cluster.Addrs
		if c.Cluster.Password != "" {
			credentials["password"] = c.Cluster.Password
		}
	case "sentinel":
		conn.Addresses = c.Sentinel.Addrs
		if c.Sentinel.Password != "" {
			credentials["password"] = c.Sentinel.Password
		}
		credentials["master_name"] = c.Sentinel.MasterName
		credentials["db"] = fmt.Sprintf("%d", c.Sentinel.Db)
	default: // standalone
		conn.Addresses = []string{c.Standalone.Addr}
		if c.Standalone.Password != "" {
			credentials["password"] = c.Standalone.Password
		}
		credentials["db"] = fmt.Sprintf("%d", c.Standalone.Db)
	}
	
	conn.Credentials = credentials
	return conn
}

// GetBenchmark 获取基准测试配置
func (c *RedisConfig) GetBenchmark() interfaces.BenchmarkConfig {
	return &c.BenchMark
}

// Validate 验证配置
func (c *RedisConfig) Validate() error {
	if c.Mode == "" {
		return fmt.Errorf("mode cannot be empty")
	}
	
	switch c.Mode {
	case "standalone":
		if c.Standalone.Addr == "" {
			return fmt.Errorf("standalone addr cannot be empty")
		}
	case "sentinel":
		if len(c.Sentinel.Addrs) == 0 {
			return fmt.Errorf("sentinel addrs cannot be empty")
		}
		if c.Sentinel.MasterName == "" {
			return fmt.Errorf("sentinel master_name cannot be empty")
		}
	case "cluster":
		if len(c.Cluster.Addrs) == 0 {
			return fmt.Errorf("cluster addrs cannot be empty")
		}
	default:
		return fmt.Errorf("unsupported mode: %s", c.Mode)
	}
	
	return c.BenchMark.Validate()
}

// Clone 克隆配置
func (c *RedisConfig) Clone() interfaces.Config {
	cloned := *c
	
	// 深拷贝切片
	if len(c.Sentinel.Addrs) > 0 {
		cloned.Sentinel.Addrs = make([]string, len(c.Sentinel.Addrs))
		copy(cloned.Sentinel.Addrs, c.Sentinel.Addrs)
	}
	
	if len(c.Cluster.Addrs) > 0 {
		cloned.Cluster.Addrs = make([]string, len(c.Cluster.Addrs))
		copy(cloned.Cluster.Addrs, c.Cluster.Addrs)
	}
	
	return &cloned
}

// ConnectionConfig 接口实现

// GetAddresses 获取地址列表
func (c *ConnectionConfigImpl) GetAddresses() []string {
	return c.Addresses
}

// GetCredentials 获取凭据
func (c *ConnectionConfigImpl) GetCredentials() map[string]string {
	return c.Credentials
}

// GetPoolConfig 获取连接池配置
func (c *ConnectionConfigImpl) GetPoolConfig() interfaces.PoolConfig {
	return &c.Pool
}

// GetTimeout 获取超时时间
func (c *ConnectionConfigImpl) GetTimeout() time.Duration {
	return c.Timeout
}

// PoolConfig 接口实现

// GetPoolSize 获取连接池大小
func (p *PoolConfigImpl) GetPoolSize() int {
	if p.PoolSize <= 0 {
		return 10
	}
	return p.PoolSize
}

// GetMinIdle 获取最小空闲连接数
func (p *PoolConfigImpl) GetMinIdle() int {
	if p.MinIdle <= 0 {
		return 2
	}
	return p.MinIdle
}

// GetMaxIdle 获取最大空闲连接数
func (p *PoolConfigImpl) GetMaxIdle() int {
	if p.MaxIdle <= 0 {
		return p.GetPoolSize()
	}
	return p.MaxIdle
}

// GetIdleTimeout 获取空闲超时时间
func (p *PoolConfigImpl) GetIdleTimeout() time.Duration {
	if p.IdleTimeout <= 0 {
		return 5 * time.Minute
	}
	return p.IdleTimeout
}

// GetConnectionTimeout 获取连接超时时间
func (p *PoolConfigImpl) GetConnectionTimeout() time.Duration {
	if p.ConnectionTimeout <= 0 {
		return 30 * time.Second
	}
	return p.ConnectionTimeout
}

// BenchmarkConfig 接口实现

// GetTotal 获取总请求数
func (b *BenchmarkConfigImpl) GetTotal() int {
	if b.Total <= 0 {
		return 100000
	}
	return b.Total
}

// GetParallels 获取并发数
func (b *BenchmarkConfigImpl) GetParallels() int {
	if b.Parallels <= 0 {
		return 50
	}
	return b.Parallels
}

// GetDataSize 获取数据大小
func (b *BenchmarkConfigImpl) GetDataSize() int {
	if b.DataSize <= 0 {
		return 3
	}
	return b.DataSize
}

// GetTTL 获取TTL
func (b *BenchmarkConfigImpl) GetTTL() time.Duration {
	if b.TTL <= 0 {
		return 120 * time.Second
	}
	return time.Duration(b.TTL) * time.Second
}

// GetReadPercent 获取读操作百分比
func (b *BenchmarkConfigImpl) GetReadPercent() int {
	if b.ReadPercent < 0 || b.ReadPercent > 100 {
		return 50
	}
	return b.ReadPercent
}

// GetRandomKeys 获取随机键范围
func (b *BenchmarkConfigImpl) GetRandomKeys() int {
	return b.RandomKeys
}

// GetTestCase 获取测试用例
func (b *BenchmarkConfigImpl) GetTestCase() string {
	if b.Case == "" {
		return "get"
	}
	return b.Case
}

// Validate 验证基准测试配置
func (b *BenchmarkConfigImpl) Validate() error {
	if b.Total <= 0 {
		return fmt.Errorf("total must be positive")
	}
	
	if b.Parallels <= 0 {
		return fmt.Errorf("parallels must be positive")
	}
	
	if b.DataSize <= 0 {
		return fmt.Errorf("data_size must be positive")
	}
	
	if b.ReadPercent < 0 || b.ReadPercent > 100 {
		return fmt.Errorf("read_percent must be between 0 and 100")
	}
	
	return nil
}

// ConfigLoader 配置加载器
type ConfigLoader struct {
	sources []ConfigSource
}

// 使用config_sources.go中定义的ConfigSource接口

// YAMLConfigSource YAML文件配置源
type YAMLConfigSource struct {
	FilePath string
}

// NewYAMLConfigSource 创建YAML配置源
func NewYAMLConfigSource(filePath string) *YAMLConfigSource {
	return &YAMLConfigSource{FilePath: filePath}
}

// Load 加载配置
func (y *YAMLConfigSource) Load() (interfaces.Config, error) {
	data, err := os.ReadFile(y.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", y.FilePath, err)
	}
	
	var configWrapper struct {
		Redis RedisConfig `yaml:"redis"`
	}
	
	err = yaml.Unmarshal(data, &configWrapper)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", y.FilePath, err)
	}
	
	return &configWrapper.Redis, nil
}

// CanLoad 检查是否可以加载
func (y *YAMLConfigSource) CanLoad() bool {
	_, err := os.Stat(y.FilePath)
	return err == nil
}

// Priority 获取优先级
func (y *YAMLConfigSource) Priority() int {
	return 1
}

// NewMultiSourceConfigLoader 创建多源配置加载器
func NewMultiSourceConfigLoader(sources ...ConfigSource) *ConfigLoader {
	return &ConfigLoader{sources: sources}
}

// Load 从多个源加载配置
func (c *ConfigLoader) Load() (interfaces.Config, error) {
	for _, source := range c.sources {
		if source.CanLoad() {
			config, err := source.Load()
			if err != nil {
				continue
			}
			return config, nil
		}
	}
	
	return nil, fmt.Errorf("no configuration source available")
}