package config

import (
	"os"
	"testing"
)

func TestLoadHttpConfigFromFile(t *testing.T) {
	// 创建临时配置文件
	content := `protocol: "http"
connection:
  base_url: "http://localhost:8080"
  timeout: "30s"
  max_idle_conns: 10
  max_conns_per_host: 10
requests:
  - method: "GET"
    path: "/api/users"
benchmark:
  total: 1000
  parallels: 5
  method: "GET"
  path: "/api/users"
auth:
  type: "none"`

	tmpFile, err := os.CreateTemp("", "http_config_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	// 测试加载配置
	loader := NewUnifiedHttpConfigLoader()
	config, err := loader.LoadConfig(tmpFile.Name(), nil)
	if err != nil {
		t.Fatalf("Failed to load config from file: %v", err)
	}

	if config == nil {
		t.Fatal("Config should not be nil")
	}

	if config.GetProtocol() != "http" {
		t.Errorf("Expected protocol 'http', got '%s'", config.GetProtocol())
	}

	benchmark := config.GetBenchmark()
	if benchmark.GetTotal() != 1000 {
		t.Errorf("Expected total 1000, got %d", benchmark.GetTotal())
	}
}

func TestLoadHttpConfigFromArgs(t *testing.T) {
	args := []string{
		"--url", "http://192.168.1.100:8080",
		"--total", "5000",
		"--parallels", "25",
		"--method", "POST",
		"--path", "/api/data",
	}

	// 测试加载配置
	loader := NewUnifiedHttpConfigLoader()
	config, err := loader.LoadConfig("", args)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// 验证配置值
	if config.GetProtocol() != "http" {
		t.Errorf("Expected protocol 'http', got '%s'", config.GetProtocol())
	}

	httpConfig, ok := config.(*HttpAdapterConfig)
	if !ok {
		t.Fatal("Config should be HttpAdapterConfig")
	}

	if httpConfig.Connection.BaseURL != "http://192.168.1.100:8080" {
		t.Errorf("Expected base URL 'http://192.168.1.100:8080', got '%s'", httpConfig.Connection.BaseURL)
	}

	benchmark := config.GetBenchmark()
	if benchmark.GetTotal() != 5000 {
		t.Errorf("Expected total 5000, got %d", benchmark.GetTotal())
	}

	if benchmark.GetParallels() != 25 {
		t.Errorf("Expected parallels 25, got %d", benchmark.GetParallels())
	}

	if httpConfig.Benchmark.Method != "POST" {
		t.Errorf("Expected method 'POST', got '%s'", httpConfig.Benchmark.Method)
	}

	if httpConfig.Benchmark.Path != "/api/data" {
		t.Errorf("Expected path '/api/data', got '%s'", httpConfig.Benchmark.Path)
	}
}

func TestHttpConfigValidation(t *testing.T) {
	config := LoadDefaultHttpConfig()

	// 测试无效的基础URL
	config.Connection.BaseURL = ""
	err := config.Validate()
	if err == nil {
		t.Error("Empty base URL should fail validation")
	}

	// 测试无效的总请求数
	config.Connection.BaseURL = "http://localhost:8080"
	config.Benchmark.Total = 0
	err = config.Validate()
	if err == nil {
		t.Error("Zero total should fail validation")
	}
}
