package config

import (
	"abc-runner/app/core/config/unified"
	"abc-runner/app/core/interfaces"
)

// UnifiedKafkaConfigLoader 统一Kafka配置加载器
type UnifiedKafkaConfigLoader struct {
	factory   *unified.ProtocolConfigFactory
	parser    unified.ConfigParser
	mapper    unified.EnvVarMapper
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
