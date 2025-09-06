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

// MetricsCollector 指标收集器接口
type MetricsCollector interface {
	RecordOperation(result *OperationResult)
	GetMetrics() *Metrics
	Reset()
	Export() map[string]interface{}
}

// Metrics 性能指标结构
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
	GetMetricsCollector() MetricsCollector
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
