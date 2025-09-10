package config

import (
	"log"

	httpconfig "abc-runner/app/adapters/http/config"
	kafkaconfig "abc-runner/app/adapters/kafka/config"
	redisconfig "abc-runner/app/adapters/redis/config"
	"abc-runner/app/core/interfaces"
)

// ConfigSource 配置源接口
type ConfigSource interface {
	Priority() int
	CanLoad() bool
	Load() (interfaces.Config, error)
}

// CreateRedisConfigSources 创建Redis配置源列表
func CreateRedisConfigSources(yamlFile string, args []string) []ConfigSource {
	// 使用Redis包中的函数创建配置源
	redisSources := redisconfig.CreateRedisConfigSources(yamlFile, args)

	// 转换为core包中的ConfigSource接口
	sources := make([]ConfigSource, len(redisSources))
	for i, source := range redisSources {
		sources[i] = source
	}

	return sources
}

// CreateHttpConfigSources 创建HTTP配置源列表
func CreateHttpConfigSources(yamlFile string, args []string) []ConfigSource {
	// 使用HTTP包中的函数创建配置源
	httpSources := httpconfig.CreateHttpConfigSources(yamlFile, args)

	// 转换为core包中的ConfigSource接口
	sources := make([]ConfigSource, len(httpSources))
	for i, source := range httpSources {
		sources[i] = &HttpConfigSourceAdapter{source: source}
	}

	return sources
}

// CreateKafkaConfigSources 创建Kafka配置源列表
func CreateKafkaConfigSources(yamlFile string, args []string) []ConfigSource {
	// 使用Kafka包中的函数创建配置源
	kafkaSources := kafkaconfig.CreateKafkaConfigSources(yamlFile, args)

	// 转换为core包中的ConfigSource接口
	sources := make([]ConfigSource, len(kafkaSources))
	for i, source := range kafkaSources {
		sources[i] = &KafkaConfigSourceAdapter{source: source}
	}

	return sources
}

// CreateConfigSourcesWithCore 创建包含核心配置的配置源列表
func CreateConfigSourcesWithCore(coreConfigPath string, protocolSources []ConfigSource) ([]ConfigSource, *CoreConfig) {
	// 加载核心配置
	coreSource := NewCoreConfigSource(coreConfigPath)
	coreConfig, err := coreSource.LoadCoreConfig()
	if err != nil {
		log.Printf("Warning: Failed to load core config: %v", err)
		// 使用默认核心配置
		loader := NewCoreConfigLoader()
		coreConfig = loader.GetDefaultConfig()
	}

	// 在协议配置源前面添加核心配置源
	allSources := append([]ConfigSource{coreSource}, protocolSources...)

	return allSources, coreConfig
}



// HttpConfigSourceAdapter HTTP配置源适配器
type HttpConfigSourceAdapter struct {
	source httpconfig.HttpConfigSource
}

// Priority 获取优先级
func (h *HttpConfigSourceAdapter) Priority() int {
	return h.source.Priority()
}

// CanLoad 检查是否可以加载
func (h *HttpConfigSourceAdapter) CanLoad() bool {
	return h.source.CanLoad()
}

// Load 加载配置
func (h *HttpConfigSourceAdapter) Load() (interfaces.Config, error) {
	httpConfig, err := h.source.Load()
	if err != nil {
		return nil, err
	}
	return httpConfig, nil
}

// KafkaConfigSourceAdapter Kafka配置源适配器
type KafkaConfigSourceAdapter struct {
	source kafkaconfig.KafkaConfigSource
}

// Priority 获取优先级
func (k *KafkaConfigSourceAdapter) Priority() int {
	return k.source.Priority()
}

// CanLoad 检查是否可以加载
func (k *KafkaConfigSourceAdapter) CanLoad() bool {
	return k.source.CanLoad()
}

// Load 加载配置
func (k *KafkaConfigSourceAdapter) Load() (interfaces.Config, error) {
	kafkaConfig, err := k.source.Load()
	if err != nil {
		return nil, err
	}
	return kafkaConfig, nil
}
