package config

import (
	"fmt"
	"os"
	"sort"

	"abc-runner/app/core/interfaces"
)

// ConfigManager 配置管理器
type ConfigManager struct {
	loader    *MultiSourceConfigLoader
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
			"config/templates/redis.yaml",
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

// MultiSourceConfigLoader 多源配置加载器
type MultiSourceConfigLoader struct {
	sources []ConfigSource
}

// NewMultiSourceConfigLoader 创建多源配置加载器
func NewMultiSourceConfigLoader(sources ...ConfigSource) *MultiSourceConfigLoader {
	return &MultiSourceConfigLoader{sources: sources}
}

// Load 从多个源加载配置
func (m *MultiSourceConfigLoader) Load() (interfaces.Config, error) {
	// 按优先级排序（高优先级在前）
	sources := make([]ConfigSource, len(m.sources))
	copy(sources, m.sources)
	sort.Slice(sources, func(i, j int) bool {
		return sources[i].Priority() > sources[j].Priority()
	})

	// 尝试从每个源加载配置
	for _, source := range sources {
		if source.CanLoad() {
			config, err := source.Load()
			if err != nil {
				continue
			}
			if config != nil {
				return config, nil
			}
		}
	}

	return nil, fmt.Errorf("no configuration source available")
}

// YAMLConfigSource YAML配置源类型定义（用于通用配置）
type YAMLConfigSource struct {
	FilePath string
}

// NewYAMLConfigSource 创建YAML配置源
func NewYAMLConfigSource(filePath string) *YAMLConfigSource {
	return &YAMLConfigSource{FilePath: filePath}
}

// Load 加载配置（默认实现返回错误）
func (y *YAMLConfigSource) Load() (interfaces.Config, error) {
	return nil, fmt.Errorf("generic YAML config source not implemented, use protocol-specific sources")
}

// CanLoad 检查是否可以加载
func (y *YAMLConfigSource) CanLoad() bool {
	_, err := os.Stat(y.FilePath)
	return err == nil
}

// Priority 获取优先级
func (y *YAMLConfigSource) Priority() int {
	return 1
}
func (m *ConfigManager) SetConfig(config interfaces.Config) {
	m.config = config
}

// LoadRedisConfig 加载Redis配置
func (m *ConfigManager) LoadRedisConfig(configPath string, args []string) error {
	sources := CreateRedisConfigSources(configPath, args)
	return m.LoadConfiguration(sources...)
}

// LoadRedisConfigFromFile 从文件加载Redis配置（向后兼容）
func LoadRedisConfigFromFile(filePath string) (interfaces.Config, error) {
	manager := NewConfigManager()
	err := manager.LoadRedisConfig(filePath, nil)
	if err != nil {
		return nil, err
	}
	return manager.GetConfig(), nil
}

// LoadRedisConfigFromArgs 从命令行参数加载Redis配置（向后兼容）
func LoadRedisConfigFromArgs(args []string) (interfaces.Config, error) {
	manager := NewConfigManager()
	err := manager.LoadRedisConfig("", args)
	if err != nil {
		return nil, err
	}
	return manager.GetConfig(), nil
}

// LoadRedisConfigFromMultipleSources 从多个源加载Redis配置（向后兼容）
func LoadRedisConfigFromMultipleSources(configPath string, args []string) (interfaces.Config, error) {
	manager := NewConfigManager()
	err := manager.LoadRedisConfig(configPath, args)
	if err != nil {
		return nil, err
	}
	return manager.GetConfig(), nil
}
