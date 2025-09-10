package utils

import (
	"context"
	"testing"
	"time"

	"abc-runner/app/core/interfaces"
)

func TestDefaultKeyGenerator(t *testing.T) {
	gen := NewDefaultKeyGenerator()
	
	// 测试生成递增键
	key1 := gen.GenerateKey("test", 0)
	key2 := gen.GenerateKey("test", 1)
	
	if key1 == key2 {
		t.Error("Expected different keys for different indices")
	}
	
	// 测试生成随机键
	randomKey1 := gen.GenerateRandomKey("test", 100)
	randomKey2 := gen.GenerateRandomKey("test", 100)
	
	if randomKey1 == randomKey2 {
		// 虽然随机键可能相同，但大概率不同
		t.Log("Warning: Generated random keys are the same")
	}
	
	// 测试获取已生成的键
	keys := gen.GetGeneratedKeys()
	if len(keys) == 0 {
		t.Error("Expected generated keys to be stored")
	}
	
	// 测试从已生成键中随机选择
	randomFromGenerated := gen.GetRandomFromGenerated()
	if randomFromGenerated == "" {
		t.Error("Expected non-empty key from generated keys")
	}
	
	// 测试重置
	gen.Reset()
	keysAfterReset := gen.GetGeneratedKeys()
	if len(keysAfterReset) != 0 {
		t.Error("Expected empty keys after reset")
	}
}

func TestOperationRegistry(t *testing.T) {
	registry := NewOperationRegistry()
	
	// 创建测试工厂
	factory := &TestOperationFactory{}
	
	// 测试注册
	registry.Register("test", factory)
	
	// 测试获取工厂
	retrievedFactory, exists := registry.GetFactory("test")
	if !exists {
		t.Error("Expected factory to exist")
	}
	
	if retrievedFactory != factory {
		t.Error("Expected same factory instance")
	}
	
	// 测试获取支持的操作
	operations := registry.GetSupportedOperations()
	if len(operations) != 1 || operations[0] != "test" {
		t.Errorf("Expected ['test'], got %v", operations)
	}
	
	// 测试创建操作
	params := map[string]interface{}{"key": "test_key"}
	operation, err := registry.CreateOperation("test", params)
	if err != nil {
		t.Fatalf("Failed to create operation: %v", err)
	}
	
	if operation.Type != "test" {
		t.Errorf("Expected operation type 'test', got '%s'", operation.Type)
	}
	
	// 测试验证操作
	err = registry.ValidateOperation("test", params)
	if err != nil {
		t.Errorf("Validation failed: %v", err)
	}
	
	// 测试不存在的操作类型
	_, err = registry.CreateOperation("nonexistent", params)
	if err == nil {
		t.Error("Expected error for nonexistent operation type")
	}
}

func TestRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()
	
	// 测试默认值
	if config.MaxRetries != 3 {
		t.Errorf("Expected max retries 3, got %d", config.MaxRetries)
	}
	
	if config.InitialDelay != 100*time.Millisecond {
		t.Errorf("Expected initial delay 100ms, got %v", config.InitialDelay)
	}
	
	// 测试延迟计算
	delay1 := config.GetDelay(0)
	delay2 := config.GetDelay(1)
	delay3 := config.GetDelay(2)
	
	if delay2 <= delay1 {
		t.Error("Expected increasing delay")
	}
	
	if delay3 <= delay2 {
		t.Error("Expected increasing delay")
	}
	
	// 测试最大延迟限制
	delay10 := config.GetDelay(10)
	if delay10 > config.MaxDelay {
		t.Errorf("Expected delay not to exceed max delay %v, got %v", config.MaxDelay, delay10)
	}
}

