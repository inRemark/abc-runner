package execution

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"abc-runner/app/core/interfaces"
)

// 测试用的mock适配器
type mockProtocolAdapter struct {
	executeCount   int64
	shouldFail     bool
	executionDelay time.Duration
}

func (m *mockProtocolAdapter) Connect(ctx context.Context, config interfaces.Config) error {
	return nil
}

func (m *mockProtocolAdapter) Execute(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	atomic.AddInt64(&m.executeCount, 1)

	// 模拟执行延迟
	if m.executionDelay > 0 {
		time.Sleep(m.executionDelay)
	}

	result := &interfaces.OperationResult{
		Success:  !m.shouldFail,
		Duration: m.executionDelay,
		IsRead:   operation.Type == "read",
		Value:    operation.Value,
	}

	if m.shouldFail {
		result.Error = fmt.Errorf("mock error")
	}

	return result, nil
}

func (m *mockProtocolAdapter) Close() error {
	return nil
}

func (m *mockProtocolAdapter) HealthCheck(ctx context.Context) error {
	return nil
}

func (m *mockProtocolAdapter) GetProtocolName() string {
	return "mock"
}

func (m *mockProtocolAdapter) GetProtocolMetrics() map[string]interface{} {
	return map[string]interface{}{
		"execute_count": atomic.LoadInt64(&m.executeCount),
	}
}

func (m *mockProtocolAdapter) GetMetricsCollector() interfaces.DefaultMetricsCollector {
	return nil // mock实现
}

// 测试用的mock指标收集器
type mockMetricsCollector struct {
	recordCount int64
	results     []*interfaces.OperationResult
}

func (m *mockMetricsCollector) Record(result *interfaces.OperationResult) {
	atomic.AddInt64(&m.recordCount, 1)
	m.results = append(m.results, result)
}

func (m *mockMetricsCollector) Snapshot() *interfaces.MetricsSnapshot[map[string]interface{}] {
	return &interfaces.MetricsSnapshot[map[string]interface{}]{
		Core: interfaces.CoreMetrics{
			Operations: interfaces.OperationMetrics{
				Total: atomic.LoadInt64(&m.recordCount),
			},
		},
		Protocol: map[string]interface{}{
			"record_count": atomic.LoadInt64(&m.recordCount),
		},
	}
}

func (m *mockMetricsCollector) Reset() {
	atomic.StoreInt64(&m.recordCount, 0)
	m.results = nil
}

func (m *mockMetricsCollector) Stop() {
	// 清理资源
}

// 测试用的mock配置
type mockBenchmarkConfig struct {
	total     int
	parallels int
	duration  time.Duration
	timeout   time.Duration
	rampUp    time.Duration
}

func (m *mockBenchmarkConfig) GetTotal() int              { return m.total }
func (m *mockBenchmarkConfig) GetParallels() int          { return m.parallels }
func (m *mockBenchmarkConfig) GetDuration() time.Duration { return m.duration }
func (m *mockBenchmarkConfig) GetTimeout() time.Duration  { return m.timeout }
func (m *mockBenchmarkConfig) GetRampUp() time.Duration   { return m.rampUp }

// 测试用的mock操作工厂
type mockOperationFactory struct {
	operationType string
}

func (m *mockOperationFactory) CreateOperation(jobID int, config BenchmarkConfig) interfaces.Operation {
	return interfaces.Operation{
		Type:  m.operationType,
		Key:   "test_key",
		Value: "test_value",
		Params: map[string]interface{}{
			"job_id": jobID,
		},
	}
}

func TestExecutionEngine_NewExecutionEngine(t *testing.T) {
	adapter := &mockProtocolAdapter{}
	collector := &mockMetricsCollector{}
	factory := &mockOperationFactory{operationType: "test"}

	engine := NewExecutionEngine(adapter, collector, factory)

	if engine == nil {
		t.Fatal("NewExecutionEngine returned nil")
	}
}

func TestExecutionEngine_RunBenchmark_Basic(t *testing.T) {
	adapter := &mockProtocolAdapter{}
	collector := &mockMetricsCollector{}
	factory := &mockOperationFactory{operationType: "test"}

	engine := NewExecutionEngine(adapter, collector, factory)
	config := &mockBenchmarkConfig{
		total:     10,
		parallels: 2,
	}

	ctx := context.Background()
	result, err := engine.RunBenchmark(ctx, config)

	if err != nil {
		t.Fatalf("RunBenchmark failed: %v", err)
	}

	if result.TotalJobs != 10 {
		t.Errorf("Expected 10 total jobs, got %d", result.TotalJobs)
	}

	if result.CompletedJobs != 10 {
		t.Errorf("Expected 10 completed jobs, got %d", result.CompletedJobs)
	}

	// 检查适配器是否被调用了正确的次数
	executeCount := atomic.LoadInt64(&adapter.executeCount)
	if executeCount != 10 {
		t.Errorf("Expected adapter to be called 10 times, got %d", executeCount)
	}

	// 检查指标收集器是否记录了正确的次数
	recordCount := atomic.LoadInt64(&collector.recordCount)
	if recordCount != 10 {
		t.Errorf("Expected metrics collector to record 10 times, got %d", recordCount)
	}
}