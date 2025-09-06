package config

import (
	"os"
	"testing"
)

func TestConfigManager(t *testing.T) {
	manager := NewConfigManager()

	// 创建测试配置源
	args := []string{"-h", "localhost", "-p", "6379", "-n", "1000"}
	cmdSource := NewRedisCommandLineConfigSource(args)

	// 测试加载配置
	err := manager.LoadConfiguration(cmdSource)
	if err != nil {
		t.Fatalf("Failed to load configuration: %v", err)
	}

	// 获取配置
	config := manager.GetConfig()
	if config == nil {
		t.Fatal("Expected non-nil config")
	}

	// 验证配置
	if config.GetProtocol() != "redis" {
		t.Errorf("Expected protocol 'redis', got '%s'", config.GetProtocol())
	}
}

func TestCreateRedisConfigSources(t *testing.T) {
	args := []string{"-h", "localhost"}
	sources := CreateRedisConfigSources("", args)

	if len(sources) == 0 {
		t.Fatal("Expected at least one source")
	}

	// 验证源的类型和优先级
	cmdLineFound := false
	envFound := false

	for _, source := range sources {
		switch source.(type) {
		case *RedisConfigSourceBridge:
			// Redis配置源桥接器
			if source.Priority() >= 100 {
				cmdLineFound = true
			} else if source.Priority() >= 70 {
				envFound = true
			}
		}
	}

	if !cmdLineFound {
		t.Error("Expected command line source to be found")
	}

	if !envFound {
		t.Error("Expected environment source to be found")
	}
}

func TestLoadRedisConfigFromFile(t *testing.T) {
	// 创建临时配置文件
	content := `redis:
  mode: "standalone"
  benchmark:
    total: 1000
    parallels: 5
    data_size: 32
    ttl: 60
    read_percent: 70
    random_keys: 100
    case: "set_get_random"
  pool:
    pool_size: 5
    min_idle: 1
  standalone:
    addr: "localhost:6379"
    password: "test"
    db: 0`

	tmpFile, err := os.CreateTemp("", "test_config_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// 测试加载配置
	config, err := LoadRedisConfigFromFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// 验证配置值
	if config.GetProtocol() != "redis" {
		t.Errorf("Expected protocol 'redis', got '%s'", config.GetProtocol())
	}

	benchmark := config.GetBenchmark()
	if benchmark.GetTotal() != 1000 {
		t.Errorf("Expected total 1000, got %d", benchmark.GetTotal())
	}
}

func TestLoadRedisConfigFromArgs(t *testing.T) {
	args := []string{
		"-h", "192.168.1.100",
		"-p", "6380",
		"-n", "5000",
		"-c", "25",
		"-d", "128",
		"-R", "80",
	}

	// 测试加载配置
	config, err := LoadRedisConfigFromArgs(args)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// 验证配置值
	if config.GetProtocol() != "redis" {
		t.Errorf("Expected protocol 'redis', got '%s'", config.GetProtocol())
	}

	benchmark := config.GetBenchmark()
	if benchmark.GetTotal() != 5000 {
		t.Errorf("Expected total 5000, got %d", benchmark.GetTotal())
	}

	if benchmark.GetParallels() != 25 {
		t.Errorf("Expected parallels 25, got %d", benchmark.GetParallels())
	}

	if benchmark.GetDataSize() != 128 {
		t.Errorf("Expected data size 128, got %d", benchmark.GetDataSize())
	}

	if benchmark.GetReadPercent() != 80 {
		t.Errorf("Expected read percent 80, got %d", benchmark.GetReadPercent())
	}
}
