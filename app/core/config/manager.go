package config

import (
	"fmt"
	"sort"

	httpconfig "abc-runner/app/adapters/http/config"
	kafkaconfig "abc-runner/app/adapters/kafka/config"
	redisconfig "abc-runner/app/adapters/redis/config"
	"abc-runner/app/core/config/unified"
	"abc-runner/app/core/interfaces"
	"abc-runner/app/core/utils"
)

// ConfigManager 配置管理器
type ConfigManager struct {
	config      interfaces.Config
	coreConfig  *CoreConfig
	validator   unified.ConfigValidator
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

// UnifiedConfigSourceAdapter 统一配置源适配器
type UnifiedConfigSourceAdapter struct {
	source ConfigSource
}

// Load 加载配置
func (u *UnifiedConfigSourceAdapter) Load() (interfaces.Config, error) {
	return u.source.Load()
}

// CanLoad 检查是否可以加载
func (u *UnifiedConfigSourceAdapter) CanLoad() bool {
	return u.source.CanLoad()
}

// Priority 获取优先级
func (u *UnifiedConfigSourceAdapter) Priority() int {
	return u.source.Priority()
}

// LoadConfiguration 加载配置
func (m *ConfigManager) LoadConfiguration(sources ...ConfigSource) error {
	// 按优先级排序配置源
	sort.Slice(sources, func(i, j int) bool {
		return sources[i].Priority() > sources[j].Priority()
	})

	// 创建统一的配置管理器
	validator := unified.NewSimpleConfigValidator()
	manager := unified.NewStandardConfigManager(validator)
	
	// 转换配置源为统一配置源
	unifiedSources := make([]unified.ConfigSource, len(sources))
	for i, source := range sources {
		unifiedSources[i] = &UnifiedConfigSourceAdapter{source: source}
	}

	err := manager.LoadConfiguration(unifiedSources...)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	config := manager.GetConfig()
	
	// 验证配置
	if err := m.validator.Validate(config); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	m.config = config
	return nil
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
	// 重新加载配置的逻辑需要根据具体实现来定
	// 这里我们简单地返回一个错误，表示此功能未实现
	return fmt.Errorf("reload configuration not implemented")
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

func (m *ConfigManager) SetConfig(config interfaces.Config) {
	m.config = config
}

// RedisConfigSourceAdapter Redis配置源适配器
type RedisConfigSourceAdapter struct {
	source redisconfig.RedisConfigSource
}

// Load 加载配置
func (r *RedisConfigSourceAdapter) Load() (interfaces.Config, error) {
	return r.source.Load()
}

// CanLoad 检查是否可以加载
func (r *RedisConfigSourceAdapter) CanLoad() bool {
	return r.source.CanLoad()
}

// Priority 获取优先级
func (r *RedisConfigSourceAdapter) Priority() int {
	return r.source.Priority()
}

// KafkaConfigSourceAdapter Kafka配置源适配器
type KafkaConfigSourceAdapter struct {
	source kafkaconfig.KafkaConfigSource
}

// Load 加载配置
func (k *KafkaConfigSourceAdapter) Load() (interfaces.Config, error) {
	return k.source.Load()
}

// CanLoad 检查是否可以加载
func (k *KafkaConfigSourceAdapter) CanLoad() bool {
	return k.source.CanLoad()
}

// Priority 获取优先级
func (k *KafkaConfigSourceAdapter) Priority() int {
	return k.source.Priority()
}

// CreateRedisConfigSourcesInCore 创建Redis配置源列表（在core包中的实现）
func CreateRedisConfigSourcesInCore(yamlFile string, args []string) []ConfigSource {
	// 使用Redis包中的函数创建配置源
	redisSources := redisconfig.CreateRedisConfigSources(yamlFile, args)
	
	// 转换为core包中的ConfigSource接口
	coreSources := make([]ConfigSource, len(redisSources))
	for i, source := range redisSources {
		coreSources[i] = &RedisConfigSourceAdapter{source: source}
	}
	
	return coreSources
}

// CreateHttpConfigSourcesInCore 创建HTTP配置源列表（在core包中的实现）
func CreateHttpConfigSourcesInCore(yamlFile string, args []string) []ConfigSource {
	// 使用HTTP包中的函数创建配置源
	httpSources := httpconfig.CreateHttpConfigSources(yamlFile, args)
	
	// 转换为core包中的ConfigSource接口
	coreSources := make([]ConfigSource, len(httpSources))
	for i, source := range httpSources {
		coreSources[i] = &HttpConfigSourceAdapter{source: source}
	}
	
	return coreSources
}

// CreateKafkaConfigSourcesInCore 创建Kafka配置源列表（在core包中的实现）
func CreateKafkaConfigSourcesInCore(yamlFile string, args []string) []ConfigSource {
	// 使用Kafka包中的函数创建配置源
	kafkaSources := kafkaconfig.CreateKafkaConfigSources(yamlFile, args)
	
	// 转换为core包中的ConfigSource接口
	coreSources := make([]ConfigSource, len(kafkaSources))
	for i, source := range kafkaSources {
		coreSources[i] = &KafkaConfigSourceAdapter{source: source}
	}
	
	return coreSources
}

// LoadRedisConfig 加载Redis配置
func (m *ConfigManager) LoadRedisConfig(configPath string, args []string) error {
	// 使用Redis包中的函数创建配置源
	redisSources := CreateRedisConfigSourcesInCore(configPath, args)
	return m.LoadConfiguration(redisSources...)
}

// LoadRedisConfigFromFile 从文件加载Redis配置（向后兼容）
func LoadRedisConfigFromFile(filePath string) (interfaces.Config, error) {
	manager := NewConfigManager(unified.NewSimpleConfigValidator()) // 使用默认验证器
	err := manager.LoadRedisConfig(filePath, nil)
	if err != nil {
		return nil, err
	}
	return manager.GetConfig(), nil
}

// LoadRedisConfigFromArgs 从命令行参数加载Redis配置（向后兼容）
func LoadRedisConfigFromArgs(args []string) (interfaces.Config, error) {
	manager := NewConfigManager(unified.NewSimpleConfigValidator()) // 使用默认验证器
	err := manager.LoadRedisConfig("", args)
	if err != nil {
		return nil, err
	}
	return manager.GetConfig(), nil
}

// LoadRedisConfigFromMultipleSources 从多个源加载Redis配置（向后兼容）
func LoadRedisConfigFromMultipleSources(configPath string, args []string) (interfaces.Config, error) {
	manager := NewConfigManager(unified.NewSimpleConfigValidator()) // 使用默认验证器
	err := manager.LoadRedisConfig(configPath, args)
	if err != nil {
		return nil, err
	}
	return manager.GetConfig(), nil
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