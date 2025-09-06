package operations

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
	
	"redis-runner/app/core/interfaces"
	httpConfig "redis-runner/app/adapters/http/config"
)

// HttpOperationFactory HTTP操作工厂
type HttpOperationFactory struct {
	config      *httpConfig.HttpAdapterConfig
	weightedReqs []weightedRequest
	totalWeight  int
	rand         *rand.Rand
}

// weightedRequest 加权请求配置
type weightedRequest struct {
	config     httpConfig.HttpRequestConfig
	weight     int
	startRange int
	endRange   int
}

// NewHttpOperationFactory 创建HTTP操作工厂
func NewHttpOperationFactory(config *httpConfig.HttpAdapterConfig) *HttpOperationFactory {
	factory := &HttpOperationFactory{
		config: config,
		rand:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	
	factory.initializeWeightedRequests()
	return factory
}

// initializeWeightedRequests 初始化加权请求
func (f *HttpOperationFactory) initializeWeightedRequests() {
	f.weightedReqs = make([]weightedRequest, 0, len(f.config.Requests))
	f.totalWeight = 0
	
	for _, reqConfig := range f.config.Requests {
		weight := reqConfig.Weight
		if weight <= 0 {
			weight = 1 // 默认权重为1
		}
		
		startRange := f.totalWeight
		f.totalWeight += weight
		endRange := f.totalWeight - 1
		
		f.weightedReqs = append(f.weightedReqs, weightedRequest{
			config:     reqConfig,
			weight:     weight,
			startRange: startRange,
			endRange:   endRange,
		})
	}
}

// CreateOperation 创建操作
func (f *HttpOperationFactory) CreateOperation(params map[string]interface{}) (interfaces.Operation, error) {
	// 选择请求配置
	reqConfig, err := f.selectRequestConfig(params)
	if err != nil {
		return interfaces.Operation{}, fmt.Errorf("failed to select request config: %w", err)
	}
	
	// 处理模板变量
	processedConfig, err := f.processTemplate(reqConfig, params)
	if err != nil {
		return interfaces.Operation{}, fmt.Errorf("failed to process template: %w", err)
	}
	
	// 创建操作
	operation := interfaces.Operation{
		Type:  f.getOperationType(processedConfig.Method),
		Key:   f.generateOperationKey(processedConfig, params),
		Value: processedConfig.Body,
		Params: map[string]interface{}{
			"method":       processedConfig.Method,
			"path":         processedConfig.Path,
			"headers":      processedConfig.Headers,
			"content_type": processedConfig.ContentType,
			"upload":       processedConfig.Upload,
			"raw_config":   processedConfig,
		},
		Metadata: map[string]string{
			"operation_factory": "http",
			"request_method":    processedConfig.Method,
			"request_path":      processedConfig.Path,
		},
	}
	
	return operation, nil
}

// selectRequestConfig 选择请求配置
func (f *HttpOperationFactory) selectRequestConfig(params map[string]interface{}) (httpConfig.HttpRequestConfig, error) {
	// 如果指定了特定的操作类型
	if opType, exists := params["operation_type"]; exists {
		if opTypeStr, ok := opType.(string); ok {
			return f.selectByOperationType(opTypeStr)
		}
	}
	
	// 如果指定了特定的方法
	if method, exists := params["method"]; exists {
		if methodStr, ok := method.(string); ok {
			return f.selectByMethod(methodStr)
		}
	}
	
	// 按权重随机选择
	return f.selectByWeight(), nil
}

// selectByOperationType 按操作类型选择
func (f *HttpOperationFactory) selectByOperationType(opType string) (httpConfig.HttpRequestConfig, error) {
	method := f.operationTypeToMethod(opType)
	return f.selectByMethod(method)
}

// selectByMethod 按方法选择
func (f *HttpOperationFactory) selectByMethod(method string) (httpConfig.HttpRequestConfig, error) {
	for _, wreq := range f.weightedReqs {
		if strings.EqualFold(wreq.config.Method, method) {
			return wreq.config, nil
		}
	}
	
	return httpConfig.HttpRequestConfig{}, fmt.Errorf("no request config found for method: %s", method)
}

// selectByWeight 按权重选择
func (f *HttpOperationFactory) selectByWeight() httpConfig.HttpRequestConfig {
	if len(f.weightedReqs) == 0 {
		return httpConfig.HttpRequestConfig{}
	}
	
	if len(f.weightedReqs) == 1 {
		return f.weightedReqs[0].config
	}
	
	randomValue := f.rand.Intn(f.totalWeight)
	
	for _, wreq := range f.weightedReqs {
		if randomValue >= wreq.startRange && randomValue <= wreq.endRange {
			return wreq.config
		}
	}
	
	// 默认返回第一个
	return f.weightedReqs[0].config
}

// processTemplate 处理模板变量
func (f *HttpOperationFactory) processTemplate(config httpConfig.HttpRequestConfig, params map[string]interface{}) (httpConfig.HttpRequestConfig, error) {
	processed := config
	
	// 处理路径中的模板变量
	processed.Path = f.replaceTemplateVariables(config.Path, params)
	
	// 处理头部中的模板变量
	if config.Headers != nil {
		processed.Headers = make(map[string]string)
		for key, value := range config.Headers {
			processed.Headers[key] = f.replaceTemplateVariables(value, params)
		}
	}
	
	// 处理请求体中的模板变量
	if config.Body != nil {
		processed.Body = f.processBodyTemplate(config.Body, params)
	}
	
	return processed, nil
}

// replaceTemplateVariables 替换模板变量
func (f *HttpOperationFactory) replaceTemplateVariables(template string, params map[string]interface{}) string {
	result := template
	
	// 替换参数变量
	for key, value := range params {
		placeholder := fmt.Sprintf("{{%s}}", key)
		replacement := fmt.Sprintf("%v", value)
		result = strings.ReplaceAll(result, placeholder, replacement)
	}
	
	// 替换随机变量
	result = f.replaceRandomVariables(result)
	
	return result
}

// replaceRandomVariables 替换随机变量
func (f *HttpOperationFactory) replaceRandomVariables(template string) string {
	result := template
	
	// 生成随机ID
	result = strings.ReplaceAll(result, "{{random.id}}", strconv.Itoa(f.rand.Intn(10000)+1))
	
	// 生成随机名称
	result = strings.ReplaceAll(result, "{{random.name}}", f.generateRandomName())
	
	// 生成随机邮箱
	result = strings.ReplaceAll(result, "{{random.email}}", f.generateRandomEmail())
	
	// 生成随机状态
	result = strings.ReplaceAll(result, "{{random.status}}", f.generateRandomStatus())
	
	// 生成随机标题
	result = strings.ReplaceAll(result, "{{random.title}}", f.generateRandomTitle())
	
	// 生成随机描述
	result = strings.ReplaceAll(result, "{{random.description}}", f.generateRandomDescription())
	
	return result
}

// processBodyTemplate 处理请求体模板
func (f *HttpOperationFactory) processBodyTemplate(body interface{}, params map[string]interface{}) interface{} {
	switch v := body.(type) {
	case string:
		return f.replaceTemplateVariables(v, params)
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, value := range v {
			result[key] = f.processBodyTemplate(value, params)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, value := range v {
			result[i] = f.processBodyTemplate(value, params)
		}
		return result
	default:
		return body
	}
}

// generateOperationKey 生成操作键
func (f *HttpOperationFactory) generateOperationKey(config httpConfig.HttpRequestConfig, params map[string]interface{}) string {
	// 基于方法和路径生成键
	key := fmt.Sprintf("%s:%s", config.Method, config.Path)
	
	// 如果有索引参数，加入键中
	if index, exists := params["index"]; exists {
		key = fmt.Sprintf("%s:%v", key, index)
	}
	
	return key
}

// getOperationType 获取操作类型
func (f *HttpOperationFactory) getOperationType(method string) string {
	return fmt.Sprintf("http_%s", strings.ToLower(method))
}

// operationTypeToMethod 操作类型转方法
func (f *HttpOperationFactory) operationTypeToMethod(opType string) string {
	if strings.HasPrefix(opType, "http_") {
		return strings.ToUpper(strings.TrimPrefix(opType, "http_"))
	}
	return strings.ToUpper(opType)
}

// GetOperationType 获取操作类型（实现OperationFactory接口）
func (f *HttpOperationFactory) GetOperationType() string {
	return "http"
}

// ValidateParams 验证参数（实现OperationFactory接口）
func (f *HttpOperationFactory) ValidateParams(params map[string]interface{}) error {
	// 基本参数验证
	if len(f.config.Requests) == 0 {
		return fmt.Errorf("no request configurations available")
	}
	
	// 如果指定了method，验证是否支持
	if method, exists := params["method"]; exists {
		if methodStr, ok := method.(string); ok {
			if !f.isMethodSupported(methodStr) {
				return fmt.Errorf("unsupported method: %s", methodStr)
			}
		}
	}
	
	return nil
}

// isMethodSupported 检查方法是否支持
func (f *HttpOperationFactory) isMethodSupported(method string) bool {
	for _, wreq := range f.weightedReqs {
		if strings.EqualFold(wreq.config.Method, method) {
			return true
		}
	}
	return false
}

// 随机数据生成方法

func (f *HttpOperationFactory) generateRandomName() string {
	names := []string{"Alice", "Bob", "Charlie", "Diana", "Edward", "Fiona", "George", "Helen"}
	return names[f.rand.Intn(len(names))]
}

func (f *HttpOperationFactory) generateRandomEmail() string {
	domains := []string{"example.com", "test.com", "demo.org"}
	name := f.generateRandomName()
	domain := domains[f.rand.Intn(len(domains))]
	return fmt.Sprintf("%s@%s", strings.ToLower(name), domain)
}

func (f *HttpOperationFactory) generateRandomStatus() string {
	statuses := []string{"active", "inactive", "pending", "completed", "cancelled"}
	return statuses[f.rand.Intn(len(statuses))]
}

func (f *HttpOperationFactory) generateRandomTitle() string {
	titles := []string{"Important Document", "Test Report", "User Manual", "Project Specification", "Meeting Notes"}
	return titles[f.rand.Intn(len(titles))]
}

func (f *HttpOperationFactory) generateRandomDescription() string {
	descriptions := []string{
		"This is a test description",
		"Sample description for testing",
		"Auto-generated content",
		"Placeholder description text",
		"Test data for benchmarking",
	}
	return descriptions[f.rand.Intn(len(descriptions))]
}