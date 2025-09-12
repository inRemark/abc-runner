package config

import (
	"abc-runner/app/core/interfaces"
	"gopkg.in/yaml.v3"
)

// HttpYAMLParser HTTP YAML解析器
type HttpYAMLParser struct {
	defaultConfig *HttpAdapterConfig
}

// NewHttpYAMLParser 创建HTTP YAML解析器
func NewHttpYAMLParser(defaultConfig *HttpAdapterConfig) *HttpYAMLParser {
	return &HttpYAMLParser{defaultConfig: defaultConfig}
}

// Parse 解析HTTP YAML配置
func (h *HttpYAMLParser) Parse(data []byte) (interfaces.Config, error) {
	var configWrapper struct {
		HTTP *HttpAdapterConfig `yaml:"http"`
	}

	if err := yaml.Unmarshal(data, &configWrapper); err != nil {
		return nil, err
	}

	if configWrapper.HTTP == nil {
		// If no http section, try to parse the whole file as http config
		config := &HttpAdapterConfig{}
		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, err
		}
		return config, nil
	}

	return configWrapper.HTTP, nil
}