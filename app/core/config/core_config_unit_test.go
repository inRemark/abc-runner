package config

import (
	"os"
	"testing"
)

func TestCoreConfigIntegration(t *testing.T) {
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
		manager := NewConfigManager()
		
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
		manager := NewConfigManager()
		
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