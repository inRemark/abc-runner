package reporting

import (
	"fmt"
	"time"

	"abc-runner/app/core/metrics"
)

// StructuredReport 结构化报告模型
type StructuredReport struct {
	// ExecutiveDashboard 高管仪表板
	Dashboard ExecutiveDashboard `json:"dashboard"`
	
	// MetricsBreakdown 指标分解
	Metrics MetricsBreakdown `json:"metrics"`
	
	// SystemHealth 系统健康状态
	System SystemHealth `json:"system"`
	
	// ContextMetadata 上下文元数据
	Context ContextMetadata `json:"context"`
}

// ExecutiveDashboard 高管仪表板
type ExecutiveDashboard struct {
	// PerformanceScore 性能评分 (0-100)
	PerformanceScore int `json:"performance_score"`
	
	// StatusIndicator 状态指示器
	StatusIndicator StatusLevel `json:"status_indicator"`
	
	// KeyInsights 关键洞察
	KeyInsights []Insight `json:"key_insights"`
	
	// Recommendations 可执行建议
	Recommendations []Recommendation `json:"recommendations"`
}

// StatusLevel 状态等级
type StatusLevel string

const (
	StatusGood     StatusLevel = "good"
	StatusWarning  StatusLevel = "warning"  
	StatusCritical StatusLevel = "critical"
)

