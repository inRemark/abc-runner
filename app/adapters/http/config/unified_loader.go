package config

import (
	"abc-runner/app/core/config/unified"
	"abc-runner/app/core/interfaces"
)

// UnifiedHttpConfigLoader 统一HTTP配置加载器
type UnifiedHttpConfigLoader struct {
	factory   *unified.ProtocolConfigFactory
	parser    unified.ConfigParser
	mapper    unified.EnvVarMapper
	argParser unified.ArgParser
}

// NewUnifiedHttpConfigLoader 创建统一HTTP配置加载器
func NewUnifiedHttpConfigLoader() *UnifiedHttpConfigLoader {
	factory := unified.NewProtocolConfigFactory("http")
	parser := NewHttpYAMLParser(LoadDefaultHttpConfig())
	mapper := NewHttpEnvVarMapper("HTTP_RUNNER")
	argParser := NewHttpArgParser()

	return &UnifiedHttpConfigLoader{
		factory:   factory,
		parser:    parser,
		mapper:    mapper,
		argParser: argParser,
	}
}

// LoadConfig 加载HTTP配置
func (u *UnifiedHttpConfigLoader) LoadConfig(configFile string, args []string) (interfaces.Config, error) {
	// Create config sources using the unified factory
	sources := u.factory.CreateConfigSources(
		configFile,
		args,
		func() interfaces.Config { return LoadDefaultHttpConfig() },
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
