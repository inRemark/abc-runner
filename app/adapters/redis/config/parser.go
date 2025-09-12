package config

import (
	"abc-runner/app/core/interfaces"
	"gopkg.in/yaml.v3"
)

// RedisYAMLParser Redis YAML解析器
type RedisYAMLParser struct {
	defaultConfig *RedisConfig
}

// NewRedisYAMLParser 创建Redis YAML解析器
func NewRedisYAMLParser(defaultConfig *RedisConfig) *RedisYAMLParser {
	return &RedisYAMLParser{defaultConfig: defaultConfig}
}

// Parse 解析Redis YAML配置
func (r *RedisYAMLParser) Parse(data []byte) (interfaces.Config, error) {
	var configWrapper struct {
		Redis *RedisConfig `yaml:"redis"`
	}

	if err := yaml.Unmarshal(data, &configWrapper); err != nil {
		return nil, err
	}

	if configWrapper.Redis == nil {
		// If no redis section, try to parse the whole file as redis config
		config := &RedisConfig{}
		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, err
		}
		return config, nil
	}

	return configWrapper.Redis, nil
}