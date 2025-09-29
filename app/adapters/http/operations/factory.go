package operations

import (
	"fmt"
	"strconv"
	"time"

	httpConfig "abc-runner/app/adapters/http/config"
	"abc-runner/app/core/execution"
	"abc-runner/app/core/interfaces"
)

// HttpOperationFactory HTTP操作工厂
type HttpOperationFactory struct {
	config   *httpConfig.HttpAdapterConfig
	testCase string
	dataSize int
}

// NewHttpOperationFactory 创建HTTP操作工厂
func NewHttpOperationFactory(config *httpConfig.HttpAdapterConfig) *HttpOperationFactory {
	return &HttpOperationFactory{
		config:   config,
		testCase: config.Benchmark.TestCase,
		dataSize: config.Benchmark.DataSize,
	}
}

// CreateOperation 创建HTTP操作
func (f *HttpOperationFactory) CreateOperation(jobID int, config execution.BenchmarkConfig) interfaces.Operation {
	// 生成操作键（URL路径）
	path := f.generatePath(jobID)
	
	// 生成操作值（请求体）
	value := f.generateRequestBody(jobID)
	
	// 创建操作特定参数
	params := map[string]interface{}{
		"job_id":       jobID,
		"data_size":    f.dataSize,
		"test_case":    f.testCase,
		"read_percent": f.config.Benchmark.ReadPercent,
		"base_url":     f.config.Connection.BaseURL,
		"timeout":      f.config.Connection.Timeout.Seconds(),
		"headers":      f.generateHeaders(jobID),
	}
	
	// 创建操作元数据
	metadata := map[string]string{
		"operation_type": f.testCase,
		"protocol":       "http",
		"job_id":         strconv.Itoa(jobID),
		"user_agent":     "abc-runner-http-client", // 默认值，因为配置中没有UserAgent字段
	}
	
	// 根据测试用例确定具体操作类型
	operationType := f.determineOperationType(jobID)
	
	return interfaces.Operation{
		Type:     operationType,
		Key:      path,
		Value:    value,
		Params:   params,
		TTL:      f.config.Benchmark.TTL,
		Metadata: metadata,
	}
}

// determineOperationType 根据测试用例和任务ID确定操作类型
func (f *HttpOperationFactory) determineOperationType(jobID int) string {
	switch f.testCase {
	case "get_post_mixed":
		// 根据读写比例决定操作类型
		if jobID%100 < f.config.Benchmark.ReadPercent {
			return "http_get"
		}
		return "http_post"
	
	case "get_only":
		return "http_get"
		
	case "post_only":
		return "http_post"
		
	case "put_only":
		return "http_put"
		
	case "delete_only":
		return "http_delete"
		
	case "patch_only":
		return "http_patch"
		
	case "head_only":
		return "http_head"
		
	case "options_only":
		return "http_options"
		
	case "crud_operations":
		// CRUD操作循环
		switch jobID % 4 {
		case 0:
			return "http_post"   // Create
		case 1:
			return "http_get"    // Read
		case 2:
			return "http_put"    // Update
		case 3:
			return "http_delete" // Delete
		}
		
	case "rest_api_test":
		// RESTful API测试
		switch jobID % 6 {
		case 0:
			return "http_get"
		case 1:
			return "http_post"
		case 2:
			return "http_put"
		case 3:
			return "http_patch"
		case 4:
			return "http_delete"
		case 5:
			return "http_head"
		}
		
	default:
		return "http_get" // 默认操作
	}
	
	return "http_get"
}

// generatePath 生成请求路径
func (f *HttpOperationFactory) generatePath(jobID int) string {
	switch f.testCase {
	case "get_post_mixed", "get_only":
		if f.config.Benchmark.RandomKeys > 0 {
			resourceID := jobID % f.config.Benchmark.RandomKeys
			return fmt.Sprintf("/api/v1/resources/%d", resourceID)
		}
		return fmt.Sprintf("/api/v1/resources/%d", jobID)
		
	case "post_only":
		return "/api/v1/resources"
		
	case "put_only", "patch_only":
		return fmt.Sprintf("/api/v1/resources/%d", jobID)
		
	case "delete_only":
		return fmt.Sprintf("/api/v1/resources/%d", jobID)
		
	case "head_only":
		return fmt.Sprintf("/api/v1/resources/%d", jobID)
		
	case "options_only":
		return "/api/v1/resources"
		
	case "crud_operations", "rest_api_test":
		switch jobID % 4 {
		case 0, 1: // POST, GET
			return "/api/v1/items"
		case 2, 3: // PUT, DELETE
			return fmt.Sprintf("/api/v1/items/%d", jobID)
		}
		
	default:
		return fmt.Sprintf("/api/test/%d", jobID)
	}
	
	return "/api/test"
}

