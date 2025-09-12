package config

import (
	"time"
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
	Enabled         bool             `yaml:"enabled"`
	MetricsInterval time.Duration    `yaml:"metrics_interval"`
	Prometheus      PrometheusConfig `yaml:"prometheus"`
	Statsd          StatsdConfig     `yaml:"statsd"`
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
	Timeout         time.Duration `yaml:"timeout"`
	KeepAlive       time.Duration `yaml:"keep_alive"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	IdleConnTimeout time.Duration `yaml:"idle_conn_timeout"`
}
