package config

import (
	"time"

	"abc-runner/app/core/execution"
)

// BenchmarkConfigAdapter Kafka基准配置适配器
type BenchmarkConfigAdapter struct {
	config *KafkaBenchmarkConfig
}

// NewBenchmarkConfigAdapter 创建Kafka基准配置适配器
func NewBenchmarkConfigAdapter(config *KafkaBenchmarkConfig) execution.BenchmarkConfig {
	return &BenchmarkConfigAdapter{config: config}
}

func (k *BenchmarkConfigAdapter) GetTotal() int {
	return k.config.GetTotal()
}

func (k *BenchmarkConfigAdapter) GetParallels() int {
	return k.config.GetParallels()
}

func (k *BenchmarkConfigAdapter) GetDuration() time.Duration {
	// Kafka配置中没有Duration字段，返回0表示使用Total模式
	return 0
}

func (k *BenchmarkConfigAdapter) GetTimeout() time.Duration {
	return k.config.GetTimeout()
}

func (k *BenchmarkConfigAdapter) GetRampUp() time.Duration {
	// Kafka配置中没有RampUp字段，返回0表示不使用渐进加载
	return 0
}
