package config

import (
	"time"
)

// RedisClientFactory Redis客户端工厂接口
type RedisClientFactory interface {
	CreateStandaloneClient(config StandAloneInfo) (interface{}, error)
	CreateSentinelClient(config SentinelInfo) (interface{}, error)
	CreateClusterClient(config ClusterInfo) (interface{}, error)
}

// OperationFactory 操作工厂接口
type OperationFactory interface {
	CreateOperation(operationType string) (Operation, error)
	RegisterOperation(operationType string, factory func() Operation)
	ListSupportedOperations() []string
}

// Operation 操作接口
type Operation interface {
	Execute(client interface{}, params OperationParams) OperationResult
	GetType() string
	Validate(params OperationParams) error
}

// OperationParams 操作参数
type OperationParams struct {
	Key       string
	Value     interface{}
	TTL       time.Duration
	Database  int
	ExtraArgs map[string]interface{}
}

// OperationResult 操作结果
type OperationResult struct {
	Success   bool
	IsRead    bool
	Duration  time.Duration
	Error     error
	Value     interface{}
	ExtraData map[string]interface{}
}

// MetricsCollector 指标收集器接口
type MetricsCollector interface {
	CollectOperation(result OperationResult)
	CollectConnection(connectionType string, success bool, duration time.Duration)
	GetMetrics() map[string]interface{}
	Reset()
}

// MetricsReporter 指标报告器接口
type MetricsReporter interface {
	Report(metrics map[string]interface{}) error
	SetFormat(format string)
	SetOutput(output string)
}
