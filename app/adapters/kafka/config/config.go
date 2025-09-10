package config

import (
	"fmt"
	"time"

	"abc-runner/app/core/interfaces"
)

// LoadDefaultKafkaConfig 加载默认Kafka配置
func LoadDefaultKafkaConfig() *KafkaAdapterConfig {
	return &KafkaAdapterConfig{
		Protocol: "kafka",
		Brokers:  []string{"localhost:9092"},
		ClientID: "abc-runner",
		Version:  "2.6.0",
		TopicConfigs: []TopicConfig{
			{
				Name:       "test-topic",
				Partitions: 1,
				Replicas:   1,
			},
		},
		Producer: ProducerConfig{
			BatchSize:    16384,
			BatchTimeout: time.Millisecond * 100,
			RetryMax:     3,
			RequiredAcks: 1,
			Compression:  "snappy",
		},
		Consumer: ConsumerConfig{
			GroupID:          "test-group",
			AutoOffsetReset:  "earliest",
			CommitInterval:   time.Second * 1,
			SessionTimeout:   time.Second * 30,
			HeartbeatTimeout: time.Second * 3,
		},
		Benchmark: KafkaBenchmarkConfig{
			Total:       1000,
			Parallels:   3,
			MessageSize: 1024,
			TestType:    "produce",
		},
		Performance: PerformanceConfig{
			ConnectionPoolSize: 10,
			ProducerPoolSize:   5,
			ConsumerPoolSize:   5,
		},
	}
}

// KafkaAdapterConfig Kafka适配器配置
type KafkaAdapterConfig struct {
	// 基础连接配置
	Protocol     string        `yaml:"protocol" json:"protocol"`           // 协议类型
	Brokers      []string      `yaml:"brokers" json:"brokers"`             // Broker地址列表
	ClientID     string        `yaml:"client_id" json:"client_id"`         // 客户端ID
	Version      string        `yaml:"version" json:"version"`             // Kafka版本
	TopicConfigs []TopicConfig `yaml:"topic_configs" json:"topic_configs"` // 主题配置

	// 生产者配置
	Producer ProducerConfig `yaml:"producer" json:"producer"`

	// 消费者配置
	Consumer ConsumerConfig `yaml:"consumer" json:"consumer"`

	// 安全配置
	Security SecurityConfig `yaml:"security" json:"security"`

	// 性能配置
	Performance PerformanceConfig `yaml:"performance" json:"performance"`

	// 基准测试配置
	Benchmark KafkaBenchmarkConfig `yaml:"benchmark" json:"benchmark"`
}

// TopicConfig 主题配置
type TopicConfig struct {
	Name       string `yaml:"name" json:"name"`             // 主题名称
	Partitions int    `yaml:"partitions" json:"partitions"` // 分区数
	Replicas   int    `yaml:"replicas" json:"replicas"`     // 副本数
}

// ProducerConfig 生产者配置
type ProducerConfig struct {
	Acks                string        `yaml:"acks" json:"acks"`                       // 确认模式: "0", "1", "all"
	Retries             int           `yaml:"retries" json:"retries"`                 // 重试次数
	BatchSize           int           `yaml:"batch_size" json:"batch_size"`           // 批处理大小
	BatchTimeout        time.Duration `yaml:"batch_timeout" json:"batch_timeout"`     // 批处理超时时间
	LingerMs            time.Duration `yaml:"linger_ms" json:"linger_ms"`             // 批处理等待时间
	Compression         string        `yaml:"compression" json:"compression"`         // 压缩算法: none, gzip, snappy, lz4, zstd
	RetryMax            int           `yaml:"retry_max" json:"retry_max"`             // 最大重试次数
	RequiredAcks        int           `yaml:"required_acks" json:"required_acks"`     // 所需确认数
	IdempotenceEnabled  bool          `yaml:"idempotence" json:"idempotence"`         // 幂等性保证
	MaxInFlightRequests int           `yaml:"max_in_flight" json:"max_in_flight"`     // 最大未确认请求数
	RequestTimeout      time.Duration `yaml:"request_timeout" json:"request_timeout"` // 请求超时
	WriteTimeout        time.Duration `yaml:"write_timeout" json:"write_timeout"`     // 写入超时
	ReadTimeout         time.Duration `yaml:"read_timeout" json:"read_timeout"`       // 读取超时
}

