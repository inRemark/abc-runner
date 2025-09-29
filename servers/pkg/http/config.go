package http

import (
	"fmt"
	"time"

	"abc-runner/servers/internal/common"
	"abc-runner/servers/pkg/interfaces"
)

// HTTPServerConfig HTTP服务端配置
type HTTPServerConfig struct {
	*common.BaseConfig `yaml:",inline"`

	// HTTP服务器配置
	ReadTimeout     time.Duration `yaml:"read_timeout" json:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout" json:"write_timeout"`
	IdleTimeout     time.Duration `yaml:"idle_timeout" json:"idle_timeout"`
	MaxHeaderBytes  int           `yaml:"max_header_bytes" json:"max_header_bytes"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" json:"shutdown_timeout"`

	// 响应配置
	Response ResponseConfig `yaml:"response" json:"response"`

	// 路由配置
	Routes []RouteConfig `yaml:"routes" json:"routes"`

	// CORS配置
	CORS CORSConfig `yaml:"cors" json:"cors"`

	// TLS配置
	TLS TLSConfig `yaml:"tls" json:"tls"`
}

// ResponseConfig 响应配置
type ResponseConfig struct {
	DefaultDelay      time.Duration `yaml:"default_delay" json:"default_delay"`
	DefaultStatusCode int           `yaml:"default_status_code" json:"default_status_code"`
	DefaultSize       int           `yaml:"default_size" json:"default_size"`
	ContentType       string        `yaml:"content_type" json:"content_type"`
	EnableCompression bool          `yaml:"enable_compression" json:"enable_compression"`
}

// RouteConfig 路由配置
type RouteConfig struct {
	Path        string                 `yaml:"path" json:"path"`
	Method      string                 `yaml:"method" json:"method"`
	StatusCode  int                    `yaml:"status_code" json:"status_code"`
	Delay       time.Duration          `yaml:"delay" json:"delay"`
	ContentType string                 `yaml:"content_type" json:"content_type"`
	Response    map[string]interface{} `yaml:"response" json:"response"`
}

// CORSConfig CORS配置
type CORSConfig struct {
	Enabled          bool     `yaml:"enabled" json:"enabled"`
	AllowedOrigins   []string `yaml:"allowed_origins" json:"allowed_origins"`
	AllowedMethods   []string `yaml:"allowed_methods" json:"allowed_methods"`
	AllowedHeaders   []string `yaml:"allowed_headers" json:"allowed_headers"`
	ExposedHeaders   []string `yaml:"exposed_headers" json:"exposed_headers"`
	AllowCredentials bool     `yaml:"allow_credentials" json:"allow_credentials"`
	MaxAge           int      `yaml:"max_age" json:"max_age"`
}

// TLSConfig TLS配置
type TLSConfig struct {
	Enabled  bool   `yaml:"enabled" json:"enabled"`
	CertFile string `yaml:"cert_file" json:"cert_file"`
	KeyFile  string `yaml:"key_file" json:"key_file"`
}

// NewHTTPServerConfig 创建HTTP服务端配置
func NewHTTPServerConfig() *HTTPServerConfig {
	return &HTTPServerConfig{
		BaseConfig: &common.BaseConfig{
			Protocol: "http",
			Host:     "localhost",
			Port:     8080,
		},
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
		IdleTimeout:     60 * time.Second,
		MaxHeaderBytes:  1048576, // 1MB
		ShutdownTimeout: 30 * time.Second,
		Response: ResponseConfig{
			DefaultDelay:      0,
			DefaultStatusCode: 200,
			DefaultSize:       1024,
			ContentType:       "application/json",
			EnableCompression: false,
		},
		Routes: []RouteConfig{
			{
				Path:        "/",
				Method:      "GET",
				StatusCode:  200,
				Delay:       0,
				ContentType: "application/json",
				Response: map[string]interface{}{
					"message":   "Hello from abc-runner HTTP test server!",
					"protocol":  "http",
					"timestamp": 0,
				},
			},
			{
				Path:        "/health",
				Method:      "GET",
				StatusCode:  200,
				Delay:       0,
				ContentType: "application/json",
				Response: map[string]interface{}{
					"status":   "ok",
					"protocol": "http",
				},
			},
		},
		CORS: CORSConfig{
			Enabled:        true,
			AllowedOrigins: []string{"*"},
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS"},
			AllowedHeaders: []string{"*"},
			MaxAge:         3600,
		},
		TLS: TLSConfig{
			Enabled: false,
		},
	}
}

