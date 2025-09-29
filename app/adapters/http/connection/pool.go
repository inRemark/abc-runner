package connection

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	httpConfig "abc-runner/app/adapters/http/config"
)

// HTTPConnectionPool HTTP连接池管理器
type HTTPConnectionPool struct {
	// HTTP客户端
	client *http.Client
	
	// 配置和状态
	config    *httpConfig.HttpAdapterConfig
	isHealthy bool
	
	// 统计信息
	activeConnections int64
	totalConnections  int64
	failedConnections int64
	requestCount      int64
	
	// 同步控制
	mutex sync.RWMutex
}

// PoolConfig HTTP连接池配置
type PoolConfig struct {
	MaxConnections       int           // 最大连接数
	MaxIdleConns         int           // 最大空闲连接数
	MaxConnsPerHost      int           // 每个主机最大连接数
	IdleConnTimeout      time.Duration // 空闲连接超时
	ConnectionTimeout    time.Duration // 连接超时
	RequestTimeout       time.Duration // 请求超时
	TLSHandshakeTimeout  time.Duration // TLS握手超时
	DisableKeepAlives    bool          // 是否禁用keep-alive
	DisableCompression   bool          // 是否禁用压缩
}

// NewHTTPConnectionPool 创建HTTP连接池
func NewHTTPConnectionPool(config *httpConfig.HttpAdapterConfig, poolConfig PoolConfig) (*HTTPConnectionPool, error) {
	// 创建自定义Transport
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   poolConfig.ConnectionTimeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          poolConfig.MaxIdleConns,
		MaxIdleConnsPerHost:   poolConfig.MaxConnsPerHost,
		MaxConnsPerHost:       poolConfig.MaxConnsPerHost,
		IdleConnTimeout:       poolConfig.IdleConnTimeout,
		TLSHandshakeTimeout:   poolConfig.TLSHandshakeTimeout,
		DisableKeepAlives:     poolConfig.DisableKeepAlives,
		DisableCompression:    poolConfig.DisableCompression,
		ExpectContinueTimeout: 1 * time.Second,
	}
	
	// 配置TLS
	// 由于原始配置结构中没有UseHTTPS字段，这里暂时跳过TLS配置
	// 未来可以根据需要添加TLS配置
	
	// 创建HTTP客户端
	client := &http.Client{
		Transport: transport,
		Timeout:   poolConfig.RequestTimeout,
	}
	
	// 不自动跟随重定向，让用户控制
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return fmt.Errorf("stopped after 10 redirects")
		}
		return nil
	}
	
	pool := &HTTPConnectionPool{
		client:    client,
		config:    config,
		isHealthy: true,
	}
	
	return pool, nil
}

// GetClient 获取HTTP客户端
func (p *HTTPConnectionPool) GetClient() *http.Client {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.client
}

// HealthCheck 健康检查
func (p *HTTPConnectionPool) HealthCheck() error {
	if p.client == nil {
		return fmt.Errorf("HTTP client not initialized")
	}
	
	// 执行简单的HEAD请求测试连接
	baseURL := p.config.Connection.BaseURL
	if baseURL == "" {
		// 如果没有配置基础URL，跳过健康检查
		return nil
	}
	
	req, err := http.NewRequest("HEAD", baseURL, nil)
	if err != nil {
		p.markUnhealthy()
		return fmt.Errorf("failed to create health check request: %w", err)
	}
	
	// 设置较短的超时时间用于健康检查
	client := &http.Client{
		Transport: p.client.Transport,
		Timeout:   5 * time.Second,
	}
	
	resp, err := client.Do(req)
	if err != nil {
		p.markUnhealthy()
		return fmt.Errorf("health check request failed: %w", err)
	}
	defer resp.Body.Close()
	
	// HTTP 2xx-4xx 状态码都认为是健康的
	if resp.StatusCode >= 200 && resp.StatusCode < 500 {
		p.markHealthy()
		return nil
	}
	
	p.markUnhealthy()
	return fmt.Errorf("health check failed with status: %d", resp.StatusCode)
}

