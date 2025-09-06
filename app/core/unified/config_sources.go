package unified

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"redis-runner/app/core/interfaces"
	"redis-runner/app/core/unified/config"
)

// FileConfigSource 文件配置源
type FileConfigSource struct {
	Path string
}

func (f *FileConfigSource) GetType() config.ConfigSourceType {
	return config.FileSource
}

func (f *FileConfigSource) GetPath() string {
	return f.Path
}

func (f *FileConfigSource) GetData() (map[string]interface{}, error) {
	// 这里应该根据文件扩展名使用相应的解析器
	// 暂时返回空数据
	return make(map[string]interface{}), nil
}

func (f *FileConfigSource) Validate() error {
	if f.Path == "" {
		return fmt.Errorf("file path cannot be empty")
	}
	
	if _, err := os.Stat(f.Path); os.IsNotExist(err) {
		return fmt.Errorf("config file does not exist: %s", f.Path)
	}
	
	return nil
}

// EnvironmentConfigSource 环境变量配置源
type EnvironmentConfigSource struct {
	Prefix string
}

func (e *EnvironmentConfigSource) GetType() config.ConfigSourceType {
	return config.EnvironmentSource
}

func (e *EnvironmentConfigSource) GetPath() string {
	return "environment"
}

func (e *EnvironmentConfigSource) GetData() (map[string]interface{}, error) {
	data := make(map[string]interface{})
	
	// 获取所有环境变量
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if len(pair) != 2 {
			continue
		}
		
		key, value := pair[0], pair[1]
		
		// 检查前缀
		if e.Prefix != "" && !strings.HasPrefix(key, e.Prefix) {
			continue
		}
		
		// 去除前缀
		if e.Prefix != "" {
			key = strings.TrimPrefix(key, e.Prefix)
		}
		
		// 转换为小写并替换下划线为点
		key = strings.ToLower(strings.ReplaceAll(key, "_", "."))
		
		data[key] = value
	}
	
	return data, nil
}

func (e *EnvironmentConfigSource) Validate() error {
	// 环境变量配置源总是有效的
	return nil
}

// CommandLineConfigSource 命令行配置源
type CommandLineConfigSource struct {
	Args map[string]string
}

func (c *CommandLineConfigSource) GetType() config.ConfigSourceType {
	return config.CommandLineSource
}

func (c *CommandLineConfigSource) GetPath() string {
	return "command_line"
}

func (c *CommandLineConfigSource) GetData() (map[string]interface{}, error) {
	data := make(map[string]interface{})
	
	for key, value := range c.Args {
		data[key] = value
	}
	
	return data, nil
}

func (c *CommandLineConfigSource) Validate() error {
	// 命令行配置源总是有效的
	return nil
}

// DefaultConfigSource 默认配置源
type DefaultConfigSource struct {
	Protocol string
}

func (d *DefaultConfigSource) GetType() config.ConfigSourceType {
	return config.DefaultSource
}

func (d *DefaultConfigSource) GetPath() string {
	return "default"
}

func (d *DefaultConfigSource) GetData() (map[string]interface{}, error) {
	return make(map[string]interface{}), nil
}

func (d *DefaultConfigSource) Validate() error {
	if d.Protocol == "" {
		return fmt.Errorf("protocol cannot be empty for default config source")
	}
	return nil
}

// RedisConfigLoader Redis配置加载器
type RedisConfigLoader struct{}

func (r *RedisConfigLoader) LoadFromSource(source config.ConfigSource) (interfaces.Config, error) {
	switch source.GetType() {
	case config.FileSource:
		return r.loadFromFile(source.GetPath())
	case config.EnvironmentSource:
		return r.loadFromEnvironment(source)
	case config.CommandLineSource:
		return r.loadFromCommandLine(source)
	case config.DefaultSource:
		return r.GetDefaultConfig(), nil
	default:
		return nil, fmt.Errorf("unsupported config source type: %s", source.GetType())
	}
}

func (r *RedisConfigLoader) GetSupportedSources() []config.ConfigSourceType {
	return []config.ConfigSourceType{
		config.FileSource,
		config.EnvironmentSource,
		config.CommandLineSource,
		config.DefaultSource,
	}
}

func (r *RedisConfigLoader) ValidateConfig(cfg interfaces.Config) error {
	conn := cfg.GetConnection()
	if conn == nil {
		return fmt.Errorf("connection config is required for Redis")
	}
	
	addresses := conn.GetAddresses()
	if len(addresses) == 0 {
		return fmt.Errorf("at least one Redis address is required")
	}
	
	// 验证Redis地址格式
	for _, addr := range addresses {
		if !strings.Contains(addr, ":") {
			return fmt.Errorf("invalid Redis address format: %s (expected host:port)", addr)
		}
	}
	
	return nil
}

