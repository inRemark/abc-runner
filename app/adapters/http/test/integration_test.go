package test

import (
	"context"
	"testing"
	"time"

	httpAdapter "abc-runner/app/adapters/http"
	httpConfig "abc-runner/app/adapters/http/config"
	"abc-runner/app/core/interfaces"
)

// TestHTTPIntegration 测试HTTP集成
func TestHTTPIntegration(t *testing.T) {
	// 创建适配器
	adapter := httpAdapter.NewHttpAdapter(&testMetricsCollector{}) // 注入指标收集器用于测试

	// 创建配置
	config := &httpConfig.HttpAdapterConfig{
		Protocol: "http",
		Connection: httpConfig.HttpConnectionConfig{
			BaseURL:         "https://httpbin.org",
			Timeout:         30 * time.Second,
			MaxIdleConns:    100,
			MaxConnsPerHost: 50,
		},
		Requests: []httpConfig.HttpRequestConfig{
			{
				Method:  "GET",
				Path:    "/get",
				Headers: map[string]string{"Accept": "application/json"},
				Weight:  100,
			},
		},
		Auth: httpConfig.HttpAuthConfig{
			Type: "none",
		},
		Benchmark: httpConfig.HttpBenchmarkConfig{
			Total:     10,
			Parallels: 2,
			Timeout:   30 * time.Second,
		},
	}

	// 连接
	ctx := context.Background()
	err := adapter.Connect(ctx, config)
	if err != nil {
		t.Logf("Connection failed (expected in some environments): %v", err)
		return
	}
	defer adapter.Close()

	// 健康检查
	err = adapter.HealthCheck(ctx)
	if err != nil {
		t.Errorf("Health check failed: %v", err)
	}

	// 获取指标
	metrics := adapter.GetProtocolMetrics()
	if metrics == nil {
		t.Error("Protocol metrics should not be nil")
	}
}

// TestHTTPExecute 测试HTTP执行
func TestHTTPExecute(t *testing.T) {
	// 创建适配器
	adapter := httpAdapter.NewHttpAdapter(&testMetricsCollector{}) // 注入指标收集器用于测试

	// 创建配置
	config := &httpConfig.HttpAdapterConfig{
		Protocol: "http",
		Connection: httpConfig.HttpConnectionConfig{
			BaseURL:         "https://httpbin.org",
			Timeout:         30 * time.Second,
			MaxIdleConns:    100,
			MaxConnsPerHost: 50,
		},
		Requests: []httpConfig.HttpRequestConfig{
			{
				Method:  "GET",
				Path:    "/get",
				Headers: map[string]string{"Accept": "application/json"},
				Weight:  100,
			},
		},
		Auth: httpConfig.HttpAuthConfig{
			Type: "none",
		},
		Benchmark: httpConfig.HttpBenchmarkConfig{
			Total:     10,
			Parallels: 2,
			Timeout:   30 * time.Second,
		},
	}

	// 连接
	ctx := context.Background()
	err := adapter.Connect(ctx, config)
	if err != nil {
		t.Logf("Connection failed (expected in some environments): %v", err)
		return
	}
	defer adapter.Close()

	// 创建操作
	operation := interfaces.Operation{
		Type: "http_get",
		Key:  "test_request",
		Params: map[string]interface{}{
			"method": "GET",
			"path":   "/get",
		},
	}

	// 执行操作
	result, err := adapter.Execute(ctx, operation)
	if err != nil {
		t.Logf("Operation failed (expected in some environments): %v", err)
		return
	}

	if result == nil {
		t.Error("Result should not be nil")
	}

	if !result.Success {
		t.Error("Operation should be successful")
	}
}