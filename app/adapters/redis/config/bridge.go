package config

import (
	"abc-runner/app/core/interfaces"
	"abc-runner/app/core/config/unified"
)

// LoadRedisConfigurationFromSources 从配置源加载Redis配置
func LoadRedisConfigurationFromSources(sources ...RedisConfigSource) (interfaces.Config, error) {
	// 创建Redis配置管理器
	manager := NewRedisConfigManager()

	// 转换为原生配置源
	nativeSources := make([]unified.ConfigSource, len(sources))
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

	// 直接返回配置，因为RedisConfig已经实现了interfaces.Config接口
	return manager.GetConfig(), nil
}

// LoadRedisConfigFromArgs 从命令行参数加载Redis配置（兼容函数）
func LoadRedisConfigFromArgs(args []string) (interfaces.Config, error) {
	manager := NewRedisConfigManager()
	err := manager.LoadFromArgs(args)
	if err != nil {
		return nil, err
	}

	// 直接返回配置，因为RedisConfig已经实现了interfaces.Config接口
	return manager.GetConfig(), nil
}

// LoadRedisConfigFromFile 从文件加载Redis配置（兼容函数）
func LoadRedisConfigFromFile(filePath string) (interfaces.Config, error) {
	manager := NewRedisConfigManager()
	err := manager.LoadFromFile(filePath)
	if err != nil {
		return nil, err
	}

	// 直接返回配置，因为RedisConfig已经实现了interfaces.Config接口
	return manager.GetConfig(), nil
}

// CreateRedisConfigManager 创建Redis配置管理器（工厂函数）
func CreateRedisConfigManager() *RedisConfigManager {
	return NewRedisConfigManager()
}
