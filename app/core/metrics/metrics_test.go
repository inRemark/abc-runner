package metrics

import (
	"testing"
	"time"

	"abc-runner/app/core/interfaces"
)

// 简化的测试，不依赖外部库
func TestBaseCollectorBasic(t *testing.T) {
	config := DefaultMetricsConfig()
	collector := NewBaseCollector(config, TestProtocolMetrics{})
	defer collector.Stop()

	// 记录一些操作
	for i := 0; i < 10; i++ {
		result := &interfaces.OperationResult{
			Success:  i%2 == 0,
			Duration: time.Duration(i+1) * time.Millisecond,
			IsRead:   i%2 == 0,
		}
		collector.Record(result)
	}

	snapshot := collector.Snapshot()
	if snapshot.Core.Operations.Total != 10 {
		t.Errorf("期望总操作数为10，实际为%d", snapshot.Core.Operations.Total)
	}

	if snapshot.Core.Operations.Success != 5 {
		t.Errorf("期望成功操作数为5，实际为%d", snapshot.Core.Operations.Success)
	}
}

func TestRingBufferBasic(t *testing.T) {
	rb := NewRingBuffer[int](3)
	
	// 测试添加元素
	rb.Push(1)
	rb.Push(2)
	rb.Push(3)
	
	if rb.Size() != 3 {
		t.Errorf("期望大小为3，实际为%d", rb.Size())
	}
	
	// 测试溢出
	rb.Push(4)
	if rb.Size() != 3 {
		t.Errorf("溢出后期望大小仍为3，实际为%d", rb.Size())
	}
}

func TestHealthCheckerBasic(t *testing.T) {
	thresholds := HealthThresholds{
		MemoryUsage: 80.0,
	}
	
	checker := NewHealthChecker(thresholds)
	
	// 测试健康状态
	metrics := SystemMetrics{
		Memory: MemoryMetrics{Usage: 50.0},
	}
	
	result := checker.Check(nil, metrics)
	if result.Status != HealthStatusHealthy {
		t.Errorf("期望健康状态，实际为%s", result.Status)
	}
	
	// 测试警告状态
	metrics.Memory.Usage = 85.0
	result = checker.Check(nil, metrics)
	if result.Status != HealthStatusWarning {
		t.Errorf("期望警告状态，实际为%s", result.Status)
	}
}

// 测试数据结构
type TestProtocolMetrics struct {
	TestField string `json:"test_field"`
}

// 基准测试
func BenchmarkBaseCollectorRecord(b *testing.B) {
	config := DefaultMetricsConfig()
	collector := NewBaseCollector(config, TestProtocolMetrics{})
	defer collector.Stop()

	result := &interfaces.OperationResult{
		Success:  true,
		Duration: 100 * time.Millisecond,
		IsRead:   true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.Record(result)
	}
}