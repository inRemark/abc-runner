package connection

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"redis-runner/app/adapters/redis/config"
)

// ClientManager Redis客户端管理器
type ClientManager struct {
	config      *config.RedisConfig
	standalone  *redis.Client
	cluster     *redis.ClusterClient
	currentMode string
	mutex       sync.RWMutex
	ctx         context.Context
}

// ClientFactory Redis客户端工厂
type ClientFactory struct {
	ctx context.Context
}

// RedisClient Redis客户端接口
type RedisClient interface {
	GetStandaloneClient() *redis.Client
	GetClusterClient() *redis.ClusterClient
	GetCmdable() redis.Cmdable
	Close() error
	Ping() error
	GetMode() string
	IsCluster() bool
}

// NewClientFactory 创建客户端工厂
func NewClientFactory() *ClientFactory {
	return &ClientFactory{
		ctx: context.Background(),
	}
}

// CreateStandaloneClient 创建单机客户端
func (f *ClientFactory) CreateStandaloneClient(config config.StandAloneInfo) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         config.Addr,
		DB:           config.Db,
		Password:     config.Password,
		PoolSize:     10, // 默认值，可以从配置中获取
		MinIdleConns: 2,  // 默认值，可以从配置中获取
	})

	// 测试连接
	_, err := client.Ping(f.ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("standalone client connection failed: %w", err)
	}

	return client, nil
}

// CreateSentinelClient 创建哨兵客户端
func (f *ClientFactory) CreateSentinelClient(config config.SentinelInfo) (*redis.Client, error) {
	client := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    config.MasterName,
		SentinelAddrs: config.Addrs,
		Password:      config.Password,
		DB:            config.Db,
		PoolSize:      10, // 默认值，可以从配置中获取
		MinIdleConns:  2,  // 默认值，可以从配置中获取
	})

	// 测试连接
	_, err := client.Ping(f.ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("sentinel client connection failed: %w", err)
	}

	return client, nil
}

// CreateClusterClient 创建集群客户端
func (f *ClientFactory) CreateClusterClient(config config.ClusterInfo) (*redis.ClusterClient, error) {
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        config.Addrs,
		Password:     config.Password,
		PoolSize:     10, // 默认值，可以从配置中获取
		MinIdleConns: 2,  // 默认值，可以从配置中获取
	})

	// 测试连接
	_, err := client.Ping(f.ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("cluster client connection failed: %w", err)
	}

	return client, nil
}

// NewClientManager 创建客户端管理器
func NewClientManager(cfg *config.RedisConfig) *ClientManager {
	return &ClientManager{
		config: cfg,
		ctx:    context.Background(),
	}
}

// Connect 连接Redis
func (cm *ClientManager) Connect() error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	factory := NewClientFactory()
	mode := cm.config.GetMode()

	switch mode {
	case "cluster":
		cluster, err := factory.CreateClusterClient(cm.config.GetClusterConfig())
		if err != nil {
			return fmt.Errorf("failed to create cluster client: %w", err)
		}
		cm.cluster = cluster
		cm.currentMode = "cluster"

	case "sentinel":
		standalone, err := factory.CreateSentinelClient(cm.config.GetSentinelConfig())
		if err != nil {
			return fmt.Errorf("failed to create sentinel client: %w", err)
		}
		cm.standalone = standalone
		cm.currentMode = "sentinel"

	default: // standalone
		standalone, err := factory.CreateStandaloneClient(cm.config.GetStandaloneConfig())
		if err != nil {
			return fmt.Errorf("failed to create standalone client: %w", err)
		}
		cm.standalone = standalone
		cm.currentMode = "standalone"
	}

	log.Printf("Redis client connected successfully in %s mode", mode)
	return nil
}

// GetStandaloneClient 获取单机/哨兵客户端
func (cm *ClientManager) GetStandaloneClient() *redis.Client {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return cm.standalone
}

// GetClusterClient 获取集群客户端
func (cm *ClientManager) GetClusterClient() *redis.ClusterClient {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return cm.cluster
}

// GetCmdable 获取通用命令接口
func (cm *ClientManager) GetCmdable() redis.Cmdable {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	if cm.currentMode == "cluster" {
		return cm.cluster
	}
	return cm.standalone
}

// IsCluster 检查是否为集群模式
func (cm *ClientManager) IsCluster() bool {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return cm.currentMode == "cluster"
}

// GetMode 获取当前模式
func (cm *ClientManager) GetMode() string {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return cm.currentMode
}

// Ping 测试连接
func (cm *ClientManager) Ping() error {
	cmdable := cm.GetCmdable()
	if cmdable == nil {
		return fmt.Errorf("no redis client available")
	}

	_, err := cmdable.Ping(cm.ctx).Result()
	if err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	return nil
}

// Close 关闭连接
func (cm *ClientManager) Close() error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	var errors []error

	if cm.standalone != nil {
		if err := cm.standalone.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close standalone client: %w", err))
		}
		cm.standalone = nil
	}

	if cm.cluster != nil {
		if err := cm.cluster.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close cluster client: %w", err))
		}
		cm.cluster = nil
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors closing clients: %v", errors)
	}

	log.Printf("Redis clients closed successfully")
	return nil
}