// Insight 洞察
type Insight struct {
	Type        InsightType `json:"type"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Impact      ImpactLevel `json:"impact"`
}

// InsightType 洞察类型
type InsightType string

const (
	InsightPerformance   InsightType = "performance"
	InsightReliability   InsightType = "reliability"
	InsightEfficiency    InsightType = "efficiency"
	InsightScalability   InsightType = "scalability"
)

// ImpactLevel 影响等级
type ImpactLevel string

const (
	ImpactHigh   ImpactLevel = "high"
	ImpactMedium ImpactLevel = "medium"
	ImpactLow    ImpactLevel = "low"
)

// Recommendation 建议
type Recommendation struct {
	Priority    Priority `json:"priority"`
	Category    string   `json:"category"`
	Action      string   `json:"action"`
	Description string   `json:"description"`
	ExpectedBenefit string `json:"expected_benefit"`
}

// Priority 优先级
type Priority string

const (
	PriorityHigh   Priority = "high"
	PriorityMedium Priority = "medium"
	PriorityLow    Priority = "low"
)

// MetricsBreakdown 指标分解
type MetricsBreakdown struct {
	// CoreOperations 核心操作指标
	CoreOperations OperationAnalysis `json:"core_operations"`
	
	// LatencyAnalysis 延迟分析
	LatencyAnalysis LatencyBreakdown `json:"latency_analysis"`
	
	// ProtocolSpecific 协议特定指标
	ProtocolSpecific interface{} `json:"protocol_specific"`
}

// OperationAnalysis 操作分析
type OperationAnalysis struct {
	TotalOperations    int64   `json:"total_operations"`
	SuccessfulOps      int64   `json:"successful_operations"`
	FailedOps          int64   `json:"failed_operations"`
	SuccessRate        float64 `json:"success_rate"`
	ErrorRate          float64 `json:"error_rate"`
	OperationsPerSecond float64 `json:"operations_per_second"`
	
	// 操作分布
	OperationTypes map[string]int64 `json:"operation_types"`
}

// LatencyBreakdown 延迟分析
type LatencyBreakdown struct {
	AverageLatency time.Duration `json:"average_latency"`
	MinLatency     time.Duration `json:"min_latency"`
	MaxLatency     time.Duration `json:"max_latency"`
	
	// 百分位延迟
	Percentiles LatencyPercentiles `json:"percentiles"`
	
	// 延迟分布
	Distribution LatencyDistribution `json:"distribution"`
}

// LatencyPercentiles 延迟百分位
type LatencyPercentiles struct {
	P50 time.Duration `json:"p50"`
	P90 time.Duration `json:"p90"`
	P95 time.Duration `json:"p95"`
	P99 time.Duration `json:"p99"`
	P999 time.Duration `json:"p999"`
}

// LatencyDistribution 延迟分布
type LatencyDistribution struct {
	Under1ms    int64 `json:"under_1ms"`
	Under5ms    int64 `json:"under_5ms"`
	Under10ms   int64 `json:"under_10ms"`
	Under50ms   int64 `json:"under_50ms"`
	Under100ms  int64 `json:"under_100ms"`
	Under500ms  int64 `json:"under_500ms"`
	Under1s     int64 `json:"under_1s"`
	Above1s     int64 `json:"above_1s"`
}

// SystemHealth 系统健康状态
type SystemHealth struct {
	// MemoryProfile 内存分析
	MemoryProfile MemoryMetrics `json:"memory_profile"`
	
	// RuntimeMetrics 运行时指标
	RuntimeMetrics RuntimeHealth `json:"runtime_metrics"`
	
	// ResourceHealth 资源健康状态
	ResourceHealth ResourceMetrics `json:"resource_health"`
}

// MemoryMetrics 内存指标
type MemoryMetrics struct {
	AllocatedMemory   int64   `json:"allocated_memory"`
	TotalAllocations  int64   `json:"total_allocations"`
	GCCount           uint32  `json:"gc_count"`
	GCPauseTotal      int64   `json:"gc_pause_total"`
	MemoryUsagePercent float64 `json:"memory_usage_percent"`
}

// RuntimeHealth 运行时健康状态
type RuntimeHealth struct {
	ActiveGoroutines  int   `json:"active_goroutines"`
	CPUUsagePercent   float64 `json:"cpu_usage_percent"`
	TestDuration      time.Duration `json:"test_duration"`
	StartTime         time.Time `json:"start_time"`
	EndTime           time.Time `json:"end_time"`
}

// ResourceMetrics 资源指标
type ResourceMetrics struct {
	MaxMemoryUsed     int64 `json:"max_memory_used"`
	MaxGoroutines     int   `json:"max_goroutines"`
	ConnectionsActive int   `json:"connections_active"`
	ConnectionsTotal  int   `json:"connections_total"`
}

// ContextMetadata 上下文元数据
type ContextMetadata struct {
	// TestConfiguration 测试配置
	TestConfiguration TestConfig `json:"test_configuration"`
	
	// Environment 环境信息
	Environment EnvInfo `json:"environment"`
	
	// ExecutionContext 执行上下文
	ExecutionContext ExecContext `json:"execution_context"`
}

// TestConfig 测试配置
type TestConfig struct {
	Protocol          string            `json:"protocol"`
	TotalOperations   int64             `json:"total_operations"`
	ConcurrentClients int               `json:"concurrent_clients"`
	TestDuration      time.Duration     `json:"test_duration"`
	Parameters        map[string]interface{} `json:"parameters"`
}

// EnvInfo 环境信息
type EnvInfo struct {
	OSName          string `json:"os_name"`
	Architecture    string `json:"architecture"`
	GoVersion       string `json:"go_version"`
	ABCRunnerVersion string `json:"abc_runner_version"`
	Hostname        string `json:"hostname"`
}

// ExecContext 执行上下文
type ExecContext struct {
	GeneratedAt     time.Time `json:"generated_at"`
	GeneratedBy     string    `json:"generated_by"`
	ReportVersion   string    `json:"report_version"`
	UniqueSessionID string    `json:"unique_session_id"`
}

// ConvertFromMetricsSnapshot 从指标快照转换为结构化报告
func ConvertFromMetricsSnapshot(snapshot *metrics.MetricsSnapshot[map[string]interface{}]) *StructuredReport {
	report := &StructuredReport{
		Dashboard: generateDashboard(snapshot),
		Metrics:   generateMetricsBreakdown(snapshot),
		System:    generateSystemHealth(snapshot),
		Context:   generateContextMetadata(snapshot),
	}
	
	return report
}

// generateDashboard 生成仪表板
func generateDashboard(snapshot *metrics.MetricsSnapshot[map[string]interface{}]) ExecutiveDashboard {
	score := calculatePerformanceScore(snapshot)
	status := determineStatusLevel(snapshot)
	insights := generateInsights(snapshot)
	recommendations := generateRecommendations(snapshot)
	
	return ExecutiveDashboard{
		PerformanceScore: score,
		StatusIndicator:  status,
		KeyInsights:      insights,
		Recommendations:  recommendations,
	}
}

// generateMetricsBreakdown 生成指标分解
func generateMetricsBreakdown(snapshot *metrics.MetricsSnapshot[map[string]interface{}]) MetricsBreakdown {
	return MetricsBreakdown{
		CoreOperations: OperationAnalysis{
			TotalOperations:     snapshot.Core.Operations.Total,
			SuccessfulOps:       snapshot.Core.Operations.Success,
			FailedOps:           snapshot.Core.Operations.Failed,
			SuccessRate:         snapshot.Core.Operations.Rate,
			ErrorRate:           float64(snapshot.Core.Operations.Failed) / float64(snapshot.Core.Operations.Total) * 100,
			OperationsPerSecond: snapshot.Core.Throughput.RPS,
			OperationTypes: map[string]int64{
				"read":  snapshot.Core.Operations.Read,
				"write": snapshot.Core.Operations.Write,
			},
		},
		LatencyAnalysis: LatencyBreakdown{
			AverageLatency: snapshot.Core.Latency.Average,
			MinLatency:     snapshot.Core.Latency.Min,
			MaxLatency:     snapshot.Core.Latency.Max,
			Percentiles: LatencyPercentiles{
				P50: snapshot.Core.Latency.P50,
				P90: snapshot.Core.Latency.P90,
				P95: snapshot.Core.Latency.P95,
				P99: snapshot.Core.Latency.P99,
			},
		},
		ProtocolSpecific: snapshot.Protocol,
	}
}

// generateSystemHealth 生成系统健康状态
func generateSystemHealth(snapshot *metrics.MetricsSnapshot[map[string]interface{}]) SystemHealth {
	return SystemHealth{
		MemoryProfile: MemoryMetrics{
			AllocatedMemory:    int64(snapshot.System.Memory.Allocated),
			TotalAllocations:   int64(snapshot.System.Memory.TotalAlloc),
			GCCount:            snapshot.System.GC.NumGC,
			GCPauseTotal:       int64(snapshot.System.GC.PauseTotal),
			MemoryUsagePercent: snapshot.System.Memory.Usage,
		},
		RuntimeMetrics: RuntimeHealth{
			ActiveGoroutines: snapshot.System.Goroutine.Active,
			TestDuration:     snapshot.Core.Duration,
			StartTime:        snapshot.Timestamp.Add(-snapshot.Core.Duration),
			EndTime:          snapshot.Timestamp,
		},
		ResourceHealth: ResourceMetrics{
			MaxMemoryUsed: int64(snapshot.System.Memory.Allocated),
			MaxGoroutines: snapshot.System.Goroutine.Active,
		},
	}
}

// generateContextMetadata 生成上下文元数据
func generateContextMetadata(snapshot *metrics.MetricsSnapshot[map[string]interface{}]) ContextMetadata {
	return ContextMetadata{
		TestConfiguration: TestConfig{
			Protocol:        getProtocolFromSnapshot(snapshot),
			TotalOperations: snapshot.Core.Operations.Total,
			TestDuration:    snapshot.Core.Duration,
			Parameters:      snapshot.Protocol,
		},
		Environment: EnvInfo{
			ABCRunnerVersion: "3.0.0",
		},
		ExecutionContext: ExecContext{
			GeneratedAt:     time.Now(),
			GeneratedBy:     "abc-runner-v3",
			ReportVersion:   "3.0.0",
			UniqueSessionID: generateSessionID(),
		},
	}
}

// Helper functions
func calculatePerformanceScore(snapshot *metrics.MetricsSnapshot[map[string]interface{}]) int {
	// 基于成功率、延迟和吞吐量计算性能评分
	successRate := snapshot.Core.Operations.Rate
	avgLatency := snapshot.Core.Latency.Average.Milliseconds()
	rps := snapshot.Core.Throughput.RPS
	
	// 简化的评分算法
	score := int(successRate * 0.4)
	
	// 延迟惩罚
	if avgLatency < 10 {
		score += 30
	} else if avgLatency < 50 {
		score += 20
	} else if avgLatency < 100 {
		score += 10
	}
	
	// 吞吐量奖励
	if rps > 1000 {
		score += 30
	} else if rps > 500 {
		score += 20
	} else if rps > 100 {
		score += 10
	}
	
	if score > 100 {
		score = 100
	}
	
	return score
}

func determineStatusLevel(snapshot *metrics.MetricsSnapshot[map[string]interface{}]) StatusLevel {
	errorRate := float64(snapshot.Core.Operations.Failed) / float64(snapshot.Core.Operations.Total) * 100
	
	if errorRate > 10 || snapshot.Core.Latency.Average.Milliseconds() > 1000 {
		return StatusCritical
	} else if errorRate > 5 || snapshot.Core.Latency.Average.Milliseconds() > 500 {
		return StatusWarning
	}
	
	return StatusGood
}

func generateInsights(snapshot *metrics.MetricsSnapshot[map[string]interface{}]) []Insight {
	var insights []Insight
	
	// 性能洞察
	if snapshot.Core.Throughput.RPS > 1000 {
		insights = append(insights, Insight{
			Type:        InsightPerformance,
			Title:       "高吞吐量性能",
			Description: "系统展现出优秀的吞吐量表现",
			Impact:      ImpactHigh,
		})
	}
	
	// 可靠性洞察
	if snapshot.Core.Operations.Rate > 99.5 {
		insights = append(insights, Insight{
			Type:        InsightReliability,
			Title:       "出色的可靠性",
			Description: "系统可靠性指标优秀，成功率超过99.5%",
			Impact:      ImpactHigh,
		})
	}
	
	return insights
}

func generateRecommendations(snapshot *metrics.MetricsSnapshot[map[string]interface{}]) []Recommendation {
	var recommendations []Recommendation
	
	errorRate := float64(snapshot.Core.Operations.Failed) / float64(snapshot.Core.Operations.Total) * 100
	
	if errorRate > 5 {
		recommendations = append(recommendations, Recommendation{
			Priority:        PriorityHigh,
			Category:        "可靠性",
			Action:          "调查并修复错误源",
			Description:     "错误率过高，需要调查根本原因",
			ExpectedBenefit: "提高系统可靠性和用户体验",
		})
	}
	
	if snapshot.Core.Latency.Average.Milliseconds() > 100 {
		recommendations = append(recommendations, Recommendation{
			Priority:        PriorityMedium,
			Category:        "性能",
			Action:          "优化延迟性能",
			Description:     "平均延迟较高，考虑优化处理逻辑",
			ExpectedBenefit: "改善响应时间和用户体验",
		})
	}
	
	return recommendations
}

func getProtocolFromSnapshot(snapshot *metrics.MetricsSnapshot[map[string]interface{}]) string {
	if protocolData, ok := snapshot.Protocol["protocol"]; ok {
		if protocol, ok := protocolData.(string); ok {
			return protocol
		}
	}
	return "unknown"
}

func generateSessionID() string {
	return fmt.Sprintf("session_%d", time.Now().Unix())
}