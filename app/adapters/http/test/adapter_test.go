package test

import (
	"context"
	"testing"
	"time"

	httpAdapter "abc-runner/app/adapters/http"
	httpConfig "abc-runner/app/adapters/http/config"
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
				Rate:    float64(success) / float64(total) * 100,
			},
		},
		Protocol:  map[string]interface{}{"test_data": "http"},
		Timestamp: time.Now(),
	}
}

func (t *testMetricsCollector) Reset() {
	t.records = nil
}

func (t *testMetricsCollector) Stop() {
	// 测试实现不需要特殊处理
}

// TestHttpAdapter 测试HTTP适配器基本功能
func TestHttpAdapter(t *testing.T) {
	// 创建测试指标收集器
	metricsCollector := &testMetricsCollector{}
	adapter := httpAdapter.NewHttpAdapter(metricsCollector)

	if adapter == nil {
		t.Fatal("Expected adapter to be created, got nil")
	}

	// 测试适配器基本属性
	if adapter.GetProtocolName() != "http" {
		t.Errorf("Expected protocol name 'http', got '%s'", adapter.GetProtocolName())
	}

	// 测试未连接状态
	if adapter.IsConnected() {
		t.Error("Adapter should not be connected initially")
	}
}

// TestHttpAdapterConnect 测试HTTP适配器连接
func TestHttpAdapterConnect(t *testing.T) {
	// 创建测试指标收集器
	metricsCollector := &testMetricsCollector{}
	adapter := httpAdapter.NewHttpAdapter(metricsCollector)
	config := createTestConfig()

	ctx := context.Background()
	err := adapter.Connect(ctx, config)

	// 注意：这个测试可能失败，因为测试环境可能没有运行HTTP服务器
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

// TestHttpAdapterMetrics 测试HTTP指标收集
func TestHttpAdapterMetrics(t *testing.T) {
	// 创建测试指标收集器
	metricsCollector := &testMetricsCollector{}
	
	// 不需要连接，直接测试指标收集
	// 模拟操作结果
	result := &interfaces.OperationResult{
		Success:  true,
		Duration: 100 * time.Millisecond,
		IsRead:   true,
	}

	metricsCollector.Record(result)

	// 检查指标更新
	snapshot := metricsCollector.Snapshot()
	if snapshot.Core.Operations.Total != 1 {
		t.Errorf("Expected total ops 1, got %d", snapshot.Core.Operations.Total)
	}
	if snapshot.Core.Operations.Success != 1 {
		t.Errorf("Expected success ops 1, got %d", snapshot.Core.Operations.Success)
	}
	if snapshot.Core.Operations.Read != 1 {
		t.Errorf("Expected read ops 1, got %d", snapshot.Core.Operations.Read)
	}
}

// createTestConfig 创建测试配置
func createTestConfig() *httpConfig.HttpAdapterConfig {
	return &httpConfig.HttpAdapterConfig{
		Protocol: "http",
		Connection: httpConfig.HttpConnectionConfig{
			BaseURL:         "https://cn.bing.com",
			Timeout:         10 * time.Second,
			MaxIdleConns:    10,
			MaxConnsPerHost: 5,
		},
		Requests: []httpConfig.HttpRequestConfig{
			{
				Method:  "GET",
				Path:    "/",
				Headers: map[string]string{"Accept": "application/json"},
				Weight:  100,
			},
		},
		Auth: httpConfig.HttpAuthConfig{
			Type: "none",
		},
		Benchmark: httpConfig.HttpBenchmarkConfig{
			Total:     1000,
			Parallels: 10,
			Timeout:   10 * time.Second,
		},
	}
}