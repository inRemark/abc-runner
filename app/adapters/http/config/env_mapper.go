package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"abc-runner/app/core/interfaces"
)

// HttpEnvVarMapper HTTP环境变量映射器
type HttpEnvVarMapper struct {
	prefix string
}

// NewHttpEnvVarMapper 创建HTTP环境变量映射器
func NewHttpEnvVarMapper(prefix string) *HttpEnvVarMapper {
	if prefix == "" {
		prefix = "HTTP_RUNNER"
	}
	return &HttpEnvVarMapper{prefix: prefix}
}

// MapEnvVarsToConfig 将环境变量映射到配置
func (h *HttpEnvVarMapper) MapEnvVarsToConfig(config interfaces.Config) error {
	httpConfig, ok := config.(*HttpAdapterConfig)
	if !ok {
		return nil // Not an HTTP config, nothing to do
	}

	// 从环境变量加载配置项
	if baseURL := os.Getenv(h.prefix + "_BASE_URL"); baseURL != "" {
		httpConfig.Connection.BaseURL = baseURL
	}

	if timeout := os.Getenv(h.prefix + "_TIMEOUT"); timeout != "" {
		if val, err := time.ParseDuration(timeout); err == nil {
			httpConfig.Connection.Timeout = val
		}
	}

	if maxIdleConns := os.Getenv(h.prefix + "_MAX_IDLE_CONNS"); maxIdleConns != "" {
		if val, err := parseInt(maxIdleConns); err == nil {
			httpConfig.Connection.MaxIdleConns = val
		}
	}

	if maxConnsPerHost := os.Getenv(h.prefix + "_MAX_CONNS_PER_HOST"); maxConnsPerHost != "" {
		if val, err := parseInt(maxConnsPerHost); err == nil {
			httpConfig.Connection.MaxConnsPerHost = val
		}
	}

	if total := os.Getenv(h.prefix + "_TOTAL"); total != "" {
		if val, err := parseInt(total); err == nil {
			httpConfig.Benchmark.Total = val
		}
	}

	if parallels := os.Getenv(h.prefix + "_PARALLELS"); parallels != "" {
		if val, err := parseInt(parallels); err == nil {
			httpConfig.Benchmark.Parallels = val
		}
	}

	if method := os.Getenv(h.prefix + "_METHOD"); method != "" {
		httpConfig.Benchmark.Method = method
	}

	if path := os.Getenv(h.prefix + "_PATH"); path != "" {
		httpConfig.Benchmark.Path = path
	}

	return nil
}

// HasRelevantEnvVars 检查是否有相关的环境变量
func (h *HttpEnvVarMapper) HasRelevantEnvVars() bool {
	// 检查是否有任何相关的环境变量
	envVars := []string{
		h.prefix + "_BASE_URL",
		h.prefix + "_TIMEOUT",
		h.prefix + "_TOTAL",
		h.prefix + "_PARALLELS",
	}

	for _, envVar := range envVars {
		if os.Getenv(envVar) != "" {
			return true
		}
	}

	return false
}

// parseInt 解析整数，忽略错误
func parseInt(s string) (int, error) {
	val, err := parseIntStrict(s)
	if err != nil {
		return 0, err
	}
	return val, nil
}

// parseIntStrict 严格解析整数
func parseIntStrict(s string) (int, error) {
	// 移除可能的空格
	s = strings.TrimSpace(s)

	// 解析整数
	return strconv.Atoi(s)
}
