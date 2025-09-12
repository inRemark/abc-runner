package config

import (
	"time"

	"abc-runner/app/core/interfaces"
)

// HttpArgParser HTTP命令行参数解析器
type HttpArgParser struct{}

// NewHttpArgParser 创建HTTP命令行参数解析器
func NewHttpArgParser() *HttpArgParser {
	return &HttpArgParser{}
}

// ParseArgs 解析命令行参数
func (h *HttpArgParser) ParseArgs(args []string, config interfaces.Config) error {
	httpConfig, ok := config.(*HttpAdapterConfig)
	if !ok {
		return nil // Not an HTTP config, nothing to do
	}

	// 从命令行参数解析配置
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--url", "-u":
			if i+1 < len(args) {
				httpConfig.Connection.BaseURL = args[i+1]
				i++
			}
		case "--timeout":
			if i+1 < len(args) {
				if t, err := time.ParseDuration(args[i+1]); err == nil {
					httpConfig.Connection.Timeout = t
				}
				i++
			}
		case "--total", "-n":
			if i+1 < len(args) {
				if val, err := parseInt(args[i+1]); err == nil {
					httpConfig.Benchmark.Total = val
				}
				i++
			}
		case "--parallels", "-c":
			if i+1 < len(args) {
				if val, err := parseInt(args[i+1]); err == nil {
					httpConfig.Benchmark.Parallels = val
				}
				i++
			}
		case "--method", "-m":
			if i+1 < len(args) {
				httpConfig.Benchmark.Method = args[i+1]
				i++
			}
		case "--path", "-p":
			if i+1 < len(args) {
				httpConfig.Benchmark.Path = args[i+1]
				i++
			}
		}
	}

	return nil
}