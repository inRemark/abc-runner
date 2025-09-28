package grpc

import (
	"fmt"
	"time"

	"abc-runner/servers/internal/common"
	"abc-runner/servers/pkg/interfaces"
)

// GRPCServerConfig gRPC服务端配置
type GRPCServerConfig struct {
	*common.BaseConfig `yaml:",inline"`
	
	// gRPC特定配置
	MaxRecvMessageSize int           `yaml:"max_recv_message_size" json:"max_recv_message_size"`
	MaxSendMessageSize int           `yaml:"max_send_message_size" json:"max_send_message_size"`
	ConnectionTimeout  time.Duration `yaml:"connection_timeout" json:"connection_timeout"`
	MaxConcurrentStreams uint32      `yaml:"max_concurrent_streams" json:"max_concurrent_streams"`
	
	// TLS配置
	TLS TLSConfig `yaml:"tls" json:"tls"`
	
	// 认证配置
	Auth AuthConfig `yaml:"auth" json:"auth"`
	
	// 健康检查配置
	HealthCheck HealthCheckConfig `yaml:"health_check" json:"health_check"`
	
	// 反射配置
	EnableReflection bool `yaml:"enable_reflection" json:"enable_reflection"`
	
	// 日志配置
	LogRequests bool `yaml:"log_requests" json:"log_requests"`
}

// TLSConfig TLS配置
type TLSConfig struct {
	Enabled  bool   `yaml:"enabled" json:"enabled"`
	CertFile string `yaml:"cert_file" json:"cert_file"`
	KeyFile  string `yaml:"key_file" json:"key_file"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	Enabled    bool   `yaml:"enabled" json:"enabled"`
	AuthToken  string `yaml:"auth_token" json:"auth_token"`
	RequireAuth bool  `yaml:"require_auth" json:"require_auth"`
}

// HealthCheckConfig 健康检查配置
type HealthCheckConfig struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
}

// NewGRPCServerConfig 创建gRPC服务端配置
func NewGRPCServerConfig() *GRPCServerConfig {
	return &GRPCServerConfig{
		BaseConfig: &common.BaseConfig{
			Protocol: "grpc",
			Host:     "localhost",
			Port:     50051,
		},
		MaxRecvMessageSize:   4 * 1024 * 1024, // 4MB
		MaxSendMessageSize:   4 * 1024 * 1024, // 4MB
		ConnectionTimeout:    30 * time.Second,
		MaxConcurrentStreams: 100,
		TLS: TLSConfig{
			Enabled: false,
		},
		Auth: AuthConfig{
			Enabled:     false,
			RequireAuth: false,
		},
		HealthCheck: HealthCheckConfig{
			Enabled: true,
		},
		EnableReflection: true,
		LogRequests:      true,
	}
}

// Validate 验证gRPC配置
func (c *GRPCServerConfig) Validate() error {
	if err := c.BaseConfig.Validate(); err != nil {
		return fmt.Errorf("base config validation failed: %w", err)
	}
	
	if c.MaxRecvMessageSize <= 0 {
		return fmt.Errorf("max_recv_message_size must be positive")
	}
	
	if c.MaxSendMessageSize <= 0 {
		return fmt.Errorf("max_send_message_size must be positive")
	}
	
	if c.ConnectionTimeout <= 0 {
		return fmt.Errorf("connection_timeout must be positive")
	}
	
	if c.MaxConcurrentStreams == 0 {
		return fmt.Errorf("max_concurrent_streams must be positive")
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

// Clone 克隆gRPC配置
func (c *GRPCServerConfig) Clone() interfaces.ServerConfig {
	clone := *c
	clone.BaseConfig = c.BaseConfig.Clone().(*common.BaseConfig)
	return &clone
}

// RequestInfo gRPC请求信息
type RequestInfo struct {
	Method     string            `json:"method"`
	Headers    map[string]string `json:"headers"`
	RemoteAddr string            `json:"remote_addr"`
	Timestamp  time.Time         `json:"timestamp"`
	Duration   time.Duration     `json:"duration"`
	Status     string            `json:"status"`
}

// ServiceMethods 支持的服务方法
var ServiceMethods = map[string]string{
	"Echo":                "echo service - returns the input message",
	"ServerStream":        "server streaming - sends multiple responses",
	"ClientStream":        "client streaming - receives multiple requests",
	"BidirectionalStream": "bidirectional streaming - full duplex communication",
	"Health":              "health check service",
}