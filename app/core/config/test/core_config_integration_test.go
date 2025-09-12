package integration

import (
	"context"
	"os"
	"testing"

	"abc-runner/app/core/config"
	"abc-runner/app/core/interfaces"
)

// CoreConfigSourceAdapter 核心配置源适配器
// 用于将 config.ConfigSource 适配为 interfaces.ConfigSource
type CoreConfigSourceAdapter struct {
	source config.ConfigSource
}

// Priority 获取优先级
func (c *CoreConfigSourceAdapter) Priority() int {
	return c.source.Priority()
}

// CanLoad 检查是否可以加载
func (c *CoreConfigSourceAdapter) CanLoad() bool {
	return c.source.CanLoad()
}

// Load 加载配置
func (c *CoreConfigSourceAdapter) Load(ctx context.Context) (interfaces.Config, error) {
	return c.source.Load()
}

// GetProtocol 获取协议名称
func (c *CoreConfigSourceAdapter) GetProtocol() string {
	return "core"
}

// MockConfigSource 实现 config.ConfigSource 接口
type MockConfigSource struct{}

func (m *MockConfigSource) Priority() int {
	return 1
}

func (m *MockConfigSource) CanLoad() bool {
	return true
}

func (m *MockConfigSource) Load() (interfaces.Config, error) {
	return nil, nil
}

// MockInterfaceConfigSource 实现 interfaces.ConfigSource 接口
type MockInterfaceConfigSource struct{}

func (m *MockInterfaceConfigSource) Priority() int {
	return 1
}

func (m *MockInterfaceConfigSource) CanLoad() bool {
	return true
}

func (m *MockInterfaceConfigSource) Load(ctx context.Context) (interfaces.Config, error) {
	return nil, nil
}

func (m *MockInterfaceConfigSource) GetProtocol() string {
	return "mock"
}

// mockConfigSourceFactory 模拟配置源工厂
type mockConfigSourceFactory struct{}

func (m *mockConfigSourceFactory) CreateRedisConfigSource() interfaces.ConfigSource {
	// 创建一个模拟的Redis配置源
	mockSource := &MockConfigSource{}
	adapter := &CoreConfigSourceAdapter{source: mockSource}
	return adapter
}

func (m *mockConfigSourceFactory) CreateHttpConfigSource() interfaces.ConfigSource {
	// 创建一个模拟的HTTP配置源
	return &MockInterfaceConfigSource{}
}

func (m *mockConfigSourceFactory) CreateKafkaConfigSource() interfaces.ConfigSource {
	// 创建一个模拟的Kafka配置源
	return &MockInterfaceConfigSource{}
}

// Validate 实现unified.ConfigValidator接口
func (m *mockConfigSourceFactory) Validate(config interfaces.Config) error {
	// 模拟验证，总是返回nil表示验证通过
	return nil
}

// TestConfigManagerIntegration 测试配置管理器集成
func TestConfigManagerIntegration(t *testing.T) {
	manager := config.NewConfigManager(&mockConfigSourceFactory{})

	if manager == nil {
		t.Error("Failed to create config manager")
	}
}

// TestCoreConfigIntegration 测试核心配置集成
func TestCoreConfigIntegration(t *testing.T) {
	// 测试核心配置加载
	loader := config.NewCoreConfigLoader()

	defaultConfig := loader.GetDefaultConfig()
	if defaultConfig == nil {
		t.Error("Failed to get default core config")
	}
}

func TestCoreConfigIntegration_Load(t *testing.T) {
	// 创建临时核心配置文件用于测试
	tempConfig := `core:
  logging:
    level: "debug"
    format: "text"
    output: "file"
    file_path: "./integration-test-logs"
  reports:
    enabled: true
    formats: ["json", "csv"]
    output_dir: "./integration-test-reports"
    file_prefix: "integration-test-benchmark"
  monitoring:
    enabled: false
    metrics_interval: "10s"
  connection:
    timeout: "60s"
    keep_alive: "60s"
    max_idle_conns: 200
    idle_conn_timeout: "120s"`

	tempFile, err := os.CreateTemp("", "core_config_integration_test_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.Write([]byte(tempConfig)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tempFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	t.Run("ConfigManagerWithCoreConfig", func(t *testing.T) {
		manager := config.NewConfigManager(nil)

		// 加载核心配置
		err := manager.LoadCoreConfiguration(tempFile.Name())
		if err != nil {
			t.Fatalf("Failed to load core configuration: %v", err)
		}

		// 获取核心配置
		coreConfig := manager.GetCoreConfig()
		if coreConfig == nil {
			t.Fatal("Core config should not be nil")
		}

		if coreConfig.Core.Logging.Level != "debug" {
			t.Errorf("Expected logging level 'debug', got '%s'", coreConfig.Core.Logging.Level)
		}

		if coreConfig.Core.Reports.FilePrefix != "integration-test-benchmark" {
			t.Errorf("Expected file prefix 'integration-test-benchmark', got '%s'", coreConfig.Core.Reports.FilePrefix)
		}

		if coreConfig.Core.Monitoring.Enabled != false {
			t.Errorf("Expected monitoring disabled, got enabled")
		}
	})

	t.Run("ConfigManagerWithDefaultCoreConfig", func(t *testing.T) {
		manager := config.NewConfigManager(nil)

		// 不加载核心配置，应该使用默认配置
		coreConfig := manager.GetCoreConfig()
		if coreConfig == nil {
			t.Fatal("Core config should not be nil")
		}

		if coreConfig.Core.Logging.Level != "info" {
			t.Errorf("Expected default logging level 'info', got '%s'", coreConfig.Core.Logging.Level)
		}

		if coreConfig.Core.Reports.OutputDir != "./reports" {
			t.Errorf("Expected default output dir './reports', got '%s'", coreConfig.Core.Reports.OutputDir)
		}
	})
}