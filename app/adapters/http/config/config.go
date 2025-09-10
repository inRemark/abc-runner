package config

import (
	"fmt"
	"time"

	"abc-runner/app/core/interfaces"
)

// LoadDefaultHttpConfig 加载默认HTTP配置
func LoadDefaultHttpConfig() *HttpAdapterConfig {
	return &HttpAdapterConfig{
		Protocol: "http",
		Connection: HttpConnectionConfig{
			BaseURL:         "http://localhost:8080",
			Timeout:         30 * time.Second,
			MaxIdleConns:    10,
			MaxConnsPerHost: 10,
			IdleConnTimeout: 90 * time.Second,
		},
		Benchmark: HttpBenchmarkConfig{
			Total:     1000,
			Parallels: 10,
			Method:    "GET",
			Path:      "/",
			Headers:   make(map[string]string),
		},
		Auth: HttpAuthConfig{
			Type: "none",
		},
	}
}

// HttpAdapterConfig HTTP适配器配置
type HttpAdapterConfig struct {
	// 基础连接配置
	Protocol   string               `yaml:"protocol" json:"protocol"`     // 协议类型
	Connection HttpConnectionConfig `yaml:"connection" json:"connection"` // 连接配置

	// 请求模板配置
	Requests []HttpRequestConfig `yaml:"requests" json:"requests"`

	// 认证配置
	Auth HttpAuthConfig `yaml:"auth" json:"auth"`

	// 文件上传配置
	Upload HttpUploadConfig `yaml:"upload" json:"upload"`

	// 基准测试配置
	Benchmark HttpBenchmarkConfig `yaml:"benchmark" json:"benchmark"`
}

// HttpConnectionConfig HTTP连接配置
type HttpConnectionConfig struct {
	BaseURL            string        `yaml:"base_url" json:"base_url"`                       // 基础URL
	Timeout            time.Duration `yaml:"timeout" json:"timeout"`                         // 请求超时
	KeepAlive          time.Duration `yaml:"keep_alive" json:"keep_alive"`                   // 长连接保持时间
	MaxIdleConns       int           `yaml:"max_idle_conns" json:"max_idle_conns"`           // 最大空闲连接数
	MaxConnsPerHost    int           `yaml:"max_conns_per_host" json:"max_conns_per_host"`   // 每个主机最大连接数
	IdleConnTimeout    time.Duration `yaml:"idle_conn_timeout" json:"idle_conn_timeout"`     // 空闲连接超时
	DisableCompression bool          `yaml:"disable_compression" json:"disable_compression"` // 禁用压缩
	TLS                HttpTLSConfig `yaml:"tls" json:"tls"`                                 // TLS配置
}

// HttpTLSConfig TLS配置
type HttpTLSConfig struct {
	InsecureSkipVerify       bool     `yaml:"insecure_skip_verify" json:"insecure_skip_verify"`               // 跳过证书验证
	MinVersion               string   `yaml:"min_version" json:"min_version"`                                 // 最低TLS版本
	MaxVersion               string   `yaml:"max_version" json:"max_version"`                                 // 最高TLS版本
	CertFile                 string   `yaml:"cert_file" json:"cert_file"`                                     // 客户端证书文件
	KeyFile                  string   `yaml:"key_file" json:"key_file"`                                       // 客户端私钥文件
	CAFile                   string   `yaml:"ca_file" json:"ca_file"`                                         // CA证书文件
	ServerName               string   `yaml:"server_name" json:"server_name"`                                 // 服务器名称(SNI)
	ClientAuth               bool     `yaml:"client_auth" json:"client_auth"`                                 // 启用客户端认证
	CipherSuites             []string `yaml:"cipher_suites" json:"cipher_suites"`                             // 密码套件
	PreferServerCipherSuites bool     `yaml:"prefer_server_cipher_suites" json:"prefer_server_cipher_suites"` // 优先服务器密码套件
	SessionTicketsDisabled   bool     `yaml:"session_tickets_disabled" json:"session_tickets_disabled"`       // 禁用会话票据
	Renegotiation            string   `yaml:"renegotiation" json:"renegotiation"`                             // 重新协商策略
}

