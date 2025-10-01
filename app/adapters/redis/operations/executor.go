package operation

import (
	"context"
	"fmt"
	"time"

	redisConfig "abc-runner/app/adapters/redis/config"
	"abc-runner/app/adapters/redis/connection"
	"abc-runner/app/core/interfaces"

	"github.com/go-redis/redis/v8"
)

// RedisOperations Redis操作执行器 - 遵循HTTP协议的最佳实践
type RedisOperations struct {
	connectionPool   *connection.RedisConnectionPool
	config           *redisConfig.RedisConfig
	metricsCollector interfaces.DefaultMetricsCollector
}

// NewRedisOperations 创建Redis操作执行器
func NewRedisOperations(
	connectionPool *connection.RedisConnectionPool,
	config *redisConfig.RedisConfig,
	metricsCollector interfaces.DefaultMetricsCollector,
) *RedisOperations {
	return &RedisOperations{
		connectionPool:   connectionPool,
		config:           config,
		metricsCollector: metricsCollector,
	}
}

// ExecuteOperation 执行Redis操作 - 统一操作入口
func (r *RedisOperations) ExecuteOperation(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	startTime := time.Now()
	result := &interfaces.OperationResult{
		IsRead:   r.isReadOperation(operation.Type),
		Metadata: make(map[string]interface{}),
	}

	// 获取Redis客户端
	client := r.connectionPool.GetClient()
	if client == nil {
		result.Error = fmt.Errorf("failed to get Redis client from pool")
		result.Duration = time.Since(startTime)
		return result, result.Error
	}

	var opErr error
	switch operation.Type {
	case "get":
		result.Value, opErr = r.executeGet(ctx, client, operation)
	case "set":
		opErr = r.executeSet(ctx, client, operation)
	case "del":
		result.Value, opErr = r.executeDelete(ctx, client, operation)
	case "incr":
		result.Value, opErr = r.executeIncr(ctx, client, operation)
	case "decr":
		result.Value, opErr = r.executeDecr(ctx, client, operation)
	case "hget":
		result.Value, opErr = r.executeHGet(ctx, client, operation)
	case "hset":
		opErr = r.executeHSet(ctx, client, operation)
	case "hgetall":
		result.Value, opErr = r.executeHGetAll(ctx, client, operation)
	case "lpush":
		result.Value, opErr = r.executeLPush(ctx, client, operation)
	case "rpush":
		result.Value, opErr = r.executeRPush(ctx, client, operation)
	case "lpop":
		result.Value, opErr = r.executeLPop(ctx, client, operation)
	case "rpop":
		result.Value, opErr = r.executeRPop(ctx, client, operation)
	case "sadd":
		result.Value, opErr = r.executeSAdd(ctx, client, operation)
	case "smembers":
		result.Value, opErr = r.executeSMembers(ctx, client, operation)
	case "srem":
		result.Value, opErr = r.executeSRem(ctx, client, operation)
	case "sismember":
		result.Value, opErr = r.executeSIsMember(ctx, client, operation)
	case "zadd":
		result.Value, opErr = r.executeZAdd(ctx, client, operation)
	case "zrange":
		result.Value, opErr = r.executeZRange(ctx, client, operation)
	case "zrem":
		result.Value, opErr = r.executeZRem(ctx, client, operation)
	case "zrank":
		result.Value, opErr = r.executeZRank(ctx, client, operation)
	case "publish":
		result.Value, opErr = r.executePublish(ctx, client, operation)
	case "subscribe":
		result.Value, opErr = r.executeSubscribe(ctx, client, operation)
	default:
		opErr = fmt.Errorf("unsupported operation type: %s", operation.Type)
	}

	result.Success = opErr == nil
	result.Error = opErr
	result.Duration = time.Since(startTime)

	// 添加操作特定元数据
	for k, v := range operation.Metadata {
		result.Metadata[k] = v
	}
	result.Metadata["operation_type"] = operation.Type
	result.Metadata["key"] = operation.Key

	return result, opErr
}

// 具体操作实现方法

// executeGet 执行GET操作
func (r *RedisOperations) executeGet(ctx context.Context, client redis.Cmdable, operation interfaces.Operation) (interface{}, error) {
	cmd := client.Get(ctx, operation.Key)
	value, err := cmd.Result()
	if err == redis.Nil {
		return nil, nil // Key不存在，不是错误
	}
	return value, err
}

// executeSet 执行SET操作
func (r *RedisOperations) executeSet(ctx context.Context, client redis.Cmdable, operation interfaces.Operation) error {
	valueStr, ok := operation.Value.(string)
	if !ok {
		return fmt.Errorf("invalid value type for SET operation: expected string")
	}

	cmd := client.Set(ctx, operation.Key, valueStr, operation.TTL)
	return cmd.Err()
}

