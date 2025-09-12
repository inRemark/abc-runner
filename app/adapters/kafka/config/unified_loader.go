package config

import (
	"abc-runner/app/core/config/unified"
	"abc-runner/app/core/interfaces"
)

// UnifiedKafkaConfigLoader 统一Kafka配置加载器
type UnifiedKafkaConfigLoader struct {
	factory *unified.ProtocolConfigFactory
	parser  unified.ConfigParser
	mapper  unified.EnvVarMapper
	argParser unified.ArgParser
}

// NewUnifiedKafkaConfigLoader 创建统一Kafka配置加载器
func NewUnifiedKafkaConfigLoader() *UnifiedKafkaConfigLoader {
	factory := unified.NewProtocolConfigFactory("kafka")
	parser := NewKafkaYAMLParser(LoadDefaultKafkaConfig())
	mapper := NewKafkaEnvVarMapper("KAFKA_RUNNER")
	argParser := NewKafkaArgParser()
	
	return &UnifiedKafkaConfigLoader{
		factory:   factory,
		parser:    parser,
		mapper:    mapper,
		argParser: argParser,
	}
}

// LoadConfig 加载Kafka配置
func (u *UnifiedKafkaConfigLoader) LoadConfig(configFile string, args []string) (interfaces.Config, error) {
	// Create config sources using the unified factory
	sources := u.factory.CreateConfigSources(
		configFile, 
		args, 
		func() interfaces.Config { return LoadDefaultKafkaConfig() },
		u.parser,
		u.mapper,
		u.argParser,
	)
	
	// Create a unified config manager
	validator := unified.NewSimpleConfigValidator()
	manager := unified.NewStandardConfigManager(validator)
	
	// Load configuration
	err := manager.LoadConfiguration(sources...)
	if err != nil {
		return nil, err
	}
	
	return manager.GetConfig(), nil
}

// CreateKafkaConfigSources 创建Kafka配置源（兼容旧接口）
func CreateKafkaConfigSources(configFile string, args []string) []KafkaConfigSource {
	// Create a mock config to get the sources
	factory := unified.NewProtocolConfigFactory("kafka")
	parser := NewKafkaYAMLParser(LoadDefaultKafkaConfig())
	mapper := NewKafkaEnvVarMapper("KAFKA_RUNNER")
	argParser := NewKafkaArgParser()
	
	sources := factory.CreateConfigSources(configFile, args, 
		func() interfaces.Config { return LoadDefaultKafkaConfig() },
		parser,
		mapper,
		argParser,
	)
	
	// Convert to KafkaConfigSource interface
	kafkaSources := make([]KafkaConfigSource, len(sources))
	for i, source := range sources {
		kafkaSources[i] = &KafkaConfigSourceAdapter{source: source}
	}
	
	return kafkaSources
}

// KafkaConfigSource Kafka配置源接口（兼容接口）
type KafkaConfigSource interface {
	Load() (interfaces.Config, error)
	CanLoad() bool
	Priority() int
}

// KafkaConfigSourceAdapter Kafka配置源适配器
type KafkaConfigSourceAdapter struct {
	source unified.ConfigSource
}

// Load 加载配置并适配为统一接口
func (k *KafkaConfigSourceAdapter) Load() (interfaces.Config, error) {
	kafkaConfig, err := k.source.Load()
	if err != nil {
		return nil, err
	}

	// 直接返回Kafka配置，因为KafkaAdapterConfig已经实现了interfaces.Config接口
	return kafkaConfig, nil
}

// CanLoad 检查是否可以加载
func (k *KafkaConfigSourceAdapter) CanLoad() bool {
	return k.source.CanLoad()
}

// Priority 获取优先级
func (k *KafkaConfigSourceAdapter) Priority() int {
	return k.source.Priority()
}