// HttpRequestConfig HTTP请求配置
type HttpRequestConfig struct {
	Method      string                `yaml:"method" json:"method"`             // 请求方法
	Path        string                `yaml:"path" json:"path"`                 // 请求路径
	Headers     map[string]string     `yaml:"headers" json:"headers"`           // 请求头
	Body        interface{}           `yaml:"body" json:"body"`                 // 请求体
	ContentType string                `yaml:"content_type" json:"content_type"` // 内容类型
	Weight      int                   `yaml:"weight" json:"weight"`             // 权重
	Upload      *HttpFileUploadConfig `yaml:"upload" json:"upload"`             // 文件上传配置
}

// HttpFileUploadConfig 文件上传配置
type HttpFileUploadConfig struct {
	Files    []FileConfig           `yaml:"files" json:"files"`         // 文件配置列表
	FormData map[string]interface{} `yaml:"form_data" json:"form_data"` // 表单数据
}

// FileConfig 文件配置
type FileConfig struct {
	Field   string `yaml:"field" json:"field"`     // 表单字段名
	Path    string `yaml:"path" json:"path"`       // 文件路径
	Pattern string `yaml:"pattern" json:"pattern"` // 文件匹配模式
}

// HttpAuthConfig HTTP认证配置
type HttpAuthConfig struct {
	Type     string `yaml:"type" json:"type"`         // 认证类型: none, basic, bearer, oauth2, mutual_tls
	Username string `yaml:"username" json:"username"` // 用户名
	Password string `yaml:"password" json:"password"` // 密码
	Token    string `yaml:"token" json:"token"`       // Token
}

// HttpUploadConfig HTTP上传配置
type HttpUploadConfig struct {
	Enable             bool          `yaml:"enable" json:"enable"`                           // 启用上传
	MaxFileSize        string        `yaml:"max_file_size" json:"max_file_size"`             // 最大文件大小
	AllowedTypes       []string      `yaml:"allowed_types" json:"allowed_types"`             // 允许的文件类型
	UploadField        string        `yaml:"upload_field" json:"upload_field"`               // 上传字段名
	ChunkSize          string        `yaml:"chunk_size" json:"chunk_size"`                   // 分块大小
	ConcurrentUploads  int           `yaml:"concurrent_uploads" json:"concurrent_uploads"`   // 并发上传数
	EnableCompression  bool          `yaml:"enable_compression" json:"enable_compression"`   // 启用压缩
	CompressionLevel   int           `yaml:"compression_level" json:"compression_level"`     // 压缩级别
	GenerateThumbnails bool          `yaml:"generate_thumbnails" json:"generate_thumbnails"` // 生成缩略图
	VirusScan          bool          `yaml:"virus_scan" json:"virus_scan"`                   // 病毒扫描
	TempDir            string        `yaml:"temp_dir" json:"temp_dir"`                       // 临时目录
	CleanupInterval    time.Duration `yaml:"cleanup_interval" json:"cleanup_interval"`       // 清理间隔
	PreserveFilename   bool          `yaml:"preserve_filename" json:"preserve_filename"`     // 保留文件名
}

// HttpBenchmarkConfig HTTP基准测试配置
type HttpBenchmarkConfig struct {
	Total              int           `yaml:"total" json:"total"`                             // 总请求数
	Parallels          int           `yaml:"parallels" json:"parallels"`                     // 并发数
	Duration           time.Duration `yaml:"duration" json:"duration"`                       // 测试持续时间
	RampUp             time.Duration `yaml:"ramp_up" json:"ramp_up"`                         // 渐进加载时间
	DataSize           int           `yaml:"data_size" json:"data_size"`                     // 数据大小
	TTL                time.Duration `yaml:"ttl" json:"ttl"`                                 // 生存时间
	ReadPercent        int           `yaml:"read_percent" json:"read_percent"`               // 读操作百分比
	RandomKeys         int           `yaml:"random_keys" json:"random_keys"`                 // 随机键范围
	TestCase           string        `yaml:"test_case" json:"test_case"`                     // 测试用例
	Timeout            time.Duration `yaml:"timeout" json:"timeout"`                         // 超时时间
	FollowRedirects    bool          `yaml:"follow_redirects" json:"follow_redirects"`       // 跟随重定向
	MaxRedirects       int           `yaml:"max_redirects" json:"max_redirects"`             // 最大重定向次数
	DisableCompression bool          `yaml:"disable_compression" json:"disable_compression"` // 禁用压缩
	EnableHTTP2        bool          `yaml:"enable_http2" json:"enable_http2"`               // 启用HTTP/2
	UserAgent          string        `yaml:"user_agent" json:"user_agent"`                   // User-Agent

	// 新增字段支持命令行配置
	Method      string            `yaml:"method" json:"method"`             // HTTP方法
	Path        string            `yaml:"path" json:"path"`                 // 请求路径
	Headers     map[string]string `yaml:"headers" json:"headers"`           // 请求头
	QueryParams map[string]string `yaml:"query_params" json:"query_params"` // 查询参数
}

