package config

import (
	"abc-runner/app/core/interfaces"
	"abc-runner/app/core/utils"
)

// CreateRedisConfigSources 创建Redis配置源（兼容旧接口）
func CreateRedisConfigSources(configFile string, args []string) []RedisConfigSource {
	sources := make([]RedisConfigSource, 0)

	// 1. 默认配置源（最低优先级）
	sources = append(sources, &RedisConfigSourceAdapter{
		source: NewDefaultConfigSource(),
	})

	// 2. YAML配置文件
	if configFile != "" {
		sources = append(sources, &RedisConfigSourceAdapter{
			source: NewYAMLConfigSource(configFile),
		})
	} else {
		// 使用统一的配置文件查找机制
		foundPath := utils.FindConfigFile("redis")
		if foundPath != "" {
			sources = append(sources, &RedisConfigSourceAdapter{
				source: NewYAMLConfigSource(foundPath),
			})
		}
	}

	// 3. 环境变量配置源
	sources = append(sources, &RedisConfigSourceAdapter{
		source: NewEnvConfigSource("ABC_RUNNER"),
	})

	// 4. 命令行参数配置源（最高优先级）
	if len(args) > 0 {
		sources = append(sources, &RedisConfigSourceAdapter{
			source: NewCommandLineConfigSource(args),
		})
	}

	return sources
}

// RedisConfigSource Redis配置源接口（兼容接口）
type RedisConfigSource interface {
	Load() (interfaces.Config, error)
	CanLoad() bool
	Priority() int
}

// RedisConfigSourceAdapter Redis配置源适配器
type RedisConfigSourceAdapter struct {
	source ConfigSource
}

// Load 加载配置并适配为统一接口
func (r *RedisConfigSourceAdapter) Load() (interfaces.Config, error) {
	redisConfig, err := r.source.Load()
	if err != nil {
		return nil, err
	}

	return NewRedisConfigAdapter(redisConfig), nil
}

// CanLoad 检查是否可以加载
func (r *RedisConfigSourceAdapter) CanLoad() bool {
	return r.source.CanLoad()
}

// Priority 获取优先级
func (r *RedisConfigSourceAdapter) Priority() int {
	return r.source.Priority()
}

// LoadRedisConfigurationFromSources 从配置源加载Redis配置
func LoadRedisConfigurationFromSources(sources ...RedisConfigSource) (interfaces.Config, error) {
	// 创建Redis配置管理器
	manager := NewRedisConfigManager()

	// 转换为原生配置源
	nativeSources := make([]ConfigSource, len(sources))
	for i, source := range sources {
		if adapter, ok := source.(*RedisConfigSourceAdapter); ok {
			nativeSources[i] = adapter.source
		}
	}

	// 加载配置
	err := manager.LoadConfiguration(nativeSources...)
	if err != nil {
		return nil, err
	}

	// 返回适配器
	return manager.GetAdapter(), nil
}

// LoadRedisConfigFromArgs 从命令行参数加载Redis配置（兼容函数）
func LoadRedisConfigFromArgs(args []string) (interfaces.Config, error) {
	manager := NewRedisConfigManager()
	err := manager.LoadFromArgs(args)
	if err != nil {
		return nil, err
	}

	return manager.GetAdapter(), nil
}

// LoadRedisConfigFromFile 从文件加载Redis配置（兼容函数）
func LoadRedisConfigFromFile(filePath string) (interfaces.Config, error) {
	manager := NewRedisConfigManager()
	err := manager.LoadFromFile(filePath)
	if err != nil {
		return nil, err
	}

	return manager.GetAdapter(), nil
}

// CreateRedisConfigManager 创建Redis配置管理器（工厂函数）
func CreateRedisConfigManager() *RedisConfigManager {
	return NewRedisConfigManager()
}
