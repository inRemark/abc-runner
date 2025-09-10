package config

import (
	"fmt"
	"sort"
)

// HttpMultiSourceConfigLoader HTTP多源配置加载器
type HttpMultiSourceConfigLoader struct {
	sources []HttpConfigSource
}

// NewHttpMultiSourceConfigLoader 创建HTTP多源配置加载器
func NewHttpMultiSourceConfigLoader(sources ...HttpConfigSource) *HttpMultiSourceConfigLoader {
	return &HttpMultiSourceConfigLoader{sources: sources}
}

// AddSource 添加配置源
func (m *HttpMultiSourceConfigLoader) AddSource(source HttpConfigSource) {
	m.sources = append(m.sources, source)
}

// GetSources 获取配置源列表
func (m *HttpMultiSourceConfigLoader) GetSources() []HttpConfigSource {
	return m.sources
}

// Load 从多个源加载配置，按优先级排序
func (m *HttpMultiSourceConfigLoader) Load() (*HttpAdapterConfig, error) {
	// 按优先级排序（高优先级在前）
	sources := make([]HttpConfigSource, len(m.sources))
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