// 实现interfaces.Config接口

// GetProtocol 获取协议名称
func (c *HttpAdapterConfig) GetProtocol() string {
	return "http"
}

// GetConnection 获取连接配置
func (c *HttpAdapterConfig) GetConnection() interfaces.ConnectionConfig {
	return &ConnectionConfigImpl{
		addresses:   []string{c.Connection.BaseURL},
		credentials: c.getCredentials(),
		poolConfig:  c.getPoolConfig(),
		timeout:     c.Connection.Timeout,
	}
}

// GetBenchmark 获取基准测试配置
func (c *HttpAdapterConfig) GetBenchmark() interfaces.BenchmarkConfig {
	return &c.Benchmark
}

// Validate 验证配置
func (c *HttpAdapterConfig) Validate() error {
	// 验证基础配置
	if c.Connection.BaseURL == "" {
		return fmt.Errorf("base_url cannot be empty")
	}

	// 验证连接配置
	if err := c.validateConnectionConfig(); err != nil {
		return fmt.Errorf("connection config validation failed: %w", err)
	}

	// 验证请求配置
	if err := c.validateRequestConfigs(); err != nil {
		return fmt.Errorf("request config validation failed: %w", err)
	}

	// 验证认证配置
	if err := c.validateAuthConfig(); err != nil {
		return fmt.Errorf("auth config validation failed: %w", err)
	}

	// 验证基准测试配置
	if err := c.validateBenchmarkConfig(); err != nil {
		return fmt.Errorf("benchmark config validation failed: %w", err)
	}

	return nil
}

// Clone 克隆配置
func (c *HttpAdapterConfig) Clone() interfaces.Config {
	clone := *c

	// 深拷贝切片和映射
	clone.Requests = make([]HttpRequestConfig, len(c.Requests))
	copy(clone.Requests, c.Requests)

	for i := range clone.Requests {
		if c.Requests[i].Headers != nil {
			clone.Requests[i].Headers = make(map[string]string)
			for k, v := range c.Requests[i].Headers {
				clone.Requests[i].Headers[k] = v
			}
		}
	}

	clone.Upload.AllowedTypes = make([]string, len(c.Upload.AllowedTypes))
	copy(clone.Upload.AllowedTypes, c.Upload.AllowedTypes)

	clone.Connection.TLS.CipherSuites = make([]string, len(c.Connection.TLS.CipherSuites))
	copy(clone.Connection.TLS.CipherSuites, c.Connection.TLS.CipherSuites)

	return &clone
}

// getCredentials 获取认证信息
func (c *HttpAdapterConfig) getCredentials() map[string]string {
	credentials := make(map[string]string)

	// 认证信息
	credentials["auth_type"] = c.Auth.Type
	if c.Auth.Username != "" {
		credentials["username"] = c.Auth.Username
	}
	if c.Auth.Password != "" {
		credentials["password"] = c.Auth.Password
	}
	if c.Auth.Token != "" {
		credentials["token"] = c.Auth.Token
	}

	// TLS认证信息
	if c.Connection.TLS.ClientAuth {
		credentials["tls_client_auth"] = "true"
		credentials["tls_cert_file"] = c.Connection.TLS.CertFile
		credentials["tls_key_file"] = c.Connection.TLS.KeyFile
		credentials["tls_ca_file"] = c.Connection.TLS.CAFile
	}

	return credentials
}

// getPoolConfig 获取连接池配置
func (c *HttpAdapterConfig) getPoolConfig() interfaces.PoolConfig {
	return &PoolConfigImpl{
		poolSize:          c.Connection.MaxConnsPerHost,
		minIdle:           c.Connection.MaxIdleConns / 2,
		maxIdle:           c.Connection.MaxIdleConns,
		idleTimeout:       c.Connection.IdleConnTimeout,
		connectionTimeout: c.Connection.Timeout,
	}
}

