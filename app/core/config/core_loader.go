package config

import (
	"fmt"
	"os"

	"abc-runner/app/core/config/unified"

	"gopkg.in/yaml.v2"
)

// UnifiedCoreConfigLoader 统一核心配置加载器
type UnifiedCoreConfigLoader struct {
	validator unified.ConfigValidator
	config    *CoreConfig
}

// NewUnifiedCoreConfigLoader 创建统一核心配置加载器
func NewUnifiedCoreConfigLoader() *UnifiedCoreConfigLoader {
	validator := unified.NewSimpleConfigValidator()
	return &UnifiedCoreConfigLoader{
		validator: validator,
	}
}

// LoadFromFile 从文件加载核心配置
func (u *UnifiedCoreConfigLoader) LoadFromFile(filePath string) (*CoreConfig, error) {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// 如果文件不存在，返回默认配置
		return u.GetDefaultConfig(), nil
	}

	// 读取文件内容
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read core config file: %w", err)
	}

	// 解析YAML
	var config CoreConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse core config file: %w", err)
	}

	u.config = &config
	return u.config, nil
}

// GetConfig 获取配置
func (u *UnifiedCoreConfigLoader) GetConfig() *CoreConfig {
	if u.config == nil {
		return u.GetDefaultConfig()
	}
	return u.config
}

// GetDefaultConfig 获取默认核心配置
func (u *UnifiedCoreConfigLoader) GetDefaultConfig() *CoreConfig {
	return &CoreConfig{
		Core: CoreConfigSection{
			Logging: LoggingConfig{
				Level:      "info",
				Format:     "json",
				Output:     "stdout",
				FilePath:   "./logs",
				MaxSize:    "100MB",
				MaxAge:     7,
				MaxBackups: 5,
				Compress:   true,
			},
			Reports: ReportsConfig{
				Enabled:             true,
				Formats:             []string{"console"},
				OutputDir:           "./reports",
				FilePrefix:          "benchmark",
				IncludeTimestamp:    true,
				EnableConsoleReport: true,
				OverwriteExisting:   false,
			},
			Monitoring: MonitoringConfig{
				Enabled:         true,
				MetricsInterval: 5 * 1000000000, // 5 seconds in nanoseconds
				Prometheus: PrometheusConfig{
					Enabled: false,
					Port:    9090,
				},
				Statsd: StatsdConfig{
					Enabled: false,
					Host:    "localhost:8125",
				},
			},
			Connection: ConnectionConfig{
				Timeout:         30 * 1000000000, // 30 seconds in nanoseconds
				KeepAlive:       30 * 1000000000, // 30 seconds in nanoseconds
				MaxIdleConns:    100,
				IdleConnTimeout: 90 * 1000000000, // 90 seconds in nanoseconds
			},
		},
	}
}

// SaveToFile 保存配置到文件
func (u *UnifiedCoreConfigLoader) SaveToFile(config *CoreConfig, filePath string) error {
	// 序列化配置
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal core config: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write core config file: %w", err)
	}

	return nil
}
