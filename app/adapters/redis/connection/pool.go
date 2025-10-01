package connection

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"

	"abc-runner/app/adapters/redis/config"
)

// RedisConnectionPool Redis连接池
type RedisConnectionPool struct {
	client redis.UniversalClient
	config *config.RedisConfig
	mutex  sync.RWMutex
}

// NewRedisConnectionPool 创建连接池
func NewRedisConnectionPool(cfg *config.RedisConfig) (*RedisConnectionPool, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	pool := &RedisConnectionPool{
		config: cfg,
	}

	client, err := pool.createClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis client: %w", err)
	}

	pool.client = client
	return pool, nil
}

// createClient 创建Redis客户端
func (p *RedisConnectionPool) createClient() (redis.UniversalClient, error) {
	options := &redis.UniversalOptions{
		PoolSize:     p.config.Pool.PoolSize,
		MinIdleConns: p.config.Pool.MinIdle,
		IdleTimeout:  p.config.Pool.IdleTimeout,
		DialTimeout:  p.config.Pool.ConnectionTimeout,
		ReadTimeout:  30 * time.Second, // 默认值
		WriteTimeout: 30 * time.Second, // 默认值
		PoolTimeout:  p.config.Pool.ConnectionTimeout,
		MaxRetries:   3, // 默认值
	}

	// 根据模式设置连接参数
	switch p.config.GetMode() {
	case "cluster":
		cluster := p.config.GetClusterConfig()
		options.Addrs = cluster.Addrs
		if cluster.Password != "" {
			options.Password = cluster.Password
		}
	case "sentinel":
		sentinel := p.config.GetSentinelConfig()
		options.Addrs = sentinel.Addrs
		options.MasterName = sentinel.MasterName
		if sentinel.Password != "" {
			options.Password = sentinel.Password
		}
		options.DB = sentinel.Db
	default: // standalone
		standalone := p.config.GetStandaloneConfig()
		options.Addrs = []string{standalone.Addr}
		if standalone.Password != "" {
			options.Password = standalone.Password
		}
		options.DB = standalone.Db
	}

	client := redis.NewUniversalClient(options)
	return client, nil
}

// GetClient 获取Redis客户端
func (p *RedisConnectionPool) GetClient() redis.UniversalClient {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.client
}

// Ping 测试连接
func (p *RedisConnectionPool) Ping(ctx context.Context) error {
	client := p.GetClient()
	if client == nil {
		return fmt.Errorf("redis client is not initialized")
	}

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}

	return nil
}

// GetStats 获取连接池统计信息
func (p *RedisConnectionPool) GetStats() map[string]interface{} {
	client := p.GetClient()
	if client == nil {
		return map[string]interface{}{
			"status": "not_initialized",
		}
	}

	stats := client.PoolStats()
	return map[string]interface{}{
		"hits":        stats.Hits,
		"misses":      stats.Misses,
		"timeouts":    stats.Timeouts,
		"total_conns": stats.TotalConns,
		"idle_conns":  stats.IdleConns,
		"stale_conns": stats.StaleConns,
	}
}

// Close 关闭连接池
func (p *RedisConnectionPool) Close() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.client != nil {
		err := p.client.Close()
		p.client = nil
		return err
	}

	return nil
}

// Reconnect 重新连接
func (p *RedisConnectionPool) Reconnect() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// 关闭现有连接
	if p.client != nil {
		_ = p.client.Close()
	}

	// 创建新连接
	client, err := p.createClient()
	if err != nil {
		return fmt.Errorf("failed to reconnect: %w", err)
	}

	p.client = client
	return nil
}

// IsConnected 检查是否已连接
func (p *RedisConnectionPool) IsConnected(ctx context.Context) bool {
	if p.client == nil {
		return false
	}

	err := p.Ping(ctx)
	return err == nil
}

// GetConfig 获取配置
func (p *RedisConnectionPool) GetConfig() *config.RedisConfig {
	return p.config
}

// SetConfig 更新配置并重新连接
func (p *RedisConnectionPool) SetConfig(cfg *config.RedisConfig) error {
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

	// 关闭现有连接
	if p.client != nil {
		_ = p.client.Close()
	}

	// 更新配置
	p.config = cfg

	// 创建新连接
	client, err := p.createClient()
	if err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	p.client = client
	return nil
}
