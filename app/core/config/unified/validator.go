package unified

import (
	"abc-runner/app/core/interfaces"
)

// SimpleConfigValidator 简单配置验证器
type SimpleConfigValidator struct{}

// NewSimpleConfigValidator 创建简单配置验证器
func NewSimpleConfigValidator() *SimpleConfigValidator {
	return &SimpleConfigValidator{}
}

// Validate 验证配置
func (s *SimpleConfigValidator) Validate(config interfaces.Config) error {
	if config == nil {
		return nil // Allow nil config for now
	}

	// Delegate to the config's own validation method
	return config.Validate()
}