// Validate 验证HTTP配置
func (c *HTTPServerConfig) Validate() error {
	if err := c.BaseConfig.Validate(); err != nil {
		return fmt.Errorf("base config validation failed: %w", err)
	}

	if c.ReadTimeout <= 0 {
		return fmt.Errorf("read_timeout must be positive")
	}

	if c.WriteTimeout <= 0 {
		return fmt.Errorf("write_timeout must be positive")
	}

	if c.IdleTimeout <= 0 {
		return fmt.Errorf("idle_timeout must be positive")
	}

	if c.MaxHeaderBytes <= 0 {
		return fmt.Errorf("max_header_bytes must be positive")
	}

	if c.Response.DefaultStatusCode < 100 || c.Response.DefaultStatusCode > 599 {
		return fmt.Errorf("default_status_code must be between 100 and 599")
	}

	if c.Response.DefaultSize < 0 {
		return fmt.Errorf("default_size cannot be negative")
	}

	// 验证路由配置
	for i, route := range c.Routes {
		if route.Path == "" {
			return fmt.Errorf("route[%d]: path cannot be empty", i)
		}
		if route.Method == "" {
			return fmt.Errorf("route[%d]: method cannot be empty", i)
		}
		if route.StatusCode < 100 || route.StatusCode > 599 {
			return fmt.Errorf("route[%d]: status_code must be between 100 and 599", i)
		}
	}

	// 验证TLS配置
	if c.TLS.Enabled {
		if c.TLS.CertFile == "" {
			return fmt.Errorf("tls enabled but cert_file is empty")
		}
		if c.TLS.KeyFile == "" {
			return fmt.Errorf("tls enabled but key_file is empty")
		}
	}

	return nil
}

// Clone 克隆HTTP配置
func (c *HTTPServerConfig) Clone() interfaces.ServerConfig {
	clone := *c
	clone.BaseConfig = c.BaseConfig.Clone().(*common.BaseConfig)

	// 深拷贝路由
	clone.Routes = make([]RouteConfig, len(c.Routes))
	copy(clone.Routes, c.Routes)

	// 深拷贝CORS配置
	clone.CORS.AllowedOrigins = make([]string, len(c.CORS.AllowedOrigins))
	copy(clone.CORS.AllowedOrigins, c.CORS.AllowedOrigins)

	clone.CORS.AllowedMethods = make([]string, len(c.CORS.AllowedMethods))
	copy(clone.CORS.AllowedMethods, c.CORS.AllowedMethods)

	clone.CORS.AllowedHeaders = make([]string, len(c.CORS.AllowedHeaders))
	copy(clone.CORS.AllowedHeaders, c.CORS.AllowedHeaders)

	clone.CORS.ExposedHeaders = make([]string, len(c.CORS.ExposedHeaders))
	copy(clone.CORS.ExposedHeaders, c.CORS.ExposedHeaders)

	return &clone
}

// RequestInfo HTTP请求信息
type RequestInfo struct {
	Method     string            `json:"method"`
	Path       string            `json:"path"`
	RemoteAddr string            `json:"remote_addr"`
	UserAgent  string            `json:"user_agent"`
	Headers    map[string]string `json:"headers"`
	Timestamp  time.Time         `json:"timestamp"`
	Duration   time.Duration     `json:"duration"`
	StatusCode int               `json:"status_code"`
	BodySize   int               `json:"body_size"`
}

// ResponseInfo HTTP响应信息
type ResponseInfo struct {
	StatusCode   int               `json:"status_code"`
	Headers      map[string]string `json:"headers"`
	BodySize     int               `json:"body_size"`
	ContentType  string            `json:"content_type"`
	ResponseTime time.Duration     `json:"response_time"`
	Compressed   bool              `json:"compressed"`
}
