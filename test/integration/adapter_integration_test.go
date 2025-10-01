// Package integration 集成测试包
// 提供多协议适配器的集成测试用例，验证系统组件间的协作能力
package integration

import (
	"context"
	"testing"
	"time"

	"abc-runner/app/adapters/http"
	httpConfig "abc-runner/app/adapters/http/config"
	"abc-runner/app/adapters/redis"
	redisConfig "abc-runner/app/adapters/redis/config"
	"abc-runner/app/adapters/tcp"
	tcpConfig "abc-runner/app/adapters/tcp/config"
	"abc-runner/app/core/interfaces"
	"abc-runner/app/core/metrics"
)

// TestAdapterIntegration 集成测试适配器基本功能
// 涵盖Redis、HTTP、TCP三种主要协议的适配器测试
func TestAdapterIntegration(t *testing.T) {
	ctx := context.Background()

	t.Run("Redis Adapter Integration", func(t *testing.T) {
		testRedisAdapter(t, ctx)
	})

	t.Run("HTTP Adapter Integration", func(t *testing.T) {
		testHTTPAdapter(t, ctx)
	})

	t.Run("TCP Adapter Integration", func(t *testing.T) {
		testTCPAdapter(t, ctx)
	})
}

// testRedisAdapter Redis适配器集成测试
// 测试场景：连接、健康检查、接口验证、指标收集
func testRedisAdapter(t *testing.T, ctx context.Context) {
	// 创建指标收集器
	metricsConfig := metrics.DefaultMetricsConfig()
	collector := metrics.NewBaseCollector(metricsConfig, map[string]interface{}{
		"protocol":  "redis",
		"test_type": "integration",
	})
	defer collector.Stop()

	// 创建Redis适配器
	adapter := redis.NewRedisAdapter(collector)

	// 创建测试配置
	config := redisConfig.NewDefaultRedisConfig()
	config.Standalone.Addr = "localhost:6379"
	config.BenchMark.Total = 10
	config.BenchMark.Parallels = 2

	// 验证配置
	if err := config.Validate(); err != nil {
		t.Fatalf("Config validation failed: %v", err)
	}

	// 测试连接（允许失败，因为可能没有Redis服务器）
	err := adapter.Connect(ctx, config)
	if err != nil {
		t.Logf("Redis connection failed (expected if no server): %v", err)
		// 继续测试其他功能
	} else {
		t.Log("✅ Redis connection successful")
		defer adapter.Close()

		// 执行健康检查
		if err := adapter.HealthCheck(ctx); err != nil {
			t.Logf("Health check failed: %v", err)
		} else {
			t.Log("✅ Redis health check passed")
		}
	}

	// 验证适配器接口实现
	validateAdapterInterface(t, adapter, "redis")

	// 检查指标收集
	snapshot := collector.Snapshot()
	if snapshot == nil {
		t.Error("Metrics snapshot should not be nil")
	}

	t.Log("✅ Redis adapter integration test completed")
}

// testHTTPAdapter HTTP适配器集成测试
// 测试场景：HTTP请求、连接管理、响应处理
func testHTTPAdapter(t *testing.T, ctx context.Context) {
	// 创建指标收集器
	metricsConfig := metrics.DefaultMetricsConfig()
	collector := metrics.NewBaseCollector(metricsConfig, map[string]interface{}{
		"protocol":  "http",
		"test_type": "integration",
	})
	defer collector.Stop()

	// 创建HTTP适配器
	adapter := http.NewHttpAdapter(collector)

	// 创建测试配置
	config := httpConfig.LoadDefaultHttpConfig()
	config.Connection.BaseURL = "https://httpbin.org"
	config.Benchmark.Total = 5
	config.Benchmark.Parallels = 2
	config.Benchmark.Path = "/get"
	config.Benchmark.Method = "GET"

	// 添加请求配置以满足验证要求
	config.Requests = []httpConfig.HttpRequestConfig{
		{
			Method: "GET",
			Path:   "/get",
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Weight: 1,
		},
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		t.Fatalf("Config validation failed: %v", err)
	}

	// 测试连接
	err := adapter.Connect(ctx, config)
	if err != nil {
		t.Logf("HTTP connection failed: %v", err)
	} else {
		t.Log("✅ HTTP connection successful")
		defer adapter.Close()

		// 执行健康检查
		if err := adapter.HealthCheck(ctx); err != nil {
			t.Logf("Health check failed: %v", err)
		} else {
			t.Log("✅ HTTP health check passed")
		}

		// 执行简单的HTTP请求测试
		operation := interfaces.Operation{
			Type: "GET",
			Key:  "/get",
			Params: map[string]interface{}{
				"test": "integration",
			},
		}

		result, err := adapter.Execute(ctx, operation)
		if err != nil {
			t.Logf("HTTP operation failed: %v", err)
		} else if result.Success {
			t.Log("✅ HTTP operation successful")
		}
	}

	// 验证适配器接口实现
	validateAdapterInterface(t, adapter, "http")

	t.Log("✅ HTTP adapter integration test completed")
}