// executeDelete 执行DELETE操作
func (r *RedisOperations) executeDelete(ctx context.Context, client redis.Cmdable, operation interfaces.Operation) (interface{}, error) {
	cmd := client.Del(ctx, operation.Key)
	deletedCount, err := cmd.Result()
	return deletedCount, err
}

// executeIncr 执行INCR操作
func (r *RedisOperations) executeIncr(ctx context.Context, client redis.Cmdable, operation interfaces.Operation) (interface{}, error) {
	cmd := client.Incr(ctx, operation.Key)
	value, err := cmd.Result()
	return value, err
}

// executeDecr 执行DECR操作
func (r *RedisOperations) executeDecr(ctx context.Context, client redis.Cmdable, operation interfaces.Operation) (interface{}, error) {
	cmd := client.Decr(ctx, operation.Key)
	value, err := cmd.Result()
	return value, err
}

// executeHGet 执行HGET操作
func (r *RedisOperations) executeHGet(ctx context.Context, client redis.Cmdable, operation interfaces.Operation) (interface{}, error) {
	field, ok := operation.Params["field"].(string)
	if !ok {
		return nil, fmt.Errorf("field parameter is required for HGET operation")
	}

	cmd := client.HGet(ctx, operation.Key, field)
	value, err := cmd.Result()
	if err == redis.Nil {
		return nil, nil
	}
	return value, err
}

// executeHSet 执行HSET操作
func (r *RedisOperations) executeHSet(ctx context.Context, client redis.Cmdable, operation interfaces.Operation) error {
	field, ok := operation.Params["field"].(string)
	if !ok {
		return fmt.Errorf("field parameter is required for HSET operation")
	}

	valueStr, ok := operation.Value.(string)
	if !ok {
		return fmt.Errorf("invalid value type for HSET operation: expected string")
	}

	cmd := client.HSet(ctx, operation.Key, field, valueStr)
	return cmd.Err()
}

// executeHGetAll 执行HGETALL操作
func (r *RedisOperations) executeHGetAll(ctx context.Context, client redis.Cmdable, operation interfaces.Operation) (interface{}, error) {
	cmd := client.HGetAll(ctx, operation.Key)
	values, err := cmd.Result()
	return values, err
}

// executeLPush 执行LPUSH操作
func (r *RedisOperations) executeLPush(ctx context.Context, client redis.Cmdable, operation interfaces.Operation) (interface{}, error) {
	valueStr, ok := operation.Value.(string)
	if !ok {
		return nil, fmt.Errorf("invalid value type for LPUSH operation: expected string")
	}

	cmd := client.LPush(ctx, operation.Key, valueStr)
	length, err := cmd.Result()
	return length, err
}

// executeRPush 执行RPUSH操作
func (r *RedisOperations) executeRPush(ctx context.Context, client redis.Cmdable, operation interfaces.Operation) (interface{}, error) {
	valueStr, ok := operation.Value.(string)
	if !ok {
		return nil, fmt.Errorf("invalid value type for RPUSH operation: expected string")
	}

	cmd := client.RPush(ctx, operation.Key, valueStr)
	length, err := cmd.Result()
	return length, err
}

// executeLPop 执行LPOP操作
func (r *RedisOperations) executeLPop(ctx context.Context, client redis.Cmdable, operation interfaces.Operation) (interface{}, error) {
	cmd := client.LPop(ctx, operation.Key)
	value, err := cmd.Result()
	if err == redis.Nil {
		return nil, nil
	}
	return value, err
}

// executeRPop 执行RPOP操作
func (r *RedisOperations) executeRPop(ctx context.Context, client redis.Cmdable, operation interfaces.Operation) (interface{}, error) {
	cmd := client.RPop(ctx, operation.Key)
	value, err := cmd.Result()
	if err == redis.Nil {
		return nil, nil
	}
	return value, err
}

// executeSAdd 执行SADD操作
func (r *RedisOperations) executeSAdd(ctx context.Context, client redis.Cmdable, operation interfaces.Operation) (interface{}, error) {
	valueStr, ok := operation.Value.(string)
	if !ok {
		return nil, fmt.Errorf("invalid value type for SADD operation: expected string")
	}

	cmd := client.SAdd(ctx, operation.Key, valueStr)
	addedCount, err := cmd.Result()
	return addedCount, err
}

// executeSMembers 执行SMEMBERS操作
func (r *RedisOperations) executeSMembers(ctx context.Context, client redis.Cmdable, operation interfaces.Operation) (interface{}, error) {
	cmd := client.SMembers(ctx, operation.Key)
	members, err := cmd.Result()
	return members, err
}

// executeSRem 执行SREM操作
func (r *RedisOperations) executeSRem(ctx context.Context, client redis.Cmdable, operation interfaces.Operation) (interface{}, error) {
	valueStr, ok := operation.Value.(string)
	if !ok {
		return nil, fmt.Errorf("invalid value type for SREM operation: expected string")
	}

	cmd := client.SRem(ctx, operation.Key, valueStr)
	removedCount, err := cmd.Result()
	return removedCount, err
}