// HealthCheck 健康检查
func (cm *ClientManager) HealthCheck() (bool, error) {
	err := cm.Ping()
	if err != nil {
		return false, err
	}
	return true, nil
}

// Reconnect 重新连接
func (cm *ClientManager) Reconnect() error {
	log.Printf("Attempting to reconnect Redis client...")

	// 关闭现有连接
	if err := cm.Close(); err != nil {
		log.Printf("Warning: failed to close existing connections: %v", err)
	}

	// 重新连接
	return cm.Connect()
}

// GetConnectionInfo 获取连接信息
func (cm *ClientManager) GetConnectionInfo() map[string]interface{} {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	info := map[string]interface{}{
		"mode":      cm.currentMode,
		"connected": cm.standalone != nil || cm.cluster != nil,
	}

	switch cm.currentMode {
	case "cluster":
		if cm.cluster != nil {
			info["addresses"] = cm.config.GetClusterConfig().Addrs
		}
	case "sentinel":
		if cm.standalone != nil {
			sentinelConfig := cm.config.GetSentinelConfig()
			info["master_name"] = sentinelConfig.MasterName
			info["sentinel_addresses"] = sentinelConfig.Addrs
			info["database"] = sentinelConfig.Db
		}
	default: // standalone
		if cm.standalone != nil {
			standaloneConfig := cm.config.GetStandaloneConfig()
			info["address"] = standaloneConfig.Addr
			info["database"] = standaloneConfig.Db
		}
	}

	return info
}

// BasicOperations 基本操作接口
type BasicOperations struct {
	manager *ClientManager
	ctx     context.Context
}

// NewBasicOperations 创建基本操作实例
func NewBasicOperations(manager *ClientManager) *BasicOperations {
	return &BasicOperations{
		manager: manager,
		ctx:     context.Background(),
	}
}

// Get 获取键值
func (ops *BasicOperations) Get(key string) (string, error) {
	cmdable := ops.manager.GetCmdable()
	if cmdable == nil {
		return "", fmt.Errorf("no redis client available")
	}

	val, err := cmdable.Get(ops.ctx, key).Result()
	if err == redis.Nil {
		return "", nil // Key does not exist
	}
	return val, err
}

// Set 设置键值
func (ops *BasicOperations) Set(key string, value interface{}, duration time.Duration) error {
	cmdable := ops.manager.GetCmdable()
	if cmdable == nil {
		return fmt.Errorf("no redis client available")
	}

	return cmdable.Set(ops.ctx, key, value, duration).Err()
}

// Del 删除键
func (ops *BasicOperations) Del(key string) error {
	cmdable := ops.manager.GetCmdable()
	if cmdable == nil {
		return fmt.Errorf("no redis client available")
	}

	return cmdable.Del(ops.ctx, key).Err()
}

// HSet 设置哈希字段
func (ops *BasicOperations) HSet(key, field string, value interface{}) error {
	cmdable := ops.manager.GetCmdable()
	if cmdable == nil {
		return fmt.Errorf("no redis client available")
	}

	return cmdable.HSet(ops.ctx, key, field, value).Err()
}

// HSetNX 仅当字段不存在时设置哈希字段
func (ops *BasicOperations) HSetNX(key, field string, value interface{}) (bool, error) {
	cmdable := ops.manager.GetCmdable()
	if cmdable == nil {
		return false, fmt.Errorf("no redis client available")
	}

	return cmdable.HSetNX(ops.ctx, key, field, value).Result()
}

// HGet 获取哈希字段值
func (ops *BasicOperations) HGet(key, field string) (string, error) {
	cmdable := ops.manager.GetCmdable()
	if cmdable == nil {
		return "", fmt.Errorf("no redis client available")
	}

	val, err := cmdable.HGet(ops.ctx, key, field).Result()
	if err == redis.Nil {
		return "", nil // Field does not exist
	}
	return val, err
}

// Publish 发布消息
func (ops *BasicOperations) Publish(channel string, message interface{}) error {
	cmdable := ops.manager.GetCmdable()
	if cmdable == nil {
		return fmt.Errorf("no redis client available")
	}

	return cmdable.Publish(ops.ctx, channel, message).Err()
}

// Subscribe 订阅频道
func (ops *BasicOperations) Subscribe(channel string, handler func(msg *redis.Message)) error {
	var pubsub *redis.PubSub

	if ops.manager.IsCluster() {
		cluster := ops.manager.GetClusterClient()
		if cluster == nil {
			return fmt.Errorf("cluster client not available")
		}
		pubsub = cluster.Subscribe(ops.ctx, channel)
	} else {
		standalone := ops.manager.GetStandaloneClient()
		if standalone == nil {
			return fmt.Errorf("standalone client not available")
		}
		pubsub = standalone.Subscribe(ops.ctx, channel)
	}

	defer pubsub.Close()

	// 处理消息
	for msg := range pubsub.Channel() {
		handler(msg)
	}

	return nil
}