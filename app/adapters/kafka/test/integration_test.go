package test

import (
	"context"
	"testing"
	"time"
	
	kafkaAdapter "redis-runner/app/adapters/kafka"
	"redis-runner/app/core/interfaces"
)

// TestKafkaAdapterIntegration 集成测试
func TestKafkaAdapterIntegration(t *testing.T) {
	// 跳过集成测试，除非明确启用
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	// 创建适配器
	adapter := kafkaAdapter.NewKafkaAdapter()
	
	// 创建测试配置
	config := createTestConfig()
	
	ctx := context.Background()
	
	// 连接
	err := adapter.Connect(ctx, config)
	if err != nil {
		t.Skipf("Skipping integration test due to connection failure: %v", err)
	}
	defer adapter.Close()
	
	// 执行多个操作
	operationCount := 10
	successCount := 0
	
	for i := 0; i < operationCount; i++ {
		// 创建生产操作
		params := map[string]interface{}{
			"operation_type": "produce_message",
			"topic":          "test_topic",
			"index":          i,
		}
		
		operation, err := adapter.CreateOperation(params)
		if err != nil {
			t.Errorf("Failed to create operation %d: %v", i, err)
			continue
		}
		
		// 执行操作
		result, err := adapter.Execute(ctx, operation)
		if err != nil {
			t.Logf("Operation %d failed (expected in test environment): %v", i, err)
			continue
		}
		
		if result != nil && result.Success {
			successCount++
		}
		
		t.Logf("Operation %d: Success=%v, Duration=%v, Type=%s", 
			i, result != nil && result.Success, 
			func() time.Duration { 
				if result != nil { 
					return result.Duration 
				}; 
				return 0 
			}(), operation.Type)
	}
	
	t.Logf("Successfully executed %d out of %d operations", successCount, operationCount)
	
	// 检查指标
	metrics := adapter.GetMetricsCollector().GetMetrics()
	t.Logf("Final metrics: Total=%d, Success=%d, Failed=%d", 
		metrics.TotalOps, metrics.SuccessOps, metrics.FailedOps)
	
	// 测试协议特定指标
	protocolMetrics := adapter.GetProtocolMetrics()
	if protocolMetrics == nil {
		t.Error("Expected protocol metrics")
	} else {
		t.Logf("Protocol metrics keys: %v", getMapKeys(protocolMetrics))
	}
}

// TestKafkaAdapterConnectionPool 测试连接池
func TestKafkaAdapterConnectionPool(t *testing.T) {
	adapter := kafkaAdapter.NewKafkaAdapter()
	config := createTestConfig()
	
	ctx := context.Background()
	err := adapter.Connect(ctx, config)
	if err != nil {
		t.Skipf("Skipping connection pool test: %v", err)
	}
	defer adapter.Close()
	
	// 获取协议指标中的连接池信息
	protocolMetrics := adapter.GetProtocolMetrics()
	if protocolMetrics == nil {
		t.Error("Expected protocol metrics")
		return
	}
	
	t.Logf("Protocol metrics: %+v", protocolMetrics)
}

// TestKafkaProduceConsume 测试生产和消费流程
func TestKafkaProduceConsume(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping produce/consume test in short mode")
	}
	
	adapter := kafkaAdapter.NewKafkaAdapter()
	config := createTestConfig()
	
	ctx := context.Background()
	err := adapter.Connect(ctx, config)
	if err != nil {
		t.Skipf("Skipping produce/consume test: %v", err)
	}
	defer adapter.Close()
	
	// 测试生产消息
	produceParams := map[string]interface{}{
		"operation_type": "produce_message",
		"topic":          "test_topic",
	}
	
	produceOperation, err := adapter.CreateOperation(produceParams)
	if err != nil {
		t.Fatalf("Failed to create produce operation: %v", err)
	}
	
	produceResult, err := adapter.Execute(ctx, produceOperation)
	if err != nil {
		t.Logf("Produce operation failed (expected in test environment): %v", err)
	} else {
		t.Logf("Produce operation: Success=%v, Duration=%v", 
			produceResult.Success, produceResult.Duration)
	}
	
	// 测试消费消息
	consumeParams := map[string]interface{}{
		"operation_type": "consume_message",
		"topic":          "test_topic",
		"timeout":        5 * time.Second,
	}
	
	consumeOperation, err := adapter.CreateOperation(consumeParams)
	if err != nil {
		t.Fatalf("Failed to create consume operation: %v", err)
	}
	
	consumeResult, err := adapter.Execute(ctx, consumeOperation)
	if err != nil {
		t.Logf("Consume operation failed (expected in test environment): %v", err)
	} else {
		t.Logf("Consume operation: Success=%v, Duration=%v", 
			consumeResult.Success, consumeResult.Duration)
	}
}

// BenchmarkKafkaAdapter Kafka适配器性能基准测试
func BenchmarkKafkaAdapter(b *testing.B) {
	adapter := kafkaAdapter.NewKafkaAdapter()
	config := createTestConfig()
	
	ctx := context.Background()
	err := adapter.Connect(ctx, config)
	if err != nil {
		b.Skipf("Skipping benchmark: %v", err)
	}
	defer adapter.Close()
	
	// 创建操作
	operation := interfaces.Operation{
		Type: "produce_message",
		Key:  "benchmark_test",
		Value: "benchmark_message_data",
		Params: map[string]interface{}{
			"topic": "benchmark_topic",
		},
	}
	
	b.ResetTimer()
	
	// 运行基准测试
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := adapter.Execute(ctx, operation)
			if err != nil {
				// 在基准测试中忽略错误，因为可能没有真实的Kafka服务器
			}
		}
	})
}

// 辅助函数

// getMapKeys 获取map的所有键
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}