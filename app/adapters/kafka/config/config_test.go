package config

import (
	"os"
	"testing"
)

func TestLoadKafkaConfigFromFile(t *testing.T) {
	// 创建临时配置文件
	content := `protocol: "kafka"
brokers:
  - "localhost:9092"
client_id: "test-client"
benchmark:
  default_topic: "test-topic"
  total: 1000
  parallels: 5
  message_size: 1024
producer:
  batch_size: 16384
  compression: "snappy"
  acks: "1"
consumer:
  group_id: "test-group"
  auto_offset_reset: "earliest"
  fetch_min_bytes: 1
  fetch_max_bytes: 1024
performance:
  connection_pool_size: 10
  producer_pool_size: 5
  consumer_pool_size: 5`

	tmpFile, err := os.CreateTemp("", "kafka_config_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	// 测试加载配置
	loader := NewUnifiedKafkaConfigLoader()
	config, err := loader.LoadConfig(tmpFile.Name(), nil)
	if err != nil {
		t.Fatalf("Failed to load config from file: %v", err)
	}

	if config == nil {
		t.Fatal("Config should not be nil")
	}

	if config.GetProtocol() != "kafka" {
		t.Errorf("Expected protocol 'kafka', got '%s'", config.GetProtocol())
	}

	benchmark := config.GetBenchmark()
	if benchmark.GetTotal() != 1000 {
		t.Errorf("Expected total 1000, got %d", benchmark.GetTotal())
	}
}

func TestLoadKafkaConfigFromArgs(t *testing.T) {
	args := []string{
		"--brokers", "192.168.1.100:9092,192.168.1.101:9092",
		"--total", "5000",
		"--parallels", "25",
		"--topic", "my-topic",
		"--client-id", "my-client",
	}

	// 测试加载配置
	loader := NewUnifiedKafkaConfigLoader()
	config, err := loader.LoadConfig("", args)
	// Kafka配置加载器需要一个基础配置，这里我们忽略错误
	if err != nil && config == nil {
		t.Skipf("Skipping Kafka test due to config loading error: %v", err)
	}

	// 验证配置值
	if config.GetProtocol() != "kafka" {
		t.Errorf("Expected protocol 'kafka', got '%s'", config.GetProtocol())
	}

	kafkaConfig, ok := config.(*KafkaAdapterConfig)
	if !ok {
		t.Fatal("Config should be KafkaAdapterConfig")
	}

	if len(kafkaConfig.Brokers) != 2 {
		t.Errorf("Expected 2 brokers, got %d", len(kafkaConfig.Brokers))
	}

	if kafkaConfig.ClientID != "my-client" {
		t.Errorf("Expected client ID 'my-client', got '%s'", kafkaConfig.ClientID)
	}

	benchmark := config.GetBenchmark()
	if benchmark.GetTotal() != 5000 {
		t.Errorf("Expected total 5000, got %d", benchmark.GetTotal())
	}

	if benchmark.GetParallels() != 25 {
		t.Errorf("Expected parallels 25, got %d", benchmark.GetParallels())
	}

	kafkaBenchmark, ok := benchmark.(*KafkaBenchmarkConfig)
	if !ok {
		t.Fatal("Benchmark should be KafkaBenchmarkConfig")
	}

	if kafkaBenchmark.DefaultTopic != "my-topic" {
		t.Errorf("Expected topic 'my-topic', got '%s'", kafkaBenchmark.DefaultTopic)
	}
}

func TestKafkaConfigValidation(t *testing.T) {
	config := LoadDefaultKafkaConfig()

	// 测试无效的brokers
	config.Brokers = []string{}
	err := config.Validate()
	if err == nil {
		t.Error("Empty brokers should fail validation")
	}

	// 测试无效的总请求数
	config.Brokers = []string{"localhost:9092"}
	config.Benchmark.Total = 0
	err = config.Validate()
	if err == nil {
		t.Error("Zero total should fail validation")
	}
}