package http

import (
	"abc-runner/app/core/interfaces"
	"abc-runner/app/core/utils"
)

// RegisterHttpOperations 注册HTTP操作
func RegisterHttpOperations(registry *utils.OperationRegistry) {
	// HTTP GET 操作
	registry.Register("http_get", &HttpGetOperationFactory{})

	// HTTP POST 操作
	registry.Register("http_post", &HttpPostOperationFactory{})

	// HTTP PUT 操作
	registry.Register("http_put", &HttpPutOperationFactory{})

	// HTTP DELETE 操作
	registry.Register("http_delete", &HttpDeleteOperationFactory{})

	// HTTP PATCH 操作
	registry.Register("http_patch", &HttpPatchOperationFactory{})

	// HTTP HEAD 操作
	registry.Register("http_head", &HttpHeadOperationFactory{})

	// HTTP OPTIONS 操作
	registry.Register("http_options", &HttpOptionsOperationFactory{})

	// 通用HTTP操作
	registry.Register("http_request", &HttpRequestOperationFactory{})
}

// HTTP操作工厂实现
type HttpGetOperationFactory struct{}

func (f *HttpGetOperationFactory) CreateOperation(params map[string]interface{}) (interfaces.Operation, error) {
	// 创建GET操作
	return interfaces.Operation{
		Type:   "http_get",
		Params: params,
		Metadata: map[string]string{
			"method": "GET",
		},
	}, nil
}

func (f *HttpGetOperationFactory) GetOperationType() string {
	return "http_get"
}

func (f *HttpGetOperationFactory) ValidateParams(params map[string]interface{}) error {
	return nil
}

type HttpPostOperationFactory struct{}

func (f *HttpPostOperationFactory) CreateOperation(params map[string]interface{}) (interfaces.Operation, error) {
	// 创建POST操作
	return interfaces.Operation{
		Type:   "http_post",
		Params: params,
		Metadata: map[string]string{
			"method": "POST",
		},
	}, nil
}

func (f *HttpPostOperationFactory) GetOperationType() string {
	return "http_post"
}

func (f *HttpPostOperationFactory) ValidateParams(params map[string]interface{}) error {
	return nil
}

type HttpPutOperationFactory struct{}

func (f *HttpPutOperationFactory) CreateOperation(params map[string]interface{}) (interfaces.Operation, error) {
	// 创建PUT操作
	return interfaces.Operation{
		Type:   "http_put",
		Params: params,
		Metadata: map[string]string{
			"method": "PUT",
		},
	}, nil
}

func (f *HttpPutOperationFactory) GetOperationType() string {
	return "http_put"
}

func (f *HttpPutOperationFactory) ValidateParams(params map[string]interface{}) error {
	return nil
}

type HttpDeleteOperationFactory struct{}

func (f *HttpDeleteOperationFactory) CreateOperation(params map[string]interface{}) (interfaces.Operation, error) {
	// 创建DELETE操作
	return interfaces.Operation{
		Type:   "http_delete",
		Params: params,
		Metadata: map[string]string{
			"method": "DELETE",
		},
	}, nil
}

func (f *HttpDeleteOperationFactory) GetOperationType() string {
	return "http_delete"
}

func (f *HttpDeleteOperationFactory) ValidateParams(params map[string]interface{}) error {
	return nil
}

type HttpPatchOperationFactory struct{}

func (f *HttpPatchOperationFactory) CreateOperation(params map[string]interface{}) (interfaces.Operation, error) {
	// 创建PATCH操作
	return interfaces.Operation{
		Type:   "http_patch",
		Params: params,
		Metadata: map[string]string{
			"method": "PATCH",
		},
	}, nil
}

func (f *HttpPatchOperationFactory) GetOperationType() string {
	return "http_patch"
}

func (f *HttpPatchOperationFactory) ValidateParams(params map[string]interface{}) error {
	return nil
}

type HttpHeadOperationFactory struct{}

func (f *HttpHeadOperationFactory) CreateOperation(params map[string]interface{}) (interfaces.Operation, error) {
	// 创建HEAD操作
	return interfaces.Operation{
		Type:   "http_head",
		Params: params,
		Metadata: map[string]string{
			"method": "HEAD",
		},
	}, nil
}

func (f *HttpHeadOperationFactory) GetOperationType() string {
	return "http_head"
}

func (f *HttpHeadOperationFactory) ValidateParams(params map[string]interface{}) error {
	return nil
}

type HttpOptionsOperationFactory struct{}

func (f *HttpOptionsOperationFactory) CreateOperation(params map[string]interface{}) (interfaces.Operation, error) {
	// 创建OPTIONS操作
	return interfaces.Operation{
		Type:   "http_options",
		Params: params,
		Metadata: map[string]string{
			"method": "OPTIONS",
		},
	}, nil
}

func (f *HttpOptionsOperationFactory) GetOperationType() string {
	return "http_options"
}

func (f *HttpOptionsOperationFactory) ValidateParams(params map[string]interface{}) error {
	return nil
}

type HttpRequestOperationFactory struct{}

func (f *HttpRequestOperationFactory) CreateOperation(params map[string]interface{}) (interfaces.Operation, error) {
	// 创建通用HTTP请求操作
	return interfaces.Operation{
		Type:   "http_request",
		Params: params,
	}, nil
}

func (f *HttpRequestOperationFactory) GetOperationType() string {
	return "http_request"
}

func (f *HttpRequestOperationFactory) ValidateParams(params map[string]interface{}) error {
	return nil
}
