package unified

import (
	"abc-runner/app/core/interfaces"
)

// ConfigSource 统一配置源接口
type ConfigSource interface {
	// Load 加载配置
	Load() (interfaces.Config, error)
	
	// CanLoad 检查是否可以加载配置
	CanLoad() bool
	
	// Priority 获取配置源优先级
	Priority() int
}

// ConfigManager 统一配置管理器接口
type ConfigManager interface {
	// LoadConfiguration 从多个配置源加载配置
	LoadConfiguration(sources ...ConfigSource) error
	
	// GetConfig 获取当前配置
	GetConfig() interfaces.Config
	
	// ReloadConfiguration 重新加载配置
	ReloadConfiguration() error
}

// Config 统一配置接口
type Config interface {
	// GetProtocol 获取协议名称
	GetProtocol() string
	
	// GetConnection 获取连接配置
	GetConnection() interfaces.ConnectionConfig
	
	// GetBenchmark 获取基准测试配置
	GetBenchmark() interfaces.BenchmarkConfig
	
	// Validate 验证配置
	Validate() error
	
	// Clone 克隆配置
	Clone() interfaces.Config
}