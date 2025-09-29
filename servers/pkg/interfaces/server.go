package interfaces

import (
	"context"
	"net"
	"time"
)

// Server 通用服务端接口
type Server interface {
	// Start 启动服务端
	Start(ctx context.Context) error

	// Stop 停止服务端
	Stop(ctx context.Context) error

	// GetProtocol 获取协议名称
	GetProtocol() string

	// GetAddress 获取监听地址
	GetAddress() string

	// GetPort 获取监听端口
	GetPort() int

	// IsRunning 检查服务端是否正在运行
	IsRunning() bool

	// HealthCheck 健康检查
	HealthCheck(ctx context.Context) error

	// GetMetrics 获取服务端指标
	GetMetrics() map[string]interface{}

	// GetConfig 获取服务端配置
	GetConfig() ServerConfig
}

// ServerConfig 通用服务端配置接口
type ServerConfig interface {
	// GetProtocol 获取协议名称
	GetProtocol() string

	// GetHost 获取监听主机
	GetHost() string

	// GetPort 获取监听端口
	GetPort() int

	// GetAddress 获取完整监听地址
	GetAddress() string

	// Validate 验证配置
	Validate() error

	// Clone 克隆配置
	Clone() ServerConfig
}

// ConnectionHandler 连接处理器接口
type ConnectionHandler interface {
	// HandleConnection 处理连接
	HandleConnection(ctx context.Context, conn net.Conn) error

	// GetProtocol 获取协议名称
	GetProtocol() string
}

// RequestHandler 请求处理器接口
type RequestHandler interface {
	// HandleRequest 处理请求
	HandleRequest(ctx context.Context, request interface{}) (interface{}, error)

	// GetSupportedOperations 获取支持的操作类型
	GetSupportedOperations() []string
}

// HealthChecker 健康检查器接口
type HealthChecker interface {
	// Check 执行健康检查
	Check(ctx context.Context) error

	// GetStatus 获取健康状态
	GetStatus() HealthStatus
}

// HealthStatus 健康状态
type HealthStatus struct {
	Status    string            `json:"status"`    // "healthy", "degraded", "unhealthy"
	Timestamp time.Time         `json:"timestamp"` // 检查时间
	Details   map[string]string `json:"details"`   // 详细信息
	Duration  time.Duration     `json:"duration"`  // 检查耗时
}

// MetricsCollector 指标收集器接口
type MetricsCollector interface {
	// RecordRequest 记录请求
	RecordRequest(protocol string, operation string, duration time.Duration, success bool)

	// RecordConnection 记录连接
	RecordConnection(protocol string, action string) // action: "open", "close"

	// RecordError 记录错误
	RecordError(protocol string, operation string, errorType string)

	// GetMetrics 获取指标快照
	GetMetrics() map[string]interface{}

	// Reset 重置指标
	Reset()
}

// Logger 日志接口
type Logger interface {
	// Debug 调试日志
	Debug(msg string, fields ...map[string]interface{})

	// Info 信息日志
	Info(msg string, fields ...map[string]interface{})

	// Warn 警告日志
	Warn(msg string, fields ...map[string]interface{})

	// Error 错误日志
	Error(msg string, err error, fields ...map[string]interface{})

	// Fatal 致命错误日志
	Fatal(msg string, err error, fields ...map[string]interface{})
}

// ConfigLoader 配置加载器接口
type ConfigLoader interface {
	// LoadConfig 加载配置
	LoadConfig(configPath string) (ServerConfig, error)

	// LoadConfigWithDefaults 加载配置并应用默认值
	LoadConfigWithDefaults(configPath string, defaults ServerConfig) (ServerConfig, error)

	// ValidateConfig 验证配置
	ValidateConfig(config ServerConfig) error
}

// ServerRegistry 服务端注册表接口
type ServerRegistry interface {
	// Register 注册服务端
	Register(server Server) error

	// Unregister 注销服务端
	Unregister(protocol string) error

	// Get 获取服务端
	Get(protocol string) (Server, error)

	// GetAll 获取所有服务端
	GetAll() []Server

	// StartAll 启动所有服务端
	StartAll(ctx context.Context) error

	// StopAll 停止所有服务端
	StopAll(ctx context.Context) error
}

// HandlerRegistry 处理器注册表接口
type HandlerRegistry interface {
	// RegisterConnectionHandler 注册连接处理器
	RegisterConnectionHandler(protocol string, handler ConnectionHandler) error

	// RegisterRequestHandler 注册请求处理器
	RegisterRequestHandler(protocol string, operation string, handler RequestHandler) error

	// GetConnectionHandler 获取连接处理器
	GetConnectionHandler(protocol string) (ConnectionHandler, error)

	// GetRequestHandler 获取请求处理器
	GetRequestHandler(protocol string, operation string) (RequestHandler, error)
}

// ServerFactory 服务端工厂接口
type ServerFactory interface {
	// CreateHTTPServer 创建HTTP服务端
	CreateHTTPServer(config ServerConfig, logger Logger, metricsCollector MetricsCollector) (Server, error)

	// CreateTCPServer 创建TCP服务端
	CreateTCPServer(config ServerConfig, logger Logger, metricsCollector MetricsCollector) (Server, error)

	// CreateUDPServer 创建UDP服务端
	CreateUDPServer(config ServerConfig, logger Logger, metricsCollector MetricsCollector) (Server, error)

	// CreateGRPCServer 创建gRPC服务端
	CreateGRPCServer(config ServerConfig, logger Logger, metricsCollector MetricsCollector) (Server, error)
}

// ServerOptions 服务端选项
type ServerOptions struct {
	Config           ServerConfig
	Logger           Logger
	MetricsCollector MetricsCollector
	HealthChecker    HealthChecker
	Context          context.Context
}
