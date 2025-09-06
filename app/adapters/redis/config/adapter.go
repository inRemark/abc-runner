package config

import (
	"fmt"
	"time"

	"redis-runner/app/core/interfaces"
)

// RedisConfigAdapter Redis配置适配器
// 将新的Redis配置接口适配到统一配置接口
type RedisConfigAdapter struct {
	redisConfig *RedisConfig
}

// NewRedisConfigAdapter 创建Redis配置适配器
func NewRedisConfigAdapter(redisConfig *RedisConfig) *RedisConfigAdapter {
	if redisConfig == nil {
		redisConfig = NewDefaultRedisConfig()
	}
	return &RedisConfigAdapter{
		redisConfig: redisConfig,
	}
}

// GetProtocol 获取协议类型
func (r *RedisConfigAdapter) GetProtocol() string {
	return r.redisConfig.GetProtocol()
}

// GetConnection 获取连接配置
func (r *RedisConfigAdapter) GetConnection() interfaces.ConnectionConfig {
	conn := r.redisConfig.GetConnection()
	return &ConnectionConfigAdapter{
		connectionConfig: conn,
	}
}

// GetBenchmark 获取基准测试配置
func (r *RedisConfigAdapter) GetBenchmark() interfaces.BenchmarkConfig {
	benchmark := r.redisConfig.GetBenchmark()
	return &BenchmarkConfigAdapter{
		benchmarkConfig: benchmark,
	}
}

// Validate 验证配置
func (r *RedisConfigAdapter) Validate() error {
	return r.redisConfig.Validate()
}

// Clone 克隆配置
func (r *RedisConfigAdapter) Clone() interfaces.Config {
	cloned := r.redisConfig.Clone()
	return NewRedisConfigAdapter(cloned)
}

// GetRedisConfig 获取原始Redis配置
func (r *RedisConfigAdapter) GetRedisConfig() *RedisConfig {
	return r.redisConfig
}

// ConnectionConfigAdapter 连接配置适配器
type ConnectionConfigAdapter struct {
	connectionConfig *ConnectionConfigImpl
}

// GetAddresses 获取地址列表
func (c *ConnectionConfigAdapter) GetAddresses() []string {
	return c.connectionConfig.GetAddresses()
}

// GetCredentials 获取凭据
func (c *ConnectionConfigAdapter) GetCredentials() map[string]string {
	return c.connectionConfig.GetCredentials()
}

// GetPoolConfig 获取连接池配置
func (c *ConnectionConfigAdapter) GetPoolConfig() interfaces.PoolConfig {
	return &PoolConfigAdapter{
		poolConfig: &c.connectionConfig.Pool,
	}
}

// GetTimeout 获取超时时间
func (c *ConnectionConfigAdapter) GetTimeout() time.Duration {
	return c.connectionConfig.GetTimeout()
}

// PoolConfigAdapter 连接池配置适配器
type PoolConfigAdapter struct {
	poolConfig *PoolConfigImpl
}

// GetPoolSize 获取连接池大小
func (p *PoolConfigAdapter) GetPoolSize() int {
	return p.poolConfig.GetPoolSize()
}

// GetMinIdle 获取最小空闲连接数
func (p *PoolConfigAdapter) GetMinIdle() int {
	return p.poolConfig.GetMinIdle()
}

// GetMaxIdle 获取最大空闲连接数
func (p *PoolConfigAdapter) GetMaxIdle() int {
	return p.poolConfig.GetMaxIdle()
}

// GetIdleTimeout 获取空闲超时时间
func (p *PoolConfigAdapter) GetIdleTimeout() time.Duration {
	return p.poolConfig.GetIdleTimeout()
}

// GetConnectionTimeout 获取连接超时时间
func (p *PoolConfigAdapter) GetConnectionTimeout() time.Duration {
	return p.poolConfig.GetConnectionTimeout()
}

// BenchmarkConfigAdapter 基准测试配置适配器
type BenchmarkConfigAdapter struct {
	benchmarkConfig *BenchmarkConfigImpl
}

// GetTotal 获取总请求数
func (b *BenchmarkConfigAdapter) GetTotal() int {
	return b.benchmarkConfig.GetTotal()
}

// GetParallels 获取并发数
func (b *BenchmarkConfigAdapter) GetParallels() int {
	return b.benchmarkConfig.GetParallels()
}

