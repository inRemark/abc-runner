package connection

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	"abc-runner/app/adapters/grpc/config"
)

// ConnectionState gRPC连接状态
type ConnectionState int32

const (
	StateIdle ConnectionState = iota
	StateConnecting
	StateReady
	StateTransientFailure
	StateShutdown
)

// GRPCConnection gRPC连接包装器
type GRPCConnection struct {
	conn       *grpc.ClientConn
	address    string
	index      int
	createTime time.Time
	lastUsed   time.Time
	useCount   int64
	state      ConnectionState
	mu         sync.RWMutex
}

// NewGRPCConnection 创建新的gRPC连接
func NewGRPCConnection(ctx context.Context, address string, index int, opts []grpc.DialOption) (*GRPCConnection, error) {
	conn, err := grpc.DialContext(ctx, address, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection: %w", err)
	}

	grpcConn := &GRPCConnection{
		conn:       conn,
		address:    address,
		index:      index,
		createTime: time.Now(),
		lastUsed:   time.Now(),
		state:      StateReady,
	}

	return grpcConn, nil
}

// GetConn 获取底层gRPC连接
func (c *GRPCConnection) GetConn() *grpc.ClientConn {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.lastUsed = time.Now()
	atomic.AddInt64(&c.useCount, 1)
	return c.conn
}

// IsHealthy 检查连接是否健康
func (c *GRPCConnection) IsHealthy() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.conn == nil {
		return false
	}

	state := c.conn.GetState()
	return state == connectivity.Ready || state == connectivity.Idle
}

// Close 关闭连接
func (c *GRPCConnection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.state = StateShutdown
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// GetStats 获取连接统计信息
func (c *GRPCConnection) GetStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return map[string]interface{}{
		"address":     c.address,
		"index":       c.index,
		"create_time": c.createTime,
		"last_used":   c.lastUsed,
		"use_count":   atomic.LoadInt64(&c.useCount),
		"state":       c.state,
	}
}

// ConnectionPool gRPC连接池
type ConnectionPool struct {
	connections []*GRPCConnection
	config      *config.GRPCConfig
	address     string
	opts        []grpc.DialOption

	// 连接管理
	mu         sync.RWMutex
	isShutdown bool

	// 负载均衡
	roundRobinIndex int64

	// 健康检查
	healthCheckInterval time.Duration
	healthCheckCancel   context.CancelFunc
}

// NewConnectionPool 创建新的连接池
func NewConnectionPool(config *config.GRPCConfig) *ConnectionPool {
	address := fmt.Sprintf("%s:%d", config.Connection.Address, config.Connection.Port)

	pool := &ConnectionPool{
		config:              config,
		address:             address,
		connections:         make([]*GRPCConnection, 0, config.Connection.Pool.PoolSize),
		healthCheckInterval: 30 * time.Second,
	}

	// 构建连接选项
	pool.buildDialOptions()

	return pool
}

// Initialize 初始化连接池
func (p *ConnectionPool) Initialize(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.isShutdown {
		return fmt.Errorf("connection pool is shut down")
	}

	poolSize := p.config.Connection.Pool.PoolSize
	log.Printf("Initializing gRPC connection pool with %d connections to %s", poolSize, p.address)

	// 创建连接
	for i := 0; i < poolSize; i++ {
		conn, err := NewGRPCConnection(ctx, p.address, i, p.opts)
		if err != nil {
			// 清理已创建的连接
			p.closeAllConnections()
			return fmt.Errorf("failed to create connection %d: %w", i, err)
		}
		p.connections = append(p.connections, conn)
	}

	// 启动健康检查
	p.startHealthCheck()

	log.Printf("Successfully initialized gRPC connection pool with %d connections", len(p.connections))
	return nil
}

