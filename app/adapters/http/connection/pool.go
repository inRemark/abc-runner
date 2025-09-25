package connection

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	httpConfig "abc-runner/app/adapters/http/config"
)

// ConnectionPool HTTP连接池
type ConnectionPool struct {
	config      *httpConfig.HttpAdapterConfig
	clients     []*http.Client
	clientIndex int
	mutex       sync.RWMutex
	poolSize    int
	idleClients chan *http.Client
	closed      bool
}

// PoolConfig 连接池配置
type PoolConfig struct {
	MaxConnections    int
	MaxIdleConns      int
	MaxConnsPerHost   int
	IdleConnTimeout   time.Duration
	ConnectionTimeout time.Duration
	DisableKeepAlives bool
}

// NewConnectionPool 创建新的连接池
func NewConnectionPool(config *httpConfig.HttpAdapterConfig, poolConfig PoolConfig) (*ConnectionPool, error) {
	pool := &ConnectionPool{
		config:      config,
		poolSize:    poolConfig.MaxConnections,
		clients:     make([]*http.Client, 0, poolConfig.MaxConnections),
		idleClients: make(chan *http.Client, poolConfig.MaxConnections),
		mutex:       sync.RWMutex{},
	}

	// 创建HTTP客户端
	for i := 0; i < poolConfig.MaxConnections; i++ {
		client, err := pool.createHTTPClient(poolConfig)
		if err != nil {
			// 清理已创建的客户端
			pool.Close()
			return nil, fmt.Errorf("failed to create HTTP client %d: %w", i, err)
		}

		pool.clients = append(pool.clients, client)
		pool.idleClients <- client
	}

	return pool, nil
}

// createHTTPClient 创建HTTP客户端
func (p *ConnectionPool) createHTTPClient(poolConfig PoolConfig) (*http.Client, error) {
	// 创建传输层配置
	transport := &http.Transport{
		MaxIdleConns:        poolConfig.MaxIdleConns,
		MaxIdleConnsPerHost: poolConfig.MaxConnsPerHost,
		IdleConnTimeout:     poolConfig.IdleConnTimeout,
		DisableKeepAlives:   poolConfig.DisableKeepAlives,
		DisableCompression:  p.config.Connection.DisableCompression,
	}

	// 配置TLS
	if p.needsTLS() {
		tlsConfig, err := p.createTLSConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to create TLS config: %w", err)
		}
		transport.TLSClientConfig = tlsConfig
	}

	// 创建客户端
	client := &http.Client{
		Transport: transport,
		Timeout:   p.config.Connection.Timeout,
	}

	// 配置重定向策略
	if p.config.Benchmark.FollowRedirects {
		client.CheckRedirect = p.createRedirectPolicy()
	} else {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	return client, nil
}

// needsTLS 检查是否需要TLS
func (p *ConnectionPool) needsTLS() bool {
	baseURL := p.config.Connection.BaseURL
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return false
	}

	return parsedURL.Scheme == "https" || p.config.Connection.TLS.ClientAuth
}

// createTLSConfig 创建TLS配置
func (p *ConnectionPool) createTLSConfig() (*tls.Config, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify:       p.config.Connection.TLS.InsecureSkipVerify,
		PreferServerCipherSuites: p.config.Connection.TLS.PreferServerCipherSuites,
		SessionTicketsDisabled:   p.config.Connection.TLS.SessionTicketsDisabled,
	}

	// 设置服务器名称
	if p.config.Connection.TLS.ServerName != "" {
		tlsConfig.ServerName = p.config.Connection.TLS.ServerName
	}

	// 设置TLS版本
	if err := p.setTLSVersions(tlsConfig); err != nil {
		return nil, err
	}

	// 设置密码套件
	if err := p.setCipherSuites(tlsConfig); err != nil {
		return nil, err
	}

	// 设置重新协商策略
	if err := p.setRenegotiation(tlsConfig); err != nil {
		return nil, err
	}

	// 加载客户端证书
	if p.config.Connection.TLS.ClientAuth {
		if err := p.loadClientCertificates(tlsConfig); err != nil {
			return nil, err
		}
	}

	// 加载CA证书
	if p.config.Connection.TLS.CAFile != "" {
		if err := p.loadCACertificates(tlsConfig); err != nil {
			return nil, err
		}
	}

	return tlsConfig, nil
}

// setTLSVersions 设置TLS版本
func (p *ConnectionPool) setTLSVersions(tlsConfig *tls.Config) error {
	// 设置最小版本
	if p.config.Connection.TLS.MinVersion != "" {
		minVersion, err := p.parseTLSVersion(p.config.Connection.TLS.MinVersion)
		if err != nil {
			return fmt.Errorf("invalid min TLS version: %w", err)
		}
		tlsConfig.MinVersion = minVersion
	}

	// 设置最大版本
	if p.config.Connection.TLS.MaxVersion != "" {
		maxVersion, err := p.parseTLSVersion(p.config.Connection.TLS.MaxVersion)
		if err != nil {
			return fmt.Errorf("invalid max TLS version: %w", err)
		}
		tlsConfig.MaxVersion = maxVersion
	}

	return nil
}

