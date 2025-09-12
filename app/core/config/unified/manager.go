package unified

import (
	"fmt"
	"sort"

	"abc-runner/app/core/interfaces"
)

// StandardConfigManager 标准配置管理器
type StandardConfigManager struct {
	validator ConfigValidator
	config    interfaces.Config
}

// ConfigValidator 配置验证器接口
type ConfigValidator interface {
	Validate(config interfaces.Config) error
}

// NewStandardConfigManager 创建标准配置管理器
func NewStandardConfigManager(validator ConfigValidator) *StandardConfigManager {
	return &StandardConfigManager{
		validator: validator,
	}
}

// LoadConfiguration 从多个配置源加载配置
func (s *StandardConfigManager) LoadConfiguration(sources ...ConfigSource) error {
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

// GetConfig 获取当前配置
func (s *StandardConfigManager) GetConfig() interfaces.Config {
	return s.config
}

// ReloadConfiguration 重新加载配置
func (s *StandardConfigManager) ReloadConfiguration() error {
	// 重新加载逻辑与LoadConfiguration相同
	// 实际实现中可能需要保存配置源信息
	return fmt.Errorf("reload not implemented")
}

// SetConfig 设置配置（主要用于测试）
func (s *StandardConfigManager) SetConfig(config interfaces.Config) {
	s.config = config
}
