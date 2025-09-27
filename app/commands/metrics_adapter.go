package commands

import (
	"time"

	"abc-runner/app/core/interfaces"
	"abc-runner/app/core/metrics"
)

// SharedMetricsAdapter 共享的指标适配器
type SharedMetricsAdapter struct {
	baseCollector *metrics.BaseCollector[map[string]interface{}]
}

func NewSharedMetricsAdapter(collector *metrics.BaseCollector[map[string]interface{}]) *SharedMetricsAdapter {
	return &SharedMetricsAdapter{
		baseCollector: collector,
	}
}

func (m *SharedMetricsAdapter) RecordOperation(result *interfaces.OperationResult) {
	if m.baseCollector != nil {
		m.baseCollector.Record(result)
	}
}

func (m *SharedMetricsAdapter) GetMetrics() *interfaces.Metrics {
	if m.baseCollector == nil {
		return &interfaces.Metrics{}
	}
	snapshot := m.baseCollector.Snapshot()
	
	// 安全计算错误率，避免NaN
	var errorRate float64
	if snapshot.Core.Operations.Total > 0 {
		errorRate = float64(snapshot.Core.Operations.Failed) / float64(snapshot.Core.Operations.Total) * 100
	}
	
	return &interfaces.Metrics{
		TotalOps:   snapshot.Core.Operations.Total,
		SuccessOps: snapshot.Core.Operations.Success,
		FailedOps:  snapshot.Core.Operations.Failed,
		ReadOps:    snapshot.Core.Operations.Read,
		WriteOps:   snapshot.Core.Operations.Write,
		AvgLatency: snapshot.Core.Latency.Average,
		MinLatency: snapshot.Core.Latency.Min,
		MaxLatency: snapshot.Core.Latency.Max,
		P90Latency: snapshot.Core.Latency.P90,
		P95Latency: snapshot.Core.Latency.P95,
		P99Latency: snapshot.Core.Latency.P99,
		ErrorRate:  errorRate,
		RPS:        int32(snapshot.Core.Throughput.RPS),
		StartTime:  time.Now().Add(-snapshot.Core.Duration),
		EndTime:    time.Now(),
		Duration:   snapshot.Core.Duration,
	}
}

func (m *SharedMetricsAdapter) Reset() {
	if m.baseCollector != nil {
		m.baseCollector.Reset()
	}
}

func (m *SharedMetricsAdapter) Export() map[string]interface{} {
	if m.baseCollector == nil {
		return make(map[string]interface{})
	}
	snapshot := m.baseCollector.Snapshot()
	return map[string]interface{}{
		"total_ops":     snapshot.Core.Operations.Total,
		"success_ops":   snapshot.Core.Operations.Success,
		"failed_ops":    snapshot.Core.Operations.Failed,
		"success_rate":  snapshot.Core.Operations.Rate,
		"rps":           snapshot.Core.Throughput.RPS,
		"avg_latency":   int64(snapshot.Core.Latency.Average),
		"p95_latency":   int64(snapshot.Core.Latency.P95),
		"p99_latency":   int64(snapshot.Core.Latency.P99),
		"protocol_data": snapshot.Protocol,
	}
}