// ConsumerConfig 消费者配置
type ConsumerConfig struct {
	GroupID            string        `yaml:"group_id" json:"group_id"`                         // 消费者组ID
	AutoOffsetReset    string        `yaml:"auto_offset_reset" json:"auto_offset_reset"`       // 自动偏移重置: earliest, latest
	CommitInterval     time.Duration `yaml:"commit_interval" json:"commit_interval"`           // 提交间隔
	SessionTimeout     time.Duration `yaml:"session_timeout" json:"session_timeout"`           // 会话超时
	HeartbeatTimeout   time.Duration `yaml:"heartbeat_timeout" json:"heartbeat_timeout"`       // 心跳超时
	EnableAutoCommit   bool          `yaml:"enable_auto_commit" json:"enable_auto_commit"`     // 启用自动提交
	AutoCommitInterval time.Duration `yaml:"auto_commit_interval" json:"auto_commit_interval"` // 自动提交间隔
	HeartbeatInterval  time.Duration `yaml:"heartbeat_interval" json:"heartbeat_interval"`     // 心跳间隔
	MaxPollRecords     int           `yaml:"max_poll_records" json:"max_poll_records"`         // 最大拉取记录数
	FetchMinBytes      int           `yaml:"fetch_min_bytes" json:"fetch_min_bytes"`           // 最小拉取字节数
	FetchMaxBytes      int           `yaml:"fetch_max_bytes" json:"fetch_max_bytes"`           // 最大拉取字节数
	FetchMaxWait       time.Duration `yaml:"fetch_max_wait" json:"fetch_max_wait"`             // 最大拉取等待时间
	ReadTimeout        time.Duration `yaml:"read_timeout" json:"read_timeout"`                 // 读取超时
	WriteTimeout       time.Duration `yaml:"write_timeout" json:"write_timeout"`               // 写入超时
	InitialOffset      string        `yaml:"initial_offset" json:"initial_offset"`             // 初始偏移: earliest, latest
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	TLS  TLSConfig  `yaml:"tls" json:"tls"`   // TLS配置
	SASL SASLConfig `yaml:"sasl" json:"sasl"` // SASL配置
}

// TLSConfig TLS配置
type TLSConfig struct {
	Enabled    bool   `yaml:"enabled" json:"enabled"`         // 是否启用TLS
	CertFile   string `yaml:"cert_file" json:"cert_file"`     // 证书文件路径
	KeyFile    string `yaml:"key_file" json:"key_file"`       // 私钥文件路径
	CaFile     string `yaml:"ca_file" json:"ca_file"`         // CA证书文件路径
	VerifySSL  bool   `yaml:"verify_ssl" json:"verify_ssl"`   // 是否验证SSL
	ServerName string `yaml:"server_name" json:"server_name"` // 服务器名称
}

// SASLConfig SASL配置
type SASLConfig struct {
	Enabled   bool   `yaml:"enabled" json:"enabled"`     // 是否启用SASL
	Mechanism string `yaml:"mechanism" json:"mechanism"` // 认证机制: PLAIN, SCRAM-SHA-256, SCRAM-SHA-512
	Username  string `yaml:"username" json:"username"`   // 用户名
	Password  string `yaml:"password" json:"password"`   // 密码
}

// PerformanceConfig 性能配置
type PerformanceConfig struct {
	ConnectionPoolSize int           `yaml:"connection_pool_size" json:"connection_pool_size"` // 连接池大小
	ProducerPoolSize   int           `yaml:"producer_pool_size" json:"producer_pool_size"`     // 生产者池大小
	ConsumerPoolSize   int           `yaml:"consumer_pool_size" json:"consumer_pool_size"`     // 消费者池大小
	MetricsInterval    time.Duration `yaml:"metrics_interval" json:"metrics_interval"`         // 指标收集间隔
}

// KafkaBenchmarkConfig 基准测试配置
type KafkaBenchmarkConfig struct {
	DefaultTopic      string           `yaml:"default_topic" json:"default_topic"`           // 默认主题
	MessageSizeRange  MessageSizeRange `yaml:"message_size_range" json:"message_size_range"` // 消息大小范围
	BatchSizes        []int            `yaml:"batch_sizes" json:"batch_sizes"`               // 批处理大小列表
	PartitionStrategy string           `yaml:"partition_strategy" json:"partition_strategy"` // 分区策略
	Total             int              `yaml:"total" json:"total"`                           // 总请求数
	Parallels         int              `yaml:"parallels" json:"parallels"`                   // 并发数
	DataSize          int              `yaml:"data_size" json:"data_size"`                   // 数据大小
	TTL               time.Duration    `yaml:"ttl" json:"ttl"`                               // 生存时间
	ReadPercent       int              `yaml:"read_percent" json:"read_percent"`             // 读操作百分比
	RandomKeys        int              `yaml:"random_keys" json:"random_keys"`               // 随机键范围
	TestCase          string           `yaml:"test_case" json:"test_case"`                   // 测试用例
	TestType          string           `yaml:"test_type" json:"test_type"`                   // 测试类型
	MessageSize       int              `yaml:"message_size" json:"message_size"`             // 消息大小
	Timeout           time.Duration    `yaml:"timeout" json:"timeout"`                       // 超时时间
}