func (r *RedisConfigLoader) GetDefaultConfig() interfaces.Config {
	return &config.UnifiedConfig{
		Protocol: "redis",
		Connection: &config.UnifiedConnectionConfig{
			Addresses: []string{"127.0.0.1:6379"},
			Credentials: map[string]string{},
			Pool: &config.UnifiedPoolConfig{
				PoolSize:          10,
				MinIdle:           1,
				MaxIdle:           5,
				IdleTimeout:       5 * time.Minute,
				ConnectionTimeout: 30 * time.Second,
				MaxLifetime:       1 * time.Hour,
				RetryInterval:     1 * time.Second,
				MaxRetries:        3,
			},
			Timeout: 30 * time.Second,
		},
		Benchmark: &config.UnifiedBenchmarkConfig{
			Total:       1000,
			Parallels:   10,
			DataSize:    1024,
			TTL:         0,
			ReadPercent: 50,
			RandomKeys:  1000,
			TestCase:    "set_get_random",
			Duration:    0,
			RateLimit:   0,
			WarmupTime:  0,
		},
		Global: &config.GlobalConfig{
			LogLevel:        "info",
			OutputFormat:    "table",
			DefaultProtocol: "redis",
			MetricsEnabled:  true,
			Debug:          false,
			Aliases:        make(map[string]string),
		},
		Metadata: make(map[string]interface{}),
	}
}

func (r *RedisConfigLoader) loadFromFile(path string) (interfaces.Config, error) {
	// 检查文件是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file does not exist: %s", path)
	}
	
	// 根据文件扩展名解析
	ext := strings.ToLower(filepath.Ext(path))
	
	switch ext {
	case ".yaml", ".yml":
		return r.loadFromYAMLFile(path)
	case ".json":
		return r.loadFromJSONFile(path)
	default:
		return nil, fmt.Errorf("unsupported config file format: %s", ext)
	}
}

func (r *RedisConfigLoader) loadFromYAMLFile(path string) (interfaces.Config, error) {
	// 暂时返回默认配置
	// 实际实现应该使用YAML库解析文件
	log.Printf("Loading Redis config from YAML file: %s", path)
	return r.GetDefaultConfig(), nil
}

func (r *RedisConfigLoader) loadFromJSONFile(path string) (interfaces.Config, error) {
	// 暂时返回默认配置
	// 实际实现应该使用JSON库解析文件
	log.Printf("Loading Redis config from JSON file: %s", path)
	return r.GetDefaultConfig(), nil
}

func (r *RedisConfigLoader) loadFromEnvironment(source config.ConfigSource) (interfaces.Config, error) {
	cfg := r.GetDefaultConfig().(*config.UnifiedConfig)
	
	data, err := source.GetData()
	if err != nil {
		return nil, err
	}
	
	// 解析环境变量到配置
	r.parseEnvironmentData(cfg, data)
	
	return cfg, nil
}

func (r *RedisConfigLoader) loadFromCommandLine(source config.ConfigSource) (interfaces.Config, error) {
	cfg := r.GetDefaultConfig().(*config.UnifiedConfig)
	
	data, err := source.GetData()
	if err != nil {
		return nil, err
	}
	
	// 解析命令行参数到配置
	r.parseCommandLineData(cfg, data)
	
	return cfg, nil
}

func (r *RedisConfigLoader) parseEnvironmentData(cfg *config.UnifiedConfig, data map[string]interface{}) {
	for key, value := range data {
		strValue := fmt.Sprintf("%v", value)
		
		switch key {
		case "redis.host":
			if cfg.Connection != nil && len(cfg.Connection.Addresses) > 0 {
				// 更新第一个地址的主机部分
				parts := strings.Split(cfg.Connection.Addresses[0], ":")
				port := "6379"
				if len(parts) > 1 {
					port = parts[1]
				}
				cfg.Connection.Addresses[0] = fmt.Sprintf("%s:%s", strValue, port)
			}
		case "redis.port":
			if cfg.Connection != nil && len(cfg.Connection.Addresses) > 0 {
				// 更新第一个地址的端口部分
				parts := strings.Split(cfg.Connection.Addresses[0], ":")
				host := "127.0.0.1"
				if len(parts) > 0 {
					host = parts[0]
				}
				cfg.Connection.Addresses[0] = fmt.Sprintf("%s:%s", host, strValue)
			}
		case "redis.password":
			if cfg.Connection != nil {
				if cfg.Connection.Credentials == nil {
					cfg.Connection.Credentials = make(map[string]string)
				}
				cfg.Connection.Credentials["password"] = strValue
			}
		case "redis.total":
			if total, err := strconv.Atoi(strValue); err == nil && cfg.Benchmark != nil {
				cfg.Benchmark.Total = total
			}
		case "redis.parallels":
			if parallels, err := strconv.Atoi(strValue); err == nil && cfg.Benchmark != nil {
				cfg.Benchmark.Parallels = parallels
			}
		}
	}
}

