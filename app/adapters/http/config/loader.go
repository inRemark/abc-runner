package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// DefaultHTTPConfig 默认HTTP配置
func DefaultHTTPConfig() *HttpAdapterConfig {
	return &HttpAdapterConfig{
		Protocol: "http",
		Connection: HttpConnectionConfig{
			BaseURL:            "http://localhost:8080",
			Timeout:            30 * time.Second,
			KeepAlive:          90 * time.Second,
			MaxIdleConns:       50,
			MaxConnsPerHost:    20,
			IdleConnTimeout:    90 * time.Second,
			DisableCompression: false,
			TLS: HttpTLSConfig{
				InsecureSkipVerify:       false,
				MinVersion:               "1.2",
				MaxVersion:               "1.3",
				PreferServerCipherSuites: true,
				SessionTicketsDisabled:   false,
				Renegotiation:            "once",
			},
		},
		Requests: []HttpRequestConfig{
			{
				Method:      "GET",
				Path:        "/",
				Headers:     map[string]string{"Accept": "application/json"},
				ContentType: "application/json",
				Weight:      100,
			},
		},
		Auth: HttpAuthConfig{
			Type: "none",
		},
		Upload: HttpUploadConfig{
			Enable:             false,
			MaxFileSize:        "100MB",
			AllowedTypes:       []string{"image/*", "application/pdf", "text/plain"},
			UploadField:        "file",
			ChunkSize:          "1MB",
			ConcurrentUploads:  3,
			EnableCompression:  true,
			CompressionLevel:   6,
			GenerateThumbnails: false,
			VirusScan:          false,
			TempDir:            "/tmp/upload",
			CleanupInterval:    1 * time.Hour,
			PreserveFilename:   true,
		},
		Benchmark: HttpBenchmarkConfig{
			Total:              100000,
			Parallels:          50,
			Duration:           5 * time.Minute,
			RampUp:             30 * time.Second,
			DataSize:           1024,
			TTL:                0,
			ReadPercent:        70,
			RandomKeys:         0,
			TestCase:           "mixed_operations",
			Timeout:            30 * time.Second,
			FollowRedirects:    true,
			MaxRedirects:       10,
			DisableCompression: false,
			EnableHTTP2:        true,
			UserAgent:          "redis-runner-http-client/1.0",
		},
	}
}

// LoadHTTPConfig 加载HTTP配置
func LoadHTTPConfig(path string) (*HttpAdapterConfig, error) {
	// 首先加载默认配置
	config := DefaultHTTPConfig()

	// 如果指定了配置文件路径，则读取配置文件
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
		}

		// 解析配置文件
		var fileConfig struct {
			HTTP *HttpAdapterConfig `yaml:"http"`
		}

		if err := yaml.Unmarshal(data, &fileConfig); err != nil {
			return nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
		}

		if fileConfig.HTTP != nil {
			// 合并配置
			config = mergeConfig(config, fileConfig.HTTP)
		}
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

// LoadHTTPConfigDefault 加载默认路径的HTTP配置
func LoadHTTPConfigDefault() (*HttpAdapterConfig, error) {
	return LoadHTTPConfig("config/templates/http.yaml")
}