// MessageSizeRange 消息大小范围
type MessageSizeRange struct {
	Min int `yaml:"min" json:"min"` // 最小大小
	Max int `yaml:"max" json:"max"` // 最大大小
}

// 实现interfaces.Config接口

// GetProtocol 获取协议名称
func (c *KafkaAdapterConfig) GetProtocol() string {
	return "kafka"
}

// GetConnection 获取连接配置
func (c *KafkaAdapterConfig) GetConnection() interfaces.ConnectionConfig {
	return &ConnectionConfigImpl{
		addresses:   c.Brokers,
		credentials: c.getCredentials(),
		poolConfig:  c.getPoolConfig(),
		timeout:     c.Benchmark.Timeout,
	}
}

// GetBenchmark 获取基准测试配置
func (c *KafkaAdapterConfig) GetBenchmark() interfaces.BenchmarkConfig {
	return &c.Benchmark
}

// Validate 验证配置
func (c *KafkaAdapterConfig) Validate() error {
	// 验证基础配置
	if len(c.Brokers) == 0 {
		return fmt.Errorf("brokers cannot be empty")
	}

	if c.ClientID == "" {
		return fmt.Errorf("client_id cannot be empty")
	}

	// 验证生产者配置
	if err := c.validateProducerConfig(); err != nil {
		return fmt.Errorf("producer config validation failed: %w", err)
	}

	// 验证消费者配置
	if err := c.validateConsumerConfig(); err != nil {
		return fmt.Errorf("consumer config validation failed: %w", err)
	}

	// 验证安全配置
	if err := c.validateSecurityConfig(); err != nil {
		return fmt.Errorf("security config validation failed: %w", err)
	}

	// 验证性能配置
	if err := c.validatePerformanceConfig(); err != nil {
		return fmt.Errorf("performance config validation failed: %w", err)
	}

	// 验证基准测试配置
	if err := c.validateBenchmarkConfig(); err != nil {
		return fmt.Errorf("benchmark config validation failed: %w", err)
	}

	return nil
}

// Clone 克隆配置
func (c *KafkaAdapterConfig) Clone() interfaces.Config {
	clone := *c
	// 深拷贝切片
	clone.Brokers = make([]string, len(c.Brokers))
	copy(clone.Brokers, c.Brokers)

	clone.Benchmark.BatchSizes = make([]int, len(c.Benchmark.BatchSizes))
	copy(clone.Benchmark.BatchSizes, c.Benchmark.BatchSizes)

	return &clone
}

// getCredentials 获取认证信息
func (c *KafkaAdapterConfig) getCredentials() map[string]string {
	credentials := make(map[string]string)

	if c.Security.SASL.Enabled {
		credentials["sasl_mechanism"] = c.Security.SASL.Mechanism
		credentials["sasl_username"] = c.Security.SASL.Username
		credentials["sasl_password"] = c.Security.SASL.Password
	}

	if c.Security.TLS.Enabled {
		credentials["tls_enabled"] = "true"
		credentials["tls_cert_file"] = c.Security.TLS.CertFile
		credentials["tls_key_file"] = c.Security.TLS.KeyFile
		credentials["tls_ca_file"] = c.Security.TLS.CaFile
	}

	return credentials
}

// getPoolConfig 获取连接池配置
func (c *KafkaAdapterConfig) getPoolConfig() interfaces.PoolConfig {
	return &PoolConfigImpl{
		poolSize:          c.Performance.ConnectionPoolSize,
		minIdle:           c.Performance.ConnectionPoolSize / 2,
		maxIdle:           c.Performance.ConnectionPoolSize,
		idleTimeout:       30 * time.Second,
		connectionTimeout: c.Benchmark.Timeout,
	}
}

