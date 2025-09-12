package config

import (
	"abc-runner/app/core/config/unified"
	"abc-runner/app/core/interfaces"
)

// UnifiedRedisConfigLoader 统一Redis配置加载器
type UnifiedRedisConfigLoader struct {
	factory   *unified.ProtocolConfigFactory
	parser    unified.ConfigParser
	mapper    unified.EnvVarMapper
	argParser unified.ArgParser
}

// NewUnifiedRedisConfigLoader 创建统一Redis配置加载器
func NewUnifiedRedisConfigLoader() *UnifiedRedisConfigLoader {
	factory := unified.NewProtocolConfigFactory("redis")
	parser := NewRedisYAMLParser(NewDefaultRedisConfig())
	mapper := NewRedisEnvVarMapper("ABC_RUNNER")
	argParser := NewRedisArgParser()

	return &UnifiedRedisConfigLoader{
		factory:   factory,
		parser:    parser,
		mapper:    mapper,
		argParser: argParser,
	}
}

// LoadConfig 加载Redis配置
func (u *UnifiedRedisConfigLoader) LoadConfig(configFile string, args []string) (interfaces.Config, error) {
	// Create config sources using the unified factory
	sources := u.factory.CreateConfigSources(
		configFile,
		args,
		func() interfaces.Config { return NewDefaultRedisConfig() },
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
