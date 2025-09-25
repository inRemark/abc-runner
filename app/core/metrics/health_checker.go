package metrics

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// healthChecker 健康检查器实现
type healthChecker struct {
	thresholds map[string]ThresholdConfig
	mutex      sync.RWMutex
}

// ThresholdConfig 阈值配置
type ThresholdConfig struct {
	Value    float64 `json:"value"`
	Severity string  `json:"severity"` // "warning", "critical"
}

// NewHealthChecker 创建健康检查器
func NewHealthChecker(thresholds HealthThresholds) HealthChecker {
	checker := &healthChecker{
		thresholds: make(map[string]ThresholdConfig),
	}

	// 设置默认阈值
	checker.thresholds["memory_usage"] = ThresholdConfig{
		Value:    thresholds.MemoryUsage,
		Severity: "warning",
	}
	checker.thresholds["gc_frequency"] = ThresholdConfig{
		Value:    float64(thresholds.GCFrequency),
		Severity: "warning",
	}
	checker.thresholds["goroutine_count"] = ThresholdConfig{
		Value:    float64(thresholds.GoroutineCount),
		Severity: "warning",
	}
	checker.thresholds["cpu_usage"] = ThresholdConfig{
		Value:    thresholds.CPUUsage,
		Severity: "warning",
	}

	// 设置关键阈值（更高的值触发critical状态）
	checker.thresholds["memory_usage_critical"] = ThresholdConfig{
		Value:    thresholds.MemoryUsage * 1.2, // 高出20%触发critical
		Severity: "critical",
	}
	checker.thresholds["goroutine_count_critical"] = ThresholdConfig{
		Value:    float64(thresholds.GoroutineCount) * 1.5, // 高出50%触发critical
		Severity: "critical",
	}

	return checker
}

// Check 执行健康检查
func (hc *healthChecker) Check(ctx context.Context, metrics SystemMetrics) *HealthCheckResult {
	result := &HealthCheckResult{
		Status:     HealthStatusHealthy,
		Message:    "System is healthy",
		Metrics:    metrics,
		Violations: make([]ThresholdViolation, 0),
		CheckedAt:  time.Now(),
	}

	hc.mutex.RLock()
	defer hc.mutex.RUnlock()

	// 检查内存使用率
	if violation := hc.checkThreshold("memory_usage", metrics.Memory.Usage); violation != nil {
		result.Violations = append(result.Violations, *violation)
	}
	if violation := hc.checkThreshold("memory_usage_critical", metrics.Memory.Usage); violation != nil {
		result.Violations = append(result.Violations, *violation)
	}

	// 检查GC频率
	if violation := hc.checkThreshold("gc_frequency", float64(metrics.GC.NumGC)); violation != nil {
		result.Violations = append(result.Violations, *violation)
	}

	// 检查协程数量
	if violation := hc.checkThreshold("goroutine_count", float64(metrics.Goroutine.Active)); violation != nil {
		result.Violations = append(result.Violations, *violation)
	}
	if violation := hc.checkThreshold("goroutine_count_critical", float64(metrics.Goroutine.Active)); violation != nil {
		result.Violations = append(result.Violations, *violation)
	}

	// 检查CPU使用率
	if violation := hc.checkThreshold("cpu_usage", metrics.CPU.Usage); violation != nil {
		result.Violations = append(result.Violations, *violation)
	}

	// 根据违规情况确定整体健康状态
	result.Status, result.Message = hc.determineOverallStatus(result.Violations)

	return result
}

// RegisterThreshold 注册阈值
func (hc *healthChecker) RegisterThreshold(metric string, threshold float64, severity string) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()

	hc.thresholds[metric] = ThresholdConfig{
		Value:    threshold,
		Severity: severity,
	}
}

// GetThresholds 获取所有阈值
func (hc *healthChecker) GetThresholds() map[string]float64 {
	hc.mutex.RLock()
	defer hc.mutex.RUnlock()

	result := make(map[string]float64)
	for metric, config := range hc.thresholds {
		result[metric] = config.Value
	}
	return result
}

// checkThreshold 检查单个阈值
func (hc *healthChecker) checkThreshold(metric string, current float64) *ThresholdViolation {
	config, exists := hc.thresholds[metric]
	if !exists {
		return nil
	}

	if current > config.Value {
		return &ThresholdViolation{
			Metric:    metric,
			Current:   current,
			Threshold: config.Value,
			Severity:  config.Severity,
		}
	}

	return nil
}

