package connection

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	httpConfig "abc-runner/app/adapters/http/config"
)

// HttpClient HTTP客户端封装
type HttpClient struct {
	client *http.Client
	config *httpConfig.HttpAdapterConfig
	pool   *HTTPConnectionPool
}

// NewHttpClient 创建HTTP客户端
func NewHttpClient(client *http.Client, config *httpConfig.HttpAdapterConfig, pool *HTTPConnectionPool) *HttpClient {
	return &HttpClient{
		client: client,
		config: config,
		pool:   pool,
	}
}

// ExecuteRequest 执行HTTP请求
func (c *HttpClient) ExecuteRequest(ctx context.Context, reqConfig httpConfig.HttpRequestConfig) (*HttpResponse, error) {
	// 构建完整URL
	fullURL, err := c.buildURL(reqConfig.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	// 准备请求体
	body, contentType, err := c.prepareRequestBody(reqConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare request body: %w", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, reqConfig.Method, fullURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	c.setRequestHeaders(req, reqConfig, contentType)

	// 设置认证
	if err := c.setAuthentication(req); err != nil {
		return nil, fmt.Errorf("failed to set authentication: %w", err)
	}

	// 执行请求
	startTime := time.Now()
	resp, err := c.client.Do(req)
	duration := time.Since(startTime)

	if err != nil {
		return &HttpResponse{
			StatusCode: 0,
			Duration:   duration,
			Error:      err,
		}, err
	}

	// 读取响应体
	respBody, err := c.readResponseBody(resp)
	if err != nil {
		resp.Body.Close()
		return &HttpResponse{
			StatusCode: resp.StatusCode,
			Duration:   duration,
			Error:      err,
		}, err
	}

	// 确保响应体被关闭
	resp.Body.Close()

	return &HttpResponse{
		StatusCode:    resp.StatusCode,
		Headers:       resp.Header,
		Body:          respBody,
		ContentLength: resp.ContentLength,
		Duration:      duration,
		Success:       c.isSuccessStatusCode(resp.StatusCode),
	}, nil
}

// buildURL 构建完整URL
func (c *HttpClient) buildURL(path string) (string, error) {
	baseURL := c.config.Connection.BaseURL

	// 解析基础URL
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}

	// 解析路径
	pathURL, err := url.Parse(path)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	// 合并URL
	fullURL := base.ResolveReference(pathURL)
	return fullURL.String(), nil
}

// prepareRequestBody 准备请求体
func (c *HttpClient) prepareRequestBody(reqConfig httpConfig.HttpRequestConfig) (io.Reader, string, error) {
	// 如果有文件上传配置
	if reqConfig.Upload != nil {
		return c.prepareMultipartBody(reqConfig)
	}

	// 如果没有body，返回空
	if reqConfig.Body == nil {
		return nil, reqConfig.ContentType, nil
	}

	// 根据Content-Type处理body
	switch reqConfig.ContentType {
	case "application/json":
		return c.prepareJSONBody(reqConfig.Body)
	case "application/x-www-form-urlencoded":
		return c.prepareFormBody(reqConfig.Body)
	case "text/plain":
		return c.prepareTextBody(reqConfig.Body)
	default:
		// 默认按JSON处理
		return c.prepareJSONBody(reqConfig.Body)
	}
}

// prepareJSONBody 准备JSON请求体
func (c *HttpClient) prepareJSONBody(body interface{}) (io.Reader, string, error) {
	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal JSON body: %w", err)
	}

	return bytes.NewBuffer(jsonData), "application/json", nil
}

// prepareFormBody 准备表单请求体
func (c *HttpClient) prepareFormBody(body interface{}) (io.Reader, string, error) {
	values := url.Values{}

	// 将body转换为map
	bodyMap, ok := body.(map[string]interface{})
	if !ok {
		return nil, "", fmt.Errorf("form body must be a map[string]interface{}")
	}

	for key, value := range bodyMap {
		values.Set(key, fmt.Sprintf("%v", value))
	}

	return strings.NewReader(values.Encode()), "application/x-www-form-urlencoded", nil
}

// prepareTextBody 准备文本请求体
func (c *HttpClient) prepareTextBody(body interface{}) (io.Reader, string, error) {
	text := fmt.Sprintf("%v", body)
	return strings.NewReader(text), "text/plain", nil
}

