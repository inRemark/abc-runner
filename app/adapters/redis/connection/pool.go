package connection

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"abc-runner/app/adapters/redis/config"
)

// ConnectionPool 连接池接口
type ConnectionPool interface {
	GetConnection() (redis.Cmdable, error)
	ReturnConnection(redis.Cmdable) error
	Close() error
	Stats() PoolStats
	HealthCheck() error
}

// PoolStats 连接池统计信息
type PoolStats struct {
	TotalConnections   int `json:"total_connections"`
	ActiveConnections  int `json:"active_connections"`
	IdleConnections    int `json:"idle_connections"`
	WaitingConnections int `json:"waiting_connections"`
	PoolHits           int `json:"pool_hits"`
	PoolMisses         int `json:"pool_misses"`
	PoolTimeouts       int `json:"pool_timeouts"`
}

// RedisConnectionPool Redis连接池实现
type RedisConnectionPool struct {
	config       *config.RedisConfig
	poolConfig   *config.PoolConfigImpl
	clientType   string
	factory      *ClientFactory
	connections  chan redis.Cmdable
	activeConns  map[redis.Cmdable]bool
	stats        PoolStats
	mutex        sync.RWMutex
	ctx          context.Context
	closed       bool
}

// PoolManager 连接池管理器
type PoolManager struct {
	pools map[string]ConnectionPool
	mutex sync.RWMutex
}

// NewRedisConnectionPool 创建Redis连接池
func NewRedisConnectionPool(cfg *config.RedisConfig) (*RedisConnectionPool, error) {
	pool := &RedisConnectionPool{
		config:      cfg,
		poolConfig:  &cfg.Pool,
		clientType:  cfg.GetMode(),
		factory:     NewClientFactory(),
		connections: make(chan redis.Cmdable, cfg.Pool.GetPoolSize()),
		activeConns: make(map[redis.Cmdable]bool),
		ctx:         context.Background(),
		closed:      false,
	}

	// 初始化连接池
	err := pool.initialize()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize connection pool: %w", err)
	}

	return pool, nil
}

// initialize 初始化连接池
func (p *RedisConnectionPool) initialize() error {
	minIdle := p.poolConfig.GetMinIdle()
	
	// 创建最小空闲连接数
	for i := 0; i < minIdle; i++ {
		conn, err := p.createConnection()
		if err != nil {
			return fmt.Errorf("failed to create initial connection %d: %w", i, err)
		}

		select {
		case p.connections <- conn:
			p.stats.IdleConnections++
		default:
			// 连接池已满，关闭连接
			p.closeConnection(conn)
		}
	}

	log.Printf("Connection pool initialized with %d idle connections", minIdle)
	return nil
}

// createConnection 创建新连接
func (p *RedisConnectionPool) createConnection() (redis.Cmdable, error) {
	switch p.clientType {
	case "cluster":
		return p.factory.CreateClusterClient(p.config.GetClusterConfig())
	case "sentinel":
		return p.factory.CreateSentinelClient(p.config.GetSentinelConfig())
	default: // standalone
		return p.factory.CreateStandaloneClient(p.config.GetStandaloneConfig())
	}
}

// GetConnection 获取连接
func (p *RedisConnectionPool) GetConnection() (redis.Cmdable, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.closed {
		return nil, fmt.Errorf("connection pool is closed")
	}

	// 尝试从池中获取空闲连接
	select {
	case conn := <-p.connections:
		p.stats.IdleConnections--
		p.stats.ActiveConnections++
		p.stats.PoolHits++
		p.activeConns[conn] = true
		return conn, nil

	default:
		// 没有空闲连接，检查是否可以创建新连接
		totalConns := p.stats.ActiveConnections + p.stats.IdleConnections
		if totalConns < p.poolConfig.GetPoolSize() {
			// 可以创建新连接
			conn, err := p.createConnection()
			if err != nil {
				p.stats.PoolMisses++
				return nil, fmt.Errorf("failed to create new connection: %w", err)
			}

			p.stats.ActiveConnections++
			p.stats.TotalConnections++
			p.activeConns[conn] = true
			return conn, nil
		}

		// 连接池已满，等待连接释放
		p.stats.WaitingConnections++
		p.mutex.Unlock()

		timeout := p.poolConfig.GetConnectionTimeout()
		ctx, cancel := context.WithTimeout(p.ctx, timeout)
		defer cancel()

		select {
		case conn := <-p.connections:
			p.mutex.Lock()
			p.stats.WaitingConnections--
			p.stats.IdleConnections--
			p.stats.ActiveConnections++
			p.stats.PoolHits++
			p.activeConns[conn] = true
			return conn, nil

		case <-ctx.Done():
			p.mutex.Lock()
			p.stats.WaitingConnections--
			p.stats.PoolTimeouts++
			return nil, fmt.Errorf("timeout waiting for connection")
		}
	}
}

