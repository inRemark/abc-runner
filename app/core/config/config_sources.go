package config

import (
	"log"

	"redis-runner/app/core/interfaces"
	redisconfig "redis-runner/app/adapters/redis/config"
	httpconfig "redis-runner/app/adapters/http/config"
	kafkaconfig "redis-runner/app/adapters/kafka/config"
)

// CreateRedisConfigSources 创建Redis配置源列表
func CreateRedisConfigSources(yamlFile string, args []string) []ConfigSource {
	sources := []ConfigSource{}
	
	// YAML文件配置（最高优先级）
	if yamlFile != "" {
		sources = append(sources, NewRedisYAMLConfigSource(yamlFile))
	}
	
	// 环境变量配置
	sources = append(sources, NewRedisEnvironmentConfigSource("REDIS_RUNNER"))
	
	// 命令行参数配置（最低优先级）
	if args != nil {
		sources = append(sources, NewRedisCommandLineConfigSource(args))
	}
	
	return sources
}

// CreateHttpConfigSources 创建HTTP配置源列表
func CreateHttpConfigSources(yamlFile string, args []string) []ConfigSource {
	sources := []ConfigSource{}
	
	// YAML文件配置（最高优先级）
	if yamlFile != "" {
		sources = append(sources, NewHttpYAMLConfigSource(yamlFile))
	}
	
	// 环境变量配置
	sources = append(sources, NewHttpEnvironmentConfigSource("HTTP_RUNNER"))
	
	// 命令行参数配置（最低优先级）
	if args != nil {
		sources = append(sources, NewHttpCommandLineConfigSource(args))
	}
	
	return sources
}

// CreateKafkaConfigSources 创建Kafka配置源列表
func CreateKafkaConfigSources(yamlFile string, args []string) []ConfigSource {
	sources := []ConfigSource{}
	
	// YAML文件配置（最高优先级）
	if yamlFile != "" {
		sources = append(sources, NewKafkaYAMLConfigSource(yamlFile))
	}
	
	// 环境变量配置
	sources = append(sources, NewKafkaEnvironmentConfigSource("KAFKA_RUNNER"))
	
	// 命令行参数配置（最低优先级）
	if args != nil {
		sources = append(sources, NewKafkaCommandLineConfigSource(args))
	}
	
	return sources
}

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
	return &HttpConfigSourceImpl{
		args:     args,
		priority: 3,
	}
}

// NewHttpEnvironmentConfigSource 创建HTTP环境变量配置源
func NewHttpEnvironmentConfigSource(prefix string) ConfigSource {
	return &HttpConfigSourceImpl{
		prefix:   prefix,
		priority: 2,
	}
}

// NewHttpYAMLConfigSource 创建HTTP YAML配置源
func NewHttpYAMLConfigSource(filePath string) ConfigSource {
	return &HttpConfigSourceImpl{
		filePath: filePath,
		priority: 1,
	}
}

// NewKafkaCommandLineConfigSource 创建Kafka命令行配置源
func NewKafkaCommandLineConfigSource(args []string) ConfigSource {
	return &KafkaConfigSourceImpl{
		args:     args,
		priority: 3,
	}
}

// NewKafkaEnvironmentConfigSource 创建Kafka环境变量配置源
func NewKafkaEnvironmentConfigSource(prefix string) ConfigSource {
	return &KafkaConfigSourceImpl{
		prefix:   prefix,
		priority: 2,
	}
}

// NewKafkaYAMLConfigSource 创建Kafka YAML配置源
func NewKafkaYAMLConfigSource(filePath string) ConfigSource {
	return &KafkaConfigSourceImpl{
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

// HttpConfigSourceImpl HTTP配置源实现
type HttpConfigSourceImpl struct {
	args     []string
	prefix   string
	filePath string
	priority int
}

// Priority 获取优先级
func (h *HttpConfigSourceImpl) Priority() int {
	return h.priority
}

// CanLoad 检查是否可以加载
func (h *HttpConfigSourceImpl) CanLoad() bool {
	if h.filePath != "" {
		// 简化实现，暂时总是返回true
		return true
	}
	return true
}

// Load 加载配置
func (h *HttpConfigSourceImpl) Load() (interfaces.Config, error) {
	if h.filePath != "" {
		// TODO: 实现YAML文件加载
		log.Printf("HTTP YAML config loading not implemented yet: %s", h.filePath)
		return httpconfig.LoadDefaultHttpConfig(), nil
	}
	
	if h.args != nil {
		// TODO: 实现命令行参数解析
		log.Println("HTTP command line config parsing not implemented yet")
		return httpconfig.LoadDefaultHttpConfig(), nil
	}
	
	// 环境变量配置
	// TODO: 实现环境变量加载
	log.Println("HTTP environment config loading not implemented yet")
	return httpconfig.LoadDefaultHttpConfig(), nil
}

// KafkaConfigSourceImpl Kafka配置源实现
type KafkaConfigSourceImpl struct {
	args     []string
	prefix   string
	filePath string
	priority int
}

// Priority 获取优先级
func (k *KafkaConfigSourceImpl) Priority() int {
	return k.priority
}

// CanLoad 检查是否可以加载
func (k *KafkaConfigSourceImpl) CanLoad() bool {
	if k.filePath != "" {
		// 简化实现，暂时总是返回true
		return true
	}
	return true
}

// Load 加载配置
func (k *KafkaConfigSourceImpl) Load() (interfaces.Config, error) {
	if k.filePath != "" {
		// TODO: 实现YAML文件加载
		log.Printf("Kafka YAML config loading not implemented yet: %s", k.filePath)
		return kafkaconfig.LoadDefaultKafkaConfig(), nil
	}
	
	if k.args != nil {
		// TODO: 实现命令行参数解析
		log.Println("Kafka command line config parsing not implemented yet")
		return kafkaconfig.LoadDefaultKafkaConfig(), nil
	}
	
	// 环境变量配置
	// TODO: 实现环境变量加载
	log.Println("Kafka environment config loading not implemented yet")
	return kafkaconfig.LoadDefaultKafkaConfig(), nil
}

// ConfigSource 配置源接口
type ConfigSource interface {
	Priority() int
	CanLoad() bool
	Load() (interfaces.Config, error)
}

// redisconfig.ConfigSource 接口定义
// 为了类型安全，我们需要确保Redis适配器中的ConfigSource接口与这里兼容