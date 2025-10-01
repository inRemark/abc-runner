package metrics

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// ConfigManager 配置管理器
type ConfigManager struct {
	config     *MetricsConfig
	configPath string
	watchers   []ConfigWatcher
}

// ConfigWatcher 配置观察者接口
type ConfigWatcher interface {
	OnConfigChanged(config *MetricsConfig) error
}

// NewConfigManager 创建配置管理器
func NewConfigManager(configPath string) *ConfigManager {
	return &ConfigManager{
		configPath: configPath,
		watchers:   make([]ConfigWatcher, 0),
	}
}

// LoadConfig 加载配置
func (cm *ConfigManager) LoadConfig() error {
	// 如果配置文件不存在，创建默认配置
	if _, err := os.Stat(cm.configPath); os.IsNotExist(err) {
		return cm.CreateDefaultConfig()
	}

	// 读取配置文件
	data, err := os.ReadFile(cm.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// 解析配置文件
	config := &MetricsConfig{}
	ext := strings.ToLower(filepath.Ext(cm.configPath))

	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, config); err != nil {
			return fmt.Errorf("failed to parse YAML config: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, config); err != nil {
			return fmt.Errorf("failed to parse JSON config: %w", err)
		}
	default:
		return fmt.Errorf("unsupported config file format: %s", ext)
	}

	// 应用默认值
	if err := cm.applyDefaults(config); err != nil {
		return fmt.Errorf("failed to apply default values: %w", err)
	}

	// 验证配置
	if err := cm.validateConfig(config); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	cm.config = config
	return nil
}

// SaveConfig 保存配置
func (cm *ConfigManager) SaveConfig(config *MetricsConfig) error {
	if config == nil {
		config = cm.config
	}

	// 确保目录存在
	dir := filepath.Dir(cm.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// 序列化配置
	var data []byte
	var err error

	ext := strings.ToLower(filepath.Ext(cm.configPath))
	switch ext {
	case ".yaml", ".yml":
		data, err = yaml.Marshal(config)
	case ".json":
		data, err = json.MarshalIndent(config, "", "  ")
	default:
		return fmt.Errorf("unsupported config file format: %s", ext)
	}

	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(cm.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	cm.config = config
	return nil
}

// GetConfig 获取当前配置
func (cm *ConfigManager) GetConfig() *MetricsConfig {
	if cm.config == nil {
		cm.config = DefaultMetricsConfig()
	}
	return cm.config
}

// UpdateConfig 更新配置
func (cm *ConfigManager) UpdateConfig(updater func(*MetricsConfig) error) error {
	if cm.config == nil {
		cm.config = DefaultMetricsConfig()
	}

	// 创建配置副本
	configCopy := *cm.config

	// 应用更新
	if err := updater(&configCopy); err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	// 验证更新后的配置
	if err := cm.validateConfig(&configCopy); err != nil {
		return fmt.Errorf("invalid updated config: %w", err)
	}

	// 保存并应用新配置
	if err := cm.SaveConfig(&configCopy); err != nil {
		return err
	}

	// 通知观察者
	for _, watcher := range cm.watchers {
		if err := watcher.OnConfigChanged(&configCopy); err != nil {
			// 记录错误但不失败
			fmt.Printf("Config watcher error: %v\n", err)
		}
	}

	return nil
}

// AddWatcher 添加配置观察者
func (cm *ConfigManager) AddWatcher(watcher ConfigWatcher) {
	cm.watchers = append(cm.watchers, watcher)
}

// CreateDefaultConfig 创建默认配置文件
func (cm *ConfigManager) CreateDefaultConfig() error {
	defaultConfig := DefaultMetricsConfig()
	return cm.SaveConfig(defaultConfig)
}

// applyDefaults 应用默认值
func (cm *ConfigManager) applyDefaults(config *MetricsConfig) error {
	defaultConfig := DefaultMetricsConfig()
	return cm.mergeDefaults(reflect.ValueOf(config).Elem(), reflect.ValueOf(defaultConfig).Elem())
}

// mergeDefaults 递归合并默认值
func (cm *ConfigManager) mergeDefaults(target, source reflect.Value) error {
	if target.Type() != source.Type() {
		return fmt.Errorf("type mismatch: %v vs %v", target.Type(), source.Type())
	}

	switch target.Kind() {
	case reflect.Struct:
		for i := 0; i < target.NumField(); i++ {
			fieldTarget := target.Field(i)
			fieldSource := source.Field(i)

			if fieldTarget.CanSet() {
				// 如果目标字段是零值，使用默认值
				if cm.isZeroValue(fieldTarget) {
					fieldTarget.Set(fieldSource)
				} else if fieldTarget.Kind() == reflect.Struct {
					// 递归处理嵌套结构
					if err := cm.mergeDefaults(fieldTarget, fieldSource); err != nil {
						return err
					}
				}
			}
		}
	case reflect.Slice:
		if target.Len() == 0 && source.Len() > 0 {
			target.Set(source)
		}
	case reflect.Map:
		if target.Len() == 0 && source.Len() > 0 {
			target.Set(source)
		}
	}

	return nil
}

// isZeroValue 检查是否为零值
func (cm *ConfigManager) isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.String:
		return v.String() == ""
	case reflect.Bool:
		return !v.Bool()
	case reflect.Slice, reflect.Map:
		return v.Len() == 0
	case reflect.Struct:
		// 对于结构体，检查所有字段是否都是零值
		for i := 0; i < v.NumField(); i++ {
			if !cm.isZeroValue(v.Field(i)) {
				return false
			}
		}
		return true
	default:
		return v.IsZero()
	}
}

