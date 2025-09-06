package config

import (
	"fmt"
	"sort"
)

// RedisConfigManager Redis配置管理器
type RedisConfigManager struct {
	loader    ConfigLoader
	validator *ConfigValidator
	config    *RedisConfig
}

// NewRedisConfigManager 创建Redis配置管理器
func NewRedisConfigManager() *RedisConfigManager {
	return &RedisConfigManager{
		validator: NewConfigValidator(),
	}
}

// LoadConfiguration 加载配置
func (m *RedisConfigManager) LoadConfiguration(sources ...ConfigSource) error {
	// 按优先级排序配置源
	sort.Slice(sources, func(i, j int) bool {
		return sources[i].Priority() > sources[j].Priority()
	})
	
	m.loader = NewMultiSourceConfigLoader(sources...)
	
	config, err := m.loader.Load()
	if err != nil {
		return fmt.Errorf("failed to load Redis configuration: %w", err)
	}
	
	// 验证配置
	if err := m.validator.Validate(config); err != nil {
		return fmt.Errorf("Redis configuration validation failed: %w", err)
	}
	
	m.config = config
	return nil
}

// LoadFromFile 从文件加载配置
func (m *RedisConfigManager) LoadFromFile(filePath string) error {
	sources := []ConfigSource{
		NewDefaultConfigSource(),
		NewYAMLConfigSource(filePath),
	}
	return m.LoadConfiguration(sources...)
}

// LoadFromArgs 从命令行参数加载配置
func (m *RedisConfigManager) LoadFromArgs(args []string) error {
	sources := []ConfigSource{
		NewDefaultConfigSource(),
		NewEnvConfigSource("REDIS_RUNNER"),
		NewCommandLineConfigSource(args),
	}
	return m.LoadConfiguration(sources...)
}

// LoadFromMultipleSources 从多个源加载配置
func (m *RedisConfigManager) LoadFromMultipleSources(configPath string, args []string) error {
	loader := CreateStandardLoader(configPath, args)
	m.loader = loader
	
	config, err := loader.Load()
	if err != nil {
		return fmt.Errorf("failed to load Redis configuration: %w", err)
	}
	
	// 验证配置
	if err := m.validator.Validate(config); err != nil {
		return fmt.Errorf("Redis configuration validation failed: %w", err)
	}
	
	m.config = config
	return nil
}

// GetConfig 获取配置
func (m *RedisConfigManager) GetConfig() *RedisConfig {
	if m.config == nil {
		return NewDefaultRedisConfig()
	}
	return m.config
}

// SetConfig 设置配置
func (m *RedisConfigManager) SetConfig(config *RedisConfig) error {
	if err := m.validator.Validate(config); err != nil {
		return fmt.Errorf("Redis configuration validation failed: %w", err)
	}
	
	m.config = config
	return nil
}

// ReloadConfiguration 重新加载配置
func (m *RedisConfigManager) ReloadConfiguration() error {
	if m.loader == nil {
		return fmt.Errorf("no configuration loader available")
	}
	
	config, err := m.loader.Load()
	if err != nil {
		return fmt.Errorf("failed to reload Redis configuration: %w", err)
	}
	
	if err := m.validator.Validate(config); err != nil {
		return fmt.Errorf("Redis configuration validation failed: %w", err)
	}
	
	m.config = config
	return nil
}

// AddValidationRule 添加验证规则
func (m *RedisConfigManager) AddValidationRule(rule ValidationRule) {
	m.validator.AddRule(rule)
}

// GetAdapter 获取适配器以兼容统一接口
func (m *RedisConfigManager) GetAdapter() *RedisConfigAdapter {
	return NewRedisConfigAdapter(m.GetConfig())
}

// ValidateConfiguration 验证当前配置
func (m *RedisConfigManager) ValidateConfiguration() error {
	if m.config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	
	return m.validator.Validate(m.config)
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