// parseTLSVersion 解析TLS版本
func (p *ConnectionPool) parseTLSVersion(version string) (uint16, error) {
	switch version {
	case "1.0":
		return tls.VersionTLS10, nil
	case "1.1":
		return tls.VersionTLS11, nil
	case "1.2":
		return tls.VersionTLS12, nil
	case "1.3":
		return tls.VersionTLS13, nil
	default:
		return 0, fmt.Errorf("unsupported TLS version: %s", version)
	}
}

// setCipherSuites 设置密码套件
func (p *ConnectionPool) setCipherSuites(tlsConfig *tls.Config) error {
	if len(p.config.Connection.TLS.CipherSuites) == 0 {
		return nil
	}

	var cipherSuites []uint16
	cipherSuiteMap := map[string]uint16{
		"TLS_AES_128_GCM_SHA256":       tls.TLS_AES_128_GCM_SHA256,
		"TLS_AES_256_GCM_SHA384":       tls.TLS_AES_256_GCM_SHA384,
		"TLS_CHACHA20_POLY1305_SHA256": tls.TLS_CHACHA20_POLY1305_SHA256,
	}

	for _, suite := range p.config.Connection.TLS.CipherSuites {
		if cipherID, exists := cipherSuiteMap[suite]; exists {
			cipherSuites = append(cipherSuites, cipherID)
		} else {
			return fmt.Errorf("unsupported cipher suite: %s", suite)
		}
	}

	tlsConfig.CipherSuites = cipherSuites
	return nil
}

// setRenegotiation 设置重新协商策略
func (p *ConnectionPool) setRenegotiation(tlsConfig *tls.Config) error {
	switch p.config.Connection.TLS.Renegotiation {
	case "never":
		tlsConfig.Renegotiation = tls.RenegotiateNever
	case "once":
		tlsConfig.Renegotiation = tls.RenegotiateOnceAsClient
	case "freely":
		tlsConfig.Renegotiation = tls.RenegotiateFreelyAsClient
	case "":
		// 使用默认值
	default:
		return fmt.Errorf("invalid renegotiation policy: %s", p.config.Connection.TLS.Renegotiation)
	}
	return nil
}

// loadClientCertificates 加载客户端证书
func (p *ConnectionPool) loadClientCertificates(tlsConfig *tls.Config) error {
	if p.config.Connection.TLS.CertFile == "" || p.config.Connection.TLS.KeyFile == "" {
		return fmt.Errorf("cert_file and key_file are required for client authentication")
	}

	cert, err := tls.LoadX509KeyPair(p.config.Connection.TLS.CertFile, p.config.Connection.TLS.KeyFile)
	if err != nil {
		return fmt.Errorf("failed to load client certificate: %w", err)
	}

	tlsConfig.Certificates = []tls.Certificate{cert}
	return nil
}

// loadCACertificates 加载CA证书
func (p *ConnectionPool) loadCACertificates(tlsConfig *tls.Config) error {
	// TODO: 实现CA证书加载
	// 这需要读取CA文件并添加到证书池
	return nil
}

// createRedirectPolicy 创建重定向策略
func (p *ConnectionPool) createRedirectPolicy() func(*http.Request, []*http.Request) error {
	return func(req *http.Request, via []*http.Request) error {
		if len(via) >= p.config.Benchmark.MaxRedirects {
			return fmt.Errorf("stopped after %d redirects", p.config.Benchmark.MaxRedirects)
		}
		return nil
	}
}

// GetClient 获取HTTP客户端
func (p *ConnectionPool) GetClient() *http.Client {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.closed {
		return nil
	}

	select {
	case client := <-p.idleClients:
		return client
	default:
		// 如果没有空闲客户端，使用轮询方式
		client := p.clients[p.clientIndex]
		p.clientIndex = (p.clientIndex + 1) % len(p.clients)
		return client
	}
}

// ReturnClient 归还HTTP客户端
func (p *ConnectionPool) ReturnClient(client *http.Client) {
	if p.closed || client == nil {
		return
	}

	select {
	case p.idleClients <- client:
		// 成功归还到池中
	default:
		// 池已满，忽略
	}
}

// GetSize 获取连接池大小
func (p *ConnectionPool) GetSize() int {
	return p.poolSize
}

// GetIdleCount 获取空闲连接数
func (p *ConnectionPool) GetIdleCount() int {
	return len(p.idleClients)
}

// GetActiveCount 获取活跃连接数
func (p *ConnectionPool) GetActiveCount() int {
	return p.poolSize - len(p.idleClients)
}

// Close 关闭连接池
func (p *ConnectionPool) Close() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true

	// 关闭空闲客户端通道
	close(p.idleClients)

	// 清理所有客户端
	for _, client := range p.clients {
		if transport, ok := client.Transport.(*http.Transport); ok {
			transport.CloseIdleConnections()
		}
	}

	// 清空客户端列表
	p.clients = nil

	return nil
}

// HealthCheck 健康检查
func (p *ConnectionPool) HealthCheck() error {
	if p.closed {
		return fmt.Errorf("connection pool is closed")
	}

	if len(p.clients) == 0 {
		return fmt.Errorf("no clients available in pool")
	}

	return nil
}

// GetStats 获取连接池统计信息
func (p *ConnectionPool) GetStats() map[string]interface{} {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	return map[string]interface{}{
		"pool_size":    p.poolSize,
		"idle_count":   len(p.idleClients),
		"active_count": p.poolSize - len(p.idleClients),
		"closed":       p.closed,
	}
}
