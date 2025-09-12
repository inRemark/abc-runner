package unified

import (
	"fmt"

	"abc-runner/app/core/interfaces"
	"gopkg.in/yaml.v3"
)

// RedisConfigParser Redis配置解析器
type RedisConfigParser struct {
	configFactory func() interfaces.Config
}

// NewRedisConfigParser 创建Redis配置解析器
func NewRedisConfigParser(configFactory func() interfaces.Config) *RedisConfigParser {
	return &RedisConfigParser{configFactory: configFactory}
}

// Parse 解析Redis配置
func (r *RedisConfigParser) Parse(data []byte) (interfaces.Config, error) {
	config := r.configFactory()
	
	var configWrapper struct {
		Redis interface{} `yaml:"redis"`
	}
	
	if err := yaml.Unmarshal(data, &configWrapper); err != nil {
		return nil, fmt.Errorf("failed to parse redis config: %w", err)
	}
	
	// Convert the parsed data to the config structure
	// This would need to be implemented based on the specific config structure
	// For now, we'll just return the default config
	return config, nil
}

// HttpConfigParser HTTP配置解析器
type HttpConfigParser struct {
	configFactory func() interfaces.Config
}

// NewHttpConfigParser 创建HTTP配置解析器
func NewHttpConfigParser(configFactory func() interfaces.Config) *HttpConfigParser {
	return &HttpConfigParser{configFactory: configFactory}
}

// Parse 解析HTTP配置
func (h *HttpConfigParser) Parse(data []byte) (interfaces.Config, error) {
	config := h.configFactory()
	
	var configWrapper struct {
		HTTP interface{} `yaml:"http"`
	}
	
	if err := yaml.Unmarshal(data, &configWrapper); err != nil {
		return nil, fmt.Errorf("failed to parse http config: %w", err)
	}
	
	// Convert the parsed data to the config structure
	// This would need to be implemented based on the specific config structure
	// For now, we'll just return the default config
	return config, nil
}

// KafkaConfigParser Kafka配置解析器
type KafkaConfigParser struct {
	configFactory func() interfaces.Config
}

// NewKafkaConfigParser 创建Kafka配置解析器
func NewKafkaConfigParser(configFactory func() interfaces.Config) *KafkaConfigParser {
	return &KafkaConfigParser{configFactory: configFactory}
}

// Parse 解析Kafka配置
func (k *KafkaConfigParser) Parse(data []byte) (interfaces.Config, error) {
	config := k.configFactory()
	
	var configWrapper struct {
		Kafka interface{} `yaml:"kafka"`
	}
	
	if err := yaml.Unmarshal(data, &configWrapper); err != nil {
		return nil, fmt.Errorf("failed to parse kafka config: %w", err)
	}
	
	// Convert the parsed data to the config structure
	// This would need to be implemented based on the specific config structure
	// For now, we'll just return the default config
	return config, nil
}