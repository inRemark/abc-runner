package operations

import (
	"context"
	"fmt"
	"time"

	httpConfig "abc-runner/app/adapters/http/config"
	"abc-runner/app/adapters/http/connection"
	"abc-runner/app/adapters/http/metrics"
	"abc-runner/app/core/interfaces"
)

// HttpOperations HTTP操作执行器
type HttpOperations struct {
	pool             *connection.ConnectionPool
	config           *httpConfig.HttpAdapterConfig
	metricsCollector *metrics.MetricsCollector
}

// NewHttpOperations 创建HTTP操作执行器
func NewHttpOperations(
	pool *connection.ConnectionPool,
	config *httpConfig.HttpAdapterConfig,
	metricsCollector *metrics.MetricsCollector,
) *HttpOperations {
	return &HttpOperations{
		pool:             pool,
		config:           config,
		metricsCollector: metricsCollector,
	}
}

// ExecuteOperation 执行HTTP操作
func (h *HttpOperations) ExecuteOperation(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	startTime := time.Now()

	// 从操作参数中提取HTTP请求配置
	reqConfig, err := h.extractRequestConfig(operation)
	if err != nil {
		return &interfaces.OperationResult{
			Success:  false,
			Duration: time.Since(startTime),
			IsRead:   h.isReadOperation(operation.Type),
			Error:    fmt.Errorf("failed to extract request config: %w", err),
		}, err
	}

	// 获取HTTP客户端
	client := h.pool.GetClient()
	if client == nil {
		return &interfaces.OperationResult{
			Success:  false,
			Duration: time.Since(startTime),
			IsRead:   h.isReadOperation(operation.Type),
			Error:    fmt.Errorf("failed to get HTTP client from pool"),
		}, fmt.Errorf("failed to get HTTP client from pool")
	}
	defer h.pool.ReturnClient(client)

	// 创建HTTP客户端封装
	httpClient := connection.NewHttpClient(client, h.config, h.pool)

	// 执行HTTP请求
	response, err := httpClient.ExecuteRequest(ctx, reqConfig)
	duration := time.Since(startTime)

	// 构建操作结果
	result := &interfaces.OperationResult{
		Success:  response != nil && response.IsSuccess(),
		Duration: duration,
		IsRead:   h.isReadOperation(operation.Type),
		Value:    h.createResultValue(response),
		Metadata: h.createResultMetadata(operation, response),
	}

	if err != nil {
		result.Error = err
		result.Success = false
	}

	// 记录HTTP特定指标
	if response != nil && h.metricsCollector != nil {
		// 使用核心接口记录指标，通过metadata传递HTTP特定信息
		operationResult := &interfaces.OperationResult{
			Success:  response.StatusCode >= 200 && response.StatusCode < 300,
			IsRead:   h.isReadOperation(operation.Type),
			Duration: duration,
			Metadata: map[string]interface{}{
				"status_code": response.StatusCode,
				"method":      reqConfig.Method,
				"url":         reqConfig.Path,
			},
		}
		h.metricsCollector.RecordOperation(operationResult)
	}

	return result, err
}

// extractRequestConfig 从操作中提取请求配置
func (h *HttpOperations) extractRequestConfig(operation interfaces.Operation) (httpConfig.HttpRequestConfig, error) {
	// 尝试从参数中获取原始配置
	if rawConfig, exists := operation.Params["raw_config"]; exists {
		if config, ok := rawConfig.(httpConfig.HttpRequestConfig); ok {
			return config, nil
		}
	}

	// 手动构建配置
	config := httpConfig.HttpRequestConfig{}

	// 提取基本参数
	if method, exists := operation.Params["method"]; exists {
		if methodStr, ok := method.(string); ok {
			config.Method = methodStr
		}
	}

	if path, exists := operation.Params["path"]; exists {
		if pathStr, ok := path.(string); ok {
			config.Path = pathStr
		}
	}

	if contentType, exists := operation.Params["content_type"]; exists {
		if contentTypeStr, ok := contentType.(string); ok {
			config.ContentType = contentTypeStr
		}
	}

	// 提取头部
	if headers, exists := operation.Params["headers"]; exists {
		if headersMap, ok := headers.(map[string]string); ok {
			config.Headers = headersMap
		}
	}

	// 提取上传配置
	if upload, exists := operation.Params["upload"]; exists {
		if uploadConfig, ok := upload.(*httpConfig.HttpFileUploadConfig); ok {
			config.Upload = uploadConfig
		}
	}

	// 设置请求体
	config.Body = operation.Value

	// 验证必要字段
	if config.Method == "" {
		return config, fmt.Errorf("method is required")
	}

	if config.Path == "" {
		return config, fmt.Errorf("path is required")
	}

	return config, nil
}

// isReadOperation 判断是否为读操作
func (h *HttpOperations) isReadOperation(operationType string) bool {
	readMethods := []string{"http_get", "http_head", "http_options"}
	for _, method := range readMethods {
		if operationType == method {
			return true
		}
	}
	return false
}

