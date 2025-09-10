package config

import (
	"os"
	"testing"
)

func TestCoreConfigLoader(t *testing.T) {
	// 创建临时核心配置文件用于测试
	tempConfig := `core:
  logging:
    level: "debug"
    format: "text"
    output: "file"
    file_path: "./test-logs"
  reports:
    enabled: true
    formats: ["json", "csv"]
    output_dir: "./test-reports"
    file_prefix: "test-benchmark"
  monitoring:
    enabled: false
    metrics_interval: "10s"
  connection:
    timeout: "60s"
    keep_alive: "60s"
    max_idle_conns: 200
    idle_conn_timeout: "120s"`

	tempFile, err := os.CreateTemp("", "core_config_test_*.yaml")
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

	t.Run("LoadFromFile", func(t *testing.T) {
		loader := NewCoreConfigLoader()
		config, err := loader.LoadFromFile(tempFile.Name())
		if err != nil {
			t.Fatalf("Failed to load config from file: %v", err)
		}

		if config.Core.Logging.Level != "debug" {
			t.Errorf("Expected logging level 'debug', got '%s'", config.Core.Logging.Level)
		}

		if config.Core.Reports.FilePrefix != "test-benchmark" {
			t.Errorf("Expected file prefix 'test-benchmark', got '%s'", config.Core.Reports.FilePrefix)
		}

		if config.Core.Monitoring.Enabled != false {
			t.Errorf("Expected monitoring disabled, got enabled")
		}
	})

	t.Run("GetDefaultConfig", func(t *testing.T) {
		loader := NewCoreConfigLoader()
		config := loader.GetDefaultConfig()

		if config.Core.Logging.Level != "info" {
			t.Errorf("Expected default logging level 'info', got '%s'", config.Core.Logging.Level)
		}

		if config.Core.Reports.OutputDir != "./reports" {
			t.Errorf("Expected default output dir './reports', got '%s'", config.Core.Reports.OutputDir)
		}
	})

	t.Run("LoadNonExistentFile", func(t *testing.T) {
		loader := NewCoreConfigLoader()
		config, err := loader.LoadFromFile("non-existent-file.yaml")
		if err != nil {
			t.Fatalf("Should not return error for non-existent file: %v", err)
		}

		// Should return default config
		if config.Core.Logging.Level != "info" {
			t.Errorf("Expected default logging level 'info', got '%s'", config.Core.Logging.Level)
		}
	})
}