// validateConnectionConfig 验证连接配置
func (c *HttpAdapterConfig) validateConnectionConfig() error {
	if c.Connection.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	if c.Connection.MaxIdleConns <= 0 {
		return fmt.Errorf("max_idle_conns must be positive")
	}

	if c.Connection.MaxConnsPerHost <= 0 {
		return fmt.Errorf("max_conns_per_host must be positive")
	}

	// 验证TLS配置
	if c.Connection.TLS.ClientAuth {
		if c.Connection.TLS.CertFile == "" {
			return fmt.Errorf("cert_file is required when client_auth is enabled")
		}
		if c.Connection.TLS.KeyFile == "" {
			return fmt.Errorf("key_file is required when client_auth is enabled")
		}
	}

	// 验证TLS版本
	validVersions := []string{"1.0", "1.1", "1.2", "1.3"}
	if c.Connection.TLS.MinVersion != "" && !contains(validVersions, c.Connection.TLS.MinVersion) {
		return fmt.Errorf("invalid min_version: %s", c.Connection.TLS.MinVersion)
	}
	if c.Connection.TLS.MaxVersion != "" && !contains(validVersions, c.Connection.TLS.MaxVersion) {
		return fmt.Errorf("invalid max_version: %s", c.Connection.TLS.MaxVersion)
	}

	// 验证重新协商策略
	if c.Connection.TLS.Renegotiation != "" {
		validRenegotiation := []string{"never", "once", "freely"}
		if !contains(validRenegotiation, c.Connection.TLS.Renegotiation) {
			return fmt.Errorf("invalid renegotiation: %s", c.Connection.TLS.Renegotiation)
		}
	}

	return nil
}

// validateRequestConfigs 验证请求配置
func (c *HttpAdapterConfig) validateRequestConfigs() error {
	if len(c.Requests) == 0 {
		return fmt.Errorf("at least one request configuration is required")
	}

	for i, req := range c.Requests {
		if err := c.validateSingleRequestConfig(req, i); err != nil {
			return err
		}
	}

	return nil
}

// validateSingleRequestConfig 验证单个请求配置
func (c *HttpAdapterConfig) validateSingleRequestConfig(req HttpRequestConfig, index int) error {
	// 验证HTTP方法
	validMethods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS", "TRACE", "CONNECT"}
	if !contains(validMethods, req.Method) {
		return fmt.Errorf("invalid method in request[%d]: %s", index, req.Method)
	}

	// 验证路径
	if req.Path == "" {
		return fmt.Errorf("path cannot be empty in request[%d]", index)
	}

	// 验证权重
	if req.Weight < 0 {
		return fmt.Errorf("weight must be non-negative in request[%d]", index)
	}

	return nil
}

// validateAuthConfig 验证认证配置
func (c *HttpAdapterConfig) validateAuthConfig() error {
	validAuthTypes := []string{"none", "basic", "bearer", "oauth2", "mutual_tls"}
	if !contains(validAuthTypes, c.Auth.Type) {
		return fmt.Errorf("invalid auth type: %s", c.Auth.Type)
	}

	// 根据认证类型验证必要字段
	switch c.Auth.Type {
	case "basic":
		if c.Auth.Username == "" || c.Auth.Password == "" {
			return fmt.Errorf("username and password are required for basic auth")
		}
	case "bearer":
		if c.Auth.Token == "" {
			return fmt.Errorf("token is required for bearer auth")
		}
	case "mutual_tls":
		if c.Connection.TLS.CertFile == "" || c.Connection.TLS.KeyFile == "" {
			return fmt.Errorf("cert_file and key_file are required for mutual TLS auth")
		}
	}

	return nil
}

// validateBenchmarkConfig 验证基准测试配置
func (c *HttpAdapterConfig) validateBenchmarkConfig() error {
	if c.Benchmark.Total <= 0 {
		return fmt.Errorf("total must be positive")
	}

	if c.Benchmark.Parallels <= 0 {
		return fmt.Errorf("parallels must be positive")
	}

	if c.Benchmark.ReadPercent < 0 || c.Benchmark.ReadPercent > 100 {
		return fmt.Errorf("read_percent must be between 0 and 100")
	}

	if c.Benchmark.MaxRedirects < 0 {
		return fmt.Errorf("max_redirects must be non-negative")
	}

	return nil
}

// contains 检查字符串切片是否包含指定元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
