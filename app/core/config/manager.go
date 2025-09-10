package config

import (
	"fmt"
	"sort"

	"abc-runner/app/core/interfaces"
	"abc-runner/app/core/utils"
)

// ConfigManager 配置管理器
type ConfigManager struct {
	loader     *MultiSourceConfigLoader
	validator  *ConfigValidator
	config     interfaces.Config
	coreConfig *CoreConfig
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

// LoadCoreConfiguration 加载核心配置
func (m *ConfigManager) LoadCoreConfiguration(coreConfigPath string) error {
	loader := NewCoreConfigLoader()

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

// GetConfig 获取配置
func (m *ConfigManager) GetConfig() interfaces.Config {
	return m.config
}

// GetCoreConfig 获取核心配置
func (m *ConfigManager) GetCoreConfig() *CoreConfig {
	if m.coreConfig == nil {
		// 返回默认核心配置
		loader := NewCoreConfigLoader()
		return loader.GetDefaultConfig()
	}
	return m.coreConfig
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

	// 验证配置
	if err := m.validator.Validate(config); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	m.config = config
	return nil
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

// ConfigSource 配置源接口 (已移至config_sources.go)
// type ConfigSource interface {
// 	Priority() int
// 	CanLoad() bool
// 	Load() (interfaces.Config, error)
// }

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

func (m *ConfigManager) SetConfig(config interfaces.Config) {
	m.config = config
}

// LoadRedisConfig 加载Redis配置
func (m *ConfigManager) LoadRedisConfig(configPath string, args []string) error {
	// 使用Redis包中的函数创建配置源
	redisSources := CreateRedisConfigSourcesInCore(configPath, args)
	return m.LoadConfiguration(redisSources...)
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

// LoadHttpConfigFromFile 从文件加载HTTP配置（向后兼容）
func LoadHttpConfigFromFile(filePath string) (interfaces.Config, error) {
	manager := NewConfigManager()
	err := manager.LoadHttpConfig(filePath, nil)
	if err != nil {
		return nil, err
	}
	return manager.GetConfig(), nil
}

// LoadHttpConfigFromArgs 从命令行参数加载HTTP配置（向后兼容）
func LoadHttpConfigFromArgs(args []string) (interfaces.Config, error) {
	manager := NewConfigManager()
	err := manager.LoadHttpConfig("", args)
	if err != nil {
		return nil, err
	}
	return manager.GetConfig(), nil
}

// LoadHttpConfigFromMultipleSources 从多个源加载HTTP配置（向后兼容）
func LoadHttpConfigFromMultipleSources(configPath string, args []string) (interfaces.Config, error) {
	manager := NewConfigManager()
	err := manager.LoadHttpConfig(configPath, args)
	if err != nil {
		return nil, err
	}
	return manager.GetConfig(), nil
}

// LoadKafkaConfigFromFile 从文件加载Kafka配置（向后兼容）
func LoadKafkaConfigFromFile(filePath string) (interfaces.Config, error) {
	manager := NewConfigManager()
	err := manager.LoadKafkaConfig(filePath, nil)
	if err != nil {
		return nil, err
	}
	return manager.GetConfig(), nil
}

// LoadKafkaConfigFromArgs 从命令行参数加载Kafka配置（向后兼容）
func LoadKafkaConfigFromArgs(args []string) (interfaces.Config, error) {
	manager := NewConfigManager()
	err := manager.LoadKafkaConfig("", args)
	if err != nil {
		return nil, err
	}
	return manager.GetConfig(), nil
}

// LoadKafkaConfigFromMultipleSources 从多个源加载Kafka配置（向后兼容）
func LoadKafkaConfigFromMultipleSources(configPath string, args []string) (interfaces.Config, error) {
	manager := NewConfigManager()
	err := manager.LoadKafkaConfig(configPath, args)
	if err != nil {
		return nil, err
	}
	return manager.GetConfig(), nil
}

// CreateRedisConfigSourcesInCore 创建Redis配置源列表（在core包中的实现）
func CreateRedisConfigSourcesInCore(yamlFile string, args []string) []ConfigSource {
	// 使用Redis包中的函数创建配置源
	redisSources := CreateRedisConfigSources(yamlFile, args)

	// 转换为core包中的ConfigSource接口
	sources := make([]ConfigSource, len(redisSources))
	for i, source := range redisSources {
		sources[i] = source
	}

	return sources
}

// CreateHttpConfigSourcesInCore 创建HTTP配置源列表（在core包中的实现）
func CreateHttpConfigSourcesInCore(yamlFile string, args []string) []ConfigSource {
	// 使用HTTP包中的函数创建配置源
	return CreateHttpConfigSources(yamlFile, args)
}

// CreateKafkaConfigSourcesInCore 创建Kafka配置源列表（在core包中的实现）
func CreateKafkaConfigSourcesInCore(yamlFile string, args []string) []ConfigSource {
	// 使用Kafka包中的函数创建配置源
	return CreateKafkaConfigSources(yamlFile, args)
}

// LoadHttpConfig 加载HTTP配置
func (m *ConfigManager) LoadHttpConfig(configPath string, args []string) error {
	// 使用HTTP包中的函数创建配置源
	httpSources := CreateHttpConfigSourcesInCore(configPath, args)
	return m.LoadConfiguration(httpSources...)
}

// LoadKafkaConfig 加载Kafka配置
func (m *ConfigManager) LoadKafkaConfig(configPath string, args []string) error {
	// 使用Kafka包中的函数创建配置源
	kafkaSources := CreateKafkaConfigSourcesInCore(configPath, args)
	return m.LoadConfiguration(kafkaSources...)
}

// ConfigValidator 配置验证器 (已移至multi_source_loader.go)
// type ConfigValidator struct {
// 	rules []ValidationRule
// }
//
// // ValidationRule 验证规则
// type ValidationRule func(interfaces.Config) error
//
// // NewConfigValidator 创建配置验证器
// func NewConfigValidator() *ConfigValidator {
// 	validator := &ConfigValidator{
// 		rules: make([]ValidationRule, 0),
// 	}
//
// 	// 添加默认验证规则
// 	validator.AddRule(validateProtocol)
// 	// 其他验证规则可以根据需要添加
//
// 	return validator
// }
//
// // AddRule 添加验证规则
// func (v *ConfigValidator) AddRule(rule ValidationRule) {
// 	v.rules = append(v.rules, rule)
// }
//
// // Validate 验证配置
// func (v *ConfigValidator) Validate(config interfaces.Config) error {
// 	for _, rule := range v.rules {
// 		if err := rule(config); err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }
//
// // validateProtocol 验证协议
// func validateProtocol(config interfaces.Config) error {
// 	protocol := config.GetProtocol()
// 	if protocol == "" {
// 		return fmt.Errorf("protocol cannot be empty")
// 	}
//
// 	// 简单的协议验证，可以根据需要扩展
// 	supportedProtocols := []string{"redis", "http", "kafka"}
// 	for _, supported := range supportedProtocols {
// 		if protocol == supported {
// 			return nil
// 		}
// 	}
//
// 	return fmt.Errorf("unsupported protocol: %s", protocol)
// }

// CommandLineConfigSource 命令行配置源 (已移至multi_source_loader.go)
// type CommandLineConfigSource struct {
// 	Args []string
// }
//
// // NewCommandLineConfigSource 创建命令行配置源
// func NewCommandLineConfigSource(args []string) *CommandLineConfigSource {
// 	return &CommandLineConfigSource{Args: args}
// }
//
// // Load 从命令行参数加载配置
// func (c *CommandLineConfigSource) Load() (interfaces.Config, error) {
// 	// 简化的实现，实际应该根据协议类型创建相应的配置
// 	// 这里返回一个基本的配置对象
// 	return nil, fmt.Errorf("command line config source not fully implemented")
// }
//
// // CanLoad 检查是否可以从命令行加载
// func (c *CommandLineConfigSource) CanLoad() bool {
// 	return len(c.Args) > 0
// }
//
// // Priority 获取优先级
// func (c *CommandLineConfigSource) Priority() int {
// 	return 3 // 命令行参数优先级最高
// }

// EnvironmentConfigSource 环境变量配置源 (已移至multi_source_loader.go)
// type EnvironmentConfigSource struct {
// 	Prefix string
// }
//
// // NewEnvironmentConfigSource 创建环境变量配置源
// func NewEnvironmentConfigSource(prefix string) *EnvironmentConfigSource {
// 	if prefix == "" {
// 		prefix = "ABC_RUNNER"
// 	}
// 	return &EnvironmentConfigSource{Prefix: prefix}
// }
//
// // Load 从环境变量加载配置
// func (e *EnvironmentConfigSource) Load() (interfaces.Config, error) {
// 	// 简化的实现，实际应该根据协议类型创建相应的配置
// 	// 这里返回一个基本的配置对象
// 	return nil, fmt.Errorf("environment config source not fully implemented")
// }
//
// // CanLoad 检查是否可以从环境变量加载
// func (e *EnvironmentConfigSource) CanLoad() bool {
// 	// 检查关键环境变量是否存在
// 	_, exists := os.LookupEnv(e.Prefix + "_PROTOCOL")
// 	return exists
// }
//
// // Priority 获取优先级
// func (e *EnvironmentConfigSource) Priority() int {
// 	return 2 // 环境变量优先级较高
// }
