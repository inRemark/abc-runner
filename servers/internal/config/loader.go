package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"abc-runner/servers/internal/common"
	"abc-runner/servers/pkg/http"
	"abc-runner/servers/pkg/interfaces"
	"gopkg.in/yaml.v3"
)

// HTTPConfigLoader HTTP配置加载器
type HTTPConfigLoader struct{}

// NewHTTPConfigLoader 创建HTTP配置加载器
func NewHTTPConfigLoader() *HTTPConfigLoader {
	return &HTTPConfigLoader{}
}

// LoadFromFile 从文件加载HTTP配置
func (loader *HTTPConfigLoader) LoadFromFile(configPath string) (*http.HTTPServerConfig, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := http.NewHTTPServerConfig()
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

// LoadConfig 实现interfaces.ConfigLoader接口
func (loader *HTTPConfigLoader) LoadConfig(configPath string) (interfaces.ServerConfig, error) {
	config, err := loader.LoadFromFile(configPath)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// LoadConfigWithDefaults 加载配置并应用默认值
func (loader *HTTPConfigLoader) LoadConfigWithDefaults(configPath string, defaults interfaces.ServerConfig) (interfaces.ServerConfig, error) {
	// 检查文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return defaults, nil
	}

	config, err := loader.LoadConfig(configPath)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// ValidateConfig 验证配置
func (loader *HTTPConfigLoader) ValidateConfig(config interfaces.ServerConfig) error {
	if httpConfig, ok := config.(*http.HTTPServerConfig); ok {
		return httpConfig.Validate()
	}
	return fmt.Errorf("invalid config type for HTTP loader")
}

// SaveToFile 保存配置到文件
func (loader *HTTPConfigLoader) SaveToFile(config *http.HTTPServerConfig, configPath string) error {
	// 确保目录存在
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := ioutil.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// TCPConfigLoader TCP配置加载器
type TCPConfigLoader struct{}

// NewTCPConfigLoader 创建TCP配置加载器
func NewTCPConfigLoader() *TCPConfigLoader {
	return &TCPConfigLoader{}
}

// LoadFromFile 从文件加载TCP配置
func (loader *TCPConfigLoader) LoadFromFile(configPath string) (interfaces.ServerConfig, error) {
	// 返回默认配置，实际应用中需要从文件加载
	return &struct {
		*common.BaseConfig
		MaxConnections int `yaml:"max_connections"`
	}{
		BaseConfig: &common.BaseConfig{
			Protocol: "tcp",
			Host:     "localhost",
			Port:     9090,
		},
		MaxConnections: 1000,
	}, nil
}

// LoadConfig 实现interfaces.ConfigLoader接口
func (loader *TCPConfigLoader) LoadConfig(configPath string) (interfaces.ServerConfig, error) {
	return loader.LoadFromFile(configPath)
}

// LoadConfigWithDefaults 加载配置并应用默认值
func (loader *TCPConfigLoader) LoadConfigWithDefaults(configPath string, defaults interfaces.ServerConfig) (interfaces.ServerConfig, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return defaults, nil
	}

	config, err := loader.LoadConfig(configPath)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// ValidateConfig 验证配置
func (loader *TCPConfigLoader) ValidateConfig(config interfaces.ServerConfig) error {
	return config.Validate()
}

// UDPConfigLoader UDP配置加载器
type UDPConfigLoader struct{}

// NewUDPConfigLoader 创建UDP配置加载器
func NewUDPConfigLoader() *UDPConfigLoader {
	return &UDPConfigLoader{}
}

// LoadFromFile 从文件加载UDP配置
func (loader *UDPConfigLoader) LoadFromFile(configPath string) (interfaces.ServerConfig, error) {
	// TODO: 实现UDP配置加载
	return nil, fmt.Errorf("UDP config loader not implemented")
}

// LoadConfig 实现interfaces.ConfigLoader接口
func (loader *UDPConfigLoader) LoadConfig(configPath string) (interfaces.ServerConfig, error) {
	return loader.LoadFromFile(configPath)
}

// LoadConfigWithDefaults 加载配置并应用默认值
func (loader *UDPConfigLoader) LoadConfigWithDefaults(configPath string, defaults interfaces.ServerConfig) (interfaces.ServerConfig, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return defaults, nil
	}

	config, err := loader.LoadConfig(configPath)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// ValidateConfig 验证配置
func (loader *UDPConfigLoader) ValidateConfig(config interfaces.ServerConfig) error {
	return config.Validate()
}

// GRPCConfigLoader gRPC配置加载器
type GRPCConfigLoader struct{}

// NewGRPCConfigLoader 创建gRPC配置加载器
func NewGRPCConfigLoader() *GRPCConfigLoader {
	return &GRPCConfigLoader{}
}

// LoadFromFile 从文件加载gRPC配置
func (loader *GRPCConfigLoader) LoadFromFile(configPath string) (interfaces.ServerConfig, error) {
	// TODO: 实现gRPC配置加载
	return nil, fmt.Errorf("gRPC config loader not implemented")
}

// LoadConfig 实现interfaces.ConfigLoader接口
func (loader *GRPCConfigLoader) LoadConfig(configPath string) (interfaces.ServerConfig, error) {
	return loader.LoadFromFile(configPath)
}

// LoadConfigWithDefaults 加载配置并应用默认值
func (loader *GRPCConfigLoader) LoadConfigWithDefaults(configPath string, defaults interfaces.ServerConfig) (interfaces.ServerConfig, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return defaults, nil
	}

	config, err := loader.LoadConfig(configPath)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// ValidateConfig 验证配置
func (loader *GRPCConfigLoader) ValidateConfig(config interfaces.ServerConfig) error {
	return config.Validate()
}

// UniversalConfigLoader 通用配置加载器
type UniversalConfigLoader struct {
	httpLoader *HTTPConfigLoader
	tcpLoader  *TCPConfigLoader
	udpLoader  *UDPConfigLoader
	grpcLoader *GRPCConfigLoader
}

// NewUniversalConfigLoader 创建通用配置加载器
func NewUniversalConfigLoader() *UniversalConfigLoader {
	return &UniversalConfigLoader{
		httpLoader: NewHTTPConfigLoader(),
		tcpLoader:  NewTCPConfigLoader(),
		udpLoader:  NewUDPConfigLoader(),
		grpcLoader: NewGRPCConfigLoader(),
	}
}

// LoadByProtocol 根据协议加载配置
func (loader *UniversalConfigLoader) LoadByProtocol(protocol, configPath string) (interfaces.ServerConfig, error) {
	switch protocol {
	case "http":
		return loader.httpLoader.LoadConfig(configPath)
	case "tcp":
		return loader.tcpLoader.LoadConfig(configPath)
	case "udp":
		return loader.udpLoader.LoadConfig(configPath)
	case "grpc":
		return loader.grpcLoader.LoadConfig(configPath)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", protocol)
	}
}

// DetectProtocol 从配置文件检测协议类型
func (loader *UniversalConfigLoader) DetectProtocol(configPath string) (string, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return "", fmt.Errorf("failed to read config file: %w", err)
	}

	// 尝试解析为通用结构体来获取协议字段
	var genericConfig struct {
		Protocol string `yaml:"protocol"`
	}

	if err := yaml.Unmarshal(data, &genericConfig); err != nil {
		return "", fmt.Errorf("failed to parse config file: %w", err)
	}

	if genericConfig.Protocol == "" {
		return "", fmt.Errorf("protocol field not found in config file")
	}

	return genericConfig.Protocol, nil
}

// ValidateConfigFile 验证配置文件
func (loader *UniversalConfigLoader) ValidateConfigFile(configPath string) error {
	protocol, err := loader.DetectProtocol(configPath)
	if err != nil {
		return err
	}

	config, err := loader.LoadByProtocol(protocol, configPath)
	if err != nil {
		return err
	}

	return config.Validate()
}