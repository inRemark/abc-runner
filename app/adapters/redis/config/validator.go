package config

import (
	"fmt"
	"strconv"
	"strings"
)

// ValidationRule 验证规则
type ValidationRule func(*RedisConfig) error

// ConfigValidator 配置验证器
type ConfigValidator struct {
	rules []ValidationRule
}

// NewConfigValidator 创建配置验证器
func NewConfigValidator() *ConfigValidator {
	validator := &ConfigValidator{}

	// 添加默认验证规则
	validator.AddRule(validateProtocol)
	validator.AddRule(validateMode)
	validator.AddRule(validateConnection)
	validator.AddRule(validateBenchmark)

	return validator
}

// AddRule 添加验证规则
func (v *ConfigValidator) AddRule(rule ValidationRule) {
	v.rules = append(v.rules, rule)
}

// Validate 验证配置
func (v *ConfigValidator) Validate(config *RedisConfig) error {
	for _, rule := range v.rules {
		if err := rule(config); err != nil {
			return err
		}
	}
	return nil
}

// validateProtocol 验证协议
func validateProtocol(config *RedisConfig) error {
	if config.Protocol == "" {
		return fmt.Errorf("protocol cannot be empty")
	}

	if config.Protocol != "redis" {
		return fmt.Errorf("unsupported protocol: %s", config.Protocol)
	}

	return nil
}

// validateMode 验证模式
func validateMode(config *RedisConfig) error {
	mode := config.GetMode()

	switch mode {
	case "standalone", "sentinel", "cluster":
		return nil
	default:
		return fmt.Errorf("unsupported mode: %s", mode)
	}
}

// validateConnection 验证连接配置
func validateConnection(config *RedisConfig) error {
	mode := config.GetMode()

	switch mode {
	case "standalone":
		if config.Standalone.Addr == "" {
			return fmt.Errorf("standalone addr cannot be empty")
		}
		// 验证地址格式
		if !isValidAddr(config.Standalone.Addr) {
			return fmt.Errorf("invalid standalone addr format: %s", config.Standalone.Addr)
		}

	case "sentinel":
		if len(config.Sentinel.Addrs) == 0 {
			return fmt.Errorf("sentinel addrs cannot be empty")
		}
		if config.Sentinel.MasterName == "" {
			return fmt.Errorf("sentinel master_name cannot be empty")
		}
		// 验证哨兵地址格式
		for _, addr := range config.Sentinel.Addrs {
			if !isValidAddr(addr) {
				return fmt.Errorf("invalid sentinel addr format: %s", addr)
			}
		}

	case "cluster":
		if len(config.Cluster.Addrs) == 0 {
			return fmt.Errorf("cluster addrs cannot be empty")
		}
		// 验证集群地址格式
		for _, addr := range config.Cluster.Addrs {
			if !isValidAddr(addr) {
				return fmt.Errorf("invalid cluster addr format: %s", addr)
			}
		}
	}

	return nil
}

// validateBenchmark 验证基准测试配置
func validateBenchmark(config *RedisConfig) error {
	benchmark := &config.BenchMark

	if benchmark.Total <= 0 {
		return fmt.Errorf("benchmark total must be positive, got: %d", benchmark.Total)
	}

	if benchmark.Parallels <= 0 {
		return fmt.Errorf("benchmark parallels must be positive, got: %d", benchmark.Parallels)
	}

	if benchmark.DataSize <= 0 {
		return fmt.Errorf("benchmark data_size must be positive, got: %d", benchmark.DataSize)
	}

	if benchmark.ReadPercent < 0 || benchmark.ReadPercent > 100 {
		return fmt.Errorf("benchmark read_percent must be between 0 and 100, got: %d", benchmark.ReadPercent)
	}

	if benchmark.RandomKeys < 0 {
		return fmt.Errorf("benchmark random_keys cannot be negative, got: %d", benchmark.RandomKeys)
	}

	if benchmark.TTL < 0 {
		return fmt.Errorf("benchmark ttl cannot be negative, got: %d", benchmark.TTL)
	}

	// 验证测试用例
	validCases := []string{"get", "set", "set_get", "set_get_random", "pub", "sub"}
	if benchmark.Case != "" {
		found := false
		for _, validCase := range validCases {
			if benchmark.Case == validCase {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("invalid test case: %s, valid cases: %v", benchmark.Case, validCases)
		}
	}

	return nil
}

// isValidAddr 验证地址格式
func isValidAddr(addr string) bool {
	if addr == "" {
		return false
	}

	// 简单的地址格式验证：host:port
	// 这里可以根据需要添加更严格的验证
	parts := strings.Split(addr, ":")
	if len(parts) != 2 {
		return false
	}

	host := parts[0]
	port := parts[1]

	if host == "" || port == "" {
		return false
	}

	// 验证端口是否为数字
	if _, err := strconv.Atoi(port); err != nil {
		return false
	}

	return true
}
