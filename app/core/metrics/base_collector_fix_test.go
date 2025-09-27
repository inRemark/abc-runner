package metrics

import (
	"testing"
	"time"
	
	"abc-runner/app/core/interfaces"
)

func TestLatencyTrackerMinValueFix(t *testing.T) {
	// 测试延迟追踪器的最小值修复
	config := LatencyConfig{
		HistorySize:     1000,
		ComputeInterval: 100 * time.Millisecond,
		SamplingRate:    1.0,
	}
	
	tracker := NewLatencyTracker(config)
	
	// 测试空数据情况
	metrics := tracker.GetMetrics()
	if metrics.Min != 0 || metrics.Max != 0 || metrics.Average != 0 {
		t.Errorf("Expected zero metrics for empty tracker, got Min=%v, Max=%v, Average=%v", 
			metrics.Min, metrics.Max, metrics.Average)
	}
	
	// 添加一些延迟数据
	tracker.Record(50 * time.Millisecond)
	tracker.Record(100 * time.Millisecond)
	tracker.Record(75 * time.Millisecond)
	
	// 等待计算间隔
	time.Sleep(200 * time.Millisecond)
	
	metrics = tracker.GetMetrics()
	
	// 验证修复后的指标
	if metrics.Min != 50*time.Millisecond {
		t.Errorf("Expected Min=50ms, got %v", metrics.Min)
	}
	
	if metrics.Max != 100*time.Millisecond {
		t.Errorf("Expected Max=100ms, got %v", metrics.Max)
	}
	
	expectedAvg := 75 * time.Millisecond
	if metrics.Average != expectedAvg {
		t.Errorf("Expected Average=75ms, got %v", metrics.Average)
	}
	
	// 验证分位数不为零
	if metrics.P50 == 0 || metrics.P90 == 0 {
		t.Errorf("Expected non-zero percentiles, got P50=%v, P90=%v", metrics.P50, metrics.P90)
	}
}

func TestSystemTrackerInitialization(t *testing.T) {
	// 测试系统追踪器的初始化修复
	config := SystemConfig{
		MonitorInterval:   time.Second,
		SnapshotRetention: 10,
		Enabled:           true,
	}
	
	tracker := NewSystemTracker(config)
	
	// 验证初始化后立即有系统指标数据
	metrics := tracker.GetMetrics()
	
	if metrics.GoroutineCount == 0 {
		t.Error("Expected non-zero goroutine count after initialization")
	}
	
	if metrics.MemoryUsage.Allocated == 0 {
		t.Error("Expected non-zero allocated memory after initialization")
	}
	
	if metrics.GCStats.NumGC < 0 {
		t.Error("Expected valid GC count after initialization")
	}
	
	if metrics.CPUUsage.Cores == 0 {
		t.Error("Expected non-zero CPU cores after initialization")
	}
}

func TestBaseCollectorIntegration(t *testing.T) {
	// 测试BaseCollector的集成修复
	config := DefaultMetricsConfig()
	protocolData := map[string]interface{}{
		"protocol": "test",
		"version":  "1.0",
	}
	
	collector := NewBaseCollector(config, protocolData)
	defer collector.Stop()
	
	// 记录一些操作结果
	results := []*interfaces.OperationResult{
		{Success: true, Duration: 10 * time.Millisecond, IsRead: true},
		{Success: true, Duration: 20 * time.Millisecond, IsRead: false},
		{Success: false, Duration: 5 * time.Millisecond, IsRead: true},
		{Success: true, Duration: 15 * time.Millisecond, IsRead: false},
	}
	
	for _, result := range results {
		collector.Record(result)
	}
	
	// 等待系统监控更新
	time.Sleep(1500 * time.Millisecond)
	
	snapshot := collector.Snapshot()
	
	// 验证操作指标
	if snapshot.Core.Operations.Total != 4 {
		t.Errorf("Expected 4 total operations, got %d", snapshot.Core.Operations.Total)
	}
	
	if snapshot.Core.Operations.Success != 3 {
		t.Errorf("Expected 3 successful operations, got %d", snapshot.Core.Operations.Success)
	}
	
	// 验证延迟指标非零
	if snapshot.Core.Latency.Min == 0 {
		t.Error("Expected non-zero minimum latency")
	}
	
	if snapshot.Core.Latency.Max == 0 {
		t.Error("Expected non-zero maximum latency")
	}
	
	if snapshot.Core.Latency.Average == 0 {
		t.Error("Expected non-zero average latency")
	}
	
	// 验证系统指标非零
	if snapshot.System.GoroutineCount == 0 {
		t.Error("Expected non-zero goroutine count in snapshot")
	}
	
	if snapshot.System.MemoryUsage.Allocated == 0 {
		t.Error("Expected non-zero allocated memory in snapshot")
	}
	
	// 验证吞吐量指标
	if snapshot.Core.Throughput.RPS == 0 {
		t.Error("Expected non-zero RPS in snapshot")
	}
}