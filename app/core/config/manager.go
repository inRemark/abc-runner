package config

import (
	"fmt"

	"abc-runner/app/core/config/unified"
	"abc-runner/app/core/interfaces"
	"abc-runner/app/core/utils"
)

// ConfigManager 配置管理器
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

// ReloadConfiguration 重新加载配置
func (m *ConfigManager) ReloadConfiguration() error {
	// 重新加载配置的逻辑需要根据具体实现来定
	// 这里我们简单地返回一个错误，表示此功能未实现
	return fmt.Errorf("reload configuration not implemented")
}