// validateProducerConfig 验证生产者配置
func (c *KafkaAdapterConfig) validateProducerConfig() error {
	// 验证acks设置
	validAcks := []string{"0", "1", "all", "-1"}
	if !contains(validAcks, c.Producer.Acks) {
		return fmt.Errorf("invalid acks value: %s, must be one of %v", c.Producer.Acks, validAcks)
	}

	// 验证压缩算法
	validCompression := []string{"none", "gzip", "snappy", "lz4", "zstd"}
	if !contains(validCompression, c.Producer.Compression) {
		return fmt.Errorf("invalid compression value: %s, must be one of %v", c.Producer.Compression, validCompression)
	}

	// 验证批处理大小
	if c.Producer.BatchSize <= 0 {
		return fmt.Errorf("batch_size must be positive, got: %d", c.Producer.BatchSize)
	}

	// 验证重试次数
	if c.Producer.Retries < 0 {
		return fmt.Errorf("retries must be non-negative, got: %d", c.Producer.Retries)
	}

	return nil
}

// validateConsumerConfig 验证消费者配置
func (c *KafkaAdapterConfig) validateConsumerConfig() error {
	// 验证偏移重置策略
	validOffsetReset := []string{"earliest", "latest"}
	if !contains(validOffsetReset, c.Consumer.AutoOffsetReset) {
		return fmt.Errorf("invalid auto_offset_reset value: %s, must be one of %v", c.Consumer.AutoOffsetReset, validOffsetReset)
	}

	// 验证消费者组ID
	if c.Consumer.GroupID == "" {
		return fmt.Errorf("consumer group_id cannot be empty")
	}

	// 验证拉取字节数设置
	if c.Consumer.FetchMinBytes <= 0 {
		return fmt.Errorf("fetch_min_bytes must be positive, got: %d", c.Consumer.FetchMinBytes)
	}

	if c.Consumer.FetchMaxBytes <= c.Consumer.FetchMinBytes {
		return fmt.Errorf("fetch_max_bytes must be greater than fetch_min_bytes")
	}

	return nil
}

// validateSecurityConfig 验证安全配置
func (c *KafkaAdapterConfig) validateSecurityConfig() error {
	// 验证SASL配置
	if c.Security.SASL.Enabled {
		validMechanisms := []string{"PLAIN", "SCRAM-SHA-256", "SCRAM-SHA-512"}
		if !contains(validMechanisms, c.Security.SASL.Mechanism) {
			return fmt.Errorf("invalid SASL mechanism: %s, must be one of %v", c.Security.SASL.Mechanism, validMechanisms)
		}

		if c.Security.SASL.Username == "" {
			return fmt.Errorf("SASL username cannot be empty when SASL is enabled")
		}

		if c.Security.SASL.Password == "" {
			return fmt.Errorf("SASL password cannot be empty when SASL is enabled")
		}
	}

	// 验证TLS配置
	if c.Security.TLS.Enabled {
		// TLS文件路径验证可选择性进行，这里只检查基本设置
	}

	return nil
}

// validatePerformanceConfig 验证性能配置
func (c *KafkaAdapterConfig) validatePerformanceConfig() error {
	if c.Performance.ConnectionPoolSize <= 0 {
		return fmt.Errorf("connection_pool_size must be positive, got: %d", c.Performance.ConnectionPoolSize)
	}

	if c.Performance.ProducerPoolSize <= 0 {
		return fmt.Errorf("producer_pool_size must be positive, got: %d", c.Performance.ProducerPoolSize)
	}

	if c.Performance.ConsumerPoolSize <= 0 {
		return fmt.Errorf("consumer_pool_size must be positive, got: %d", c.Performance.ConsumerPoolSize)
	}

	return nil
}

// validateBenchmarkConfig 验证基准测试配置
func (c *KafkaAdapterConfig) validateBenchmarkConfig() error {
	if c.Benchmark.DefaultTopic == "" {
		return fmt.Errorf("default_topic cannot be empty")
	}

	if c.Benchmark.Total <= 0 {
		return fmt.Errorf("total must be positive, got: %d", c.Benchmark.Total)
	}

	if c.Benchmark.Parallels <= 0 {
		return fmt.Errorf("parallels must be positive, got: %d", c.Benchmark.Parallels)
	}

	if c.Benchmark.ReadPercent < 0 || c.Benchmark.ReadPercent > 100 {
		return fmt.Errorf("read_percent must be between 0 and 100, got: %d", c.Benchmark.ReadPercent)
	}

	return nil
}

// contains 检查字符串切片是否包含指定元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
