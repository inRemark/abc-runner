package test

import (
	"context"
	"testing"
	"time"
	
	httpAdapter "redis-runner/app/adapters/http"
	httpConfig "redis-runner/app/adapters/http/config"
	"redis-runner/app/core/interfaces"
)

// TestHttpAdapterIntegration 集成测试
func TestHttpAdapterIntegration(t *testing.T) {
	// 跳过集成测试，除非明确启用
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	// 创建适配器
	adapter := httpAdapter.NewHttpAdapter()
	
	// 创建测试配置
	config := createTestConfig()
	
	// 修改配置为测试环境
	config.Connection.BaseURL = "http://httpbin.org"
	config.Connection.Timeout = 30 * time.Second
	config.Requests = []httpConfig.HttpRequestConfig{
		{
			Method:      "GET",
			Path:        "/get",
			Headers:     map[string]string{"Accept": "application/json"},
			ContentType: "application/json",
			Weight:      60,
		},
		{
			Method:      "POST",
			Path:        "/post",
			Headers:     map[string]string{"Content-Type": "application/json"},
			ContentType: "application/json",
			Body:        map[string]interface{}{"test": "data", "timestamp": "{{random.id}}"},
			Weight:      40,
		},
	}
	
	ctx := context.Background()
	
	// 连接
	err := adapter.Connect(ctx, config)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer adapter.Close()
	
	// 执行多个操作
	operationCount := 10
	for i := 0; i < operationCount; i++ {
		// 创建操作
		params := map[string]interface{}{
			"index": i,
		}
		
		operation, err := adapter.CreateOperation(params)
		if err != nil {
			t.Fatalf("Failed to create operation %d: %v", i, err)
		}
		
		// 执行操作
		result, err := adapter.Execute(ctx, operation)
		if err != nil {
			t.Errorf("Failed to execute operation %d: %v", i, err)
			continue
		}
		
		if result == nil {
			t.Errorf("Got nil result for operation %d", i)
			continue
		}
		
		t.Logf("Operation %d: Success=%v, Duration=%v, Type=%s", 
			i, result.Success, result.Duration, operation.Type)
	}
	
	// 检查指标
	metrics := adapter.GetMetricsCollector().GetMetrics()
	t.Logf("Final metrics: Total=%d, Success=%d, Failed=%d, RPS=%d", 
		metrics.TotalOps, metrics.SuccessOps, metrics.FailedOps, metrics.RPS)
	
	if metrics.TotalOps == 0 {
		t.Error("Expected some operations to be recorded")
	}
	
	// 生成报告
	report := adapter.GenerateSimpleReport()
	t.Logf("Performance Report:\n%s", report)
	
	// 测试协议特定指标
	protocolMetrics := adapter.GetProtocolMetrics()
	if protocolMetrics == nil {
		t.Error("Expected protocol metrics")
	}
}

// TestHttpAdapterConnectionPool 测试连接池
func TestHttpAdapterConnectionPool(t *testing.T) {
	adapter := httpAdapter.NewHttpAdapter()
	config := createTestConfig()
	
	ctx := context.Background()
	err := adapter.Connect(ctx, config)
	if err != nil {
		t.Skipf("Skipping connection pool test: %v", err)
	}
	defer adapter.Close()
	
	// 获取连接池统计
	poolStats := adapter.GetConnectionPoolStats()
	if poolStats == nil {
		t.Error("Expected connection pool stats")
		return
	}
	
	t.Logf("Connection pool stats: %+v", poolStats)
	
	// 验证连接池大小
	if poolSize, exists := poolStats["pool_size"]; exists {
		if poolSize.(int) <= 0 {
			t.Error("Expected positive pool size")
		}
	}
}

// BenchmarkHttpAdapter HTTP适配器性能基准测试
func BenchmarkHttpAdapter(b *testing.B) {
	adapter := httpAdapter.NewHttpAdapter()
	config := createTestConfig()
	
	ctx := context.Background()
	err := adapter.Connect(ctx, config)
	if err != nil {
		b.Skipf("Skipping benchmark: %v", err)
	}
	defer adapter.Close()
	
	// 创建操作
	operation := interfaces.Operation{
		Type: "http_get",
		Key:  "benchmark_test",
		Params: map[string]interface{}{
			"method": "GET",
			"path":   "/get",
		},
	}
	
	b.ResetTimer()
	
	// 运行基准测试
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := adapter.Execute(ctx, operation)
			if err != nil {
				b.Errorf("Execute failed: %v", err)
			}
		}
	})
}