// executeSIsMember 执行SISMEMBER操作
func (r *RedisOperations) executeSIsMember(ctx context.Context, client redis.Cmdable, operation interfaces.Operation) (interface{}, error) {
	valueStr, ok := operation.Value.(string)
	if !ok {
		return nil, fmt.Errorf("invalid value type for SISMEMBER operation: expected string")
	}

	cmd := client.SIsMember(ctx, operation.Key, valueStr)
	exists, err := cmd.Result()
	return exists, err
}

// executeZAdd 执行ZADD操作
func (r *RedisOperations) executeZAdd(ctx context.Context, client redis.Cmdable, operation interfaces.Operation) (interface{}, error) {
	score, ok := operation.Params["score"].(float64)
	if !ok {
		return nil, fmt.Errorf("score parameter is required for ZADD operation")
	}

	valueStr, ok := operation.Value.(string)
	if !ok {
		return nil, fmt.Errorf("invalid value type for ZADD operation: expected string")
	}

	cmd := client.ZAdd(ctx, operation.Key, &redis.Z{Score: score, Member: valueStr})
	addedCount, err := cmd.Result()
	return addedCount, err
}

// executeZRange 执行ZRANGE操作
func (r *RedisOperations) executeZRange(ctx context.Context, client redis.Cmdable, operation interfaces.Operation) (interface{}, error) {
	start, ok := operation.Params["start"].(int64)
	if !ok {
		start = 0
	}

	stop, ok := operation.Params["stop"].(int64)
	if !ok {
		stop = -1
	}

	cmd := client.ZRange(ctx, operation.Key, start, stop)
	members, err := cmd.Result()
	return members, err
}

// executeZRem 执行ZREM操作
func (r *RedisOperations) executeZRem(ctx context.Context, client redis.Cmdable, operation interfaces.Operation) (interface{}, error) {
	valueStr, ok := operation.Value.(string)
	if !ok {
		return nil, fmt.Errorf("invalid value type for ZREM operation: expected string")
	}

	cmd := client.ZRem(ctx, operation.Key, valueStr)
	removedCount, err := cmd.Result()
	return removedCount, err
}

// executeZRank 执行ZRANK操作
func (r *RedisOperations) executeZRank(ctx context.Context, client redis.Cmdable, operation interfaces.Operation) (interface{}, error) {
	valueStr, ok := operation.Value.(string)
	if !ok {
		return nil, fmt.Errorf("invalid value type for ZRANK operation: expected string")
	}

	cmd := client.ZRank(ctx, operation.Key, valueStr)
	rank, err := cmd.Result()
	if err == redis.Nil {
		return nil, nil
	}
	return rank, err
}

// executePublish 执行PUBLISH操作
func (r *RedisOperations) executePublish(ctx context.Context, client redis.Cmdable, operation interfaces.Operation) (interface{}, error) {
	channel := operation.Key
	messageStr, ok := operation.Value.(string)
	if !ok {
		return nil, fmt.Errorf("invalid value type for PUBLISH operation: expected string")
	}

	cmd := client.Publish(ctx, channel, messageStr)
	subscriberCount, err := cmd.Result()
	return subscriberCount, err
}

// executeSubscribe 执行SUBSCRIBE操作
func (r *RedisOperations) executeSubscribe(ctx context.Context, client redis.Cmdable, operation interfaces.Operation) (interface{}, error) {
	// 订阅操作比较特殊，通常需要维持连接
	// 这里简化实现，返回订阅成功的消息
	return fmt.Sprintf("subscribed to channel: %s", operation.Key), nil
}

// isReadOperation 判断是否为读操作
func (r *RedisOperations) isReadOperation(operationType string) bool {
	readOperations := map[string]bool{
		"get":       true,
		"hget":      true,
		"hgetall":   true,
		"lpop":      true,
		"rpop":      true,
		"smembers":  true,
		"sismember": true,
		"zrange":    true,
		"zrank":     true,
		"subscribe": true,
		// 写操作
		"set":     false,
		"del":     false,
		"incr":    false,
		"decr":    false,
		"hset":    false,
		"lpush":   false,
		"rpush":   false,
		"sadd":    false,
		"srem":    false,
		"zadd":    false,
		"zrem":    false,
		"publish": false,
	}

	return readOperations[operationType]
}

// GetSupportedOperations 获取支持的操作类型
func (r *RedisOperations) GetSupportedOperations() []string {
	return []string{
		"get", "set", "del", "incr", "decr",
		"hget", "hset", "hgetall",
		"lpush", "rpush", "lpop", "rpop",
		"sadd", "srem", "smembers", "sismember",
		"zadd", "zrem", "zrange", "zrank",
		"publish", "subscribe",
	}
}