// createResultValue 创建结果值
func (h *HttpOperations) createResultValue(response *connection.HttpResponse) interface{} {
	if response == nil {
		return nil
	}

	return map[string]interface{}{
		"status_code":    response.StatusCode,
		"headers":        response.Headers,
		"body":           string(response.Body),
		"content_length": response.ContentLength,
		"duration":       response.Duration,
		"success":        response.Success,
	}
}

// createResultMetadata 创建结果元数据
func (h *HttpOperations) createResultMetadata(operation interfaces.Operation, response *connection.HttpResponse) map[string]interface{} {
	metadata := make(map[string]interface{})

	// 复制操作元数据
	for k, v := range operation.Metadata {
		metadata[k] = v
	}

	// 添加响应元数据
	if response != nil {
		metadata["response_status"] = response.StatusCode
		metadata["response_size"] = len(response.Body)
		metadata["response_duration"] = response.Duration.Nanoseconds()

		// 添加响应头信息
		if contentType := response.GetHeader("Content-Type"); contentType != "" {
			metadata["response_content_type"] = contentType
		}

		if server := response.GetHeader("Server"); server != "" {
			metadata["response_server"] = server
		}
	}

	return metadata
}

// ExecuteGetOperation 执行GET操作
func (h *HttpOperations) ExecuteGetOperation(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	// 确保操作类型为GET
	operation.Type = "http_get"
	if operation.Params == nil {
		operation.Params = make(map[string]interface{})
	}
	operation.Params["method"] = "GET"

	return h.ExecuteOperation(ctx, operation)
}

// ExecutePostOperation 执行POST操作
func (h *HttpOperations) ExecutePostOperation(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	// 确保操作类型为POST
	operation.Type = "http_post"
	if operation.Params == nil {
		operation.Params = make(map[string]interface{})
	}
	operation.Params["method"] = "POST"

	return h.ExecuteOperation(ctx, operation)
}

// ExecutePutOperation 执行PUT操作
func (h *HttpOperations) ExecutePutOperation(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	// 确保操作类型为PUT
	operation.Type = "http_put"
	if operation.Params == nil {
		operation.Params = make(map[string]interface{})
	}
	operation.Params["method"] = "PUT"

	return h.ExecuteOperation(ctx, operation)
}

// ExecutePatchOperation 执行PATCH操作
func (h *HttpOperations) ExecutePatchOperation(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	// 确保操作类型为PATCH
	operation.Type = "http_patch"
	if operation.Params == nil {
		operation.Params = make(map[string]interface{})
	}
	operation.Params["method"] = "PATCH"

	return h.ExecuteOperation(ctx, operation)
}

// ExecuteDeleteOperation 执行DELETE操作
func (h *HttpOperations) ExecuteDeleteOperation(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	// 确保操作类型为DELETE
	operation.Type = "http_delete"
	if operation.Params == nil {
		operation.Params = make(map[string]interface{})
	}
	operation.Params["method"] = "DELETE"

	return h.ExecuteOperation(ctx, operation)
}

// ExecuteHeadOperation 执行HEAD操作
func (h *HttpOperations) ExecuteHeadOperation(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	// 确保操作类型为HEAD
	operation.Type = "http_head"
	if operation.Params == nil {
		operation.Params = make(map[string]interface{})
	}
	operation.Params["method"] = "HEAD"

	return h.ExecuteOperation(ctx, operation)
}

// ExecuteOptionsOperation 执行OPTIONS操作
func (h *HttpOperations) ExecuteOptionsOperation(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	// 确保操作类型为OPTIONS
	operation.Type = "http_options"
	if operation.Params == nil {
		operation.Params = make(map[string]interface{})
	}
	operation.Params["method"] = "OPTIONS"

	return h.ExecuteOperation(ctx, operation)
}

// ExecuteTraceOperation 执行TRACE操作
func (h *HttpOperations) ExecuteTraceOperation(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	// 确保操作类型为TRACE
	operation.Type = "http_trace"
	if operation.Params == nil {
		operation.Params = make(map[string]interface{})
	}
	operation.Params["method"] = "TRACE"

	return h.ExecuteOperation(ctx, operation)
}

// ExecuteConnectOperation 执行CONNECT操作
func (h *HttpOperations) ExecuteConnectOperation(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	// 确保操作类型为CONNECT
	operation.Type = "http_connect"
	if operation.Params == nil {
		operation.Params = make(map[string]interface{})
	}
	operation.Params["method"] = "CONNECT"

	return h.ExecuteOperation(ctx, operation)
}

// ExecuteUploadOperation 执行文件上传操作
func (h *HttpOperations) ExecuteUploadOperation(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	// 确保操作类型为文件上传
	operation.Type = "http_upload"
	if operation.Params == nil {
		operation.Params = make(map[string]interface{})
	}
	operation.Params["method"] = "POST"

	// 确保有上传配置
	if _, exists := operation.Params["upload"]; !exists {
		return &interfaces.OperationResult{
			Success:  false,
			Duration: 0,
			IsRead:   false,
			Error:    fmt.Errorf("upload configuration is required for upload operation"),
		}, fmt.Errorf("upload configuration is required for upload operation")
	}

	return h.ExecuteOperation(ctx, operation)
}
