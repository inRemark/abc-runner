package test

import (
	"context"
	"testing"
	"time"

	kafkaAdapter "abc-runner/app/adapters/kafka"
	kafkaConfig "abc-runner/app/adapters/kafka/config"
	"abc-runner/app/core/interfaces"
)

// 测试用的简单指标收集器
type testMetricsCollector struct {
	records []*interfaces.OperationResult
}

func (t *testMetricsCollector) Record(result *interfaces.OperationResult) {
	t.records = append(t.records, result)
}

func (t *testMetricsCollector) Snapshot() *interfaces.MetricsSnapshot[map[string]interface{}] {
	// 计算基本指标
	var total, success, failed, read, write int64
	for _, record := range t.records {
		total++
		if record.Success {
			success++
		} else {
			failed++
		}
		if record.IsRead {
			read++
		} else {
			write++
		}
	}
	
	return &interfaces.MetricsSnapshot[map[string]interface{}]{
		Core: interfaces.CoreMetrics{
			Operations: interfaces.OperationMetrics{
				Total:   total,
				Success: success,
				Failed:  failed,
				Read:    read,
				Write:   write,
				Rate:    func() float64 { if total > 0 { return float64(success) / float64(total) * 100 } else { return 0 } }(),
			},
		},
		Protocol:  map[string]interface{}{"test_data": "kafka"},
		Timestamp: time.Now(),
	}
}

func (t *testMetricsCollector) Reset() {
	t.records = nil
}

func (t *testMetricsCollector) Stop() {
	// 测试实现不需要特殊处理
}

// TestKafkaAdapter 测试Kafka适配器基本功能
func TestKafkaAdapter(t *testing.T) {
	adapter := kafkaAdapter.NewKafkaAdapter(&testMetricsCollector{}) // 注入指标收集器用于测试

	// 测试适配器基本属性
	if adapter.GetProtocolName() != "kafka" {
		t.Errorf("Expected protocol name 'kafka', got '%s'", adapter.GetProtocolName())
	}

	// 测试未连接状态
	if adapter.IsConnected() {
		t.Error("Adapter should not be connected initially")
	}
}

// TestKafkaAdapterConnect 测试Kafka适配器连接
func TestKafkaAdapterConnect(t *testing.T) {
	adapter := kafkaAdapter.NewKafkaAdapter(&testMetricsCollector{}) // 注入指标收集器用于测试
	config := createTestConfig()

	ctx := context.Background()
	err := adapter.Connect(ctx, config)

	// 注意：这个测试可能失败，因为测试环境可能没有运行Kafka服务器
	if err != nil {
		t.Logf("Connection failed (expected in test environment): %v", err)
		return
	}

	if !adapter.IsConnected() {
		t.Error("Adapter should be connected after successful connect")
	}

	// 测试连接关闭
	err = adapter.Close()
	if err != nil {
		t.Errorf("Failed to close adapter: %v", err)
	}
}

// TestKafkaAdapterMetrics 测试Kafka指标收集
func TestKafkaAdapterMetrics(t *testing.T) {
	adapter := kafkaAdapter.NewKafkaAdapter(&testMetricsCollector{}) // 注入指标收集器用于测试
	config := createTestConfig()

	ctx := context.Background()
	err := adapter.Connect(ctx, config)
	if err != nil {
		t.Skipf("Skipping metrics test due to connection failure: %v", err)
	}
	defer adapter.Close()

	// 获取指标收集器
	metricsCollector := adapter.GetMetricsCollector()
	if metricsCollector == nil {
		t.Error("Metrics collector should not be nil")
		return
	}

	// 模拟操作结果
	result := &interfaces.OperationResult{
		Success:  true,
		Duration: 100 * time.Millisecond,
		IsRead:   false,
	}

	metricsCollector.Record(result)

	// 检查指标更新
	snapshot := metricsCollector.Snapshot()
	if snapshot.Core.Operations.Total != 1 {
		t.Errorf("Expected total ops 1, got %d", snapshot.Core.Operations.Total)
	}
}

// TestKafkaOperationFactory 测试Kafka操作工厂
func TestKafkaOperationFactory(t *testing.T) {
	adapter := kafkaAdapter.NewKafkaAdapter(&testMetricsCollector{}) // 注入指标收集器用于测试
	config := createTestConfig()

	ctx := context.Background()
	err := adapter.Connect(ctx, config)
	if err != nil {
		t.Skipf("Skipping operation factory test due to connection failure: %v", err)
	}
	defer adapter.Close()

	// 测试操作工厂
	factory := adapter.GetOperationFactory()
	if factory == nil {
		t.Error("Operation factory should not be nil")
		return
	}

	// 创建操作
	params := map[string]interface{}{
		"operation_type": "produce_message",
		"topic":          "test_topic",
	}

	operation, err := factory.CreateOperation(params)
	if err != nil {
		t.Errorf("Failed to create operation: %v", err)
		return
	}

	// 验证操作
	if operation.Type != "produce_message" {
		t.Errorf("Expected operation type 'produce_message', got '%s'", operation.Type)
	}

	if topic, exists := operation.Params["topic"]; !exists || topic != "test_topic" {
		t.Error("Expected topic 'test_topic' in operation params")
	}
}

// createTestConfig 创建测试配置
func createTestConfig() *kafkaConfig.KafkaAdapterConfig {
	return &kafkaConfig.KafkaAdapterConfig{
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
}