// validateConfig 验证配置
func (cm *ConfigManager) validateConfig(config *MetricsConfig) error {
	// 验证延迟配置
	if config.Latency.HistorySize <= 0 {
		return fmt.Errorf("latency.history_size must be positive")
	}
	if config.Latency.SamplingRate < 0 || config.Latency.SamplingRate > 1 {
		return fmt.Errorf("latency.sampling_rate must be between 0 and 1")
	}
	if config.Latency.ComputeInterval <= 0 {
		return fmt.Errorf("latency.compute_interval must be positive")
	}

	// 验证吞吐量配置
	if config.Throughput.WindowSize <= 0 {
		return fmt.Errorf("throughput.window_size must be positive")
	}
	if config.Throughput.UpdateInterval <= 0 {
		return fmt.Errorf("throughput.update_interval must be positive")
	}

	// 验证系统配置
	if config.System.MonitorInterval <= 0 {
		return fmt.Errorf("system.monitor_interval must be positive")
	}
	if config.System.SnapshotRetention <= 0 {
		return fmt.Errorf("system.snapshot_retention must be positive")
	}

	// 验证健康阈值
	thresholds := config.System.HealthThresholds
	if thresholds.MemoryUsage < 0 || thresholds.MemoryUsage > 100 {
		return fmt.Errorf("health_thresholds.memory_usage must be between 0 and 100")
	}
	if thresholds.CPUUsage < 0 || thresholds.CPUUsage > 100 {
		return fmt.Errorf("health_thresholds.cpu_usage must be between 0 and 100")
	}

	// 验证存储配置
	if config.Storage.MemoryLimit <= 0 {
		return fmt.Errorf("storage.memory_limit must be positive")
	}
	if config.Storage.FlushInterval <= 0 {
		return fmt.Errorf("storage.flush_interval must be positive")
	}

	// 验证导出配置
	if config.Export.Interval <= 0 {
		return fmt.Errorf("export.interval must be positive")
	}

	return nil
}

// LoadFromEnvironment 从环境变量加载配置
func (cm *ConfigManager) LoadFromEnvironment(prefix string) error {
	if cm.config == nil {
		cm.config = DefaultMetricsConfig()
	}

	return cm.loadEnvRecursive(reflect.ValueOf(cm.config).Elem(), prefix)
}

// loadEnvRecursive 递归加载环境变量
func (cm *ConfigManager) loadEnvRecursive(v reflect.Value, prefix string) error {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		if !field.CanSet() {
			continue
		}

		// 构建环境变量名
		envName := prefix + "_" + strings.ToUpper(fieldType.Name)
		envValue := os.Getenv(envName)

		switch field.Kind() {
		case reflect.Struct:
			// 递归处理嵌套结构
			if err := cm.loadEnvRecursive(field, envName); err != nil {
				return err
			}
		case reflect.String:
			if envValue != "" {
				field.SetString(envValue)
			}
		case reflect.Int, reflect.Int32, reflect.Int64:
			if envValue != "" {
				if val, err := strconv.ParseInt(envValue, 10, 64); err == nil {
					field.SetInt(val)
				}
			}
		case reflect.Uint, reflect.Uint32, reflect.Uint64:
			if envValue != "" {
				if val, err := strconv.ParseUint(envValue, 10, 64); err == nil {
					field.SetUint(val)
				}
			}
		case reflect.Float64:
			if envValue != "" {
				if val, err := strconv.ParseFloat(envValue, 64); err == nil {
					field.SetFloat(val)
				}
			}
		case reflect.Bool:
			if envValue != "" {
				if val, err := strconv.ParseBool(envValue); err == nil {
					field.SetBool(val)
				}
			}
		case reflect.Slice:
			if envValue != "" && field.Type().Elem().Kind() == reflect.String {
				// 解析逗号分隔的字符串列表
				values := strings.Split(envValue, ",")
				slice := reflect.MakeSlice(field.Type(), len(values), len(values))
				for j, val := range values {
					slice.Index(j).SetString(strings.TrimSpace(val))
				}
				field.Set(slice)
			}
		case reflect.TypeOf(time.Duration(0)).Kind():
			if envValue != "" {
				if val, err := time.ParseDuration(envValue); err == nil {
					field.Set(reflect.ValueOf(val))
				}
			}
		}
	}

	return nil
}

