package config

import (
	"fmt"

	"redis-runner/app/core/interfaces"
	redisconfig "redis-runner/app/adapters/redis/config"
)

// 简化版本的配置源实现，用于支持新架构

// NewRedisCommandLineConfigSource 创建Redis命令行配置源
func NewRedisCommandLineConfigSource(args []string) ConfigSource {
	return &RedisConfigSourceBridge{
		source: redisconfig.NewCommandLineConfigSource(args),
	}
}

// NewRedisEnvironmentConfigSource 创建Redis环境变量配置源
func NewRedisEnvironmentConfigSource(prefix string) ConfigSource {
	return &RedisConfigSourceBridge{
		source: redisconfig.NewEnvConfigSource(prefix),
	}
}

// NewRedisYAMLConfigSource 创建Redis YAML配置源
func NewRedisYAMLConfigSource(filePath string) ConfigSource {
	return &RedisConfigSourceBridge{
		source: redisconfig.NewYAMLConfigSource(filePath),
	}
}

// NewHttpCommandLineConfigSource 创建HTTP命令行配置源
func NewHttpCommandLineConfigSource(args []string) ConfigSource {
	return &DummyConfigSource{
		protocol: "http",
		args:     args,
		priority: 3,
	}
}

// NewHttpEnvironmentConfigSource 创建HTTP环境变量配置源
func NewHttpEnvironmentConfigSource(prefix string) ConfigSource {
	return &DummyConfigSource{
		protocol: "http",
		prefix:   prefix,
		priority: 2,
	}
}

// NewHttpYAMLConfigSource 创建HTTP YAML配置源
func NewHttpYAMLConfigSource(filePath string) ConfigSource {
	return &DummyConfigSource{
		protocol: "http",
		filePath: filePath,
		priority: 1,
	}
}

// NewKafkaCommandLineConfigSource 创建Kafka命令行配置源
func NewKafkaCommandLineConfigSource(args []string) ConfigSource {
	return &DummyConfigSource{
		protocol: "kafka",
		args:     args,
		priority: 3,
	}
}

// NewKafkaEnvironmentConfigSource 创建Kafka环境变量配置源
func NewKafkaEnvironmentConfigSource(prefix string) ConfigSource {
	return &DummyConfigSource{
		protocol: "kafka",
		prefix:   prefix,
		priority: 2,
	}
}

// NewKafkaYAMLConfigSource 创建Kafka YAML配置源
func NewKafkaYAMLConfigSource(filePath string) ConfigSource {
	return &DummyConfigSource{
		protocol: "kafka",
		filePath: filePath,
		priority: 1,
	}
}

// RedisConfigSourceBridge Redis配置源桥接器
type RedisConfigSourceBridge struct {
	source redisconfig.ConfigSource
}

// Priority 获取优先级
func (r *RedisConfigSourceBridge) Priority() int {
	return r.source.Priority()
}

// CanLoad 检查是否可以加载
func (r *RedisConfigSourceBridge) CanLoad() bool {
	return r.source.CanLoad()
}

// Load 加载配置并适配为统一接口
func (r *RedisConfigSourceBridge) Load() (interfaces.Config, error) {
	redisConfig, err := r.source.Load()
	if err != nil {
		return nil, err
	}
	return redisconfig.NewRedisConfigAdapter(redisConfig), nil
}

// DummyConfigSource 临时的配置源实现（用于其他协议）
type DummyConfigSource struct {
	protocol string
	args     []string
	prefix   string
	filePath string
	priority int
}

// Priority 获取优先级
func (d *DummyConfigSource) Priority() int {
	return d.priority
}

// CanLoad 检查是否可以加载
func (d *DummyConfigSource) CanLoad() bool {
	// 简化实现，总是返回true
	return true
}

// Load 加载配置
func (d *DummyConfigSource) Load() (interfaces.Config, error) {
	// 简化实现，返回错误提示暂未实现
	return nil, fmt.Errorf("config source for %s not fully implemented yet", d.protocol)
}

// ConfigSource 配置源接口
type ConfigSource interface {
	Priority() int
	CanLoad() bool
	Load() (interfaces.Config, error)
}

// redisconfig.ConfigSource 接口定义
// 为了类型安全，我们需要确保Redis适配器中的ConfigSource接口与这里兼容