package config

import (
	"fmt"
	"sort"

	"abc-runner/app/core/config/unified"
	"abc-runner/app/core/interfaces"
)

// KafkaConfigManager Kafka配置管理器
type KafkaConfigManager struct {
	manager unified.ConfigManager
	config  interfaces.Config
}

// NewKafkaConfigManager 创建Kafka配置管理器
func NewKafkaConfigManager() *KafkaConfigManager {
	validator := unified.NewSimpleConfigValidator()
	manager := unified.NewStandardConfigManager(validator)
	return &KafkaConfigManager{
		manager: manager,
	}
}

// LoadConfiguration 加载配置
func (m *KafkaConfigManager) LoadConfiguration(sources ...unified.ConfigSource) error {
	// 按优先级排序配置源
	sortedSources := make([]unified.ConfigSource, len(sources))
	copy(sortedSources, sources)

	sort.Slice(sortedSources, func(i, j int) bool {
		return sortedSources[i].Priority() > sortedSources[j].Priority()
	})

	err := m.manager.LoadConfiguration(sortedSources...)
	if err != nil {
		return fmt.Errorf("failed to load Kafka configuration: %w", err)
	}

	m.config = m.manager.GetConfig()
	return nil
}

// LoadFromFile 从文件加载配置
func (m *KafkaConfigManager) LoadFromFile(filePath string) error {
	loader := NewUnifiedKafkaConfigLoader()
	config, err := loader.LoadConfig(filePath, nil)
	if err != nil {
		return err
	}

	m.config = config
	return nil
}

// LoadFromArgs 从命令行参数加载配置
func (m *KafkaConfigManager) LoadFromArgs(args []string) error {
	loader := NewUnifiedKafkaConfigLoader()
	config, err := loader.LoadConfig("", args)
	if err != nil {
		return err
	}

	m.config = config
	return nil
}

// LoadFromMultipleSources 从多个源加载配置
func (m *KafkaConfigManager) LoadFromMultipleSources(configPath string, args []string) error {
	loader := NewUnifiedKafkaConfigLoader()
	config, err := loader.LoadConfig(configPath, args)
	if err != nil {
		return fmt.Errorf("failed to load Kafka configuration: %w", err)
	}

	m.config = config
	return nil
}

// GetConfig 获取配置
func (m *KafkaConfigManager) GetConfig() *KafkaAdapterConfig {
	if m.config == nil {
		return LoadDefaultKafkaConfig()
	}

	if kafkaConfig, ok := m.config.(*KafkaAdapterConfig); ok {
		return kafkaConfig
	}

	return LoadDefaultKafkaConfig()
}

// SetConfig 设置配置
func (m *KafkaConfigManager) SetConfig(config *KafkaAdapterConfig) error {
	m.config = config
	return nil
}

// ReloadConfiguration 重新加载配置
func (m *KafkaConfigManager) ReloadConfiguration() error {
	return fmt.Errorf("reload not implemented in unified config system")
}

// ValidateConfiguration 验证当前配置
func (m *KafkaConfigManager) ValidateConfiguration() error {
	if m.config == nil {
		return fmt.Errorf("no configuration loaded")
	}

	return m.config.Validate()
}

// IsConfigurationLoaded 检查配置是否已加载
func (m *KafkaConfigManager) IsConfigurationLoaded() bool {
	return m.config != nil
}

// GetConnectionInfo 获取连接信息摘要
func (m *KafkaConfigManager) GetConnectionInfo() map[string]interface{} {
	config := m.GetConfig()
	info := make(map[string]interface{})

	info["protocol"] = config.GetProtocol()
	info["brokers"] = config.Brokers
	info["topic"] = config.Benchmark.DefaultTopic

	return info
}

// GetBenchmarkInfo 获取基准测试信息摘要
func (m *KafkaConfigManager) GetBenchmarkInfo() map[string]interface{} {
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
