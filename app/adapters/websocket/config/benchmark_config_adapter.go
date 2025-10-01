package config

import (
	"time"

	"abc-runner/app/core/execution"
)

// BenchmarkConfigAdapter WebSocket基准配置适配器
type BenchmarkConfigAdapter struct {
	config *BenchmarkConfig
}

// NewBenchmarkConfigAdapter 创建WebSocket基准配置适配器
func NewBenchmarkConfigAdapter(config *BenchmarkConfig) execution.BenchmarkConfig {
	return &BenchmarkConfigAdapter{
		config: config,
	}
}

// GetTotal 获取总操作数
func (b *BenchmarkConfigAdapter) GetTotal() int {
	return b.config.Total
}

// GetParallels 获取并发数
func (b *BenchmarkConfigAdapter) GetParallels() int {
	return b.config.Parallels
}

// GetDuration 获取测试持续时间
func (b *BenchmarkConfigAdapter) GetDuration() time.Duration {
	return b.config.Duration
}

// GetTimeout 获取操作超时时间
func (b *BenchmarkConfigAdapter) GetTimeout() time.Duration {
	return b.config.TTL
}

// GetRampUp 获取渐进加载时间
func (b *BenchmarkConfigAdapter) GetRampUp() time.Duration {
	// WebSocket暂不支持渐进加载，返回0
	return 0
}
