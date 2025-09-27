package metrics

import (
	"time"

	"abc-runner/app/core/interfaces"
)

// 使用来自 interfaces 包中的 SystemMetrics 和 MetricsCollector
type SystemMetrics = interfaces.SystemMetrics
type MetricsCollector[T any] = interfaces.MetricsCollector[T]
type MetricsSnapshot[T any] = interfaces.MetricsSnapshot[T]
type CoreMetrics = interfaces.CoreMetrics
type OperationMetrics = interfaces.OperationMetrics
type LatencyMetrics = interfaces.LatencyMetrics
type ThroughputMetrics = interfaces.ThroughputMetrics
type DefaultMetricsCollector = interfaces.DefaultMetricsCollector
type DefaultMetricsSnapshot = interfaces.DefaultMetricsSnapshot



// MetricsConfig 指标配置
type MetricsConfig struct {
	// Latency 延迟相关配置
	Latency LatencyConfig `json:"latency"`

	// Throughput 吞吐量相关配置
	Throughput ThroughputConfig `json:"throughput"`

	// System 系统监控配置
	System SystemConfig `json:"system"`

	// Storage 存储配置
	Storage StorageConfig `json:"storage"`

	// Export 导出配置
	Export ExportConfig `json:"export"`
}

// LatencyConfig 延迟配置
type LatencyConfig struct {
	// HistorySize 延迟历史记录大小(环形缓冲区大小)
	HistorySize int `json:"history_size" default:"10000"`

	// Percentiles 需要计算的分位数
	Percentiles []float64 `json:"percentiles" default:"[0.5,0.9,0.95,0.99]"`

	// SamplingRate 采样率(0.0-1.0)
	SamplingRate float64 `json:"sampling_rate" default:"1.0"`

	// ComputeInterval 计算间隔
	ComputeInterval time.Duration `json:"compute_interval" default:"1s"`
}

// ThroughputConfig 吞吐量配置
type ThroughputConfig struct {
	// WindowSize 时间窗口大小
	WindowSize time.Duration `json:"window_size" default:"60s"`

	// UpdateInterval 更新间隔
	UpdateInterval time.Duration `json:"update_interval" default:"1s"`
}

// SystemConfig 系统监控配置
type SystemConfig struct {
	// MonitorInterval 监控间隔
	MonitorInterval time.Duration `json:"monitor_interval" default:"1s"`

	// HealthThresholds 健康阈值
	HealthThresholds HealthThresholds `json:"health_thresholds"`

	// SnapshotRetention 快照保留数量
	SnapshotRetention int `json:"snapshot_retention" default:"100"`

	// Enabled 是否启用系统监控
	Enabled bool `json:"enabled" default:"true"`
}

// HealthThresholds 健康阈值
type HealthThresholds struct {
	// MemoryUsage 内存使用率阈值(%)
	MemoryUsage float64 `json:"memory_usage" default:"80.0"`

	// GCFrequency GC频率阈值(次数)
	GCFrequency uint32 `json:"gc_frequency" default:"100"`

	// GoroutineCount 协程数量阈值
	GoroutineCount int `json:"goroutine_count" default:"1000"`

	// CPUUsage CPU使用率阈值(%)
	CPUUsage float64 `json:"cpu_usage" default:"80.0"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	// MemoryLimit 内存限制(bytes)
	MemoryLimit int64 `json:"memory_limit" default:"104857600"` // 100MB

	// UseCompression 是否使用压缩
	UseCompression bool `json:"use_compression" default:"false"`

	// FlushInterval 刷新间隔
	FlushInterval time.Duration `json:"flush_interval" default:"5s"`
}

// ExportConfig 导出配置
type ExportConfig struct {
	// Format 导出格式
	Format []string `json:"format" default:"[\"json\"]"`

	// Interval 导出间隔
	Interval time.Duration `json:"interval" default:"10s"`

	// Enabled 是否启用自动导出
	Enabled bool `json:"enabled" default:"false"`
}

// MetricsCollectorFactory 指标收集器工厂接口
type MetricsCollectorFactory[T any] interface {
	// CreateCollector 创建指标收集器
	CreateCollector(config *MetricsConfig) (MetricsCollector[T], error)

	// GetProtocolType 获取协议类型
	GetProtocolType() string
}

// MetricsAggregator 指标聚合器接口
type MetricsAggregator interface {
	// Aggregate 聚合多个指标快照
	Aggregate(snapshots ...interface{}) (*AggregatedMetrics, error)

	// AddSnapshot 添加快照
	AddSnapshot(snapshot interface{}) error

	// GetAggregated 获取聚合结果
	GetAggregated() *AggregatedMetrics
}

// AggregatedMetrics 聚合指标
type AggregatedMetrics struct {
	// Core 聚合的核心指标
	Core CoreMetrics `json:"core"`

	// Protocols 各协议指标
	Protocols map[string]interface{} `json:"protocols"`

	// System 聚合的系统指标
	System SystemMetrics `json:"system"`

	// Summary 汇总信息
	Summary AggregationSummary `json:"summary"`
}

// AggregationSummary 聚合汇总
type AggregationSummary struct {
	// TotalSnapshots 快照总数
	TotalSnapshots int `json:"total_snapshots"`

	// Protocols 包含的协议
	Protocols []string `json:"protocols"`

	// TimeRange 时间范围
	TimeRange TimeRange `json:"time_range"`

	// AggregatedAt 聚合时间
	AggregatedAt time.Time `json:"aggregated_at"`
}

// TimeRange 时间范围
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// 注意：HealthStatus、HealthCheckResult、HealthChecker 现在定义在 advanced_health_checker.go 中