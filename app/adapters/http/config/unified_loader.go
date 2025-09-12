package config

import (
	"abc-runner/app/core/config/unified"
	"abc-runner/app/core/interfaces"
)

// UnifiedHttpConfigLoader 统一HTTP配置加载器
type UnifiedHttpConfigLoader struct {
	factory *unified.ProtocolConfigFactory
	parser  unified.ConfigParser
	mapper  unified.EnvVarMapper
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

// CreateHttpConfigSources 创建HTTP配置源（兼容旧接口）
func CreateHttpConfigSources(configFile string, args []string) []HttpConfigSource {
	// Create a mock config to get the sources
	factory := unified.NewProtocolConfigFactory("http")
	parser := NewHttpYAMLParser(LoadDefaultHttpConfig())
	mapper := NewHttpEnvVarMapper("HTTP_RUNNER")
	argParser := NewHttpArgParser()
	
	sources := factory.CreateConfigSources(configFile, args, 
		func() interfaces.Config { return LoadDefaultHttpConfig() },
		parser,
		mapper,
		argParser,
	)
	
	// Convert to HttpConfigSource interface
	httpSources := make([]HttpConfigSource, len(sources))
	for i, source := range sources {
		httpSources[i] = &HttpConfigSourceAdapter{source: source}
	}
	
	return httpSources
}

// HttpConfigSource HTTP配置源接口（兼容接口）
type HttpConfigSource interface {
	Load() (interfaces.Config, error)
	CanLoad() bool
	Priority() int
}

// HttpConfigSourceAdapter HTTP配置源适配器
type HttpConfigSourceAdapter struct {
	source unified.ConfigSource
}

// Load 加载配置并适配为统一接口
func (h *HttpConfigSourceAdapter) Load() (interfaces.Config, error) {
	httpConfig, err := h.source.Load()
	if err != nil {
		return nil, err
	}

	// 直接返回HTTP配置，因为HttpAdapterConfig已经实现了interfaces.Config接口
	return httpConfig, nil
}

// CanLoad 检查是否可以加载
func (h *HttpConfigSourceAdapter) CanLoad() bool {
	return h.source.CanLoad()
}

// Priority 获取优先级
func (h *HttpConfigSourceAdapter) Priority() int {
	return h.source.Priority()
}