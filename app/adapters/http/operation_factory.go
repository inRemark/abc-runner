package http

import (
	"fmt"

	"abc-runner/app/adapters/http/config"
	"abc-runner/app/core/execution"
	"abc-runner/app/core/interfaces"
)

// OperationFactory HTTP操作工厂
type OperationFactory struct {
	config *config.HttpAdapterConfig
}

// NewOperationFactory 创建HTTP操作工厂
func NewOperationFactory(config *config.HttpAdapterConfig) execution.OperationFactory {
	return &OperationFactory{config: config}
}

func (h *OperationFactory) CreateOperation(jobID int, benchmarkConfig execution.BenchmarkConfig) interfaces.Operation {
	// 从配置中选择请求配置
	var reqConfig config.HttpRequestConfig
	if len(h.config.Requests) > 0 {
		// 轮询选择请求配置
		reqConfig = h.config.Requests[jobID%len(h.config.Requests)]
	} else {
		// 默认GET请求
		reqConfig = config.HttpRequestConfig{
			Method: "GET",
			Path:   "/",
		}
	}

	return interfaces.Operation{
		Type: "http_request",
		Key:  fmt.Sprintf("http_job_%d", jobID),
		Params: map[string]interface{}{
			"method":  reqConfig.Method,
			"path":    reqConfig.Path,
			"headers": reqConfig.Headers,
			"body":    reqConfig.Body,
			"weight":  reqConfig.Weight,
		},
	}
}
