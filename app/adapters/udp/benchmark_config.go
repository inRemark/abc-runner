package udp

import (
	"abc-runner/app/core/execution"
	"time"
)

// SimpleBenchmarkConfig 简单基准测试配置
type SimpleBenchmarkConfig struct {
	total     int
	parallels int
	duration  time.Duration
	timeout   time.Duration
	rampUp    time.Duration
}

// NewSimpleBenchmarkConfig 创建简单基准测试配置
func NewSimpleBenchmarkConfig(total, parallels int, duration time.Duration) *SimpleBenchmarkConfig {
	return &SimpleBenchmarkConfig{
		total:     total,
		parallels: parallels,
		duration:  duration,
		timeout:   30 * time.Second,
		rampUp:    0,
	}
}

// GetTotal 获取总操作数
func (c *SimpleBenchmarkConfig) GetTotal() int {
	return c.total
}

// GetParallels 获取并发数
func (c *SimpleBenchmarkConfig) GetParallels() int {
	return c.parallels
}

// GetDuration 获取测试持续时间
func (c *SimpleBenchmarkConfig) GetDuration() time.Duration {
	return c.duration
}

// GetTimeout 获取操作超时时间
func (c *SimpleBenchmarkConfig) GetTimeout() time.Duration {
	return c.timeout
}

// GetRampUp 获取渐进加载时间
func (c *SimpleBenchmarkConfig) GetRampUp() time.Duration {
	return c.rampUp
}

// 确保实现了接口
var _ execution.BenchmarkConfig = (*SimpleBenchmarkConfig)(nil)