// prepareMultipartBody 准备multipart请求体
func (c *HttpClient) prepareMultipartBody(reqConfig httpConfig.HttpRequestConfig) (io.Reader, string, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// 添加表单数据
	if reqConfig.Upload.FormData != nil {
		for key, value := range reqConfig.Upload.FormData {
			if err := writer.WriteField(key, fmt.Sprintf("%v", value)); err != nil {
				return nil, "", fmt.Errorf("failed to write form field %s: %w", key, err)
			}
		}
	}

	// 添加文件
	for _, fileConfig := range reqConfig.Upload.Files {
		if err := c.addFileToMultipart(writer, fileConfig); err != nil {
			return nil, "", fmt.Errorf("failed to add file %s: %w", fileConfig.Path, err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, "", fmt.Errorf("failed to close multipart writer: %w", err)
	}

	return &buf, writer.FormDataContentType(), nil
}

// addFileToMultipart 向multipart添加文件
func (c *HttpClient) addFileToMultipart(writer *multipart.Writer, fileConfig httpConfig.FileConfig) error {
	// 如果指定了pattern，需要匹配文件
	if fileConfig.Pattern != "" {
		matches, err := filepath.Glob(filepath.Join(fileConfig.Path, fileConfig.Pattern))
		if err != nil {
			return fmt.Errorf("failed to glob pattern %s: %w", fileConfig.Pattern, err)
		}

		if len(matches) == 0 {
			return fmt.Errorf("no files found matching pattern %s", fileConfig.Pattern)
		}

		// 上传第一个匹配的文件
		return c.addSingleFileToMultipart(writer, fileConfig.Field, matches[0])
	}

	// 直接上传指定文件
	return c.addSingleFileToMultipart(writer, fileConfig.Field, fileConfig.Path)
}

// addSingleFileToMultipart 向multipart添加单个文件
func (c *HttpClient) addSingleFileToMultipart(writer *multipart.Writer, fieldName, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	// 创建文件字段
	part, err := writer.CreateFormFile(fieldName, filepath.Base(filePath))
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	// 复制文件内容
	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	return nil
}

// setRequestHeaders 设置请求头
func (c *HttpClient) setRequestHeaders(req *http.Request, reqConfig httpConfig.HttpRequestConfig, contentType string) {
	// 设置User-Agent
	if c.config.Benchmark.UserAgent != "" {
		req.Header.Set("User-Agent", c.config.Benchmark.UserAgent)
	}

	// 设置Content-Type
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	// 设置自定义头部
	for key, value := range reqConfig.Headers {
		req.Header.Set(key, value)
	}
}

// setAuthentication 设置认证
func (c *HttpClient) setAuthentication(req *http.Request) error {
	switch c.config.Auth.Type {
	case "none":
		// 无需认证
		return nil
	case "basic":
		req.SetBasicAuth(c.config.Auth.Username, c.config.Auth.Password)
		return nil
	case "bearer":
		req.Header.Set("Authorization", "Bearer "+c.config.Auth.Token)
		return nil
	case "oauth2":
		// TODO: 实现OAuth2认证
		return fmt.Errorf("OAuth2 authentication not implemented yet")
	case "mutual_tls":
		// TLS认证在传输层处理
		return nil
	default:
		return fmt.Errorf("unsupported authentication type: %s", c.config.Auth.Type)
	}
}

// readResponseBody 读取响应体
func (c *HttpClient) readResponseBody(resp *http.Response) ([]byte, error) {
	// 限制读取大小以防止内存耗尽
	const maxBodySize = 10 * 1024 * 1024 // 10MB

	limitedReader := io.LimitReader(resp.Body, maxBodySize)
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}

// isSuccessStatusCode 检查是否为成功状态码
func (c *HttpClient) isSuccessStatusCode(statusCode int) bool {
	return statusCode >= 200 && statusCode < 300
}

// HttpResponse HTTP响应结构
type HttpResponse struct {
	StatusCode    int
	Headers       http.Header
	Body          []byte
	ContentLength int64
	Duration      time.Duration
	Success       bool
	Error         error
}

// String 返回响应的字符串表示
func (r *HttpResponse) String() string {
	if r.Error != nil {
		return fmt.Sprintf("Error: %v, Duration: %v", r.Error, r.Duration)
	}

	return fmt.Sprintf("Status: %d, Length: %d bytes, Duration: %v",
		r.StatusCode, len(r.Body), r.Duration)
}

// IsSuccess 检查请求是否成功
func (r *HttpResponse) IsSuccess() bool {
	return r.Success && r.Error == nil
}

// GetBodyString 获取响应体字符串
func (r *HttpResponse) GetBodyString() string {
	return string(r.Body)
}

// GetHeader 获取响应头
func (r *HttpResponse) GetHeader(key string) string {
	return r.Headers.Get(key)
}
