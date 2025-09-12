package config

import (
	"fmt"
	"sort"

	"abc-runner/app/core/config/unified"
	"abc-runner/app/core/interfaces"
)

// HttpConfigManager HTTP配置管理器
type HttpConfigManager struct {
	manager unified.ConfigManager
	config  interfaces.Config
}

// NewHttpConfigManager 创建HTTP配置管理器
func NewHttpConfigManager() *HttpConfigManager {
	validator := unified.NewSimpleConfigValidator()
	manager := unified.NewStandardConfigManager(validator)
	return &HttpConfigManager{
		manager: manager,
	}
}

// LoadConfiguration 加载配置
func (m *HttpConfigManager) LoadConfiguration(sources ...unified.ConfigSource) error {
	// 按优先级排序配置源
	sortedSources := make([]unified.ConfigSource, len(sources))
	copy(sortedSources, sources)

	sort.Slice(sortedSources, func(i, j int) bool {
		return sortedSources[i].Priority() > sortedSources[j].Priority()
	})

	err := m.manager.LoadConfiguration(sortedSources...)
	if err != nil {
		return fmt.Errorf("failed to load HTTP configuration: %w", err)
	}

	m.config = m.manager.GetConfig()
	return nil
}

// LoadFromFile 从文件加载配置
func (m *HttpConfigManager) LoadFromFile(filePath string) error {
	loader := NewUnifiedHttpConfigLoader()
	config, err := loader.LoadConfig(filePath, nil)
	if err != nil {
		return err
	}

	m.config = config
	return nil
}

// LoadFromArgs 从命令行参数加载配置
func (m *HttpConfigManager) LoadFromArgs(args []string) error {
	loader := NewUnifiedHttpConfigLoader()
	config, err := loader.LoadConfig("", args)
	if err != nil {
		return err
	}

	m.config = config
	return nil
}

// LoadFromMultipleSources 从多个源加载配置
func (m *HttpConfigManager) LoadFromMultipleSources(configPath string, args []string) error {
	loader := NewUnifiedHttpConfigLoader()
	config, err := loader.LoadConfig(configPath, args)
	if err != nil {
		return fmt.Errorf("failed to load HTTP configuration: %w", err)
	}

	m.config = config
	return nil
}

// GetConfig 获取配置
func (m *HttpConfigManager) GetConfig() *HttpAdapterConfig {
	if m.config == nil {
		return LoadDefaultHttpConfig()
	}

	if httpConfig, ok := m.config.(*HttpAdapterConfig); ok {
		return httpConfig
	}

	return LoadDefaultHttpConfig()
}

// SetConfig 设置配置
func (m *HttpConfigManager) SetConfig(config *HttpAdapterConfig) error {
	m.config = config
	return nil
}

// ReloadConfiguration 重新加载配置
func (m *HttpConfigManager) ReloadConfiguration() error {
	return fmt.Errorf("reload not implemented in unified config system")
}

// ValidateConfiguration 验证当前配置
func (m *HttpConfigManager) ValidateConfiguration() error {
	if m.config == nil {
		return fmt.Errorf("no configuration loaded")
	}

	return m.config.Validate()
}

// IsConfigurationLoaded 检查配置是否已加载
func (m *HttpConfigManager) IsConfigurationLoaded() bool {
	return m.config != nil
}

// GetConnectionInfo 获取连接信息摘要
func (m *HttpConfigManager) GetConnectionInfo() map[string]interface{} {
	config := m.GetConfig()
	info := make(map[string]interface{})

	info["protocol"] = config.GetProtocol()
	info["base_url"] = config.Connection.BaseURL
	info["method"] = config.Benchmark.Method

	return info
}

// GetBenchmarkInfo 获取基准测试信息摘要
func (m *HttpConfigManager) GetBenchmarkInfo() map[string]interface{} {
	config := m.GetConfig()
	benchmark := config.GetBenchmark()

	return map[string]interface{}{
		"total":        benchmark.GetTotal(),
		"parallels":    benchmark.GetParallels(),
		"data_size":    benchmark.GetDataSize(),
		"read_percent": benchmark.GetReadPercent(),
		"random_keys":  benchmark.GetRandomKeys(),
		"test_case":    benchmark.GetTestCase(),
	}
}
