package http

import (
	"time"

	"abc-runner/app/adapters/http/config"
	"abc-runner/app/core/execution"
)

// BenchmarkConfigAdapter HTTP基准配置适配器
type BenchmarkConfigAdapter struct {
	config *config.HttpBenchmarkConfig
}

// NewBenchmarkConfigAdapter 创建HTTP基准配置适配器
func NewBenchmarkConfigAdapter(config *config.HttpBenchmarkConfig) execution.BenchmarkConfig {
	return &BenchmarkConfigAdapter{config: config}
}

func (h *BenchmarkConfigAdapter) GetTotal() int {
	if h.config.Total <= 0 {
		return 1000
	}
	return h.config.Total
}

func (h *BenchmarkConfigAdapter) GetParallels() int {
	if h.config.Parallels <= 0 {
		return 10
	}
	return h.config.Parallels
}

func (h *BenchmarkConfigAdapter) GetDuration() time.Duration {
	return h.config.Duration
}

func (h *BenchmarkConfigAdapter) GetTimeout() time.Duration {
	if h.config.Timeout <= 0 {
		return 30 * time.Second
	}
	return h.config.Timeout
}

func (h *BenchmarkConfigAdapter) GetRampUp() time.Duration {
	return h.config.RampUp
}
