package execution

import (
	"time"

	httpConfig "abc-runner/app/adapters/http/config"
	kafkaConfig "abc-runner/app/adapters/kafka/config"
	"abc-runner/app/core/interfaces"
)

// HttpBenchmarkConfigAdapter HTTP基准配置适配器
type HttpBenchmarkConfigAdapter struct {
	config *httpConfig.HttpBenchmarkConfig
}

// NewHttpBenchmarkConfigAdapter 创建HTTP基准配置适配器
func NewHttpBenchmarkConfigAdapter(config *httpConfig.HttpBenchmarkConfig) BenchmarkConfig {
	return &HttpBenchmarkConfigAdapter{config: config}
}

func (h *HttpBenchmarkConfigAdapter) GetTotal() int {
	if h.config.Total <= 0 {
		return 1000
	}
	return h.config.Total
}

func (h *HttpBenchmarkConfigAdapter) GetParallels() int {
	if h.config.Parallels <= 0 {
		return 10
	}
	return h.config.Parallels
}

func (h *HttpBenchmarkConfigAdapter) GetDuration() time.Duration {
	return h.config.Duration
}

func (h *HttpBenchmarkConfigAdapter) GetTimeout() time.Duration {
	if h.config.Timeout <= 0 {
		return 30 * time.Second
	}
	return h.config.Timeout
}

func (h *HttpBenchmarkConfigAdapter) GetRampUp() time.Duration {
	return h.config.RampUp
}

// RedisBenchmarkConfigAdapter Redis基准配置适配器
type RedisBenchmarkConfigAdapter struct {
	config interfaces.BenchmarkConfig
}

// NewRedisBenchmarkConfigAdapter 创建Redis基准配置适配器
func NewRedisBenchmarkConfigAdapter(config interfaces.BenchmarkConfig) BenchmarkConfig {
	return &RedisBenchmarkConfigAdapter{config: config}
}

func (r *RedisBenchmarkConfigAdapter) GetTotal() int {
	return r.config.GetTotal()
}

func (r *RedisBenchmarkConfigAdapter) GetParallels() int {
	return r.config.GetParallels()
}

func (r *RedisBenchmarkConfigAdapter) GetDuration() time.Duration {
	// Redis配置中没有Duration字段，返回0表示使用Total模式
	return 0
}

func (r *RedisBenchmarkConfigAdapter) GetTimeout() time.Duration {
	// Redis配置中没有直接的Timeout，使用默认值
	return 30 * time.Second
}

func (r *RedisBenchmarkConfigAdapter) GetRampUp() time.Duration {
	// Redis配置中没有RampUp字段，返回0表示不使用渐进加载
	return 0
}

// KafkaBenchmarkConfigAdapter Kafka基准配置适配器
type KafkaBenchmarkConfigAdapter struct {
	config *kafkaConfig.KafkaBenchmarkConfig
}

// NewKafkaBenchmarkConfigAdapter 创建Kafka基准配置适配器
func NewKafkaBenchmarkConfigAdapter(config *kafkaConfig.KafkaBenchmarkConfig) BenchmarkConfig {
	return &KafkaBenchmarkConfigAdapter{config: config}
}

func (k *KafkaBenchmarkConfigAdapter) GetTotal() int {
	return k.config.GetTotal()
}

func (k *KafkaBenchmarkConfigAdapter) GetParallels() int {
	return k.config.GetParallels()
}

func (k *KafkaBenchmarkConfigAdapter) GetDuration() time.Duration {
	// Kafka配置中没有Duration字段，返回0表示使用Total模式
	return 0
}

func (k *KafkaBenchmarkConfigAdapter) GetTimeout() time.Duration {
	return k.config.GetTimeout()
}

func (k *KafkaBenchmarkConfigAdapter) GetRampUp() time.Duration {
	// Kafka配置中没有RampUp字段，返回0表示不使用渐进加载
	return 0
}