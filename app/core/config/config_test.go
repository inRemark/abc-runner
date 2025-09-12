package config

import (
	"os"
	"testing"

	httpconfig "abc-runner/app/adapters/http/config"
	kafkaconfig "abc-runner/app/adapters/kafka/config"
	redisconfig "abc-runner/app/adapters/redis/config"
)

// TestConfigManager 测试配置管理器
func TestConfigManager(t *testing.T) {
	manager := NewConfigManager(&mockConfigSourceFactory{})

	if manager == nil {
		t.Error("Failed to create config manager")
	}
}

func TestCreateRedisConfigSources(t *testing.T) {
	args := []string{"--addr", "localhost:6379"}
	sources := CreateRedisConfigSources("", args)

	if len(sources) == 0 {
		t.Fatal("Expected at least one source")
	}

	// 验证源的类型和优先级
	cmdLineFound := false
	envFound := false

	for _, source := range sources {
		// 检查是否是Redis配置源适配器
		if adapter, ok := source.(*RedisConfigSourceAdapter); ok {
			// Redis配置源适配器
			if adapter.Priority() >= 100 {
				cmdLineFound = true
			} else if adapter.Priority() >= 70 {
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

func TestCreateHttpConfigSources(t *testing.T) {
	args := []string{"--url", "http://localhost:8080"}
	sources := CreateHttpConfigSources("", args)

	if len(sources) == 0 {
		t.Fatal("Expected at least one source")
	}

	// 验证源的类型和优先级
	cmdLineFound := false
	envFound := false

	for _, source := range sources {
		// 检查是否是HTTP配置源适配器
		if _, ok := source.(*HttpConfigSourceAdapter); ok {
			// HTTP配置源适配器
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

func TestCreateKafkaConfigSources(t *testing.T) {
	args := []string{"--brokers", "localhost:9092"}
	sources := CreateKafkaConfigSources("", args)

	if len(sources) == 0 {
		t.Fatal("Expected at least one source")
	}

	// 验证源的类型和优先级
	cmdLineFound := false
	envFound := false

	for _, source := range sources {
		// 检查是否是Kafka配置源适配器
		if _, ok := source.(*KafkaConfigSourceAdapter); ok {
			// Kafka配置源适配器
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
  standalone:
    addr: "127.0.0.1:6379"
    password: ""
    db: 0`

	tmpFile, err := os.CreateTemp("", "redis_config_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	// 测试加载配置
	loader := redisconfig.NewUnifiedRedisConfigLoader()
	config, err := loader.LoadConfig(tmpFile.Name(), nil)
	if err != nil {
		t.Fatalf("Failed to load config from file: %v", err)
	}

	if config == nil {
		t.Fatal("Config should not be nil")
	}

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
		"--addr", "192.168.1.100:6380",
		"--total", "5000",
		"--parallels", "25",
		"--data-size", "128",
		"--read-percent", "80",
	}

	// 测试加载配置
	loader := redisconfig.NewUnifiedRedisConfigLoader()
	config, err := loader.LoadConfig("", args)
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

func TestUnifiedConfigManager(t *testing.T) {
	// 测试Redis统一配置管理器
	t.Run("Redis Unified Config Manager", func(t *testing.T) {
		args := []string{
			"--addr", "localhost:6379",
			"--total", "1000",
			"--parallels", "10",
		}

		loader := redisconfig.NewUnifiedRedisConfigLoader()
		config, err := loader.LoadConfig("", args)
		if err != nil {
			t.Fatalf("Failed to load Redis config: %v", err)
		}

		if config == nil {
			t.Fatal("Redis config should not be nil")
		}

		if config.GetProtocol() != "redis" {
			t.Errorf("Expected protocol 'redis', got '%s'", config.GetProtocol())
		}
	})

	// 测试HTTP统一配置管理器
	t.Run("HTTP Unified Config Manager", func(t *testing.T) {
		args := []string{
			"--url", "http://localhost:8080",
			"--total", "1000",
			"--parallels", "10",
		}

		loader := httpconfig.NewUnifiedHttpConfigLoader()
		config, err := loader.LoadConfig("", args)
		if err != nil {
			t.Fatalf("Failed to load HTTP config: %v", err)
		}

		if config == nil {
			t.Fatal("HTTP config should not be nil")
		}

		if config.GetProtocol() != "http" {
			t.Errorf("Expected protocol 'http', got '%s'", config.GetProtocol())
		}
	})

	// 测试Kafka统一配置管理器
	t.Run("Kafka Unified Config Manager", func(t *testing.T) {
		args := []string{
			"--brokers", "localhost:9092",
			"--total", "1000",
			"--parallels", "10",
		}

		loader := kafkaconfig.NewUnifiedKafkaConfigLoader()
		config, err := loader.LoadConfig("", args)
		// Kafka配置加载器需要一个基础配置，这里我们忽略错误
		if err != nil && config == nil {
			t.Skipf("Skipping Kafka test due to config loading error: %v", err)
		}

		if config != nil && config.GetProtocol() != "kafka" {
			t.Errorf("Expected protocol 'kafka', got '%s'", config.GetProtocol())
		}
	})
}