package config

import (
	"time"

	"abc-runner/app/core/execution"
	"abc-runner/app/core/interfaces"
)

// BenchmarkConfigAdapter TCP基准测试配置适配器
type BenchmarkConfigAdapter struct {
	config interfaces.BenchmarkConfig
}

// NewBenchmarkConfigAdapter 创建TCP基准测试配置适配器
func NewBenchmarkConfigAdapter(config interfaces.BenchmarkConfig) *BenchmarkConfigAdapter {
	return &BenchmarkConfigAdapter{
		config: config,
	}
}

// GetTotal 获取总操作数
func (b *BenchmarkConfigAdapter) GetTotal() int {
	return b.config.GetTotal()
}

// GetParallels 获取并发数
func (b *BenchmarkConfigAdapter) GetParallels() int {
	return b.config.GetParallels()
}

// GetDuration 获取测试持续时间
func (b *BenchmarkConfigAdapter) GetDuration() time.Duration {
	// 尝试从原始配置获取持续时间
	if tcpBenchConfig, ok := b.config.(*BenchmarkConfig); ok {
		return tcpBenchConfig.Duration
	}
	// 默认持续时间
	return 60 * time.Second
}

// GetTimeout 获取操作超时时间
func (b *BenchmarkConfigAdapter) GetTimeout() time.Duration {
	// TCP特定的超时时间，通常设置为30秒
	return 30 * time.Second
}

// GetRampUp 获取渐进加载时间
func (b *BenchmarkConfigAdapter) GetRampUp() time.Duration {
	// TCP协议建议启用渐进加载以减少服务器冲击
	return 5 * time.Second
}

// GetTestCase 获取测试用例类型
func (b *BenchmarkConfigAdapter) GetTestCase() string {
	return b.config.GetTestCase()
}

// GetDataSize 获取数据包大小
func (b *BenchmarkConfigAdapter) GetDataSize() int {
	return b.config.GetDataSize()
}

// GetReadPercent 获取读操作百分比
func (b *BenchmarkConfigAdapter) GetReadPercent() int {
	return b.config.GetReadPercent()
}

// 确保实现了execution.BenchmarkConfig接口
var _ execution.BenchmarkConfig = (*BenchmarkConfigAdapter)(nil)
