package unified

import (
	"abc-runner/app/core/interfaces"
	"fmt"
	"sort"
)

// ConfigValidator 配置验证器接口
type ConfigValidator interface {
	Validate(config interfaces.Config) error
}

// NewStandardConfigManager 创建标准配置管理器（兼容性函数）
func NewStandardConfigManager(validator ConfigValidator) ConfigManager {
	return &StandardConfigManagerImpl{
		validator: validator,
	}
}

// StandardConfigManagerImpl 标准配置管理器实现
type StandardConfigManagerImpl struct {
	validator ConfigValidator
	config    interfaces.Config
}

// 实现 ConfigManager 接口
func (s *StandardConfigManagerImpl) LoadConfiguration(sources ...ConfigSource) error {
	// 按优先级排序配置源（低优先级在前，高优先级在后）
	sort.Slice(sources, func(i, j int) bool {
		return sources[i].Priority() < sources[j].Priority()
	})

	// 依次加载配置，后面的配置源会覆盖前面的配置
	var baseConfig interfaces.Config
	var lastError error
	for _, source := range sources {
		if source.CanLoad() {
			// 如果是环境变量或命令行参数源，设置基础配置
			if envSource, ok := source.(*EnvConfigSource); ok && baseConfig != nil {
				envSource.SetConfig(baseConfig)
			}
			if argSource, ok := source.(*ArgConfigSource); ok && baseConfig != nil {
				argSource.SetConfig(baseConfig)
			}
			config, err := source.Load()
			if err != nil {
				lastError = err
				continue
			}
			// 验证配置
			if err := s.validator.Validate(config); err != nil {
				lastError = err
				continue
			}
			// 更新基础配置
			baseConfig = config
		}
	}
	if baseConfig != nil {
		s.config = baseConfig
		return nil
	}
	if lastError != nil {
		return lastError
	}
	return fmt.Errorf("no configuration source could be loaded")
}

func (s *StandardConfigManagerImpl) GetConfig() interfaces.Config {
	return s.config
}

func (s *StandardConfigManagerImpl) ReloadConfiguration() error {
	return fmt.Errorf("reload not implemented")
}

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
