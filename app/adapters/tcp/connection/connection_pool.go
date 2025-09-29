package connection

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"abc-runner/app/adapters/tcp/config"
)

// ConnectionPool TCP连接池
type ConnectionPool struct {
	connections chan net.Conn
	mu          sync.RWMutex
	closed      bool
	config      *config.TCPConfig
	activeCount int64
	address     string

	// 性能统计
	createdCount    int64 // 已创建连接数
	validationCount int64 // 验证次数
	validationFails int64 // 验证失败次数
	getTimeouts     int64 // 获取连接超时次数
}

// NewConnectionPool 创建新的连接池
func NewConnectionPool(cfg *config.TCPConfig) (*ConnectionPool, error) {
	address := fmt.Sprintf("%s:%d", cfg.Connection.Address, cfg.Connection.Port)

	pool := &ConnectionPool{
		connections: make(chan net.Conn, cfg.Connection.Pool.PoolSize),
		config:      cfg,
		address:     address,
		closed:      false,
	}

	// 预创建最小空闲连接
	for i := 0; i < cfg.Connection.Pool.MinIdle; i++ {
		conn, err := pool.createConnection()
		if err != nil {
			// 如果无法创建连接，清理已创建的连接并返回错误
			pool.Close()
			return nil, fmt.Errorf("failed to create initial connection %d: %w", i, err)
		}

		select {
		case pool.connections <- conn:
		default:
			// 连接池已满，关闭连接
			conn.Close()
		}
	}

	return pool, nil
}

// createConnection 创建新连接
func (p *ConnectionPool) createConnection() (net.Conn, error) {
	dialer := &net.Dialer{
		Timeout:   p.config.Connection.Pool.ConnectionTimeout,
		KeepAlive: p.config.Connection.KeepAlivePeriod,
	}

	conn, err := dialer.Dial("tcp", p.address)
	if err != nil {
		return nil, fmt.Errorf("failed to dial %s: %w", p.address, err)
	}

	// 配置TCP选项
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		// 设置TCP选项
		if err := p.configureTCPConnection(tcpConn); err != nil {
			conn.Close()
			return nil, fmt.Errorf("failed to configure TCP connection: %w", err)
		}
	}

	atomic.AddInt64(&p.activeCount, 1)
	atomic.AddInt64(&p.createdCount, 1)
	return conn, nil
}

// configureTCPConnection 配置TCP连接选项
func (p *ConnectionPool) configureTCPConnection(conn *net.TCPConn) error {
	// 设置NoDelay
	if err := conn.SetNoDelay(p.config.TCPSpecific.NoDelay); err != nil {
		return fmt.Errorf("failed to set NoDelay: %w", err)
	}

	// 设置KeepAlive
	if p.config.Connection.KeepAlive {
		if err := conn.SetKeepAlive(true); err != nil {
			return fmt.Errorf("failed to set KeepAlive: %w", err)
		}

		if err := conn.SetKeepAlivePeriod(p.config.Connection.KeepAlivePeriod); err != nil {
			return fmt.Errorf("failed to set KeepAlivePeriod: %w", err)
		}
	}

	// 设置Linger
	if p.config.TCPSpecific.LingerTimeout >= 0 {
		if err := conn.SetLinger(p.config.TCPSpecific.LingerTimeout); err != nil {
			return fmt.Errorf("failed to set Linger: %w", err)
		}
	}

	return nil
}

// GetConnection 从池中获取连接
func (p *ConnectionPool) GetConnection() (net.Conn, error) {
	p.mu.RLock()
	if p.closed {
		p.mu.RUnlock()
		return nil, fmt.Errorf("connection pool is closed")
	}
	p.mu.RUnlock()

	// 尝试从池中获取现有连接
	select {
	case conn := <-p.connections:
		// 检查连接是否仍然有效
		if p.isConnectionValid(conn) {
			return conn, nil
		}
		// 连接无效，关闭并创建新连接
		conn.Close()
		atomic.AddInt64(&p.activeCount, -1)
	default:
		// 池中没有可用连接
	}

	// 创建新连接
	conn, err := p.createConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to create new connection: %w", err)
	}

	return conn, nil
}

