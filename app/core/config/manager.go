package config

import (
	"abc-runner/app/core/config/unified"
	"abc-runner/app/core/interfaces"
	"abc-runner/app/core/utils"
	"fmt"
	"sort"
)

// ConfigManager 统一配置管理器（合并了StandardConfigManager功能）
type ConfigManager struct {
	config     interfaces.Config
	coreConfig *CoreConfig
	validator  unified.ConfigValidator
}

// NewConfigManager 创建配置管理器
func NewConfigManager(validator unified.ConfigValidator) *ConfigManager {
	if validator == nil {
		validator = unified.NewSimpleConfigValidator()
	}
	return &ConfigManager{
		validator: validator,
	}
}

// GetConfig 获取配置
func (m *ConfigManager) GetConfig() interfaces.Config {
	return m.config
}

// SetConfig 设置配置（主要用于测试）
func (m *ConfigManager) SetConfig(config interfaces.Config) {
	m.config = config
}

// LoadCoreConfiguration 加载核心配置
func (m *ConfigManager) LoadCoreConfiguration(coreConfigPath string) error {
	loader := NewUnifiedCoreConfigLoader()

	// 如果没有指定核心配置路径，使用统一的查找机制
	if coreConfigPath == "" {
		coreConfigPath = utils.FindCoreConfigFile()
		// 如果找不到，使用默认路径
		if coreConfigPath == "" {
			coreConfigPath = "config/core.yaml"
		}
	}

	coreConfig, err := loader.LoadFromFile(coreConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load core configuration: %w", err)
	}

	m.coreConfig = coreConfig
	return nil
}

// GetCoreConfig 获取核心配置
func (m *ConfigManager) GetCoreConfig() *CoreConfig {
	if m.coreConfig == nil {
		// 返回默认核心配置
		loader := NewUnifiedCoreConfigLoader()
		return loader.GetDefaultConfig()
	}
	return m.coreConfig
}

// LoadConfiguration 从多个配置源加载配置（合并自StandardConfigManager）
func (m *ConfigManager) LoadConfiguration(sources ...unified.ConfigSource) error {
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
			if envSource, ok := source.(*unified.EnvConfigSource); ok && baseConfig != nil {
				envSource.SetConfig(baseConfig)
			}
			if argSource, ok := source.(*unified.ArgConfigSource); ok && baseConfig != nil {
				argSource.SetConfig(baseConfig)
			}
			config, err := source.Load()
			if err != nil {
				lastError = err
				continue
			}
			// 验证配置
			if err := m.validator.Validate(config); err != nil {
				lastError = err
				continue
			}
			// 更新基础配置
			baseConfig = config
		}
	}
	if baseConfig != nil {
		m.config = baseConfig
		return nil
	}
	if lastError != nil {
		return lastError
	}
	return fmt.Errorf("no configuration source could be loaded")
}

// ReloadConfiguration 重新加载配置
func (m *ConfigManager) ReloadConfiguration() error {
	// 重新加载逻辑与LoadConfiguration相同
	// 实际实现中可能需要保存配置源信息
	return fmt.Errorf("reload not implemented")
}
