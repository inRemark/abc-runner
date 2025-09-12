package config

import "abc-runner/app/core/interfaces"

// mockConfigSourceFactory 模拟配置源工厂
type mockConfigSourceFactory struct{}

func (m *mockConfigSourceFactory) CreateRedisConfigSource() interfaces.ConfigSource {
	return nil
}

func (m *mockConfigSourceFactory) CreateHttpConfigSource() interfaces.ConfigSource {
	return nil
}

func (m *mockConfigSourceFactory) CreateKafkaConfigSource() interfaces.ConfigSource {
	return nil
}

// Validate 实现unified.ConfigValidator接口
func (m *mockConfigSourceFactory) Validate(config interfaces.Config) error {
	// 模拟验证，总是返回nil表示验证通过
	return nil
}