// ReturnConnection 将连接返回到池中
func (p *ConnectionPool) ReturnConnection(conn net.Conn) {
	if conn == nil {
		return
	}

	p.mu.RLock()
	closed := p.closed
	p.mu.RUnlock()

	if closed {
		conn.Close()
		atomic.AddInt64(&p.activeCount, -1)
		return
	}

	// 检查连接是否仍然有效
	if !p.isConnectionValid(conn) {
		conn.Close()
		atomic.AddInt64(&p.activeCount, -1)
		return
	}

	// 尝试将连接返回到池中
	select {
	case p.connections <- conn:
		// 成功返回到池中
	default:
		// 池已满，关闭连接
		conn.Close()
		atomic.AddInt64(&p.activeCount, -1)
	}
}

// isConnectionValid 检查连接是否有效
func (p *ConnectionPool) isConnectionValid(conn net.Conn) bool {
	atomic.AddInt64(&p.validationCount, 1)

	if conn == nil {
		atomic.AddInt64(&p.validationFails, 1)
		return false
	}

	// 设置短超时进行连接测试
	originalDeadline := time.Now().Add(100 * time.Millisecond)
	if err := conn.SetDeadline(originalDeadline); err != nil {
		atomic.AddInt64(&p.validationFails, 1)
		return false
	}

	// 尝试写入0字节来测试连接（更轻量级的检查）
	_, err := conn.Write([]byte{})
	if err != nil {
		// 如果错误是超时，连接可能仍然有效
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			// 重置deadline并返回有效
			conn.SetDeadline(time.Time{})
			return true
		}
		atomic.AddInt64(&p.validationFails, 1)
		return false
	}

	// 重置deadline
	if err := conn.SetDeadline(time.Time{}); err != nil {
		atomic.AddInt64(&p.validationFails, 1)
		return false
	}

	return true
}

// Close 关闭连接池
func (p *ConnectionPool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true

	// 关闭所有池中的连接
	close(p.connections)
	for conn := range p.connections {
		if conn != nil {
			conn.Close()
			atomic.AddInt64(&p.activeCount, -1)
		}
	}

	return nil
}

// ActiveConnections 获取活跃连接数
func (p *ConnectionPool) ActiveConnections() int64 {
	return atomic.LoadInt64(&p.activeCount)
}

// AvailableConnections 获取可用连接数
func (p *ConnectionPool) AvailableConnections() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return 0
	}

	return len(p.connections)
}

// Stats 获取连接池统计信息
func (p *ConnectionPool) Stats() map[string]interface{} {
	return map[string]interface{}{
		"active_connections":      p.ActiveConnections(),
		"available_connections":   p.AvailableConnections(),
		"pool_size":               p.config.Connection.Pool.PoolSize,
		"min_idle":                p.config.Connection.Pool.MinIdle,
		"max_idle":                p.config.Connection.Pool.MaxIdle,
		"closed":                  p.closed,
		"created_count":           atomic.LoadInt64(&p.createdCount),
		"validation_count":        atomic.LoadInt64(&p.validationCount),
		"validation_fails":        atomic.LoadInt64(&p.validationFails),
		"get_timeouts":            atomic.LoadInt64(&p.getTimeouts),
		"validation_success_rate": p.getValidationSuccessRate(),
	}
}

// getValidationSuccessRate 获取验证成功率
func (p *ConnectionPool) getValidationSuccessRate() float64 {
	totalValidations := atomic.LoadInt64(&p.validationCount)
	if totalValidations == 0 {
		return 100.0
	}
	failedValidations := atomic.LoadInt64(&p.validationFails)
	successRate := float64(totalValidations-failedValidations) / float64(totalValidations) * 100.0
	return successRate
}
