package main

import (
	"context"
	"fmt"
	"log"
	"time"

	httpAdapter "abc-runner/app/adapters/http"
	httpConfig "abc-runner/app/adapters/http/config"
	"abc-runner/app/core/interfaces"
)

func main() {
	// 创建HTTP适配器
	adapter := httpAdapter.NewHttpAdapter(nil) // 注入nil指标收集器用于示例

	// 创建测试配置
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

	// 连接适配器
	ctx := context.Background()
	err := adapter.Connect(ctx, config)
	if err != nil {
		log.Printf("Failed to connect: %v", err)
		// 在示例中继续执行，即使连接失败
	} else {
		defer adapter.Close()
		fmt.Println("Connected to HTTP server")
	}

	// 创建测试操作
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
		log.Printf("Operation failed: %v", err)
	} else {
		fmt.Printf("Operation successful: %+v\n", result)
	}

	// 获取指标
	metrics := adapter.GetProtocolMetrics()
	fmt.Printf("Protocol metrics: %+v\n", metrics)

	// 获取指标收集器
	metricsCollector := adapter.GetMetricsCollector()
	if metricsCollector != nil {
		metricsSnapshot := metricsCollector.Snapshot()
		fmt.Printf("Metrics snapshot: %+v\n", metricsSnapshot)
	}
}