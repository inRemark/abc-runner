package main

import (
	"context"
	"fmt"
	"log"
	"time"

	httpAdapter "abc-runner/app/adapters/http"
	httpConfig "abc-runner/app/adapters/http/config"
)

func main() {
	fmt.Println("=== HTTP Adapter Demo ===")

	// 创建HTTP适配器
	adapter := httpAdapter.NewHttpAdapter()
	fmt.Printf("✓ Created HTTP adapter: %s\n", adapter.GetProtocolName())

	// 加载默认配置
	config, err := httpConfig.LoadHTTPConfigDefault()
	if err != nil {
		log.Printf("Could not load default config, using test config: %v", err)
		// 使用测试配置
		config = &httpConfig.HttpAdapterConfig{
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
					Path:    "/search?q=redis+runner",
					Headers: map[string]string{"Accept": "application/json"},
					Weight:  100,
				},
			},
			Auth: httpConfig.HttpAuthConfig{Type: "none"},
			Benchmark: httpConfig.HttpBenchmarkConfig{
				Total:     10,
				Parallels: 2,
				Timeout:   10 * time.Second,
			},
		}
	}

	fmt.Println("✓ Configuration loaded")

	// 连接适配器
	ctx := context.Background()
	if err := adapter.Connect(ctx, config); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer adapter.Close()

	fmt.Println("✓ Connected successfully")
	fmt.Println("✓ HTTP Adapter implementation completed!")
}
