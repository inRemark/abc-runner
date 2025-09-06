package redis

import (
	"context"
	"fmt"
	"sync"
	"time"

	"redis-runner/app/core/base"
	redisconfig "redis-runner/app/adapters/redis/config"
	"redis-runner/app/core/interfaces"

	"github.com/go-redis/redis/v8"
)

// RedisAdapter Redis协议适配器
type RedisAdapter struct {
	*base.BaseAdapter
	client           redis.Cmdable
	clusterClient    *redis.ClusterClient
	standaloneClient *redis.Client
	mode             string
	config           *redisconfig.RedisConfig
	mutex            sync.RWMutex
}

// NewRedisAdapter 创建Redis适配器
func NewRedisAdapter() *RedisAdapter {
	return &RedisAdapter{
		BaseAdapter: base.NewBaseAdapter("redis"),
	}
}

// Connect 初始化连接
func (r *RedisAdapter) Connect(ctx context.Context, cfg interfaces.Config) error {
	// 提取Redis配置
	var redisConfig *redisconfig.RedisConfig
	if adapter, ok := cfg.(*redisconfig.RedisConfigAdapter); ok {
		redisConfig = adapter.GetRedisConfig()
	} else {
		// 如果不是适配器，尝试转换
		var err error
		redisConfig, err = redisconfig.ExtractRedisConfig(cfg)
		if err != nil {
			return fmt.Errorf("failed to extract Redis config: %w", err)
		}
	}

	if err := r.ValidateConfig(cfg); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	r.config = redisConfig
	r.mode = redisConfig.GetMode()

	var client redis.Cmdable
	var err error

	switch r.mode {
	case "standalone":
		client, err = r.createStandaloneClient(redisConfig)
	case "sentinel":
		client, err = r.createSentinelClient(redisConfig)
	case "cluster":
		client, err = r.createClusterClient(redisConfig)
	default:
		return fmt.Errorf("unsupported Redis mode: %s", r.mode)
	}

	if err != nil {
		return fmt.Errorf("failed to create Redis client: %w", err)
	}

	r.client = client

	// 测试连接
	if err := r.HealthCheck(ctx); err != nil {
		return fmt.Errorf("connection health check failed: %w", err)
	}

	r.SetConnected(true)
	r.SetConfig(cfg)

	return nil
}

// Execute 执行操作
func (r *RedisAdapter) Execute(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	if !r.IsConnected() {
		return nil, fmt.Errorf("Redis adapter is not connected")
	}

	start := time.Now()
	result := &interfaces.OperationResult{
		IsRead: r.isReadOperation(operation.Type),
	}

	switch operation.Type {
	case "get":
		value, err := r.executeGet(ctx, operation)
		result.Value = value
		result.Error = err
		result.Success = err == nil

	case "set":
		err := r.executeSet(ctx, operation)
		result.Error = err
		result.Success = err == nil

	case "del":
		count, err := r.executeDelete(ctx, operation)
		result.Value = count
		result.Error = err
		result.Success = err == nil

	case "hget":
		value, err := r.executeHGet(ctx, operation)
		result.Value = value
		result.Error = err
		result.Success = err == nil

	case "hset":
		err := r.executeHSet(ctx, operation)
		result.Error = err
		result.Success = err == nil

	case "pub":
		err := r.executePublish(ctx, operation)
		result.Error = err
		result.Success = err == nil

	case "set_get_random":
		err := r.executeSetGetRandom(ctx, operation)
		result.Error = err
		result.Success = err == nil

	default:
		return nil, fmt.Errorf("unsupported operation type: %s", operation.Type)
	}

	result.Duration = time.Since(start)

	// 更新协议指标
	r.updateOperationMetrics(operation.Type, result.Success, result.Duration)

	return result, nil
}

