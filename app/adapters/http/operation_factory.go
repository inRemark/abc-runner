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

	// 使用具体的HTTP方法作为操作类型
	operationType := reqConfig.Method
	if operationType == "" {
		operationType = "GET"
	}

	// 构建请求路径
	requestPath := reqConfig.Path
	if requestPath == "" {
		requestPath = "/"
	}

	// 构建操作对象
	return interfaces.Operation{
		Type:  operationType, // 使用HTTP方法作为操作类型
		Key:   requestPath,   // 使用路径作为键
		Value: reqConfig.Body, // 请求体
		Params: map[string]interface{}{
			"method":   reqConfig.Method,
			"path":     reqConfig.Path,
			"headers":  reqConfig.Headers,
			"body":     reqConfig.Body,
			"weight":   reqConfig.Weight,
			"job_id":   jobID,
		},
		Metadata: map[string]string{
			"protocol":      "http",
			"operation_type": operationType,
			"job_id":        fmt.Sprintf("%d", jobID),
			"path":          requestPath,
		},
	}
}
