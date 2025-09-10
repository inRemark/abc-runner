package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v2"
)

// ConfigLoader 配置加载器
type ConfigLoader struct {
	defaultConfig *KafkaAdapterConfig
}

// NewConfigLoader 创建配置加载器
func NewConfigLoader() *ConfigLoader {
	return &ConfigLoader{
		defaultConfig: getDefaultConfig(),
	}
}

// LoadFromFile 从文件加载配置
func (l *ConfigLoader) LoadFromFile(filePath string) (*KafkaAdapterConfig, error) {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", filePath)
	}

	// 读取文件内容
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 解析YAML
	var config KafkaAdapterConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// 合并默认配置
	mergedConfig := l.mergeWithDefaults(&config)

	// 环境变量替换
	l.replaceEnvVars(mergedConfig)

	// 验证配置
	if err := mergedConfig.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return mergedConfig, nil
}

// LoadFromDefault 加载默认配置
func (l *ConfigLoader) LoadFromDefault() *KafkaAdapterConfig {
	config := *l.defaultConfig
	l.replaceEnvVars(&config)
	return &config
}

// LoadFromCommandLine 从命令行参数创建配置
func (l *ConfigLoader) LoadFromCommandLine(params map[string]interface{}) (*KafkaAdapterConfig, error) {
	config := *l.defaultConfig

	// 从参数中提取配置
	if broker, ok := params["broker"].(string); ok && broker != "" {
		config.Brokers = []string{broker}
	}

	if topic, ok := params["topic"].(string); ok && topic != "" {
		config.Benchmark.DefaultTopic = topic
	}

	if groupID, ok := params["group_id"].(string); ok && groupID != "" {
		config.Consumer.GroupID = groupID
	}

	if acks, ok := params["acks"].(int); ok {
		config.Producer.Acks = fmt.Sprintf("%d", acks)
	}

	if n, ok := params["n"].(int); ok && n > 0 {
		config.Benchmark.Total = n
	}

	if c, ok := params["c"].(int); ok && c > 0 {
		config.Benchmark.Parallels = c
	}

	if d, ok := params["d"].(int); ok && d > 0 {
		config.Benchmark.DataSize = d
	}

	// 环境变量替换
	l.replaceEnvVars(&config)

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// mergeWithDefaults 与默认配置合并
func (l *ConfigLoader) mergeWithDefaults(config *KafkaAdapterConfig) *KafkaAdapterConfig {
	merged := *l.defaultConfig

	// 合并基础配置
	if len(config.Brokers) > 0 {
		merged.Brokers = config.Brokers
	}

	if config.ClientID != "" {
		merged.ClientID = config.ClientID
	}

	if config.Version != "" {
		merged.Version = config.Version
	}

	// 合并生产者配置
	merged.Producer = l.mergeProducerConfig(&merged.Producer, &config.Producer)

	// 合并消费者配置
	merged.Consumer = l.mergeConsumerConfig(&merged.Consumer, &config.Consumer)

	// 合并安全配置
	merged.Security = l.mergeSecurityConfig(&merged.Security, &config.Security)

	// 合并性能配置
	merged.Performance = l.mergePerformanceConfig(&merged.Performance, &config.Performance)

	// 合并基准测试配置
	merged.Benchmark = l.mergeBenchmarkConfig(&merged.Benchmark, &config.Benchmark)

	return &merged
}

// mergeProducerConfig 合并生产者配置
func (l *ConfigLoader) mergeProducerConfig(default_, config *ProducerConfig) ProducerConfig {
	merged := *default_

	if config.Acks != "" {
		merged.Acks = config.Acks
	}
	if config.Retries > 0 {
		merged.Retries = config.Retries
	}
	if config.BatchSize > 0 {
		merged.BatchSize = config.BatchSize
	}
	if config.LingerMs > 0 {
		merged.LingerMs = config.LingerMs
	}
	if config.Compression != "" {
		merged.Compression = config.Compression
	}
	if config.IdempotenceEnabled {
		merged.IdempotenceEnabled = config.IdempotenceEnabled
	}
	if config.MaxInFlightRequests > 0 {
		merged.MaxInFlightRequests = config.MaxInFlightRequests
	}
	if config.RequestTimeout > 0 {
		merged.RequestTimeout = config.RequestTimeout
	}
	if config.WriteTimeout > 0 {
		merged.WriteTimeout = config.WriteTimeout
	}
	if config.ReadTimeout > 0 {
		merged.ReadTimeout = config.ReadTimeout
	}

	return merged
}

// mergeConsumerConfig 合并消费者配置
func (l *ConfigLoader) mergeConsumerConfig(default_, config *ConsumerConfig) ConsumerConfig {
	merged := *default_

	if config.GroupID != "" {
		merged.GroupID = config.GroupID
	}
	if config.AutoOffsetReset != "" {
		merged.AutoOffsetReset = config.AutoOffsetReset
	}
	if config.EnableAutoCommit {
		merged.EnableAutoCommit = config.EnableAutoCommit
	}
	if config.AutoCommitInterval > 0 {
		merged.AutoCommitInterval = config.AutoCommitInterval
	}
	if config.SessionTimeout > 0 {
		merged.SessionTimeout = config.SessionTimeout
	}
	if config.HeartbeatInterval > 0 {
		merged.HeartbeatInterval = config.HeartbeatInterval
	}
	if config.MaxPollRecords > 0 {
		merged.MaxPollRecords = config.MaxPollRecords
	}
	if config.FetchMinBytes > 0 {
		merged.FetchMinBytes = config.FetchMinBytes
	}
	if config.FetchMaxBytes > 0 {
		merged.FetchMaxBytes = config.FetchMaxBytes
	}
	if config.FetchMaxWait > 0 {
		merged.FetchMaxWait = config.FetchMaxWait
	}
	if config.ReadTimeout > 0 {
		merged.ReadTimeout = config.ReadTimeout
	}
	if config.WriteTimeout > 0 {
		merged.WriteTimeout = config.WriteTimeout
	}
	if config.InitialOffset != "" {
		merged.InitialOffset = config.InitialOffset
	}

	return merged
}

// mergeSecurityConfig 合并安全配置
func (l *ConfigLoader) mergeSecurityConfig(default_, config *SecurityConfig) SecurityConfig {
	merged := *default_

	// TLS配置合并
	if config.TLS.Enabled {
		merged.TLS = config.TLS
	}

	// SASL配置合并
	if config.SASL.Enabled {
		merged.SASL = config.SASL
	}

	return merged
}

// mergePerformanceConfig 合并性能配置
func (l *ConfigLoader) mergePerformanceConfig(default_, config *PerformanceConfig) PerformanceConfig {
	merged := *default_

	if config.ConnectionPoolSize > 0 {
		merged.ConnectionPoolSize = config.ConnectionPoolSize
	}
	if config.ProducerPoolSize > 0 {
		merged.ProducerPoolSize = config.ProducerPoolSize
	}
	if config.ConsumerPoolSize > 0 {
		merged.ConsumerPoolSize = config.ConsumerPoolSize
	}
	if config.MetricsInterval > 0 {
		merged.MetricsInterval = config.MetricsInterval
	}

	return merged
}

// mergeBenchmarkConfig 合并基准测试配置
func (l *ConfigLoader) mergeBenchmarkConfig(default_, config *KafkaBenchmarkConfig) KafkaBenchmarkConfig {
	merged := *default_

	if config.DefaultTopic != "" {
		merged.DefaultTopic = config.DefaultTopic
	}
	if config.MessageSizeRange.Min > 0 {
		merged.MessageSizeRange.Min = config.MessageSizeRange.Min
	}
	if config.MessageSizeRange.Max > 0 {
		merged.MessageSizeRange.Max = config.MessageSizeRange.Max
	}
	if len(config.BatchSizes) > 0 {
		merged.BatchSizes = config.BatchSizes
	}
	if config.PartitionStrategy != "" {
		merged.PartitionStrategy = config.PartitionStrategy
	}
	if config.Total > 0 {
		merged.Total = config.Total
	}
	if config.Parallels > 0 {
		merged.Parallels = config.Parallels
	}
	if config.DataSize > 0 {
		merged.DataSize = config.DataSize
	}
	if config.TTL > 0 {
		merged.TTL = config.TTL
	}
	if config.ReadPercent >= 0 {
		merged.ReadPercent = config.ReadPercent
	}
	if config.RandomKeys > 0 {
		merged.RandomKeys = config.RandomKeys
	}
	if config.TestCase != "" {
		merged.TestCase = config.TestCase
	}
	if config.Timeout > 0 {
		merged.Timeout = config.Timeout
	}

	return merged
}

// replaceEnvVars 替换环境变量
func (l *ConfigLoader) replaceEnvVars(config *KafkaAdapterConfig) {
	// 替换SASL用户名和密码中的环境变量
	config.Security.SASL.Username = l.expandEnvVars(config.Security.SASL.Username)
	config.Security.SASL.Password = l.expandEnvVars(config.Security.SASL.Password)

	// 替换TLS文件路径中的环境变量
	config.Security.TLS.CertFile = l.expandEnvVars(config.Security.TLS.CertFile)
	config.Security.TLS.KeyFile = l.expandEnvVars(config.Security.TLS.KeyFile)
	config.Security.TLS.CaFile = l.expandEnvVars(config.Security.TLS.CaFile)
}

// expandEnvVars 展开环境变量
func (l *ConfigLoader) expandEnvVars(s string) string {
	if s == "" {
		return s
	}

	// 支持 ${VAR} 和 $VAR 格式
	return os.ExpandEnv(s)
}

// getDefaultConfig 获取默认配置
func getDefaultConfig() *KafkaAdapterConfig {
	return &KafkaAdapterConfig{
		Brokers:  []string{"localhost:9092"},
		ClientID: "abc-runner-kafka-client",
		Version:  "2.8.0",

		Producer: ProducerConfig{
			Acks:                "all",
			Retries:             3,
			BatchSize:           16384, // 16KB
			LingerMs:            5 * time.Millisecond,
			Compression:         "snappy",
			IdempotenceEnabled:  true,
			MaxInFlightRequests: 5,
			RequestTimeout:      30 * time.Second,
			WriteTimeout:        10 * time.Second,
			ReadTimeout:         10 * time.Second,
		},

		Consumer: ConsumerConfig{
			GroupID:            "abc-runner-group",
			AutoOffsetReset:    "latest",
			EnableAutoCommit:   true,
			AutoCommitInterval: 1 * time.Second,
			SessionTimeout:     30 * time.Second,
			HeartbeatInterval:  3 * time.Second,
			MaxPollRecords:     500,
			FetchMinBytes:      1024,     // 1KB
			FetchMaxBytes:      52428800, // 50MB
			FetchMaxWait:       500 * time.Millisecond,
			ReadTimeout:        10 * time.Second,
			WriteTimeout:       10 * time.Second,
			InitialOffset:      "latest",
		},

		Security: SecurityConfig{
			TLS: TLSConfig{
				Enabled:   false,
				VerifySSL: true,
			},
			SASL: SASLConfig{
				Enabled:   false,
				Mechanism: "SCRAM-SHA-512",
			},
		},

		Performance: PerformanceConfig{
			ConnectionPoolSize: 10,
			ProducerPoolSize:   5,
			ConsumerPoolSize:   5,
			MetricsInterval:    5 * time.Second,
		},

		Benchmark: KafkaBenchmarkConfig{
			DefaultTopic: "benchmark-topic",
			MessageSizeRange: MessageSizeRange{
				Min: 100,
				Max: 10240,
			},
			BatchSizes:        []int{1, 10, 100, 1000},
			PartitionStrategy: "round_robin",
			Total:             100000,
			Parallels:         50,
			DataSize:          1024,
			TTL:               0,
			ReadPercent:       50,
			RandomKeys:        10000,
			TestCase:          "produce",
			Timeout:           30 * time.Second,
		},
	}
}

// SaveToFile 保存配置到文件
func (l *ConfigLoader) SaveToFile(config *KafkaAdapterConfig, filePath string) error {
	// 创建目录（如果不存在）
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 序列化配置
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// ValidateFile 验证配置文件
func (l *ConfigLoader) ValidateFile(filePath string) error {
	_, err := l.LoadFromFile(filePath)
	return err
}