// executeGet 执行GET操作
func (r *RedisAdapter) executeGet(ctx context.Context, operation interfaces.Operation) (string, error) {
	value, err := r.client.Get(ctx, operation.Key).Result()
	if err == redis.Nil {
		return "", nil // 键不存在，返回空字符串而不是错误
	}
	return value, err
}

// executeSet 执行SET操作
func (r *RedisAdapter) executeSet(ctx context.Context, operation interfaces.Operation) error {
	value := fmt.Sprintf("%v", operation.Value)
	return r.client.Set(ctx, operation.Key, value, operation.TTL).Err()
}

// executeDelete 执行DELETE操作
func (r *RedisAdapter) executeDelete(ctx context.Context, operation interfaces.Operation) (int64, error) {
	return r.client.Del(ctx, operation.Key).Result()
}

// executeHGet 执行HGET操作
func (r *RedisAdapter) executeHGet(ctx context.Context, operation interfaces.Operation) (string, error) {
	field, ok := operation.Params["field"].(string)
	if !ok {
		return "", fmt.Errorf("field parameter is required for HGET operation")
	}

	value, err := r.client.HGet(ctx, operation.Key, field).Result()
	if err == redis.Nil {
		return "", nil
	}
	return value, err
}

// executeHSet 执行HSET操作
func (r *RedisAdapter) executeHSet(ctx context.Context, operation interfaces.Operation) error {
	field, ok := operation.Params["field"].(string)
	if !ok {
		return fmt.Errorf("field parameter is required for HSET operation")
	}

	value := fmt.Sprintf("%v", operation.Value)
	return r.client.HSet(ctx, operation.Key, field, value).Err()
}

// executePublish 执行PUBLISH操作
func (r *RedisAdapter) executePublish(ctx context.Context, operation interfaces.Operation) error {
	channel, ok := operation.Params["channel"].(string)
	if !ok {
		channel = "my_channel" // 默认频道
	}

	value := fmt.Sprintf("%v", operation.Value)
	return r.client.Publish(ctx, channel, value).Err()
}

// executeSetGetRandom 执行混合读写操作
func (r *RedisAdapter) executeSetGetRandom(ctx context.Context, operation interfaces.Operation) error {
	readPercent, _ := operation.Params["read_percent"].(int)
	if readPercent <= 0 {
		readPercent = 50
	}

	// 基于时间生成随机数
	if time.Now().UnixNano()%100 < int64(readPercent) {
		// 执行读操作
		_, err := r.executeGet(ctx, operation)
		return err
	} else {
		// 执行写操作
		return r.executeSet(ctx, operation)
	}
}

// createStandaloneClient 创建单机客户端
func (r *RedisAdapter) createStandaloneClient(config *redisconfig.RedisConfig) (redis.Cmdable, error) {
	standalone := config.GetStandaloneConfig()
	
	client := redis.NewClient(&redis.Options{
		Addr:         standalone.Addr,
		Password:     standalone.Password,
		DB:           standalone.Db,
		PoolSize:     config.Pool.GetPoolSize(),
		MinIdleConns: config.Pool.GetMinIdle(),
		MaxRetries:   3,
		DialTimeout:  config.Pool.GetConnectionTimeout(),
		IdleTimeout:  config.Pool.GetIdleTimeout(),
	})

	r.standaloneClient = client
	return client, nil
}

// createSentinelClient 创建哨兵客户端
func (r *RedisAdapter) createSentinelClient(config *redisconfig.RedisConfig) (redis.Cmdable, error) {
	sentinel := config.GetSentinelConfig()
	
	client := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    sentinel.MasterName,
		SentinelAddrs: sentinel.Addrs,
		Password:      sentinel.Password,
		DB:            sentinel.Db,
		PoolSize:      config.Pool.GetPoolSize(),
		MinIdleConns:  config.Pool.GetMinIdle(),
		MaxRetries:    3,
		DialTimeout:   config.Pool.GetConnectionTimeout(),
		IdleTimeout:   config.Pool.GetIdleTimeout(),
	})

	r.standaloneClient = client
	return client, nil
}

