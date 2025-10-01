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
		return nil, fmt.Errorf("Redis adapter is not connected")
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

// executeRedisOperation 执行具体的Redis操作
func (r *RedisAdapter) executeRedisOperation(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	startTime := time.Now()
	result := &interfaces.OperationResult{
		IsRead: r.isReadOperation(operation.Type),
	}

	var err error
	switch operation.Type {
	case "get":
		result.Value, err = r.executeGet(ctx, operation)
	case "set":
		err = r.executeSet(ctx, operation)
	case "del":
		result.Value, err = r.executeDelete(ctx, operation)
	case "incr":
		result.Value, err = r.executeIncr(ctx, operation)
	case "decr":
		result.Value, err = r.executeDecr(ctx, operation)
	case "hget":
		result.Value, err = r.executeHGet(ctx, operation)
	case "hset":
		err = r.executeHSet(ctx, operation)
	case "hgetall":
		result.Value, err = r.executeHGetAll(ctx, operation)
	case "lpush":
		result.Value, err = r.executeLPush(ctx, operation)
	case "rpush":
		result.Value, err = r.executeRPush(ctx, operation)
	case "lpop":
		result.Value, err = r.executeLPop(ctx, operation)
	case "rpop":
		result.Value, err = r.executeRPop(ctx, operation)
	case "sadd":
		result.Value, err = r.executeSAdd(ctx, operation)
	case "smembers":
		result.Value, err = r.executeSMembers(ctx, operation)
	case "srem":
		result.Value, err = r.executeSRem(ctx, operation)
	case "sismember":
		result.Value, err = r.executeSIsMember(ctx, operation)
	case "zadd":
		result.Value, err = r.executeZAdd(ctx, operation)
	case "zrange":
		result.Value, err = r.executeZRange(ctx, operation)
	case "zrem":
		result.Value, err = r.executeZRem(ctx, operation)
	case "zrank":
		result.Value, err = r.executeZRank(ctx, operation)
	case "publish":
		result.Value, err = r.executePublish(ctx, operation)
	case "subscribe":
		result.Value, err = r.executeSubscribe(ctx, operation)
	default:
		err = fmt.Errorf("unsupported operation type: %s", operation.Type)
	}

	result.Duration = time.Since(startTime)
	result.Success = err == nil
	result.Error = err

	// 设置结果元数据
	result.Metadata = map[string]interface{}{
		"operation_type": operation.Type,
		"key":            operation.Key,
		"duration_ms":    result.Duration.Milliseconds(),
	}

	return result, nil
}

// Redis操作实现方法
func (r *RedisAdapter) executeGet(ctx context.Context, operation interfaces.Operation) (interface{}, error) {
	value, err := r.client.Get(ctx, operation.Key).Result()
	if err == redis.Nil {
		return nil, nil // Key不存在返回nil而不是错误
	}
	return value, err
}

func (r *RedisAdapter) executeSet(ctx context.Context, operation interfaces.Operation) error {
	valueStr := fmt.Sprintf("%v", operation.Value)
	ttl := operation.TTL
	if ttl <= 0 {
		ttl = 0 // 永不过期
	}
	return r.client.Set(ctx, operation.Key, valueStr, ttl).Err()
}

func (r *RedisAdapter) executeDelete(ctx context.Context, operation interfaces.Operation) (interface{}, error) {
	count, err := r.client.Del(ctx, operation.Key).Result()
	return count, err
}

func (r *RedisAdapter) executeIncr(ctx context.Context, operation interfaces.Operation) (interface{}, error) {
	value, err := r.client.Incr(ctx, operation.Key).Result()
	return value, err
}

func (r *RedisAdapter) executeDecr(ctx context.Context, operation interfaces.Operation) (interface{}, error) {
	value, err := r.client.Decr(ctx, operation.Key).Result()
	return value, err
}

func (r *RedisAdapter) executeHSet(ctx context.Context, operation interfaces.Operation) error {
	if hashValue, ok := operation.Value.(map[string]interface{}); ok {
		return r.client.HMSet(ctx, operation.Key, hashValue).Err()
	}
	return fmt.Errorf("invalid hash value format")
}

func (r *RedisAdapter) executeHGet(ctx context.Context, operation interfaces.Operation) (interface{}, error) {
	field := "field1" // 默认字段
	if fieldParam, exists := operation.Params["field"]; exists {
		if fieldStr, ok := fieldParam.(string); ok {
			field = fieldStr
		}
	}

	value, err := r.client.HGet(ctx, operation.Key, field).Result()
	if err == redis.Nil {
		return nil, nil
	}
	return value, err
}

func (r *RedisAdapter) executeHGetAll(ctx context.Context, operation interfaces.Operation) (interface{}, error) {
	values, err := r.client.HGetAll(ctx, operation.Key).Result()
	return values, err
}

func (r *RedisAdapter) executeLPush(ctx context.Context, operation interfaces.Operation) (interface{}, error) {
	if listValue, ok := operation.Value.([]string); ok {
		interfaces := make([]interface{}, len(listValue))
		for i, v := range listValue {
			interfaces[i] = v
		}
		count, err := r.client.LPush(ctx, operation.Key, interfaces...).Result()
		return count, err
	}
	return nil, fmt.Errorf("invalid list value format")
}

func (r *RedisAdapter) executeRPush(ctx context.Context, operation interfaces.Operation) (interface{}, error) {
	if listValue, ok := operation.Value.([]string); ok {
		interfaces := make([]interface{}, len(listValue))
		for i, v := range listValue {
			interfaces[i] = v
		}
		count, err := r.client.RPush(ctx, operation.Key, interfaces...).Result()
		return count, err
	}
	return nil, fmt.Errorf("invalid list value format")
}

