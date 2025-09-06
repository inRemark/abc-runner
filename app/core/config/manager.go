package config

import (
	"fmt"
	"sort"

	"redis-runner/app/core/interfaces"
)

// ConfigManager 配置管理器
type ConfigManager struct {
	loader    *ConfigLoader
	validator *ConfigValidator
	config    interfaces.Config
}

// NewConfigManager 创建配置管理器
func NewConfigManager() *ConfigManager {
	return &ConfigManager{
		validator: NewConfigValidator(),
	}
}

// LoadConfiguration 加载配置
func (m *ConfigManager) LoadConfiguration(sources ...ConfigSource) error {
	// 按优先级排序配置源
	sort.Slice(sources, func(i, j int) bool {
		return sources[i].Priority() > sources[j].Priority()
	})
	
	m.loader = NewMultiSourceConfigLoader(sources...)
	
	config, err := m.loader.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	
	// 验证配置
	if err := m.validator.Validate(config); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}
	
	m.config = config
	return nil
}

// GetConfig 获取配置
func (m *ConfigManager) GetConfig() interfaces.Config {
	return m.config
}

// ReloadConfiguration 重新加载配置
func (m *ConfigManager) ReloadConfiguration() error {
	if m.loader == nil {
		return fmt.Errorf("no configuration loader available")
	}
	
	config, err := m.loader.Load()
	if err != nil {
		return fmt.Errorf("failed to reload configuration: %w", err)
	}
	
	if err := m.validator.Validate(config); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}
	
	m.config = config
	return nil
}

// AddValidationRule 添加验证规则
func (m *ConfigManager) AddValidationRule(rule ValidationRule) {
	m.validator.AddRule(rule)
}

// CreateDefaultSources 创建默认配置源
func CreateDefaultSources(configFile string, args []string) []ConfigSource {
	sources := make([]ConfigSource, 0)
	
	// 1. 命令行参数（最高优先级）
	if len(args) > 0 {
		sources = append(sources, NewCommandLineConfigSource(args))
	}
	
	// 2. 环境变量
	sources = append(sources, NewEnvironmentConfigSource("REDIS_RUNNER"))
	
	// 3. YAML配置文件（最低优先级）
	if configFile != "" {
		sources = append(sources, NewYAMLConfigSource(configFile))
	} else {
		// 尝试默认路径
		defaultPaths := []string{
			"conf/redis.yaml",
			"conf/redis-config.yaml", 
			"redis.yaml",
			"config.yaml",
		}
		
		for _, path := range defaultPaths {
			yamlSource := NewYAMLConfigSource(path)
			if yamlSource.CanLoad() {
				sources = append(sources, yamlSource)
				break
			}
		}
	}
	
	return sources
}

// LoadFromFile 从文件加载配置（向后兼容）
func LoadFromFile(filePath string) (interfaces.Config, error) {
	source := NewYAMLConfigSource(filePath)
	return source.Load()
}

// LoadFromArgs 从命令行参数加载配置（向后兼容）
func LoadFromArgs(args []string) (interfaces.Config, error) {
	source := NewCommandLineConfigSource(args)
	return source.Load()
}

// LoadFromEnv 从环境变量加载配置
func LoadFromEnv(prefix string) (interfaces.Config, error) {
	source := NewEnvironmentConfigSource(prefix)
	return source.Load()
}