// createClusterClient 创建集群客户端
func (r *RedisAdapter) createClusterClient(config *redisconfig.RedisConfig) (redis.Cmdable, error) {
	cluster := config.GetClusterConfig()
	
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        cluster.Addrs,
		Password:     cluster.Password,
		PoolSize:     config.Pool.GetPoolSize(),
		MinIdleConns: config.Pool.GetMinIdle(),
		MaxRetries:   3,
		DialTimeout:  config.Pool.GetConnectionTimeout(),
		IdleTimeout:  config.Pool.GetIdleTimeout(),
	})

	r.clusterClient = client
	return client, nil
}

// Close 关闭连接
func (r *RedisAdapter) Close() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	var err error

	if r.standaloneClient != nil {
		err = r.standaloneClient.Close()
		r.standaloneClient = nil
	}

	if r.clusterClient != nil {
		if closeErr := r.clusterClient.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
		r.clusterClient = nil
	}

	r.client = nil
	r.SetConnected(false)

	return err
}

// HealthCheck 健康检查
func (r *RedisAdapter) HealthCheck(ctx context.Context) error {
	if r.client == nil {
		return fmt.Errorf("Redis client is not initialized")
	}

	_, err := r.client.Ping(ctx).Result()
	return err
}

// isReadOperation 判断是否为读操作
func (r *RedisAdapter) isReadOperation(operationType string) bool {
	readOps := []string{"get", "hget", "hgetall", "lrange", "smembers", "zrange", "exists", "ttl"}
	for _, op := range readOps {
		if op == operationType {
			return true
		}
	}
	return false
}

// updateOperationMetrics 更新操作指标
func (r *RedisAdapter) updateOperationMetrics(operationType string, success bool, duration time.Duration) {
	r.UpdateMetric("last_operation_type", operationType)
	r.UpdateMetric("last_operation_success", success)
	r.UpdateMetric("last_operation_duration", duration.Nanoseconds())

	// 更新连接池状态
	if r.standaloneClient != nil {
		poolStats := r.standaloneClient.PoolStats()
		r.UpdateMetric("pool_hits", poolStats.Hits)
		r.UpdateMetric("pool_misses", poolStats.Misses)
		r.UpdateMetric("pool_timeouts", poolStats.Timeouts)
		r.UpdateMetric("pool_total_conns", poolStats.TotalConns)
		r.UpdateMetric("pool_idle_conns", poolStats.IdleConns)
		r.UpdateMetric("pool_stale_conns", poolStats.StaleConns)
	}

	if r.clusterClient != nil {
		poolStats := r.clusterClient.PoolStats()
		r.UpdateMetric("cluster_pool_hits", poolStats.Hits)
		r.UpdateMetric("cluster_pool_misses", poolStats.Misses)
		r.UpdateMetric("cluster_pool_timeouts", poolStats.Timeouts)
		r.UpdateMetric("cluster_pool_total_conns", poolStats.TotalConns)
		r.UpdateMetric("cluster_pool_idle_conns", poolStats.IdleConns)
		r.UpdateMetric("cluster_pool_stale_conns", poolStats.StaleConns)
	}
}

// GetClient 获取Redis客户端（用于高级操作）
func (r *RedisAdapter) GetClient() redis.Cmdable {
	return r.client
}

// GetMode 获取连接模式
func (r *RedisAdapter) GetMode() string {
	return r.mode
}

// Subscribe 订阅频道（用于Sub操作）
func (r *RedisAdapter) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	if r.clusterClient != nil {
		return r.clusterClient.Subscribe(ctx, channels...)
	}
	if r.standaloneClient != nil {
		return r.standaloneClient.Subscribe(ctx, channels...)
	}
	return nil
}

// GetMetricsCollector 获取指标收集器
func (r *RedisAdapter) GetMetricsCollector() interfaces.MetricsCollector {
	return r.BaseAdapter.GetMetricsCollector()
}
