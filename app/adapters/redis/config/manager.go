package config

import (
	"fmt"
	"sort"

	"abc-runner/app/core/config/unified"
	"abc-runner/app/core/interfaces"
)

// RedisConfigManager Redis配置管理器
type RedisConfigManager struct {
	manager unified.ConfigManager
	config  interfaces.Config
}

// NewRedisConfigManager 创建Redis配置管理器
func NewRedisConfigManager() *RedisConfigManager {
	validator := unified.NewSimpleConfigValidator()
	manager := unified.NewStandardConfigManager(validator)
	return &RedisConfigManager{
		manager: manager,
	}
}

// LoadConfiguration 加载配置
func (m *RedisConfigManager) LoadConfiguration(sources ...unified.ConfigSource) error {
	// 按优先级排序配置源
	sortedSources := make([]unified.ConfigSource, len(sources))
	copy(sortedSources, sources)

	sort.Slice(sortedSources, func(i, j int) bool {
		return sortedSources[i].Priority() > sortedSources[j].Priority()
	})

	err := m.manager.LoadConfiguration(sortedSources...)
	if err != nil {
		return fmt.Errorf("failed to load Redis configuration: %w", err)
	}

	m.config = m.manager.GetConfig()
	return nil
}

// LoadFromFile 从文件加载配置
func (m *RedisConfigManager) LoadFromFile(filePath string) error {
	loader := NewUnifiedRedisConfigLoader()
	config, err := loader.LoadConfig(filePath, nil)
	if err != nil {
		return err
	}

	m.config = config
	return nil
}

// LoadFromArgs 从命令行参数加载配置
func (m *RedisConfigManager) LoadFromArgs(args []string) error {
	loader := NewUnifiedRedisConfigLoader()
	config, err := loader.LoadConfig("", args)
	if err != nil {
		return err
	}

	m.config = config
	return nil
}

// LoadFromMultipleSources 从多个源加载配置
func (m *RedisConfigManager) LoadFromMultipleSources(configPath string, args []string) error {
	loader := NewUnifiedRedisConfigLoader()
	config, err := loader.LoadConfig(configPath, args)
	if err != nil {
		return fmt.Errorf("failed to load Redis configuration: %w", err)
	}

	m.config = config
	return nil
}

// GetConfig 获取配置
func (m *RedisConfigManager) GetConfig() *RedisConfig {
	if m.config == nil {
		return NewDefaultRedisConfig()
	}

	if redisConfig, ok := m.config.(*RedisConfig); ok {
		return redisConfig
	}

	return NewDefaultRedisConfig()
}

// SetConfig 设置配置
func (m *RedisConfigManager) SetConfig(config *RedisConfig) error {
	m.config = config
	return nil
}

// ReloadConfiguration 重新加载配置
func (m *RedisConfigManager) ReloadConfiguration() error {
	return fmt.Errorf("reload not implemented in unified config system")
}

// ValidateConfiguration 验证当前配置
func (m *RedisConfigManager) ValidateConfiguration() error {
	if m.config == nil {
		return fmt.Errorf("no configuration loaded")
	}

	return m.config.Validate()
}

// IsConfigurationLoaded 检查配置是否已加载
func (m *RedisConfigManager) IsConfigurationLoaded() bool {
	return m.config != nil
}

// GetConnectionInfo 获取连接信息摘要
func (m *RedisConfigManager) GetConnectionInfo() map[string]interface{} {
	config := m.GetConfig()
	info := make(map[string]interface{})

	info["protocol"] = config.GetProtocol()
	info["mode"] = config.GetMode()

	switch config.GetMode() {
	case "standalone":
		info["addr"] = config.Standalone.Addr
		info["db"] = config.Standalone.Db
	case "sentinel":
		info["addrs"] = config.Sentinel.Addrs
		info["master_name"] = config.Sentinel.MasterName
		info["db"] = config.Sentinel.Db
	case "cluster":
		info["addrs"] = config.Cluster.Addrs
	}

	return info
}

// GetBenchmarkInfo 获取基准测试信息摘要
func (m *RedisConfigManager) GetBenchmarkInfo() map[string]interface{} {
	config := m.GetConfig()
	benchmark := config.GetBenchmark()

	return map[string]interface{}{
		"total":        benchmark.GetTotal(),
		"parallels":    benchmark.GetParallels(),
		"data_size":    benchmark.GetDataSize(),
		"ttl":          benchmark.GetTTL(),
		"read_percent": benchmark.GetReadPercent(),
		"random_keys":  benchmark.GetRandomKeys(),
		"test_case":    benchmark.GetTestCase(),
	}
}