// testTCPAdapter TCP适配器集成测试
// 测试场景：TCP连接、数据传输、连接管理
func testTCPAdapter(t *testing.T, ctx context.Context) {
	// 创建指标收集器
	metricsConfig := metrics.DefaultMetricsConfig()
	collector := metrics.NewBaseCollector(metricsConfig, map[string]interface{}{
		"protocol":  "tcp",
		"test_type": "integration",
	})
	defer collector.Stop()

	// 创建TCP适配器
	adapter := tcp.NewTCPAdapter(collector)

	// 创建测试配置
	config := tcpConfig.NewDefaultTCPConfig()
	config.Connection.Address = "localhost"
	config.Connection.Port = 8080
	config.BenchMark.Total = 5
	config.BenchMark.Parallels = 2
	config.BenchMark.TestCase = "echo_test"

	// 验证配置
	if err := config.Validate(); err != nil {
		t.Fatalf("Config validation failed: %v", err)
	}

	// 测试连接（允许失败，因为可能没有TCP服务器）
	err := adapter.Connect(ctx, config)
	if err != nil {
		t.Logf("TCP connection failed (expected if no server): %v", err)
	} else {
		t.Log("✅ TCP connection successful")
		defer adapter.Close()

		// 执行健康检查
		if err := adapter.HealthCheck(ctx); err != nil {
			t.Logf("Health check failed: %v", err)
		} else {
			t.Log("✅ TCP health check passed")
		}
	}

	// 验证适配器接口实现
	validateAdapterInterface(t, adapter, "tcp")

	t.Log("✅ TCP adapter integration test completed")
}

// validateAdapterInterface 验证适配器接口实现的完整性
// 检查协议名称、指标收集器、协议指标等核心功能
func validateAdapterInterface(t *testing.T, adapter interfaces.ProtocolAdapter, expectedProtocol string) {
	// 检查协议名称
	protocolName := adapter.GetProtocolName()
	if protocolName != expectedProtocol {
		t.Errorf("Expected protocol '%s', got '%s'", expectedProtocol, protocolName)
	}

	// 检查指标收集器
	collector := adapter.GetMetricsCollector()
	if collector == nil {
		t.Error("Metrics collector should not be nil")
	}

	// 检查协议指标
	metrics := adapter.GetProtocolMetrics()
	if metrics == nil {
		t.Error("Protocol metrics should not be nil")
	}

	t.Logf("✅ %s adapter interface validation passed", expectedProtocol)
}

// TestConfigValidation 配置验证集成测试
// 验证各协议配置的正确性和错误处理能力
func TestConfigValidation(t *testing.T) {
	t.Run("Redis Config Validation", func(t *testing.T) {
		config := redisConfig.NewDefaultRedisConfig()

		// 测试有效配置
		if err := config.Validate(); err != nil {
			t.Errorf("Valid config should pass validation: %v", err)
		}

		// 测试无效配置
		config.Standalone.Addr = ""
		if err := config.Validate(); err == nil {
			t.Error("Invalid config should fail validation")
		}

		// 测试配置克隆
		config = redisConfig.NewDefaultRedisConfig()
		cloned := config.Clone()
		if cloned == nil {
			t.Error("Cloned config should not be nil")
		}

		t.Log("✅ Redis config validation test passed")
	})

	t.Run("HTTP Config Validation", func(t *testing.T) {
		config := httpConfig.LoadDefaultHttpConfig()

		// 添加需要的请求配置
		config.Requests = []httpConfig.HttpRequestConfig{
			{
				Method: "GET",
				Path:   "/test",
				Weight: 1,
			},
		}

		// 测试有效配置
		if err := config.Validate(); err != nil {
			t.Errorf("Valid config should pass validation: %v", err)
		}

		// 测试无效配置
		config.Connection.BaseURL = ""
		if err := config.Validate(); err == nil {
			t.Error("Invalid config should fail validation")
		}

		t.Log("✅ HTTP config validation test passed")
	})

	t.Run("TCP Config Validation", func(t *testing.T) {
		config := tcpConfig.NewDefaultTCPConfig()

		// 测试有效配置
		if err := config.Validate(); err != nil {
			t.Errorf("Valid config should pass validation: %v", err)
		}

		// 测试无效配置
		config.Connection.Address = ""
		if err := config.Validate(); err == nil {
			t.Error("Invalid config should fail validation")
		}

		t.Log("✅ TCP config validation test passed")
	})
}

// TestMetricsIntegration 指标收集集成测试
// 验证指标收集器的功能和数据统计准确性
func TestMetricsIntegration(t *testing.T) {
	metricsConfig := metrics.DefaultMetricsConfig()
	collector := metrics.NewBaseCollector(metricsConfig, map[string]interface{}{
		"protocol":  "test",
		"test_type": "integration",
	})
	defer collector.Stop()

	// 模拟操作结果
	result := &interfaces.OperationResult{
		Success:  true,
		Duration: 100 * time.Millisecond,
		IsRead:   true,
		Value:    "test_value",
	}

	// 记录操作
	collector.Record(result)

	// 获取快照
	snapshot := collector.Snapshot()
	if snapshot == nil {
		t.Fatal("Snapshot should not be nil")
	}

	// 检查核心指标
	if snapshot.Core.Operations.Total == 0 {
		t.Error("Total operations should be greater than 0")
	}

	if snapshot.Core.Operations.Success == 0 {
		t.Error("Success operations should be greater than 0")
	}

	if snapshot.Core.Operations.Read == 0 {
		t.Error("Read operations should be greater than 0")
	}

	t.Log("✅ Metrics integration test passed")
}

// BenchmarkAdapterCreation 适配器创建性能基准测试
// 评估各适配器的创建性能，用于性能调优参考
func BenchmarkAdapterCreation(b *testing.B) {
	metricsConfig := metrics.DefaultMetricsConfig()
	collector := metrics.NewBaseCollector(metricsConfig, map[string]interface{}{
		"protocol":  "benchmark",
		"test_type": "performance",
	})
	defer collector.Stop()

	b.Run("Redis Adapter Creation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			adapter := redis.NewRedisAdapter(collector)
			_ = adapter
		}
	})

	b.Run("HTTP Adapter Creation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			adapter := http.NewHttpAdapter(collector)
			_ = adapter
		}
	})

	b.Run("TCP Adapter Creation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			adapter := tcp.NewTCPAdapter(collector)
			_ = adapter
		}
	})
}