// GetDataSize 获取数据大小
func (b *BenchmarkConfigAdapter) GetDataSize() int {
	return b.benchmarkConfig.GetDataSize()
}

// GetTTL 获取TTL
func (b *BenchmarkConfigAdapter) GetTTL() time.Duration {
	return b.benchmarkConfig.GetTTL()
}

// GetReadPercent 获取读操作百分比
func (b *BenchmarkConfigAdapter) GetReadPercent() int {
	return b.benchmarkConfig.GetReadPercent()
}

// GetRandomKeys 获取随机键范围
func (b *BenchmarkConfigAdapter) GetRandomKeys() int {
	return b.benchmarkConfig.GetRandomKeys()
}

// GetTestCase 获取测试用例
func (b *BenchmarkConfigAdapter) GetTestCase() string {
	return b.benchmarkConfig.GetTestCase()
}

// AdaptRedisConfig 将Redis配置适配为统一配置接口
func AdaptRedisConfig(redisConfig *RedisConfig) interfaces.Config {
	return NewRedisConfigAdapter(redisConfig)
}

// ExtractRedisConfig 从统一配置接口提取Redis配置
func ExtractRedisConfig(config interfaces.Config) (*RedisConfig, error) {
	if adapter, ok := config.(*RedisConfigAdapter); ok {
		return adapter.GetRedisConfig(), nil
	}
	
	// 如果不是适配器，尝试转换为Redis配置
	// 这里可以根据接口信息重新构建Redis配置
	redisConfig := NewDefaultRedisConfig()
	redisConfig.Protocol = config.GetProtocol()
	
	conn := config.GetConnection()
	addresses := conn.GetAddresses()
	credentials := conn.GetCredentials()
	
	// 根据地址和凭据判断模式
	if len(addresses) > 1 {
		redisConfig.Mode = "cluster"
		redisConfig.Cluster.Addrs = addresses
		if password, exists := credentials["password"]; exists {
			redisConfig.Cluster.Password = password
		}
	} else if masterName, exists := credentials["master_name"]; exists {
		redisConfig.Mode = "sentinel"
		redisConfig.Sentinel.Addrs = addresses
		redisConfig.Sentinel.MasterName = masterName
		if password, exists := credentials["password"]; exists {
			redisConfig.Sentinel.Password = password
		}
		if db, exists := credentials["db"]; exists {
			if dbInt, err := fmt.Sscanf(db, "%d", &redisConfig.Sentinel.Db); err == nil && dbInt == 1 {
				// DB已设置
			}
		}
	} else {
		redisConfig.Mode = "standalone"
		if len(addresses) > 0 {
			redisConfig.Standalone.Addr = addresses[0]
		}
		if password, exists := credentials["password"]; exists {
			redisConfig.Standalone.Password = password
		}
		if db, exists := credentials["db"]; exists {
			if dbInt, err := fmt.Sscanf(db, "%d", &redisConfig.Standalone.Db); err == nil && dbInt == 1 {
				// DB已设置
			}
		}
	}
	
	// 设置基准测试配置
	benchmark := config.GetBenchmark()
	redisConfig.BenchMark.Total = benchmark.GetTotal()
	redisConfig.BenchMark.Parallels = benchmark.GetParallels()
	redisConfig.BenchMark.DataSize = benchmark.GetDataSize()
	redisConfig.BenchMark.TTL = int(benchmark.GetTTL().Seconds())
	redisConfig.BenchMark.ReadPercent = benchmark.GetReadPercent()
	redisConfig.BenchMark.RandomKeys = benchmark.GetRandomKeys()
	redisConfig.BenchMark.Case = benchmark.GetTestCase()
	
	// 设置连接池配置
	poolConfig := conn.GetPoolConfig()
	redisConfig.Pool.PoolSize = poolConfig.GetPoolSize()
	redisConfig.Pool.MinIdle = poolConfig.GetMinIdle()
	redisConfig.Pool.MaxIdle = poolConfig.GetMaxIdle()
	redisConfig.Pool.IdleTimeout = poolConfig.GetIdleTimeout()
	redisConfig.Pool.ConnectionTimeout = poolConfig.GetConnectionTimeout()
	
	return redisConfig, nil
}