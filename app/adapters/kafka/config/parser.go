package config

import (
	"abc-runner/app/core/interfaces"
	"gopkg.in/yaml.v2"
)

// KafkaYAMLParser Kafka YAML解析器
type KafkaYAMLParser struct {
	defaultConfig *KafkaAdapterConfig
}

// NewKafkaYAMLParser 创建Kafka YAML解析器
func NewKafkaYAMLParser(defaultConfig *KafkaAdapterConfig) *KafkaYAMLParser {
	return &KafkaYAMLParser{defaultConfig: defaultConfig}
}

// Parse 解析Kafka YAML配置
func (k *KafkaYAMLParser) Parse(data []byte) (interfaces.Config, error) {
	var configWrapper struct {
		Kafka *KafkaAdapterConfig `yaml:"kafka"`
	}

	if err := yaml.Unmarshal(data, &configWrapper); err != nil {
		return nil, err
	}

	if configWrapper.Kafka == nil {
		// If no kafka section, try to parse the whole file as kafka config
		config := &KafkaAdapterConfig{}
		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, err
		}
		return config, nil
	}

	return configWrapper.Kafka, nil
}