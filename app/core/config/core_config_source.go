package config

import (
	"abc-runner/app/core/interfaces"
)

// CoreConfigSource 核心配置源
type CoreConfigSource struct {
	FilePath string
	priority int
}

// NewCoreConfigSource 创建核心配置源
func NewCoreConfigSource(filePath string) *CoreConfigSource {
	return &CoreConfigSource{
		FilePath: filePath,
		priority: 10, // 核心配置优先级较高，但低于命令行参数
	}
}

// Priority 获取优先级
func (c *CoreConfigSource) Priority() int {
	return c.priority
}

// CanLoad 检查是否可以加载
func (c *CoreConfigSource) CanLoad() bool {
	// 核心配置是可选的，总是可以加载（即使文件不存在也会使用默认配置）
	return true
}

// Load 加载配置
func (c *CoreConfigSource) Load() (interfaces.Config, error) {
	// CoreConfigSource主要用于加载核心配置，而不是协议配置
	// 这里返回nil表示这不是一个协议配置源
	return nil, nil
}

// LoadCoreConfig 加载核心配置（专用方法）
func (c *CoreConfigSource) LoadCoreConfig() (*CoreConfig, error) {
	loader := NewUnifiedCoreConfigLoader()
	return loader.LoadFromFile(c.FilePath)
}