// GetStats 获取连接池统计信息
func (p *HTTPConnectionPool) GetStats() map[string]interface{} {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	
	stats := map[string]interface{}{
		"is_healthy":          p.isHealthy,
		"active_connections":  p.activeConnections,
		"total_connections":   p.totalConnections,
		"failed_connections":  p.failedConnections,
		"request_count":       p.requestCount,
	}
	
	// 添加客户端配置信息
	if p.client != nil && p.client.Transport != nil {
		if transport, ok := p.client.Transport.(*http.Transport); ok {
			stats["transport_stats"] = map[string]interface{}{
				"max_idle_conns":         transport.MaxIdleConns,
				"max_idle_conns_per_host": transport.MaxIdleConnsPerHost,
				"max_conns_per_host":     transport.MaxConnsPerHost,
				"idle_conn_timeout":      transport.IdleConnTimeout.String(),
				"disable_keep_alives":    transport.DisableKeepAlives,
				"disable_compression":    transport.DisableCompression,
			}
		}
	}
	
	return stats
}

// Close 关闭连接池
func (p *HTTPConnectionPool) Close() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	if p.client != nil && p.client.Transport != nil {
		if transport, ok := p.client.Transport.(*http.Transport); ok {
			transport.CloseIdleConnections()
		}
		p.client = nil
	}
	
	p.isHealthy = false
	return nil
}

// IsHealthy 检查连接池是否健康
func (p *HTTPConnectionPool) IsHealthy() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.isHealthy
}

// IncrementRequestCount 增加请求计数
func (p *HTTPConnectionPool) IncrementRequestCount() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.requestCount++
}

// IncrementConnectionCount 增加连接计数
func (p *HTTPConnectionPool) IncrementConnectionCount() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.totalConnections++
	p.activeConnections++
}

// DecrementConnectionCount 减少活动连接计数
func (p *HTTPConnectionPool) DecrementConnectionCount() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if p.activeConnections > 0 {
		p.activeConnections--
	}
}

// IncrementFailedConnectionCount 增加失败连接计数
func (p *HTTPConnectionPool) IncrementFailedConnectionCount() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.failedConnections++
}

// GetConfig 获取HTTP配置
func (p *HTTPConnectionPool) GetConfig() *httpConfig.HttpAdapterConfig {
	return p.config
}

// 内部方法

func (p *HTTPConnectionPool) markHealthy() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.isHealthy = true
}

func (p *HTTPConnectionPool) markUnhealthy() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.isHealthy = false
}

// SetTimeout 设置请求超时
func (p *HTTPConnectionPool) SetTimeout(timeout time.Duration) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if p.client != nil {
		p.client.Timeout = timeout
	}
}

// GetTimeout 获取请求超时
func (p *HTTPConnectionPool) GetTimeout() time.Duration {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	if p.client != nil {
		return p.client.Timeout
	}
	return 0
}

// HttpConnectionPool HTTP连接池类型别名
type HttpConnectionPool = HTTPConnectionPool

// NewHttpConnectionPool 创建HTTP连接池（适配器版本）
func NewHttpConnectionPool(config *httpConfig.HttpAdapterConfig) (*HTTPConnectionPool, error) {
	// 使用默认的池配置
	poolConfig := PoolConfig{
		MaxConnections:       100,
		MaxIdleConns:         config.Connection.MaxIdleConns,
		MaxConnsPerHost:      config.Connection.MaxConnsPerHost,
		IdleConnTimeout:      config.Connection.IdleConnTimeout,
		ConnectionTimeout:    config.Connection.Timeout,
		RequestTimeout:       config.Connection.Timeout,
		TLSHandshakeTimeout:  10 * time.Second,
		DisableKeepAlives:    false,
		DisableCompression:   false,
	}
	
	return NewHTTPConnectionPool(config, poolConfig)
}