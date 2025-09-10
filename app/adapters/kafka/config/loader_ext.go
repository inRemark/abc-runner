package config

import (
	"fmt"
	"sort"
)

// KafkaMultiSourceConfigLoader Kafka多源配置加载器
type KafkaMultiSourceConfigLoader struct {
	sources []KafkaConfigSource
}

// NewKafkaMultiSourceConfigLoader 创建Kafka多源配置加载器
func NewKafkaMultiSourceConfigLoader(sources ...KafkaConfigSource) *KafkaMultiSourceConfigLoader {
	return &KafkaMultiSourceConfigLoader{sources: sources}
}

// AddSource 添加配置源
func (m *KafkaMultiSourceConfigLoader) AddSource(source KafkaConfigSource) {
	m.sources = append(m.sources, source)
}

// GetSources 获取配置源列表
func (m *KafkaMultiSourceConfigLoader) GetSources() []KafkaConfigSource {
	return m.sources
}

// Load 从多个源加载配置，按优先级排序
func (m *KafkaMultiSourceConfigLoader) Load() (*KafkaAdapterConfig, error) {
	// 按优先级排序（高优先级在前）
	sources := make([]KafkaConfigSource, len(m.sources))
	copy(sources, m.sources)
	sort.Slice(sources, func(i, j int) bool {
		return sources[i].Priority() > sources[j].Priority()
	})

	// 尝试从每个源加载配置
	for _, source := range sources {
		if source.CanLoad() {
			config, err := source.Load()
			if err != nil {
				continue
			}
			if config != nil {
				return config, nil
			}
		}
	}

	return nil, fmt.Errorf("no configuration source available")
}