// GetConnection 获取可用连接
func (p *ConnectionPool) GetConnection() (*GRPCConnection, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.isShutdown {
		return nil, fmt.Errorf("connection pool is shut down")
	}

	if len(p.connections) == 0 {
		return nil, fmt.Errorf("no connections available")
	}

	// 使用轮询策略
	index := atomic.AddInt64(&p.roundRobinIndex, 1) % int64(len(p.connections))
	conn := p.connections[index]

	// 检查连接健康状态
	if !conn.IsHealthy() {
		log.Printf("Connection %d is unhealthy, attempting to find healthy connection", index)
		// 尝试找到健康的连接
		for i := 0; i < len(p.connections); i++ {
			testIndex := (int(index) + i) % len(p.connections)
			testConn := p.connections[testIndex]
			if testConn.IsHealthy() {
				return testConn, nil
			}
		}
		return nil, fmt.Errorf("no healthy connections available")
	}

	return conn, nil
}

// GetStats 获取连接池统计信息
func (p *ConnectionPool) GetStats() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := map[string]interface{}{
		"address":           p.address,
		"pool_size":         len(p.connections),
		"configured_size":   p.config.Connection.Pool.PoolSize,
		"is_shutdown":       p.isShutdown,
		"round_robin_index": atomic.LoadInt64(&p.roundRobinIndex),
	}

	// 添加每个连接的状态
	connectionStats := make([]map[string]interface{}, len(p.connections))
	healthyCount := 0
	for i, conn := range p.connections {
		connectionStats[i] = conn.GetStats()
		if conn.IsHealthy() {
			healthyCount++
		}
	}

	stats["connections"] = connectionStats
	stats["healthy_connections"] = healthyCount

	return stats
}

// Close 关闭连接池
func (p *ConnectionPool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.isShutdown {
		return nil
	}

	log.Printf("Closing gRPC connection pool")

	// 停止健康检查
	if p.healthCheckCancel != nil {
		p.healthCheckCancel()
	}

	// 关闭所有连接
	p.closeAllConnections()

	p.isShutdown = true
	log.Printf("gRPC connection pool closed")

	return nil
}

// 内部方法

// buildDialOptions 构建连接选项
func (p *ConnectionPool) buildDialOptions() {
	var opts []grpc.DialOption

	// TLS配置
	if p.config.GRPCSpecific.TLS.Enabled {
		var creds credentials.TransportCredentials
		if p.config.GRPCSpecific.TLS.InsecureSkipVerify {
			creds = insecure.NewCredentials()
		} else {
			// 可以根据需要添加实际的TLS配置
			creds = insecure.NewCredentials()
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Keep-alive配置
	if p.config.Connection.KeepAlive {
		kacp := keepalive.ClientParameters{
			Time:                p.config.Connection.KeepAlivePeriod,
			Timeout:             10 * time.Second,
			PermitWithoutStream: true,
		}
		opts = append(opts, grpc.WithKeepaliveParams(kacp))
	}

	// 负载均衡配置
	opts = append(opts, grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"loadBalancingConfig": [{"%s":{}}]}`,
		p.config.GRPCSpecific.LoadBalancing)))

	// 最大消息大小
	if p.config.GRPCSpecific.MaxMessageSize > 0 {
		opts = append(opts, grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(p.config.GRPCSpecific.MaxMessageSize),
			grpc.MaxCallSendMsgSize(p.config.GRPCSpecific.MaxMessageSize),
		))
	}

	// 连接超时
	opts = append(opts, grpc.WithTimeout(p.config.Connection.Timeout))

	p.opts = opts
}

// closeAllConnections 关闭所有连接
func (p *ConnectionPool) closeAllConnections() {
	for i, conn := range p.connections {
		if conn != nil {
			if err := conn.Close(); err != nil {
				log.Printf("Error closing connection %d: %v", i, err)
			}
		}
	}
	p.connections = p.connections[:0]
}

// startHealthCheck 启动健康检查
func (p *ConnectionPool) startHealthCheck() {
	ctx, cancel := context.WithCancel(context.Background())
	p.healthCheckCancel = cancel

	go func() {
		ticker := time.NewTicker(p.healthCheckInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				p.performHealthCheck()
			}
		}
	}()
}

// performHealthCheck 执行健康检查
func (p *ConnectionPool) performHealthCheck() {
	p.mu.RLock()
	connections := p.connections
	p.mu.RUnlock()

	if p.isShutdown {
		return
	}

	for _, conn := range connections {
		if !conn.IsHealthy() {
			log.Printf("Detected unhealthy connection %d to %s", conn.index, conn.address)
			// 这里可以实现连接重建逻辑
		}
	}
}
