package redis

import (
	"context"
	"fmt"
	"sync"
	"time"

	redisConfig "abc-runner/app/adapters/redis/config"
	"abc-runner/app/adapters/redis/connection"
	operation "abc-runner/app/adapters/redis/operations"
	"abc-runner/app/core/interfaces"

	"github.com/go-redis/redis/v8"
)

// RedisAdapter Redis协议适配器 - 遵循HTTP协议最佳实践
type RedisAdapter struct {
	// 核心组件
	connectionPool  *connection.RedisConnectionPool
	redisOperations *operation.RedisOperations
	client          redis.Cmdable
	config          *redisConfig.RedisConfig

	// 指标收集器
	metricsCollector interfaces.DefaultMetricsCollector

	// 状态管理
	isConnected bool
	mutex       sync.RWMutex

	// 统计信息
	totalOperations   int64
	successOperations int64
	failedOperations  int64
	startTime         time.Time
}

// NewRedisAdapter 创建Redis适配器 - 新架构
func NewRedisAdapter(metricsCollector interfaces.DefaultMetricsCollector) *RedisAdapter {
	if metricsCollector == nil {
		panic("metricsCollector cannot be nil - dependency injection required")
	}

	return &RedisAdapter{
		metricsCollector: metricsCollector,
		startTime:        time.Now(),
	}
}

// Connect 初始化连接
func (r *RedisAdapter) Connect(ctx context.Context, config interfaces.Config) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// 类型断言配置
	redisConfig, ok := config.(*redisConfig.RedisConfig)
	if !ok {
		return fmt.Errorf("invalid config type for Redis adapter: expected *redisConfig.RedisConfig, got %T", config)
	}

	// 验证配置
	if err := redisConfig.Validate(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	r.config = redisConfig

	// 创建连接池
	pool, err := connection.NewRedisConnectionPool(redisConfig)
	if err != nil {
		return fmt.Errorf("failed to create Redis connection pool: %w", err)
	}

	r.connectionPool = pool

	// 创建Redis操作执行器
	r.redisOperations = operation.NewRedisOperations(pool, redisConfig, r.metricsCollector)

	// 获取客户端
	client := pool.GetClient()
	if client == nil {
		return fmt.Errorf("failed to get Redis client from pool")
	}

	r.client = client

	// 执行健康检查
	if err := r.HealthCheck(ctx); err != nil {
		return fmt.Errorf("initial health check failed: %w", err)
	}

	r.isConnected = true
	return nil
}

// Execute 执行Redis操作 - 使用RedisOperations执行器
func (r *RedisAdapter) Execute(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	if !r.IsConnected() {
		return nil, fmt.Errorf("redis adapter is not connected")
	}

	// 统计操作数
	r.incrementTotalOperations()

	// 执行操作
	result, err := r.redisOperations.ExecuteOperation(ctx, operation)

	// 更新统计信息
	if err != nil || (result != nil && !result.Success) {
		r.incrementFailedOperations()
	} else {
		r.incrementSuccessOperations()
	}

	// 注意：不要在这里调用 r.metricsCollector.Record(result)
	// 因为执行引擎会负责记录指标，避免重复计数

	return result, err
}

// Close 关闭连接
func (r *RedisAdapter) Close() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.connectionPool != nil {
		if err := r.connectionPool.Close(); err != nil {
			return fmt.Errorf("failed to close Redis connection pool: %w", err)
		}
		r.connectionPool = nil
	}

	r.client = nil
	r.isConnected = false

	return nil
}

// HealthCheck 健康检查
func (r *RedisAdapter) HealthCheck(ctx context.Context) error {
	if !r.isConnected {
		return fmt.Errorf("adapter not connected")
	}

	if r.client == nil {
		return fmt.Errorf("redis client not initialized")
	}

	// 执行ping操作进行健康检查
	cmd := r.client.Ping(ctx)
	return cmd.Err()
}

// GetProtocolName 获取协议名称
func (r *RedisAdapter) GetProtocolName() string {
	return "redis"
}

// GetMetricsCollector 获取指标收集器
func (r *RedisAdapter) GetMetricsCollector() interfaces.DefaultMetricsCollector {
	return r.metricsCollector
}

// GetProtocolMetrics 获取Redis特定指标
func (r *RedisAdapter) GetProtocolMetrics() map[string]interface{} {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	metrics := map[string]interface{}{
		"protocol":           "redis",
		"is_connected":       r.isConnected,
		"total_operations":   r.totalOperations,
		"success_operations": r.successOperations,
		"failed_operations":  r.failedOperations,
		"uptime_seconds":     time.Since(r.startTime).Seconds(),
	}

	// 添加连接池统计信息
	if r.connectionPool != nil {
		poolStats := r.connectionPool.GetStats()
		metrics["connection_pool"] = poolStats
	}

	// 添加配置信息
	if r.config != nil {
		connectionConfig := r.config.GetConnection()
		metrics["config"] = map[string]interface{}{
			"addresses": connectionConfig.GetAddresses(),
			"mode":      r.config.GetMode(),
		}
	}

	return metrics
}

// IsConnected 检查连接状态
func (r *RedisAdapter) IsConnected() bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.isConnected
}

// incrementTotalOperations 增加总操作数
func (r *RedisAdapter) incrementTotalOperations() {
	r.totalOperations++
}

// incrementSuccessOperations 增加成功操作数
func (r *RedisAdapter) incrementSuccessOperations() {
	r.successOperations++
}

// incrementFailedOperations 增加失败操作数
func (r *RedisAdapter) incrementFailedOperations() {
	r.failedOperations++
}

// GetOperationStats 获取操作统计信息
func (r *RedisAdapter) GetOperationStats() map[string]interface{} {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var successRate float64
	if r.totalOperations > 0 {
		successRate = float64(r.successOperations) / float64(r.totalOperations) * 100
	}

	return map[string]interface{}{
		"total_operations":   r.totalOperations,
		"success_operations": r.successOperations,
		"failed_operations":  r.failedOperations,
		"success_rate":       successRate,
		"uptime":             time.Since(r.startTime),
	}
}

// 确保实现了ProtocolAdapter接口
var _ interfaces.ProtocolAdapter = (*RedisAdapter)(nil)