// mergeConfig 合并配置
func mergeConfig(base, override *HttpAdapterConfig) *HttpAdapterConfig {
	result := *base

	// 合并连接配置
	if override.Connection.BaseURL != "" {
		result.Connection.BaseURL = override.Connection.BaseURL
	}
	if override.Connection.Timeout > 0 {
		result.Connection.Timeout = override.Connection.Timeout
	}
	if override.Connection.KeepAlive > 0 {
		result.Connection.KeepAlive = override.Connection.KeepAlive
	}
	if override.Connection.MaxIdleConns > 0 {
		result.Connection.MaxIdleConns = override.Connection.MaxIdleConns
	}
	if override.Connection.MaxConnsPerHost > 0 {
		result.Connection.MaxConnsPerHost = override.Connection.MaxConnsPerHost
	}
	if override.Connection.IdleConnTimeout > 0 {
		result.Connection.IdleConnTimeout = override.Connection.IdleConnTimeout
	}
	result.Connection.DisableCompression = override.Connection.DisableCompression

	// 合并TLS配置
	if override.Connection.TLS.MinVersion != "" {
		result.Connection.TLS.MinVersion = override.Connection.TLS.MinVersion
	}
	if override.Connection.TLS.MaxVersion != "" {
		result.Connection.TLS.MaxVersion = override.Connection.TLS.MaxVersion
	}
	if override.Connection.TLS.CertFile != "" {
		result.Connection.TLS.CertFile = override.Connection.TLS.CertFile
	}
	if override.Connection.TLS.KeyFile != "" {
		result.Connection.TLS.KeyFile = override.Connection.TLS.KeyFile
	}
	if override.Connection.TLS.CAFile != "" {
		result.Connection.TLS.CAFile = override.Connection.TLS.CAFile
	}
	if override.Connection.TLS.ServerName != "" {
		result.Connection.TLS.ServerName = override.Connection.TLS.ServerName
	}
	if len(override.Connection.TLS.CipherSuites) > 0 {
		result.Connection.TLS.CipherSuites = override.Connection.TLS.CipherSuites
	}
	if override.Connection.TLS.Renegotiation != "" {
		result.Connection.TLS.Renegotiation = override.Connection.TLS.Renegotiation
	}
	result.Connection.TLS.InsecureSkipVerify = override.Connection.TLS.InsecureSkipVerify
	result.Connection.TLS.ClientAuth = override.Connection.TLS.ClientAuth
	result.Connection.TLS.PreferServerCipherSuites = override.Connection.TLS.PreferServerCipherSuites
	result.Connection.TLS.SessionTicketsDisabled = override.Connection.TLS.SessionTicketsDisabled

	// 合并请求配置
	if len(override.Requests) > 0 {
		result.Requests = override.Requests
	}

	// 合并认证配置
	if override.Auth.Type != "" {
		result.Auth = override.Auth
	}

	// 合并上传配置
	if override.Upload.Enable {
		result.Upload = override.Upload
	}

	// 合并基准测试配置
	if override.Benchmark.Total > 0 {
		result.Benchmark.Total = override.Benchmark.Total
	}
	if override.Benchmark.Parallels > 0 {
		result.Benchmark.Parallels = override.Benchmark.Parallels
	}
	if override.Benchmark.Duration > 0 {
		result.Benchmark.Duration = override.Benchmark.Duration
	}
	if override.Benchmark.RampUp > 0 {
		result.Benchmark.RampUp = override.Benchmark.RampUp
	}
	if override.Benchmark.DataSize > 0 {
		result.Benchmark.DataSize = override.Benchmark.DataSize
	}
	if override.Benchmark.TTL > 0 {
		result.Benchmark.TTL = override.Benchmark.TTL
	}
	if override.Benchmark.ReadPercent >= 0 {
		result.Benchmark.ReadPercent = override.Benchmark.ReadPercent
	}
	if override.Benchmark.RandomKeys >= 0 {
		result.Benchmark.RandomKeys = override.Benchmark.RandomKeys
	}
	if override.Benchmark.TestCase != "" {
		result.Benchmark.TestCase = override.Benchmark.TestCase
	}
	if override.Benchmark.Timeout > 0 {
		result.Benchmark.Timeout = override.Benchmark.Timeout
	}
	if override.Benchmark.MaxRedirects >= 0 {
		result.Benchmark.MaxRedirects = override.Benchmark.MaxRedirects
	}
	if override.Benchmark.UserAgent != "" {
		result.Benchmark.UserAgent = override.Benchmark.UserAgent
	}
	result.Benchmark.FollowRedirects = override.Benchmark.FollowRedirects
	result.Benchmark.DisableCompression = override.Benchmark.DisableCompression
	result.Benchmark.EnableHTTP2 = override.Benchmark.EnableHTTP2

	return &result
}

// SaveHTTPConfig 保存HTTP配置到文件
func SaveHTTPConfig(config *HttpAdapterConfig, path string) error {
	// 包装配置
	wrapper := struct {
		HTTP *HttpAdapterConfig `yaml:"http"`
	}{
		HTTP: config,
	}

	data, err := yaml.Marshal(wrapper)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file %s: %w", path, err)
	}

	return nil
}
