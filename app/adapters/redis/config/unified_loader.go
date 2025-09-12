package config

import (
	"abc-runner/app/core/config/unified"
	"abc-runner/app/core/interfaces"
)

// UnifiedRedisConfigLoader 统一Redis配置加载器
type UnifiedRedisConfigLoader struct {
	factory *unified.ProtocolConfigFactory
	parser  unified.ConfigParser
	mapper  unified.EnvVarMapper
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

// CreateRedisConfigSources 创建Redis配置源（兼容旧接口）
func CreateRedisConfigSources(configFile string, args []string) []RedisConfigSource {
	// Create a mock config to get the sources
	factory := unified.NewProtocolConfigFactory("redis")
	parser := NewRedisYAMLParser(NewDefaultRedisConfig())
	mapper := NewRedisEnvVarMapper("ABC_RUNNER")
	argParser := NewRedisArgParser()
	
	sources := factory.CreateConfigSources(configFile, args, 
		func() interfaces.Config { return NewDefaultRedisConfig() },
		parser,
		mapper,
		argParser,
	)
	
	// Convert to RedisConfigSource interface
	redisSources := make([]RedisConfigSource, len(sources))
	for i, source := range sources {
		redisSources[i] = &RedisConfigSourceAdapter{source: source}
	}
	
	return redisSources
}

// RedisConfigSource Redis配置源接口（兼容接口）
type RedisConfigSource interface {
	Load() (interfaces.Config, error)
	CanLoad() bool
	Priority() int
}

// RedisConfigSourceAdapter Redis配置源适配器
type RedisConfigSourceAdapter struct {
	source unified.ConfigSource
}

// Load 加载配置并适配为统一接口
func (r *RedisConfigSourceAdapter) Load() (interfaces.Config, error) {
	redisConfig, err := r.source.Load()
	if err != nil {
		return nil, err
	}

	// 直接返回Redis配置，因为RedisConfig已经实现了interfaces.Config接口
	return redisConfig, nil
}

// CanLoad 检查是否可以加载
func (r *RedisConfigSourceAdapter) CanLoad() bool {
	return r.source.CanLoad()
}

// Priority 获取优先级
func (r *RedisConfigSourceAdapter) Priority() int {
	return r.source.Priority()
}