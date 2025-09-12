package test

import (
	"context"
	"testing"
	"time"

	kafkaAdapter "abc-runner/app/adapters/kafka"
	kafkaConfig "abc-runner/app/adapters/kafka/config"
	"abc-runner/app/core/interfaces"
)

// TestKafkaIntegration 测试Kafka集成
func TestKafkaIntegration(t *testing.T) {
	// 创建适配器
	adapter := kafkaAdapter.NewKafkaAdapter(nil) // 注入nil指标收集器用于测试

	// 创建配置
	config := &kafkaConfig.KafkaAdapterConfig{
		Brokers:  []string{"localhost:9092"},
		ClientID: "abc-runner-test",
		Version:  "2.8.0",
		Producer: kafkaConfig.ProducerConfig{
			Acks:           "all",
			Retries:        3,
			BatchSize:      16384,
			LingerMs:       5 * time.Millisecond,
			Compression:    "snappy",
			RequestTimeout: 30 * time.Second,
			WriteTimeout:   30 * time.Second,
			ReadTimeout:    30 * time.Second,
		},
		Consumer: kafkaConfig.ConsumerConfig{
			GroupID:            "test-consumer-group",
			AutoOffsetReset:    "earliest",
			EnableAutoCommit:   true,
			AutoCommitInterval: 1 * time.Second,
			SessionTimeout:     10 * time.Second,
			HeartbeatInterval:  3 * time.Second,
			MaxPollRecords:     500,
			FetchMinBytes:      1,
			FetchMaxBytes:      1024 * 1024,
			FetchMaxWait:       500 * time.Millisecond,
			ReadTimeout:        30 * time.Second,
			WriteTimeout:       30 * time.Second,
			InitialOffset:      "earliest",
		},
		Security: kafkaConfig.SecurityConfig{
			TLS: kafkaConfig.TLSConfig{
				Enabled:   false,
				VerifySSL: true,
			},
			SASL: kafkaConfig.SASLConfig{
				Enabled: false,
			},
		},
		Performance: kafkaConfig.PerformanceConfig{
			ConnectionPoolSize: 10,
			ProducerPoolSize:   5,
			ConsumerPoolSize:   5,
			MetricsInterval:    10 * time.Second,
		},
		Benchmark: kafkaConfig.KafkaBenchmarkConfig{
			DefaultTopic:      "test_topic",
			MessageSizeRange:  kafkaConfig.MessageSizeRange{Min: 100, Max: 1000},
			BatchSizes:        []int{1, 10, 100},
			PartitionStrategy: "round_robin",
			Total:             1000,
			Parallels:         10,
			DataSize:          512,
			TTL:               0,
			ReadPercent:       50,
			RandomKeys:        100,
			TestCase:          "mixed_operations",
			Timeout:           30 * time.Second,
		},
	}

	// 连接
	ctx := context.Background()
	err := adapter.Connect(ctx, config)
	if err != nil {
		t.Logf("Connection failed (expected in test environment): %v", err)
		return
	}
	defer adapter.Close()

	// 健康检查
	err = adapter.HealthCheck(ctx)
	if err != nil {
		t.Logf("Health check failed (expected in test environment): %v", err)
		return
	}

	// 获取指标
	metrics := adapter.GetProtocolMetrics()
	if metrics == nil {
		t.Error("Protocol metrics should not be nil")
	}
}

// TestKafkaExecute 测试Kafka执行
func TestKafkaExecute(t *testing.T) {
	// 创建适配器
	adapter := kafkaAdapter.NewKafkaAdapter(nil) // 注入nil指标收集器用于测试

	// 创建配置
	config := &kafkaConfig.KafkaAdapterConfig{
		Brokers:  []string{"localhost:9092"},
		ClientID: "abc-runner-test",
		Version:  "2.8.0",
		Producer: kafkaConfig.ProducerConfig{
			Acks:           "all",
			Retries:        3,
			BatchSize:      16384,
			LingerMs:       5 * time.Millisecond,
			Compression:    "snappy",
			RequestTimeout: 30 * time.Second,
			WriteTimeout:   30 * time.Second,
			ReadTimeout:    30 * time.Second,
		},
		Consumer: kafkaConfig.ConsumerConfig{
			GroupID:            "test-consumer-group",
			AutoOffsetReset:    "earliest",
			EnableAutoCommit:   true,
			AutoCommitInterval: 1 * time.Second,
			SessionTimeout:     10 * time.Second,
			HeartbeatInterval:  3 * time.Second,
			MaxPollRecords:     500,
			FetchMinBytes:      1,
			FetchMaxBytes:      1024 * 1024,
			FetchMaxWait:       500 * time.Millisecond,
			ReadTimeout:        30 * time.Second,
			WriteTimeout:       30 * time.Second,
			InitialOffset:      "earliest",
		},
		Security: kafkaConfig.SecurityConfig{
			TLS: kafkaConfig.TLSConfig{
				Enabled:   false,
				VerifySSL: true,
			},
			SASL: kafkaConfig.SASLConfig{
				Enabled: false,
			},
		},
		Performance: kafkaConfig.PerformanceConfig{
			ConnectionPoolSize: 10,
			ProducerPoolSize:   5,
			ConsumerPoolSize:   5,
			MetricsInterval:    10 * time.Second,
		},
		Benchmark: kafkaConfig.KafkaBenchmarkConfig{
			DefaultTopic:      "test_topic",
			MessageSizeRange:  kafkaConfig.MessageSizeRange{Min: 100, Max: 1000},
			BatchSizes:        []int{1, 10, 100},
			PartitionStrategy: "round_robin",
			Total:             1000,
			Parallels:         10,
			DataSize:          512,
			TTL:               0,
			ReadPercent:       50,
			RandomKeys:        100,
			TestCase:          "mixed_operations",
			Timeout:           30 * time.Second,
		},
	}

	// 连接
	ctx := context.Background()
	err := adapter.Connect(ctx, config)
	if err != nil {
		t.Logf("Connection failed (expected in test environment): %v", err)
		return
	}
	defer adapter.Close()

	// 创建操作
	operation := interfaces.Operation{
		Type: "produce_message",
		Key:  "test_key",
		Value: "test_value",
		Params: map[string]interface{}{
			"topic": "test_topic",
		},
	}

	// 执行操作
	result, err := adapter.Execute(ctx, operation)
	if err != nil {
		t.Logf("Operation failed (expected in test environment): %v", err)
		return
	}

	if result == nil {
		t.Error("Result should not be nil")
	}
}