func (r *RedisAdapter) executeLPop(ctx context.Context, operation interfaces.Operation) (interface{}, error) {
	value, err := r.client.LPop(ctx, operation.Key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	return value, err
}

func (r *RedisAdapter) executeRPop(ctx context.Context, operation interfaces.Operation) (interface{}, error) {
	value, err := r.client.RPop(ctx, operation.Key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	return value, err
}

func (r *RedisAdapter) executeSAdd(ctx context.Context, operation interfaces.Operation) (interface{}, error) {
	if setValue, ok := operation.Value.([]string); ok {
		interfaces := make([]interface{}, len(setValue))
		for i, v := range setValue {
			interfaces[i] = v
		}
		count, err := r.client.SAdd(ctx, operation.Key, interfaces...).Result()
		return count, err
	}
	return nil, fmt.Errorf("invalid set value format")
}

func (r *RedisAdapter) executeSMembers(ctx context.Context, operation interfaces.Operation) (interface{}, error) {
	members, err := r.client.SMembers(ctx, operation.Key).Result()
	return members, err
}

func (r *RedisAdapter) executeSRem(ctx context.Context, operation interfaces.Operation) (interface{}, error) {
	member := "default_member"
	if memberParam, exists := operation.Params["member"]; exists {
		if memberStr, ok := memberParam.(string); ok {
			member = memberStr
		}
	}

	count, err := r.client.SRem(ctx, operation.Key, member).Result()
	return count, err
}

func (r *RedisAdapter) executeSIsMember(ctx context.Context, operation interfaces.Operation) (interface{}, error) {
	member := "default_member"
	if memberParam, exists := operation.Params["member"]; exists {
		if memberStr, ok := memberParam.(string); ok {
			member = memberStr
		}
	}

	isMember, err := r.client.SIsMember(ctx, operation.Key, member).Result()
	return isMember, err
}

func (r *RedisAdapter) executeZAdd(ctx context.Context, operation interfaces.Operation) (interface{}, error) {
	if zsetValue, ok := operation.Value.(map[string]interface{}); ok {
		if members, exists := zsetValue["members"]; exists {
			if scores, exists := zsetValue["scores"]; exists {
				if memberList, ok := members.([]string); ok {
					if scoreList, ok := scores.([]float64); ok {
						if len(memberList) == len(scoreList) {
							var zValues []*redis.Z
							for i := 0; i < len(memberList); i++ {
								zValues = append(zValues, &redis.Z{
									Score:  scoreList[i],
									Member: memberList[i],
								})
							}
							count, err := r.client.ZAdd(ctx, operation.Key, zValues...).Result()
							return count, err
						}
					}
				}
			}
		}
	}
	return nil, fmt.Errorf("invalid zset value format")
}

func (r *RedisAdapter) executeZRange(ctx context.Context, operation interfaces.Operation) (interface{}, error) {
	start := int64(0)
	stop := int64(9)

	if startParam, exists := operation.Params["start"]; exists {
		if startInt, ok := startParam.(int); ok {
			start = int64(startInt)
		}
	}

	if stopParam, exists := operation.Params["stop"]; exists {
		if stopInt, ok := stopParam.(int); ok {
			stop = int64(stopInt)
		}
	}

	members, err := r.client.ZRange(ctx, operation.Key, start, stop).Result()
	return members, err
}

func (r *RedisAdapter) executeZRem(ctx context.Context, operation interfaces.Operation) (interface{}, error) {
	member := "default_member"
	if memberParam, exists := operation.Params["member"]; exists {
		if memberStr, ok := memberParam.(string); ok {
			member = memberStr
		}
	}

	count, err := r.client.ZRem(ctx, operation.Key, member).Result()
	return count, err
}

func (r *RedisAdapter) executeZRank(ctx context.Context, operation interfaces.Operation) (interface{}, error) {
	member := "default_member"
	if memberParam, exists := operation.Params["member"]; exists {
		if memberStr, ok := memberParam.(string); ok {
			member = memberStr
		}
	}

	rank, err := r.client.ZRank(ctx, operation.Key, member).Result()
	if err == redis.Nil {
		return nil, nil
	}
	return rank, err
}

func (r *RedisAdapter) executePublish(ctx context.Context, operation interfaces.Operation) (interface{}, error) {
	message := fmt.Sprintf("%v", operation.Value)
	count, err := r.client.Publish(ctx, operation.Key, message).Result()
	return count, err
}

func (r *RedisAdapter) executeSubscribe(ctx context.Context, operation interfaces.Operation) (interface{}, error) {
	return map[string]interface{}{
		"channel": operation.Key,
		"status":  "subscribed",
		"message": "subscription simulated (limited by redis.Cmdable interface)",
	}, nil
}

// isReadOperation 判断是否为读操作
func (r *RedisAdapter) isReadOperation(operationType string) bool {
	readOps := []string{
		"get", "hget", "hgetall", "lpop", "rpop", "smembers", "sismember",
		"zrange", "zrank", "subscribe",
	}

	for _, readOp := range readOps {
		if readOp == operationType {
			return true
		}
	}
	return false
}

// HealthCheck 健康检查
func (r *RedisAdapter) HealthCheck(ctx context.Context) error {
	if !r.isConnected {
		return fmt.Errorf("adapter not connected")
	}

	if r.client == nil {
		return fmt.Errorf("Redis client not initialized")
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
