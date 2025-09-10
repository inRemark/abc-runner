package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

// CoreConfig 核心配置结构体
type CoreConfig struct {
	Core CoreConfigSection `yaml:"core"`
}

// CoreConfigSection 核心配置部分
type CoreConfigSection struct {
	Logging    LoggingConfig    `yaml:"logging"`
	Reports    ReportsConfig    `yaml:"reports"`
	Monitoring MonitoringConfig `yaml:"monitoring"`
	Connection ConnectionConfig `yaml:"connection"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level      string `yaml:"level"`
	Format     string `yaml:"format"`
	Output     string `yaml:"output"`
	FilePath   string `yaml:"file_path"`
	MaxSize    string `yaml:"max_size"`
	MaxAge     int    `yaml:"max_age"`
	MaxBackups int    `yaml:"max_backups"`
	Compress   bool   `yaml:"compress"`
}

// ReportsConfig 报告配置
type ReportsConfig struct {
	Enabled             bool     `yaml:"enabled"`
	Formats             []string `yaml:"formats"`
	OutputDir           string   `yaml:"output_dir"`
	FilePrefix          string   `yaml:"file_prefix"`
	IncludeTimestamp    bool     `yaml:"include_timestamp"`
	EnableConsoleReport bool     `yaml:"enable_console_report"`
	OverwriteExisting   bool     `yaml:"overwrite_existing"`
}

// MonitoringConfig 监控配置
type MonitoringConfig struct {
	Enabled         bool              `yaml:"enabled"`
	MetricsInterval time.Duration     `yaml:"metrics_interval"`
	Prometheus      PrometheusConfig  `yaml:"prometheus"`
	Statsd          StatsdConfig      `yaml:"statsd"`
}

// PrometheusConfig Prometheus配置
type PrometheusConfig struct {
	Enabled bool `yaml:"enabled"`
	Port    int  `yaml:"port"`
}

// StatsdConfig StatsD配置
type StatsdConfig struct {
	Enabled bool   `yaml:"enabled"`
	Host    string `yaml:"host"`
}

// ConnectionConfig 连接配置
type ConnectionConfig struct {
	Timeout           time.Duration `yaml:"timeout"`
	KeepAlive         time.Duration `yaml:"keep_alive"`
	MaxIdleConns      int           `yaml:"max_idle_conns"`
	IdleConnTimeout   time.Duration `yaml:"idle_conn_timeout"`
}

// CoreConfigLoader 核心配置加载器
type CoreConfigLoader struct {
	config *CoreConfig
}

// NewCoreConfigLoader 创建核心配置加载器
func NewCoreConfigLoader() *CoreConfigLoader {
	return &CoreConfigLoader{
		config: &CoreConfig{},
	}
}

// LoadFromFile 从文件加载核心配置
func (c *CoreConfigLoader) LoadFromFile(filePath string) (*CoreConfig, error) {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// 如果文件不存在，返回默认配置
		return c.GetDefaultConfig(), nil
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

	c.config = &config
	return c.config, nil
}

// GetConfig 获取配置
func (c *CoreConfigLoader) GetConfig() *CoreConfig {
	if c.config == nil {
		return c.GetDefaultConfig()
	}
	return c.config
}

// GetDefaultConfig 获取默认核心配置
func (c *CoreConfigLoader) GetDefaultConfig() *CoreConfig {
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
				MetricsInterval: 5 * time.Second,
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
				Timeout:           30 * time.Second,
				KeepAlive:         30 * time.Second,
				MaxIdleConns:      100,
				IdleConnTimeout:   90 * time.Second,
			},
		},
	}
}

// SaveToFile 保存配置到文件
func (c *CoreConfigLoader) SaveToFile(config *CoreConfig, filePath string) error {
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