// generateRequestBody 生成请求体
func (f *HttpOperationFactory) generateRequestBody(jobID int) interface{} {
	switch f.testCase {
	case "get_only", "delete_only", "head_only", "options_only":
		return nil // 这些方法通常不需要请求体
		
	case "post_only", "put_only", "patch_only":
		return f.generateJSONBody(jobID)
		
	case "get_post_mixed":
		// POST请求需要请求体
		return f.generateJSONBody(jobID)
		
	case "crud_operations", "rest_api_test":
		// CREATE和UPDATE操作需要请求体
		opType := jobID % 4
		if opType == 0 || opType == 2 { // POST或PUT
			return f.generateJSONBody(jobID)
		}
		return nil
		
	default:
		return f.generateJSONBody(jobID)
	}
}

// generateJSONBody 生成JSON请求体
func (f *HttpOperationFactory) generateJSONBody(jobID int) map[string]interface{} {
	body := map[string]interface{}{
		"id":        jobID,
		"name":      fmt.Sprintf("test_item_%d", jobID),
		"timestamp": time.Now().Unix(),
		"job_id":    jobID,
	}
	
	// 根据数据大小生成额外的数据
	if f.dataSize > 0 {
		// 生成指定大小的数据字段
		dataContent := make([]byte, f.dataSize)
		pattern := fmt.Sprintf("HTTP_DATA_%d_", jobID)
		patternBytes := []byte(pattern)
		
		for i := 0; i < f.dataSize; i++ {
			dataContent[i] = patternBytes[i%len(patternBytes)]
		}
		
		body["data"] = string(dataContent)
	}
	
	// 添加更多字段以模拟真实场景
	body["description"] = fmt.Sprintf("Description for item %d", jobID)
	body["category"] = fmt.Sprintf("category_%d", jobID%10)
	body["status"] = "active"
	body["version"] = 1
	
	return body
}

// generateHeaders 生成请求头
func (f *HttpOperationFactory) generateHeaders(jobID int) map[string]string {
	headers := map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json",
		"User-Agent":   "abc-runner-http-client", // 默认值
	}
	
	// 根据测试用例添加特定头部
	switch f.testCase {
	case "post_only", "put_only", "patch_only":
		headers["X-API-Version"] = "v1"
		headers["X-Client-Type"] = "performance-test"
		
	case "get_only":
		headers["Cache-Control"] = "no-cache"
		
	case "rest_api_test":
		headers["X-Test-Type"] = "rest-api"
		headers["X-Job-ID"] = strconv.Itoa(jobID)
	}
	
	return headers
}

// GetOperationType 获取操作类型
func (f *HttpOperationFactory) GetOperationType() string {
	return f.testCase
}

// GetConfig 获取配置信息
func (f *HttpOperationFactory) GetConfig() *httpConfig.HttpAdapterConfig {
	return f.config
}

// GetSupportedOperations 获取支持的操作类型
func (f *HttpOperationFactory) GetSupportedOperations() []string {
	return []string{
		"get_post_mixed", "get_only", "post_only", "put_only", "delete_only",
		"patch_only", "head_only", "options_only", "crud_operations", "rest_api_test",
	}
}

// GetSupportedHTTPMethods 获取支持的HTTP方法
func (f *HttpOperationFactory) GetSupportedHTTPMethods() []string {
	return []string{
		"http_get", "http_post", "http_put", "http_delete",
		"http_patch", "http_head", "http_options",
	}
}

// ValidateTestCase 验证测试用例是否支持
func (f *HttpOperationFactory) ValidateTestCase(testCase string) error {
	supportedCases := f.GetSupportedOperations()
	for _, supported := range supportedCases {
		if testCase == supported {
			return nil
		}
	}
	return fmt.Errorf("unsupported test case: %s, supported: %v", testCase, supportedCases)
}

// GetOperationMetadata 获取操作元数据
func (f *HttpOperationFactory) GetOperationMetadata(operationType string) map[string]interface{} {
	metadata := map[string]interface{}{
		"protocol":     "http",
		"operation":    operationType,
		"test_case":    f.testCase,
		"is_read":      f.isReadOperation(operationType),
		"base_url":     f.config.Connection.BaseURL,
		"timeout":      f.config.Connection.Timeout.String(),
	}
	
	return metadata
}

// isReadOperation 判断是否为读操作
func (f *HttpOperationFactory) isReadOperation(operationType string) bool {
	readOps := []string{"http_get", "http_head", "http_options"}
	for _, readOp := range readOps {
		if readOp == operationType {
			return true
		}
	}
	return false
}

// 确保实现了execution.OperationFactory接口
var _ execution.OperationFactory = (*HttpOperationFactory)(nil)