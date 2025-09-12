package config

import (
	"testing"
	"time"
)

// TestCoreConfigLoader 测试核心配置加载器
func TestCoreConfigLoader(t *testing.T) {
	loader := NewUnifiedCoreConfigLoader()

	if loader == nil {
		t.Fatal("Failed to create core config loader")
	}

	// 测试默认配置
	defaultConfig := loader.GetDefaultConfig()
	if defaultConfig == nil {
		t.Fatal("Failed to get default config")
	}

	// 验证默认配置值
	if defaultConfig.Core.Logging.Level != "info" {
		t.Errorf("Expected logging level 'info', got '%s'", defaultConfig.Core.Logging.Level)
	}

	if defaultConfig.Core.Reports.Enabled != true {
		t.Error("Expected reports to be enabled by default")
	}

	if defaultConfig.Core.Monitoring.MetricsInterval != 5*time.Second {
		t.Errorf("Expected metrics interval 5s, got %v", defaultConfig.Core.Monitoring.MetricsInterval)
	}
}

// TestCoreConfigLoaderFromFile 测试从文件加载核心配置
func TestCoreConfigLoaderFromFile(t *testing.T) {
	loader := NewUnifiedCoreConfigLoader()
	if loader == nil {
		t.Fatal("Failed to create core config loader")
	}

	// 测试加载不存在的文件（应该返回默认配置）
	config, err := loader.LoadFromFile("nonexistent.yaml")
	if err != nil {
		t.Errorf("LoadFromFile should not return error for nonexistent file: %v", err)
	}

	if config == nil {
		t.Fatal("LoadFromFile should return default config for nonexistent file")
	}
}