// ReturnConnection 归还连接
func (p *RedisConnectionPool) ReturnConnection(conn redis.Cmdable) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.closed {
		p.closeConnection(conn)
		return nil
	}

	// 检查连接是否有效
	if !p.activeConns[conn] {
		return fmt.Errorf("connection not from this pool")
	}

	delete(p.activeConns, conn)
	p.stats.ActiveConnections--

	// 检查连接健康状态
	if err := p.pingConnection(conn); err != nil {
		// 连接不健康，关闭并创建新连接
		p.closeConnection(conn)
		p.stats.TotalConnections--

		// 尝试创建新连接补充池
		go p.replenishPool()
		return nil
	}

	// 尝试将连接放回池中
	select {
	case p.connections <- conn:
		p.stats.IdleConnections++
	default:
		// 池已满，关闭连接
		p.closeConnection(conn)
		p.stats.TotalConnections--
	}

	return nil
}

// replenishPool 补充连接池
func (p *RedisConnectionPool) replenishPool() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.closed {
		return
	}

	minIdle := p.poolConfig.GetMinIdle()
	currentIdle := p.stats.IdleConnections

	if currentIdle < minIdle {
		needed := minIdle - currentIdle
		maxPool := p.poolConfig.GetPoolSize()
		totalConns := p.stats.ActiveConnections + p.stats.IdleConnections

		for i := 0; i < needed && totalConns < maxPool; i++ {
			conn, err := p.createConnection()
			if err != nil {
				log.Printf("Failed to replenish connection pool: %v", err)
				break
			}

			select {
			case p.connections <- conn:
				p.stats.IdleConnections++
				p.stats.TotalConnections++
				totalConns++
			default:
				p.closeConnection(conn)
				break
			}
		}
	}
}

// pingConnection 检查连接健康状态
func (p *RedisConnectionPool) pingConnection(conn redis.Cmdable) error {
	ctx, cancel := context.WithTimeout(p.ctx, 5*time.Second)
	defer cancel()

	_, err := conn.Ping(ctx).Result()
	return err
}

// closeConnection 关闭连接
func (p *RedisConnectionPool) closeConnection(conn redis.Cmdable) {
	switch client := conn.(type) {
	case *redis.Client:
		client.Close()
	case *redis.ClusterClient:
		client.Close()
	}
}

// Close 关闭连接池
func (p *RedisConnectionPool) Close() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true

	// 关闭所有空闲连接
	close(p.connections)
	for conn := range p.connections {
		p.closeConnection(conn)
	}

	// 关闭所有活跃连接
	for conn := range p.activeConns {
		p.closeConnection(conn)
	}

	log.Printf("Connection pool closed successfully")
	return nil
}

// Stats 获取连接池统计信息
func (p *RedisConnectionPool) Stats() PoolStats {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.stats
}

// HealthCheck 健康检查
func (p *RedisConnectionPool) HealthCheck() error {
	conn, err := p.GetConnection()
	if err != nil {
		return fmt.Errorf("failed to get connection for health check: %w", err)
	}
	defer p.ReturnConnection(conn)

	return p.pingConnection(conn)
}

// NewPoolManager 创建连接池管理器
func NewPoolManager() *PoolManager {
	return &PoolManager{
		pools: make(map[string]ConnectionPool),
	}
}

// CreatePool 创建连接池
func (pm *PoolManager) CreatePool(name string, cfg *config.RedisConfig) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if _, exists := pm.pools[name]; exists {
		return fmt.Errorf("pool %s already exists", name)
	}

	pool, err := NewRedisConnectionPool(cfg)
	if err != nil {
		return fmt.Errorf("failed to create pool %s: %w", name, err)
	}

	pm.pools[name] = pool
	log.Printf("Created connection pool: %s", name)
	return nil
}

// GetPool 获取连接池
func (pm *PoolManager) GetPool(name string) (ConnectionPool, error) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	pool, exists := pm.pools[name]
	if !exists {
		return nil, fmt.Errorf("pool %s not found", name)
	}

	return pool, nil
}

// ClosePool 关闭指定连接池
func (pm *PoolManager) ClosePool(name string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	pool, exists := pm.pools[name]
	if !exists {
		return fmt.Errorf("pool %s not found", name)
	}

	err := pool.Close()
	if err != nil {
		return fmt.Errorf("failed to close pool %s: %w", name, err)
	}

	delete(pm.pools, name)
	log.Printf("Closed connection pool: %s", name)
	return nil
}

// CloseAll 关闭所有连接池
func (pm *PoolManager) CloseAll() error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	var errors []error

	for name, pool := range pm.pools {
		if err := pool.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close pool %s: %w", name, err))
		}
	}

	pm.pools = make(map[string]ConnectionPool)

	if len(errors) > 0 {
		return fmt.Errorf("errors closing pools: %v", errors)
	}

	log.Printf("All connection pools closed")
	return nil
}

// GetAllStats 获取所有连接池统计信息
func (pm *PoolManager) GetAllStats() map[string]PoolStats {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	stats := make(map[string]PoolStats)
	for name, pool := range pm.pools {
		stats[name] = pool.Stats()
	}

	return stats
}

// HealthCheckAll 检查所有连接池健康状态
func (pm *PoolManager) HealthCheckAll() map[string]error {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	results := make(map[string]error)
	for name, pool := range pm.pools {
		results[name] = pool.HealthCheck()
	}

	return results
}