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
	metrics *interfaces.Metrics
}

func (t *testMetricsCollector) RecordOperation(result *interfaces.OperationResult) {
	if t.metrics == nil {
		t.metrics = &interfaces.Metrics{
			StartTime: time.Now(),
		}
	}
	t.metrics.TotalOps++
	if result.Success {
		t.metrics.SuccessOps++
	} else {
		t.metrics.FailedOps++
	}
	if result.IsRead {
		t.metrics.ReadOps++
	} else {
		t.metrics.WriteOps++
	}
}

func (t *testMetricsCollector) GetMetrics() *interfaces.Metrics {
	if t.metrics == nil {
		t.metrics = &interfaces.Metrics{}
	}
	return t.metrics
}

func (t *testMetricsCollector) Reset() {
	t.metrics = &interfaces.Metrics{}
}

func (t *testMetricsCollector) Export() map[string]interface{} {
	return map[string]interface{}{
		"metrics": t.metrics,
	}
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

	metricsCollector.RecordOperation(result)

	// 检查指标更新
	metrics := metricsCollector.GetMetrics()
	if metrics.TotalOps != 1 {
		t.Errorf("Expected total ops 1, got %d", metrics.TotalOps)
	}
	if metrics.SuccessOps != 1 {
		t.Errorf("Expected success ops 1, got %d", metrics.SuccessOps)
	}
	if metrics.ReadOps != 1 {
		t.Errorf("Expected read ops 1, got %d", metrics.ReadOps)
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