// determineOverallStatus 确定整体健康状态
func (hc *healthChecker) determineOverallStatus(violations []ThresholdViolation) (HealthStatus, string) {
	if len(violations) == 0 {
		return HealthStatusHealthy, "All metrics are within normal ranges"
	}

	var criticalCount, warningCount int
	var messages []string

	for _, violation := range violations {
		switch violation.Severity {
		case "critical":
			criticalCount++
			messages = append(messages, fmt.Sprintf("CRITICAL: %s (%.2f > %.2f)", 
				violation.Metric, violation.Current, violation.Threshold))
		case "warning":
			warningCount++
			messages = append(messages, fmt.Sprintf("WARNING: %s (%.2f > %.2f)", 
				violation.Metric, violation.Current, violation.Threshold))
		}
	}

	if criticalCount > 0 {
		return HealthStatusCritical, fmt.Sprintf("System has %d critical issues: %v", 
			criticalCount, messages)
	}

	if warningCount > 0 {
		return HealthStatusWarning, fmt.Sprintf("System has %d warnings: %v", 
			warningCount, messages)
	}

	return HealthStatusUnknown, "Unknown health status"
}

// MetricsAggregatorImpl 指标聚合器实现
type MetricsAggregatorImpl struct {
	snapshots []interface{}
	mutex     sync.RWMutex
}

// NewMetricsAggregator 创建指标聚合器
func NewMetricsAggregator() MetricsAggregator {
	return &MetricsAggregatorImpl{
		snapshots: make([]interface{}, 0),
	}
}

// Aggregate 聚合多个指标快照
func (ma *MetricsAggregatorImpl) Aggregate(snapshots ...interface{}) (*AggregatedMetrics, error) {
	ma.mutex.Lock()
	defer ma.mutex.Unlock()

	if len(snapshots) == 0 {
		return nil, fmt.Errorf("no snapshots provided for aggregation")
	}

	// 初始化聚合结果
	aggregated := &AggregatedMetrics{
		Protocols: make(map[string]interface{}),
		Summary: AggregationSummary{
			TotalSnapshots: len(snapshots),
			Protocols:      make([]string, 0),
			AggregatedAt:   time.Now(),
		},
	}

	var (
		totalOps     int64
		totalSuccess int64
		totalFailed  int64
		totalRead    int64
		totalWrite   int64
		latencies    []time.Duration
		allDurations []time.Duration
		startTime    time.Time
		endTime      time.Time
		protocolSet  = make(map[string]bool)
	)

	// 聚合各个快照
	for i, snapshot := range snapshots {
		switch s := snapshot.(type) {
		case *MetricsSnapshot[interface{}]:
			ma.aggregateGenericSnapshot(s, &totalOps, &totalSuccess, &totalFailed, 
				&totalRead, &totalWrite, &latencies, &allDurations, aggregated, protocolSet)
			
			if i == 0 || s.Timestamp.Before(startTime) {
				startTime = s.Timestamp
			}
			if i == 0 || s.Timestamp.After(endTime) {
				endTime = s.Timestamp
			}

		default:
			// 尝试使用反射处理其他类型的快照
			if err := ma.aggregateUnknownSnapshot(snapshot, aggregated, protocolSet); err != nil {
				return nil, fmt.Errorf("failed to aggregate snapshot %d: %w", i, err)
			}
		}
	}

	// 计算聚合的核心指标
	ma.calculateAggregatedCore(aggregated, totalOps, totalSuccess, totalFailed, 
		totalRead, totalWrite, latencies, allDurations)

	// 设置汇总信息
	for protocol := range protocolSet {
		aggregated.Summary.Protocols = append(aggregated.Summary.Protocols, protocol)
	}
	aggregated.Summary.TimeRange = TimeRange{
		Start: startTime,
		End:   endTime,
	}

	return aggregated, nil
}

// AddSnapshot 添加快照
func (ma *MetricsAggregatorImpl) AddSnapshot(snapshot interface{}) error {
	ma.mutex.Lock()
	defer ma.mutex.Unlock()

	ma.snapshots = append(ma.snapshots, snapshot)
	return nil
}

// GetAggregated 获取聚合结果
func (ma *MetricsAggregatorImpl) GetAggregated() *AggregatedMetrics {
	ma.mutex.RLock()
	defer ma.mutex.RUnlock()

	if len(ma.snapshots) == 0 {
		return &AggregatedMetrics{
			Protocols: make(map[string]interface{}),
			Summary: AggregationSummary{
				TotalSnapshots: 0,
				Protocols:      []string{},
				AggregatedAt:   time.Now(),
			},
		}
	}

	// 聚合所有已存储的快照
	result, _ := ma.Aggregate(ma.snapshots...)
	return result
}

