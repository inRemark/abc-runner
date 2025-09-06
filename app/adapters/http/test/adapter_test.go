package test

import (
	"context"
	"testing"
	"time"

	httpAdapter "redis-runner/app/adapters/http"
	httpConfig "redis-runner/app/adapters/http/config"
	"redis-runner/app/core/interfaces"
)

// TestHttpAdapter 测试HTTP适配器基本功能
func TestHttpAdapter(t *testing.T) {
	adapter := httpAdapter.NewHttpAdapter()

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
	adapter := httpAdapter.NewHttpAdapter()
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
	adapter := httpAdapter.NewHttpAdapter()
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
		IsRead:   true,
	}

	metricsCollector.RecordOperation(result)

	// 检查指标更新
	metrics := metricsCollector.GetMetrics()
	if metrics.TotalOps != 1 {
		t.Errorf("Expected total ops 1, got %d", metrics.TotalOps)
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