// GetConfigAsYAML 获取YAML格式的配置
func (cm *ConfigManager) GetConfigAsYAML() (string, error) {
	if cm.config == nil {
		cm.config = DefaultMetricsConfig()
	}

	data, err := yaml.Marshal(cm.config)
	if err != nil {
		return "", fmt.Errorf("failed to marshal config to YAML: %w", err)
	}

	return string(data), nil
}

// GetConfigAsJSON 获取JSON格式的配置
func (cm *ConfigManager) GetConfigAsJSON() (string, error) {
	if cm.config == nil {
		cm.config = DefaultMetricsConfig()
	}

	data, err := json.MarshalIndent(cm.config, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal config to JSON: %w", err)
	}

	return string(data), nil
}

// ConfigTemplate 配置模板生成器
type ConfigTemplate struct {
	Description string                 `json:"description" yaml:"description"`
	Version     string                 `json:"version" yaml:"version"`
	Config      *MetricsConfig         `json:"config" yaml:"config"`
	Examples    map[string]interface{} `json:"examples" yaml:"examples"`
}

// GenerateTemplate 生成配置模板
func (cm *ConfigManager) GenerateTemplate() *ConfigTemplate {
	return &ConfigTemplate{
		Description: "ABC Runner Metrics Configuration Template",
		Version:     "0.2.0",
		Config:      DefaultMetricsConfig(),
		Examples: map[string]interface{}{
			"high_performance": map[string]interface{}{
				"latency": map[string]interface{}{
					"history_size":     50000,
					"sampling_rate":    0.5,
					"compute_interval": "500ms",
				},
				"system": map[string]interface{}{
					"monitor_interval": "500ms",
					"enabled":          true,
				},
			},
			"memory_optimized": map[string]interface{}{
				"latency": map[string]interface{}{
					"history_size":  5000,
					"sampling_rate": 0.1,
				},
				"storage": map[string]interface{}{
					"memory_limit":    "50MB",
					"use_compression": true,
					"flush_interval":  "1s",
				},
			},
			"detailed_monitoring": map[string]interface{}{
				"system": map[string]interface{}{
					"monitor_interval":   "100ms",
					"snapshot_retention": 1000,
					"health_thresholds": map[string]interface{}{
						"memory_usage":    60.0,
						"goroutine_count": 500,
					},
				},
				"export": map[string]interface{}{
					"enabled":  true,
					"interval": "5s",
					"format":   []string{"json", "csv"},
				},
			},
		},
	}
}

// ConfigValidator 配置验证器
type ConfigValidator struct {
	rules []ValidationRule
}

// ValidationRule 验证规则
type ValidationRule struct {
	Name        string
	Description string
	Validator   func(*MetricsConfig) error
}

// NewConfigValidator 创建配置验证器
func NewConfigValidator() *ConfigValidator {
	validator := &ConfigValidator{
		rules: make([]ValidationRule, 0),
	}

	// 添加默认验证规则
	validator.AddRule("latency_history_size", "Latency history size validation", func(config *MetricsConfig) error {
		if config.Latency.HistorySize < 100 {
			return fmt.Errorf("latency history size too small (minimum: 100)")
		}
		if config.Latency.HistorySize > 1000000 {
			return fmt.Errorf("latency history size too large (maximum: 1000000)")
		}
		return nil
	})

	validator.AddRule("memory_limit", "Memory limit validation", func(config *MetricsConfig) error {
		if config.Storage.MemoryLimit < 1024*1024 { // 1MB
			return fmt.Errorf("memory limit too small (minimum: 1MB)")
		}
		if config.Storage.MemoryLimit > 1024*1024*1024 { // 1GB
			return fmt.Errorf("memory limit too large (maximum: 1GB)")
		}
		return nil
	})

	validator.AddRule("monitor_interval", "Monitor interval validation", func(config *MetricsConfig) error {
		if config.System.MonitorInterval < 100*time.Millisecond {
			return fmt.Errorf("monitor interval too small (minimum: 100ms)")
		}
		if config.System.MonitorInterval > 60*time.Second {
			return fmt.Errorf("monitor interval too large (maximum: 60s)")
		}
		return nil
	})

	return validator
}

// AddRule 添加验证规则
func (cv *ConfigValidator) AddRule(name, description string, validator func(*MetricsConfig) error) {
	cv.rules = append(cv.rules, ValidationRule{
		Name:        name,
		Description: description,
		Validator:   validator,
	})
}

// Validate 验证配置
func (cv *ConfigValidator) Validate(config *MetricsConfig) error {
	var errors []string

	for _, rule := range cv.rules {
		if err := rule.Validator(config); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", rule.Name, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

// GetRules 获取所有验证规则
func (cv *ConfigValidator) GetRules() []ValidationRule {
	return cv.rules
}
