package interfaces

import (
	"context"
	"time"
)

// ProtocolAdapter 协议适配器统一接口
type ProtocolAdapter interface {
	// Connect 初始化连接
	Connect(ctx context.Context, config Config) error

	// Execute 执行操作并返回结果
	Execute(ctx context.Context, operation Operation) (*OperationResult, error)

	// Close 关闭连接
	Close() error

	// GetProtocolMetrics 获取协议特定指标
	GetProtocolMetrics() map[string]interface{}

	// HealthCheck 健康检查
	HealthCheck(ctx context.Context) error

	// GetProtocolName 获取协议名称
	GetProtocolName() string

	// GetMetricsCollector 获取指标收集器
	GetMetricsCollector() DefaultMetricsCollector
}

// Operation 操作定义
type Operation struct {
	Type     string                 `json:"type"`     // 操作类型 (get, set, create, delete, etc.)
	Key      string                 `json:"key"`      // 操作的键
	Value    interface{}            `json:"value"`    // 操作的值
	Params   map[string]interface{} `json:"params"`   // 附加参数
	TTL      time.Duration          `json:"ttl"`      // 生存时间
	Metadata map[string]string      `json:"metadata"` // 元数据
}

// OperationResult 操作执行结果
type OperationResult struct {
	Success  bool                   `json:"success"`  // 是否成功
	Duration time.Duration          `json:"duration"` // 执行时间
	IsRead   bool                   `json:"is_read"`  // 是否为读操作
	Error    error                  `json:"error"`    // 错误信息
	Value    interface{}            `json:"value"`    // 返回值
	Metadata map[string]interface{} `json:"metadata"` // 结果元数据
}

// Config 统一配置接口
type Config interface {
	GetProtocol() string
	GetConnection() ConnectionConfig
	GetBenchmark() BenchmarkConfig
	Validate() error
	Clone() Config
}

// ConnectionConfig 连接配置接口
type ConnectionConfig interface {
	GetAddresses() []string
	GetCredentials() map[string]string
	GetPoolConfig() PoolConfig
	GetTimeout() time.Duration
}

// BenchmarkConfig 基准测试配置接口
type BenchmarkConfig interface {
	GetTotal() int
	GetParallels() int
	GetDataSize() int
	GetTTL() time.Duration
	GetReadPercent() int
	GetRandomKeys() int
	GetTestCase() string
}

// PoolConfig 连接池配置接口
type PoolConfig interface {
	GetPoolSize() int
	GetMinIdle() int
	GetMaxIdle() int
	GetIdleTimeout() time.Duration
	GetConnectionTimeout() time.Duration
}

// MetricsCollector 泛型指标收集器接口
type MetricsCollector[T any] interface {
	// Record 记录操作结果
	Record(result *OperationResult)

	// Snapshot 获取当前指标快照
	Snapshot() *MetricsSnapshot[T]

	// Reset 重置所有指标
	Reset()

	// Stop 停止收集器
	Stop()
}

// DefaultMetricsCollector 默认指标收集器类型（map[string]interface{}）
type DefaultMetricsCollector = MetricsCollector[map[string]interface{}]

// DefaultMetricsSnapshot 默认指标快照类型
type DefaultMetricsSnapshot = MetricsSnapshot[map[string]interface{}]

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
	Min          time.Duration `json:"min"`           // 最小延迟
	Max          time.Duration `json:"max"`           // 最大延迟
	Average      time.Duration `json:"average"`       // 平均延迟
	P50          time.Duration `json:"p50"`           // P50延迟
	P90          time.Duration `json:"p90"`           // P90延迟
	P95          time.Duration `json:"p95"`           // P95延迟
	P99          time.Duration `json:"p99"`           // P99延迟
	StdDeviation time.Duration `json:"std_deviation"` // 标准差
}

// ThroughputMetrics 吞吐量指标
type ThroughputMetrics struct {
	RPS      float64 `json:"rps"`       // 每秒请求数
	ReadRPS  float64 `json:"read_rps"`  // 每秒读请求数
	WriteRPS float64 `json:"write_rps"` // 每秒写请求数
}

// SystemMetrics 系统监控指标
type SystemMetrics struct {
	MemoryUsage    MemoryMetrics    `json:"memory"`     // 内存使用情况
	GCStats        GCMetrics        `json:"gc"`        // GC统计
	GoroutineCount int              `json:"goroutines"` // 协程数量
	CPUUsage       CPUMetrics       `json:"cpu"`       // CPU使用情况
}

// MemoryMetrics 内存指标
type MemoryMetrics struct {
	Allocated   uint64 `json:"allocated"`   // 已分配内存
	InUse       uint64 `json:"in_use"`      // 正在使用的内存
	TotalAlloc  uint64 `json:"total_alloc"` // 总分配内存
	Sys         uint64 `json:"sys"`         // 系统内存
	GCReleased  uint64 `json:"gc_released"` // GC释放的内存
}

// GCMetrics GC指标
type GCMetrics struct {
	LastGC       time.Time     `json:"last_gc"`       // 最后一次GC时间
	NumGC        uint32        `json:"num_gc"`        // GC次数
	TotalPause   time.Duration `json:"total_pause"`   // 总暂停时间
	AveragePause time.Duration `json:"average_pause"` // 平均暂停时间
}

// CPUMetrics CPU指标
type CPUMetrics struct {
	UsagePercent float64 `json:"usage_percent"` // CPU使用率百分比
	Cores        int     `json:"cores"`         // CPU核心数
}

// Metrics 兼容性结构（向前兼容）
type Metrics struct {
	RPS        int32         `json:"rps"`         // 每秒请求数
	TotalOps   int64         `json:"total_ops"`   // 总操作数
	SuccessOps int64         `json:"success_ops"` // 成功操作数
	FailedOps  int64         `json:"failed_ops"`  // 失败操作数
	ReadOps    int64         `json:"read_ops"`    // 读操作数
	WriteOps   int64         `json:"write_ops"`   // 写操作数
	AvgLatency time.Duration `json:"avg_latency"` // 平均延迟
	MinLatency time.Duration `json:"min_latency"` // 最小延迟
	MaxLatency time.Duration `json:"max_latency"` // 最大延迟
	P90Latency time.Duration `json:"p90_latency"` // P90延迟
	P95Latency time.Duration `json:"p95_latency"` // P95延迟
	P99Latency time.Duration `json:"p99_latency"` // P99延迟
	ErrorRate  float64       `json:"error_rate"`  // 错误率
	StartTime  time.Time     `json:"start_time"`  // 开始时间
	EndTime    time.Time     `json:"end_time"`    // 结束时间
	Duration   time.Duration `json:"duration"`    // 总耗时
}

// OperationFactory 操作工厂接口
type OperationFactory interface {
	CreateOperation(params map[string]interface{}) (Operation, error)
	GetOperationType() string
	ValidateParams(params map[string]interface{}) error
}

// TestContext 测试上下文接口
type TestContext interface {
	GetAdapter() ProtocolAdapter
	GetConfig() Config
	GetMetricsCollector() DefaultMetricsCollector
	GetKeyGenerator() KeyGenerator
	Cancel()
	IsCancelled() bool
}

// KeyGenerator 键生成器接口
type KeyGenerator interface {
	GenerateKey(operationType string, index int64) string
	GenerateRandomKey(operationType string, maxRange int) string
	GetGeneratedKeys() []string
	Reset()
}
