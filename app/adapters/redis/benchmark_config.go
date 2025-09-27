package redis

import (
	"time"

	"abc-runner/app/core/execution"
	"abc-runner/app/core/interfaces"
)

// BenchmarkConfigAdapter Redis基准配置适配器
type BenchmarkConfigAdapter struct {
	config interfaces.BenchmarkConfig
}

// NewBenchmarkConfigAdapter 创建Redis基准配置适配器
func NewBenchmarkConfigAdapter(config interfaces.BenchmarkConfig) execution.BenchmarkConfig {
	return &BenchmarkConfigAdapter{config: config}
}

func (r *BenchmarkConfigAdapter) GetTotal() int {
	return r.config.GetTotal()
}

func (r *BenchmarkConfigAdapter) GetParallels() int {
	return r.config.GetParallels()
}

func (r *BenchmarkConfigAdapter) GetDuration() time.Duration {
	// Redis配置中没有Duration字段，返回0表示使用Total模式
	return 0
}

func (r *BenchmarkConfigAdapter) GetTimeout() time.Duration {
	// Redis配置中没有直接的Timeout，使用默认值
	return 30 * time.Second
}

func (r *BenchmarkConfigAdapter) GetRampUp() time.Duration {
	// Redis配置中没有RampUp字段，返回0表示不使用渐进加载
	return 0
}
