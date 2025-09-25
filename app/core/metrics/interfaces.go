package metrics

import (
	"context"
	"time"

	"abc-runner/app/core/interfaces"
)

// MetricsCollector 新的泛型指标收集器接口
type MetricsCollector[T any] interface {
	// Record 记录操作结果
	Record(result *interfaces.OperationResult)

	// Snapshot 获取当前指标快照
	Snapshot() *MetricsSnapshot[T]

	// Reset 重置所有指标
	Reset()
}

// MetricsSnapshot 指标快照结构
type MetricsSnapshot[T any] struct {
	// Core 通用核心指标
	Core CoreMetrics `json:"core"`

	// Protocol 协议特定指标
	Protocol T `json:"protocol"`

	// System 系统监控指标
	System SystemMetrics `json:"system"`

	// Timestamp 快照时间戳
	Timestamp time.Time `json:"timestamp"`
}

// CoreMetrics 核心通用指标
type CoreMetrics struct {
	// Operations 操作指标
	Operations OperationMetrics `json:"operations"`

	// Latency 延迟指标
	Latency LatencyMetrics `json:"latency"`

	// Throughput 吞吐量指标
	Throughput ThroughputMetrics `json:"throughput"`

	// Duration 测试持续时间
	Duration time.Duration `json:"duration"`
}

// OperationMetrics 操作指标
type OperationMetrics struct {
	Total   int64   `json:"total"`         // 总操作数
	Success int64   `json:"success"`       // 成功操作数
	Failed  int64   `json:"failed"`        // 失败操作数
	Read    int64   `json:"read"`          // 读操作数
	Write   int64   `json:"write"`         // 写操作数
	Rate    float64 `json:"success_rate"`  // 成功率 (%)
}

// LatencyMetrics 延迟指标
type LatencyMetrics struct {
	Min         time.Duration `json:"min"`          // 最小延迟
	Max         time.Duration `json:"max"`          // 最大延迟
	Average     time.Duration `json:"average"`      // 平均延迟
	P50         time.Duration `json:"p50"`          // P50分位数
	P90         time.Duration `json:"p90"`          // P90分位数
	P95         time.Duration `json:"p95"`          // P95分位数
	P99         time.Duration `json:"p99"`          // P99分位数
	StdDeviation time.Duration `json:"std_dev"`     // 标准差
}

// ThroughputMetrics 吞吐量指标
type ThroughputMetrics struct {
	RPS     float64 `json:"rps"`      // 每秒请求数
	ReadRPS float64 `json:"read_rps"` // 每秒读请求数
	WriteRPS float64 `json:"write_rps"` // 每秒写请求数
}

// SystemMetrics 系统监控指标
type SystemMetrics struct {
	Memory    MemoryMetrics    `json:"memory"`    // 内存指标
	GC        GCMetrics        `json:"gc"`        // GC指标
	Goroutine GoroutineMetrics `json:"goroutine"` // 协程指标
	CPU       CPUMetrics       `json:"cpu"`       // CPU指标
}

// MemoryMetrics 内存指标
type MemoryMetrics struct {
	Allocated   uint64  `json:"allocated"`    // 已分配内存(bytes)
	TotalAlloc  uint64  `json:"total_alloc"`  // 累计分配内存(bytes)
	Sys         uint64  `json:"sys"`          // 系统内存(bytes)
	NumGC       uint32  `json:"num_gc"`       // GC次数
	Usage       float64 `json:"usage"`        // 内存使用率(%)
}

// GCMetrics GC指标
type GCMetrics struct {
	NumGC        uint32        `json:"num_gc"`         // GC次数
	PauseTotal   time.Duration `json:"pause_total"`    // 总暂停时间
	PauseAvg     time.Duration `json:"pause_avg"`      // 平均暂停时间
	LastPause    time.Duration `json:"last_pause"`     // 最后暂停时间
	ForcedGC     uint32        `json:"forced_gc"`      // 强制GC次数
}

// GoroutineMetrics 协程指标
type GoroutineMetrics struct {
	Active int `json:"active"` // 活跃协程数
	Peak   int `json:"peak"`   // 峰值协程数
}

// CPUMetrics CPU指标
type CPUMetrics struct {
	Usage   float64 `json:"usage"`   // CPU使用率(%)
	Cores   int     `json:"cores"`   // CPU核心数
}

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

// HealthStatus 健康状态
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusWarning   HealthStatus = "warning"
	HealthStatusCritical  HealthStatus = "critical"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// HealthCheckResult 健康检查结果
type HealthCheckResult struct {
	Status     HealthStatus          `json:"status"`
	Message    string                `json:"message"`
	Metrics    SystemMetrics         `json:"metrics"`
	Violations []ThresholdViolation  `json:"violations"`
	CheckedAt  time.Time             `json:"checked_at"`
}

// ThresholdViolation 阈值违反
type ThresholdViolation struct {
	Metric    string  `json:"metric"`
	Current   float64 `json:"current"`
	Threshold float64 `json:"threshold"`
	Severity  string  `json:"severity"`
}

// HealthChecker 健康检查器接口
type HealthChecker interface {
	// Check 执行健康检查
	Check(ctx context.Context, metrics SystemMetrics) *HealthCheckResult

	// RegisterThreshold 注册阈值
	RegisterThreshold(metric string, threshold float64, severity string)

	// GetThresholds 获取所有阈值
	GetThresholds() map[string]float64
}