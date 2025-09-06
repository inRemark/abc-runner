package config

import (
	"time"

	"redis-runner/app/core/interfaces"
)

// ConnectionConfigImpl 连接配置实现
type ConnectionConfigImpl struct {
	addresses   []string
	credentials map[string]string
	poolConfig  interfaces.PoolConfig
	timeout     time.Duration
}

// GetAddresses 获取地址列表
func (c *ConnectionConfigImpl) GetAddresses() []string {
	return c.addresses
}

// GetCredentials 获取认证信息
func (c *ConnectionConfigImpl) GetCredentials() map[string]string {
	return c.credentials
}

// GetPoolConfig 获取连接池配置
func (c *ConnectionConfigImpl) GetPoolConfig() interfaces.PoolConfig {
	return c.poolConfig
}

// GetTimeout 获取超时时间
func (c *ConnectionConfigImpl) GetTimeout() time.Duration {
	return c.timeout
}

// PoolConfigImpl 连接池配置实现
type PoolConfigImpl struct {
	poolSize          int
	minIdle           int
	maxIdle           int
	idleTimeout       time.Duration
	connectionTimeout time.Duration
}

// GetPoolSize 获取连接池大小
func (p *PoolConfigImpl) GetPoolSize() int {
	return p.poolSize
}

// GetMinIdle 获取最小空闲连接数
func (p *PoolConfigImpl) GetMinIdle() int {
	return p.minIdle
}

// GetMaxIdle 获取最大空闲连接数
func (p *PoolConfigImpl) GetMaxIdle() int {
	return p.maxIdle
}

// GetIdleTimeout 获取空闲超时时间
func (p *PoolConfigImpl) GetIdleTimeout() time.Duration {
	return p.idleTimeout
}

// GetConnectionTimeout 获取连接超时时间
func (p *PoolConfigImpl) GetConnectionTimeout() time.Duration {
	return p.connectionTimeout
}

// 实现BenchmarkConfig接口的方法

// GetTotal 获取总请求数
func (b *HttpBenchmarkConfig) GetTotal() int {
	return b.Total
}

// GetParallels 获取并发数
func (b *HttpBenchmarkConfig) GetParallels() int {
	return b.Parallels
}

// GetDataSize 获取数据大小
func (b *HttpBenchmarkConfig) GetDataSize() int {
	return b.DataSize
}

// GetTTL 获取生存时间
func (b *HttpBenchmarkConfig) GetTTL() time.Duration {
	return b.TTL
}

// GetReadPercent 获取读操作百分比
func (b *HttpBenchmarkConfig) GetReadPercent() int {
	return b.ReadPercent
}

// GetRandomKeys 获取随机键范围
func (b *HttpBenchmarkConfig) GetRandomKeys() int {
	return b.RandomKeys
}

// GetTestCase 获取测试用例
func (b *HttpBenchmarkConfig) GetTestCase() string {
	return b.TestCase
}

// GetTimeout 获取超时时间
func (b *HttpBenchmarkConfig) GetTimeout() time.Duration {
	return b.Timeout
}