func TestProgressTracker(t *testing.T) {
	total := int64(1000)
	tracker := NewProgressTracker(total)
	
	// 初始状态
	current, totalReturned, percentage, _ := tracker.GetProgress()
	if current != 0 {
		t.Errorf("Expected initial current 0, got %d", current)
	}
	
	if totalReturned != total {
		t.Errorf("Expected total %d, got %d", total, totalReturned)
	}
	
	if percentage != 0 {
		t.Errorf("Expected initial percentage 0, got %.2f", percentage)
	}
	
	if tracker.IsCompleted() {
		t.Error("Expected tracker not to be completed initially")
	}
	
	// 更新进度
	tracker.Update(100)
	current, _, percentage, _ = tracker.GetProgress()
	
	if current != 100 {
		t.Errorf("Expected current 100, got %d", current)
	}
	
	expectedPercentage := 10.0
	if percentage != expectedPercentage {
		t.Errorf("Expected percentage %.2f, got %.2f", expectedPercentage, percentage)
	}
	
	// 完成进度
	tracker.Update(900)
	if !tracker.IsCompleted() {
		t.Error("Expected tracker to be completed")
	}
}

func TestTestContext(t *testing.T) {
	// 创建模拟组件
	adapter := &TestAdapter{}
	config := &TestConfig{}
	metricsCollector := &TestMetricsCollector{}
	keyGenerator := NewDefaultKeyGenerator()
	
	// 创建测试上下文
	ctx := NewTestContext(adapter, config, metricsCollector, keyGenerator)
	
	// 测试getter方法
	if ctx.GetAdapter() != adapter {
		t.Error("Expected same adapter instance")
	}
	
	if ctx.GetConfig() != config {
		t.Error("Expected same config instance")
	}
	
	if ctx.GetMetricsCollector() != metricsCollector {
		t.Error("Expected same metrics collector instance")
	}
	
	if ctx.GetKeyGenerator() != keyGenerator {
		t.Error("Expected same key generator instance")
	}
	
	// 测试取消功能
	if ctx.IsCancelled() {
		t.Error("Expected context not to be cancelled initially")
	}
	
	ctx.Cancel()
	if !ctx.IsCancelled() {
		t.Error("Expected context to be cancelled after Cancel()")
	}
}

// 测试辅助类型

type TestOperationFactory struct{}

func (f *TestOperationFactory) CreateOperation(params map[string]interface{}) (interfaces.Operation, error) {
	key, _ := params["key"].(string)
	return interfaces.Operation{
		Type: "test",
		Key:  key,
	}, nil
}

func (f *TestOperationFactory) GetOperationType() string {
	return "test"
}

func (f *TestOperationFactory) ValidateParams(params map[string]interface{}) error {
	return nil
}

type TestAdapter struct{}

func (a *TestAdapter) Connect(ctx context.Context, config interfaces.Config) error {
	return nil
}

func (a *TestAdapter) Execute(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	return &interfaces.OperationResult{Success: true}, nil
}

func (a *TestAdapter) Close() error {
	return nil
}

func (a *TestAdapter) GetProtocolMetrics() map[string]interface{} {
	return make(map[string]interface{})
}

func (a *TestAdapter) HealthCheck(ctx context.Context) error {
	return nil
}

func (a *TestAdapter) GetProtocolName() string {
	return "test"
}

type TestConfig struct{}

func (c *TestConfig) GetProtocol() string {
	return "test"
}

func (c *TestConfig) GetConnection() interfaces.ConnectionConfig {
	return nil
}

func (c *TestConfig) GetBenchmark() interfaces.BenchmarkConfig {
	return nil
}

func (c *TestConfig) Validate() error {
	return nil
}

func (c *TestConfig) Clone() interfaces.Config {
	return &TestConfig{}
}

type TestMetricsCollector struct{}

func (m *TestMetricsCollector) RecordOperation(result *interfaces.OperationResult) {}

func (m *TestMetricsCollector) GetMetrics() *interfaces.Metrics {
	return &interfaces.Metrics{}
}

func (m *TestMetricsCollector) Reset() {}

func (m *TestMetricsCollector) Export() map[string]interface{} {
	return make(map[string]interface{})
}