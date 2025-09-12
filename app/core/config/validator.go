package config

import (
	"fmt"

	"abc-runner/app/core/interfaces"
)

// ValidationRule 验证规则
type ValidationRule func(interfaces.Config) error

// ConfigValidator 配置验证器
type ConfigValidator struct {
	rules []ValidationRule
}

// NewConfigValidator 创建配置验证器
func NewConfigValidator() *ConfigValidator {
	validator := &ConfigValidator{
		rules: make([]ValidationRule, 0),
	}

	// 添加默认验证规则
	validator.AddRule(validateProtocol)

	return validator
}

// AddRule 添加验证规则
func (v *ConfigValidator) AddRule(rule ValidationRule) {
	v.rules = append(v.rules, rule)
}

// Validate 验证配置
func (v *ConfigValidator) Validate(config interfaces.Config) error {
	for _, rule := range v.rules {
		if err := rule(config); err != nil {
			return err
		}
	}
	return nil
}

// validateProtocol 验证协议
func validateProtocol(config interfaces.Config) error {
	protocol := config.GetProtocol()
	if protocol == "" {
		return fmt.Errorf("protocol cannot be empty")
	}

	// 支持的协议列表
	supportedProtocols := []string{"redis", "http", "kafka"}
	for _, supported := range supportedProtocols {
		if protocol == supported {
			return nil
		}
	}

	return fmt.Errorf("unsupported protocol: %s", protocol)
}