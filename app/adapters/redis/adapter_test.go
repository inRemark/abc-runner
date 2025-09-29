package redis

import (
	"context"
	"testing"
	"time"

	redisConfig "abc-runner/app/adapters/redis/config"
	"abc-runner/app/core/metrics"
	"abc-runner/app/core/utils"
)

// TestRedisAdapterWithOperationsFactory 测试Redis适配器使用operations.go工厂
func TestRedisAdapterWithOperationsFactory(t *testing.T) {
	// 创建指标收集器
	metricsConfig := metrics.DefaultMetricsConfig()
	collector := metrics.NewBaseCollector(metricsConfig, map[string]interface{}{
		"protocol": "redis",
	})
	defer collector.Stop()

	// 创建Redis适配器
	adapter := NewRedisAdapter(collector)

	// 验证适配器创建成功
	if adapter == nil {
		t.Fatal("Failed to create Redis adapter")
	}

	// 验证协议名称
	if adapter.GetProtocolName() != "redis" {
		t.Errorf("Expected protocol name 'redis', got '%s'", adapter.GetProtocolName())
	}

	// 验证指标收集器
	if adapter.GetMetricsCollector() != collector {
		t.Error("Metrics collector not set correctly")
	}

	// 验证初始状态
	if adapter.IsConnected() {
		t.Error("Adapter should not be connected initially")
	}
}

// TestRedisAdapterOperationsRegistry 测试操作注册表功能
func TestRedisAdapterOperationsRegistry(t *testing.T) {
	// 创建指标收集器
	metricsConfig := metrics.DefaultMetricsConfig()
	collector := metrics.NewBaseCollector(metricsConfig, map[string]interface{}{
		"protocol": "redis",
	})
	defer collector.Stop()

	// 创建Redis适配器
	adapter := NewRedisAdapter(collector)

	// 创建测试配置 - 使用正确的配置结构
	config := redisConfig.NewDefaultRedisConfig()
	config.Mode = "standalone"
	config.Standalone.Addr = "127.0.0.1:6379"
	config.Standalone.Password = ""
	config.Standalone.Db = 0
	config.Pool.PoolSize = 10
	config.Pool.MinIdle = 2
	config.Pool.IdleTimeout = 300 * time.Second
	config.Pool.ConnectionTimeout = 30 * time.Second

	// 尝试连接（可能失败，但应该初始化操作注册表）
	ctx := context.Background()
	adapter.Connect(ctx, config)

	// 验证操作注册表已初始化
	registry := adapter.GetOperationRegistry()
	if registry == nil {
		t.Fatal("Operation registry should be initialized after Connect")
	}

	// 验证支持的操作类型
	supportedOps := adapter.GetSupportedOperationTypes()
	expectedOps := []string{
		"get", "set", "del", "set_get_random", "pub", "hset", "hget",
		"incr", "decr", "lpush", "rpush", "lpop", "rpop",
		"sadd", "smembers", "srem", "sismember",
		"zadd", "zrem", "zrange", "zrank",
		"hmset", "hmget", "hgetall", "sub",
	}

	if len(supportedOps) != len(expectedOps) {
		t.Errorf("Expected %d supported operations, got %d", len(expectedOps), len(supportedOps))
	}

	// 验证特定操作存在
	hasGet := false
	hasSet := false
	for _, op := range supportedOps {
		if op == "get" {
			hasGet = true
		}
		if op == "set" {
			hasSet = true
		}
	}

	if !hasGet {
		t.Error("GET operation should be supported")
	}
	if !hasSet {
		t.Error("SET operation should be supported")
	}
}

// TestCreateOperationFromFactory 测试从工厂创建操作
func TestCreateOperationFromFactory(t *testing.T) {
	// 创建指标收集器
	metricsConfig := metrics.DefaultMetricsConfig()
	collector := metrics.NewBaseCollector(metricsConfig, map[string]interface{}{
		"protocol": "redis",
	})
	defer collector.Stop()

	// 创建Redis适配器
	adapter := NewRedisAdapter(collector)

	// 创建测试配置
	config := redisConfig.NewDefaultRedisConfig()
	config.Mode = "standalone"
	config.Standalone.Addr = "127.0.0.1:6379"

	// 连接以初始化操作注册表
	ctx := context.Background()
	adapter.Connect(ctx, config)

	// 创建键生成器
	keyGen := utils.NewDefaultKeyGenerator()

	// 测试创建GET操作
	params := map[string]interface{}{
		"key_generator": keyGen,
		"random_keys":   10,
	}

	operation, err := adapter.CreateOperationFromFactory("get", params)
	if err != nil {
		t.Fatalf("Failed to create GET operation: %v", err)
	}

	if operation.Type != "get" {
		t.Errorf("Expected operation type 'get', got '%s'", operation.Type)
	}

	if operation.Key == "" {
		t.Error("Operation key should not be empty")
	}

	// 测试参数验证
	err = adapter.ValidateOperationParams("get", params)
	if err != nil {
		t.Errorf("GET operation params should be valid: %v", err)
	}

	// 测试无效操作类型
	_, err = adapter.CreateOperationFromFactory("invalid_op", params)
	if err == nil {
		t.Error("Should fail for invalid operation type")
	}
}

// TestValidateOperation 测试操作验证
func TestValidateOperation(t *testing.T) {
	// 创建指标收集器
	metricsConfig := metrics.DefaultMetricsConfig()
	collector := metrics.NewBaseCollector(metricsConfig, map[string]interface{}{
		"protocol": "redis",
	})
	defer collector.Stop()

	// 创建Redis适配器
	adapter := NewRedisAdapter(collector)

	// 测试支持的操作
	supportedOps := []string{"get", "set", "del", "incr", "decr", "hget", "hset"}
	for _, op := range supportedOps {
		err := adapter.ValidateOperation(op)
		if err != nil {
			t.Errorf("Operation '%s' should be supported, got error: %v", op, err)
		}
	}

	// 测试不支持的操作
	err := adapter.ValidateOperation("unsupported_operation")
	if err == nil {
		t.Error("Should fail for unsupported operation")
	}
}