func (r *RedisConfigLoader) parseCommandLineData(cfg *config.UnifiedConfig, data map[string]interface{}) {
	for key, value := range data {
		strValue := fmt.Sprintf("%v", value)
		
		switch key {
		case "h", "host":
			if cfg.Connection != nil && len(cfg.Connection.Addresses) > 0 {
				parts := strings.Split(cfg.Connection.Addresses[0], ":")
				port := "6379"
				if len(parts) > 1 {
					port = parts[1]
				}
				cfg.Connection.Addresses[0] = fmt.Sprintf("%s:%s", strValue, port)
			}
		case "p", "port":
			if cfg.Connection != nil && len(cfg.Connection.Addresses) > 0 {
				parts := strings.Split(cfg.Connection.Addresses[0], ":")
				host := "127.0.0.1"
				if len(parts) > 0 {
					host = parts[0]
				}
				cfg.Connection.Addresses[0] = fmt.Sprintf("%s:%s", host, strValue)
			}
		case "a", "auth":
			if cfg.Connection != nil {
				if cfg.Connection.Credentials == nil {
					cfg.Connection.Credentials = make(map[string]string)
				}
				cfg.Connection.Credentials["password"] = strValue
			}
		case "n", "total":
			if total, err := strconv.Atoi(strValue); err == nil && cfg.Benchmark != nil {
				cfg.Benchmark.Total = total
			}
		case "c", "connections":
			if parallels, err := strconv.Atoi(strValue); err == nil && cfg.Benchmark != nil {
				cfg.Benchmark.Parallels = parallels
			}
		}
	}
}

// HttpConfigLoader HTTP配置加载器
type HttpConfigLoader struct{}

func (h *HttpConfigLoader) LoadFromSource(source config.ConfigSource) (interfaces.Config, error) {
	// HTTP配置加载逻辑（简化实现）
	return h.GetDefaultConfig(), nil
}

func (h *HttpConfigLoader) GetSupportedSources() []config.ConfigSourceType {
	return []config.ConfigSourceType{
		config.FileSource,
		config.CommandLineSource,
		config.DefaultSource,
	}
}

func (h *HttpConfigLoader) ValidateConfig(cfg interfaces.Config) error {
	conn := cfg.GetConnection()
	if conn == nil {
		return fmt.Errorf("connection config is required for HTTP")
	}
	
	addresses := conn.GetAddresses()
	if len(addresses) == 0 {
		return fmt.Errorf("at least one HTTP endpoint is required")
	}
	
	return nil
}

func (h *HttpConfigLoader) GetDefaultConfig() interfaces.Config {
	return &config.UnifiedConfig{
		Protocol: "http",
		Connection: &config.UnifiedConnectionConfig{
			Addresses: []string{"http://localhost:8080"},
			Pool: &config.UnifiedPoolConfig{
				PoolSize:          10,
				MinIdle:           1,
				MaxIdle:           5,
				IdleTimeout:       5 * time.Minute,
				ConnectionTimeout: 30 * time.Second,
			},
			Timeout: 30 * time.Second,
		},
		Benchmark: &config.UnifiedBenchmarkConfig{
			Total:     1000,
			Parallels: 10,
			TestCase:  "GET",
		},
		Metadata: make(map[string]interface{}),
	}
}

// KafkaConfigLoader Kafka配置加载器
type KafkaConfigLoader struct{}

func (k *KafkaConfigLoader) LoadFromSource(source config.ConfigSource) (interfaces.Config, error) {
	// Kafka配置加载逻辑（简化实现）
	return k.GetDefaultConfig(), nil
}

func (k *KafkaConfigLoader) GetSupportedSources() []config.ConfigSourceType {
	return []config.ConfigSourceType{
		config.FileSource,
		config.CommandLineSource,
		config.DefaultSource,
	}
}

func (k *KafkaConfigLoader) ValidateConfig(cfg interfaces.Config) error {
	conn := cfg.GetConnection()
	if conn == nil {
		return fmt.Errorf("connection config is required for Kafka")
	}
	
	addresses := conn.GetAddresses()
	if len(addresses) == 0 {
		return fmt.Errorf("at least one Kafka broker is required")
	}
	
	return nil
}

func (k *KafkaConfigLoader) GetDefaultConfig() interfaces.Config {
	return &config.UnifiedConfig{
		Protocol: "kafka",
		Connection: &config.UnifiedConnectionConfig{
			Addresses: []string{"localhost:9092"},
			Pool: &config.UnifiedPoolConfig{
				PoolSize:          3,
				MinIdle:           1,
				MaxIdle:           3,
				ConnectionTimeout: 30 * time.Second,
			},
			Timeout: 30 * time.Second,
		},
		Benchmark: &config.UnifiedBenchmarkConfig{
			Total:     1000,
			Parallels: 3,
			TestCase:  "produce_consume",
		},
		Metadata: make(map[string]interface{}),
	}
}