// aggregateGenericSnapshot 聚合泛型快照
func (ma *MetricsAggregatorImpl) aggregateGenericSnapshot(
	snapshot *MetricsSnapshot[interface{}],
	totalOps, totalSuccess, totalFailed, totalRead, totalWrite *int64,
	latencies, allDurations *[]time.Duration,
	aggregated *AggregatedMetrics,
	protocolSet map[string]bool) {

	core := snapshot.Core
	*totalOps += core.Operations.Total
	*totalSuccess += core.Operations.Success
	*totalFailed += core.Operations.Failed
	*totalRead += core.Operations.Read
	*totalWrite += core.Operations.Write

	*allDurations = append(*allDurations, core.Duration)

	// 这里需要从协议数据中提取延迟信息
	// 由于泛型限制，这里使用简化处理
	*latencies = append(*latencies, core.Latency.Average)

	// 设置协议标识
	protocolType := "unknown"
	if snapshot.Protocol != nil {
		protocolType = fmt.Sprintf("%T", snapshot.Protocol)
	}
	protocolSet[protocolType] = true
	aggregated.Protocols[protocolType] = snapshot.Protocol
}

// aggregateUnknownSnapshot 聚合未知类型快照
func (ma *MetricsAggregatorImpl) aggregateUnknownSnapshot(
	snapshot interface{},
	aggregated *AggregatedMetrics,
	protocolSet map[string]bool) error {

	// 这里可以使用反射来处理未知类型的快照
	// 为了简化，我们直接将其存储在协议映射中
	snapshotType := fmt.Sprintf("%T", snapshot)
	protocolSet[snapshotType] = true
	aggregated.Protocols[snapshotType] = snapshot

	return nil
}

// calculateAggregatedCore 计算聚合的核心指标
func (ma *MetricsAggregatorImpl) calculateAggregatedCore(
	aggregated *AggregatedMetrics,
	totalOps, totalSuccess, totalFailed, totalRead, totalWrite int64,
	latencies, allDurations []time.Duration) {

	// 计算操作指标
	var successRate float64
	if totalOps > 0 {
		successRate = float64(totalSuccess) / float64(totalOps) * 100.0
	}

	aggregated.Core.Operations = OperationMetrics{
		Total:   totalOps,
		Success: totalSuccess,
		Failed:  totalFailed,
		Read:    totalRead,
		Write:   totalWrite,
		Rate:    successRate,
	}

	// 计算延迟指标
	if len(latencies) > 0 {
		aggregated.Core.Latency = ma.calculateAggregatedLatency(latencies)
	}

	// 计算吞吐量指标
	if len(allDurations) > 0 {
		totalDuration := ma.sumDurations(allDurations)
		if totalDuration > 0 {
			aggregated.Core.Throughput = ThroughputMetrics{
				RPS:      float64(totalOps) / totalDuration.Seconds(),
				ReadRPS:  float64(totalRead) / totalDuration.Seconds(),
				WriteRPS: float64(totalWrite) / totalDuration.Seconds(),
			}
		}
		aggregated.Core.Duration = totalDuration
	}
}

// calculateAggregatedLatency 计算聚合延迟指标
func (ma *MetricsAggregatorImpl) calculateAggregatedLatency(latencies []time.Duration) LatencyMetrics {
	if len(latencies) == 0 {
		return LatencyMetrics{}
	}

	// 排序延迟数据
	sortedLatencies := make([]time.Duration, len(latencies))
	copy(sortedLatencies, latencies)

	var total time.Duration
	min := sortedLatencies[0]
	max := sortedLatencies[0]

	for _, lat := range sortedLatencies {
		total += lat
		if lat < min {
			min = lat
		}
		if lat > max {
			max = lat
		}
	}

	avg := total / time.Duration(len(latencies))

	// 计算分位数
	percentiles := ma.calculatePercentiles(sortedLatencies)

	return LatencyMetrics{
		Min:     min,
		Max:     max,
		Average: avg,
		P50:     percentiles[50],
		P90:     percentiles[90],
		P95:     percentiles[95],
		P99:     percentiles[99],
	}
}

// calculatePercentiles 计算分位数
func (ma *MetricsAggregatorImpl) calculatePercentiles(sortedLatencies []time.Duration) map[int]time.Duration {
	percentiles := make(map[int]time.Duration)
	
	for _, p := range []int{50, 90, 95, 99} {
		index := int(float64(len(sortedLatencies)) * float64(p) / 100.0)
		if index >= len(sortedLatencies) {
			index = len(sortedLatencies) - 1
		}
		if index < 0 {
			index = 0
		}
		percentiles[p] = sortedLatencies[index]
	}
	
	return percentiles
}

// sumDurations 计算持续时间总和
func (ma *MetricsAggregatorImpl) sumDurations(durations []time.Duration) time.Duration {
	var total time.Duration
	for _, d := range durations {
		total += d
	}
	return total
}