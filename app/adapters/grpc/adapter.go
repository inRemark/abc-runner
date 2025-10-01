package grpc

import (
	"abc-runner/app/adapters/grpc/config"
	"abc-runner/app/adapters/grpc/connection"
	"abc-runner/app/adapters/grpc/operations"
	"abc-runner/app/core/interfaces"
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// GRPCAdapter gRPC协议适配器 - 遵循统一架构模式
// 职责：连接管理、状态维护、健康检查
type GRPCAdapter struct {
	config           *config.GRPCConfig
	connectionPool   *connection.ConnectionPool
	grpcOperations   *operations.GRPCOperations
	metricsCollector interfaces.DefaultMetricsCollector
	mu               sync.RWMutex
	isConnected      bool

	// 统计信息
	totalCalls      int64
	successfulCalls int64
	failedCalls     int64
	totalLatency    time.Duration
	startTime       time.Time
}

// NewGRPCAdapter 创建新的gRPC适配器
func NewGRPCAdapter(metricsCollector interfaces.DefaultMetricsCollector) *GRPCAdapter {
	return &GRPCAdapter{
		metricsCollector: metricsCollector,
		isConnected:      false,
		startTime:        time.Now(),
	}
}

// Connect 连接到gRPC服务器
func (adapter *GRPCAdapter) Connect(ctx context.Context, cfg interfaces.Config) error {
	adapter.mu.Lock()
	defer adapter.mu.Unlock()

	// 转换配置
	grpcConfig, err := adapter.parseConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to parse gRPC config: %w", err)
	}

	adapter.config = grpcConfig

	// 验证配置
	if err := adapter.config.Validate(); err != nil {
		return fmt.Errorf("invalid gRPC config: %w", err)
	}

	// 创建连接池
	adapter.connectionPool = connection.NewConnectionPool(adapter.config)
	if err := adapter.connectionPool.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize connection pool: %w", err)
	}

	// 创建gRPC操作执行器
	adapter.grpcOperations = operations.NewGRPCOperations(adapter.connectionPool, adapter.config, adapter.metricsCollector)

	adapter.isConnected = true
	log.Printf("Successfully connected to gRPC server: %s:%d with %d connections",
		adapter.config.Connection.Address, adapter.config.Connection.Port,
		adapter.config.Connection.Pool.PoolSize)

	return nil
}

// Execute 执行操作 - 使用执行器处理
func (adapter *GRPCAdapter) Execute(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	if !adapter.isConnected {
		return nil, fmt.Errorf("adapter not connected")
	}

	startTime := time.Now()

	// 委托给gRPC操作执行器处理
	result, err := adapter.grpcOperations.ExecuteOperation(ctx, operation)

	// 记录指标（仅作为适配器内部统计）
	duration := time.Since(startTime)
	adapter.recordMetrics(operation.Type, duration, err == nil)

	if err != nil {
		return nil, fmt.Errorf("gRPC operation failed: %w", err)
	}

	return result, nil
}

// Close 关闭适配器
func (adapter *GRPCAdapter) Close() error {
	adapter.mu.Lock()
	defer adapter.mu.Unlock()

	if !adapter.isConnected {
		return nil
	}

	// 关闭连接池
	if adapter.connectionPool != nil {
		if err := adapter.connectionPool.Close(); err != nil {
			log.Printf("Error closing connection pool: %v", err)
		}
	}

	adapter.isConnected = false
	log.Println("gRPC adapter closed")
	return nil
}

// GetProtocolMetrics 获取协议指标
func (adapter *GRPCAdapter) GetProtocolMetrics() map[string]interface{} {
	adapter.mu.RLock()
	defer adapter.mu.RUnlock()

	var avgLatency float64
	if adapter.totalCalls > 0 {
		avgLatency = float64(adapter.totalLatency.Nanoseconds()) / float64(adapter.totalCalls) / 1e6
	}

	successRate := float64(0)
	if adapter.totalCalls > 0 {
		successRate = float64(adapter.successfulCalls) / float64(adapter.totalCalls) * 100
	}

	metrics := map[string]interface{}{
		"protocol":         "grpc",
		"total_calls":      adapter.totalCalls,
		"successful_calls": adapter.successfulCalls,
		"failed_calls":     adapter.failedCalls,
		"success_rate":     successRate,
		"avg_latency_ms":   avgLatency,
		"uptime":           time.Since(adapter.startTime),
		"service_name":     adapter.config.GRPCSpecific.ServiceName,
		"method_name":      adapter.config.GRPCSpecific.MethodName,
	}

	// 添加连接池指标
	if adapter.connectionPool != nil {
		poolStats := adapter.connectionPool.GetStats()
		metrics["connection_pool"] = poolStats
	}

	return metrics
}

// HealthCheck 健康检查
func (adapter *GRPCAdapter) HealthCheck(ctx context.Context) error {
	if !adapter.isConnected {
		return fmt.Errorf("adapter not connected")
	}

	if adapter.connectionPool == nil {
		return fmt.Errorf("connection pool not initialized")
	}

	// 尝试获取一个连接来验证健康状态
	connWrapper, err := adapter.connectionPool.GetConnection()
	if err != nil {
		return fmt.Errorf("failed to get healthy connection: %w", err)
	}

	if !connWrapper.IsHealthy() {
		return fmt.Errorf("connection is not healthy")
	}

	return nil
}

// GetProtocolName 获取协议名称
func (adapter *GRPCAdapter) GetProtocolName() string {
	return "grpc"
}

// GetMetricsCollector 获取指标收集器
func (adapter *GRPCAdapter) GetMetricsCollector() interfaces.DefaultMetricsCollector {
	return adapter.metricsCollector
}

// 内部辅助方法

// parseConfig 解析配置
func (adapter *GRPCAdapter) parseConfig(cfg interfaces.Config) (*config.GRPCConfig, error) {
	// 如果传入的是gRPC配置类型，直接使用
	if gCfg, ok := cfg.(*config.GRPCConfig); ok {
		return gCfg, nil
	}

	// 使用默认配置
	grpcConfig := config.NewDefaultGRPCConfig()
	return grpcConfig, nil
}

// recordMetrics 记录指标（仅作为适配器内部统计）
func (adapter *GRPCAdapter) recordMetrics(operationType string, duration time.Duration, success bool) {
	adapter.mu.Lock()
	defer adapter.mu.Unlock()

	adapter.totalCalls++
	adapter.totalLatency += duration

	if success {
		adapter.successfulCalls++
	} else {
		adapter.failedCalls++
	}

	// 注意：不要在这里调用 adapter.metricsCollector.Record(result)
	// 因为执行引擎会负责记录指标，避免重复计数
	// 这里只做适配器内部的统计
}
