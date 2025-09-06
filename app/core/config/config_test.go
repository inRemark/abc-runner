package config

import (
	"os"
	"testing"
	"time"
)

func TestRedisConfig_GetProtocol(t *testing.T) {
	config := &RedisConfig{}

	// 测试默认协议
	if protocol := config.GetProtocol(); protocol != "redis" {
		t.Errorf("Expected protocol 'redis', got '%s'", protocol)
	}

	// 测试自定义协议
	config.Protocol = "custom"
	if protocol := config.GetProtocol(); protocol != "custom" {
		t.Errorf("Expected protocol 'custom', got '%s'", protocol)
	}
}

func TestRedisConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *RedisConfig
		wantErr bool
	}{
		{
			name: "valid standalone config",
			config: &RedisConfig{
				Mode: "standalone",
				Standalone: StandAloneInfo{
					Addr: "localhost:6379",
				},
				BenchMark: BenchmarkConfigImpl{
					Total:       1000,
					Parallels:   10,
					DataSize:    64,
					ReadPercent: 50,
				},
			},
			wantErr: false,
		},
		{
			name: "empty mode",
			config: &RedisConfig{
				Mode: "",
			},
			wantErr: true,
		},
		{
			name: "unsupported mode",
			config: &RedisConfig{
				Mode: "invalid",
			},
			wantErr: true,
		},
		{
			name: "empty standalone addr",
			config: &RedisConfig{
				Mode: "standalone",
				Standalone: StandAloneInfo{
					Addr: "",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("RedisConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBenchmarkConfigImpl_GetDefaults(t *testing.T) {
	config := &BenchmarkConfigImpl{}

	// 测试默认值
	if total := config.GetTotal(); total != 100000 {
		t.Errorf("Expected default total 100000, got %d", total)
	}

	if parallels := config.GetParallels(); parallels != 50 {
		t.Errorf("Expected default parallels 50, got %d", parallels)
	}

	if dataSize := config.GetDataSize(); dataSize != 3 {
		t.Errorf("Expected default data size 3, got %d", dataSize)
	}

	if ttl := config.GetTTL(); ttl != 120*time.Second {
		t.Errorf("Expected default TTL 120s, got %v", ttl)
	}
}

func TestBenchmarkConfigImpl_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *BenchmarkConfigImpl
		wantErr bool
	}{
		{
			name: "valid config",
			config: &BenchmarkConfigImpl{
				Total:       1000,
				Parallels:   10,
				DataSize:    64,
				ReadPercent: 50,
			},
			wantErr: false,
		},
		{
			name: "zero total",
			config: &BenchmarkConfigImpl{
				Total:     0,
				Parallels: 10,
				DataSize:  64,
			},
			wantErr: true,
		},
		{
			name: "negative parallels",
			config: &BenchmarkConfigImpl{
				Total:     1000,
				Parallels: -1,
				DataSize:  64,
			},
			wantErr: true,
		},
		{
			name: "invalid read percent",
			config: &BenchmarkConfigImpl{
				Total:       1000,
				Parallels:   10,
				DataSize:    64,
				ReadPercent: 150,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("BenchmarkConfigImpl.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestYAMLConfigSource(t *testing.T) {
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

	// 测试YAML配置源
	source := NewYAMLConfigSource(tmpFile.Name())

	// 测试CanLoad
	if !source.CanLoad() {
		t.Error("Expected CanLoad() to return true")
	}

	// 测试Load
	config, err := source.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	redisConfig, ok := config.(*RedisConfig)
	if !ok {
		t.Fatal("Expected *RedisConfig")
	}

	// 验证配置值
	if redisConfig.Mode != "standalone" {
		t.Errorf("Expected mode 'standalone', got '%s'", redisConfig.Mode)
	}

	if redisConfig.BenchMark.Total != 1000 {
		t.Errorf("Expected total 1000, got %d", redisConfig.BenchMark.Total)
	}

	if redisConfig.Standalone.Addr != "localhost:6379" {
		t.Errorf("Expected addr 'localhost:6379', got '%s'", redisConfig.Standalone.Addr)
	}
}

func TestEnvironmentConfigSource(t *testing.T) {
	// 设置环境变量
	os.Setenv("REDIS_RUNNER_MODE", "cluster")
	os.Setenv("REDIS_RUNNER_TOTAL", "2000")
	os.Setenv("REDIS_RUNNER_PARALLELS", "20")
	defer func() {
		os.Unsetenv("REDIS_RUNNER_MODE")
		os.Unsetenv("REDIS_RUNNER_TOTAL")
		os.Unsetenv("REDIS_RUNNER_PARALLELS")
	}()

	source := NewEnvironmentConfigSource("REDIS_RUNNER")

	// 测试CanLoad
	if !source.CanLoad() {
		t.Error("Expected CanLoad() to return true")
	}

	// 测试Load
	config, err := source.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	redisConfig, ok := config.(*RedisConfig)
	if !ok {
		t.Fatal("Expected *RedisConfig")
	}

	// 验证配置值
	if redisConfig.Mode != "cluster" {
		t.Errorf("Expected mode 'cluster', got '%s'", redisConfig.Mode)
	}

	if redisConfig.BenchMark.Total != 2000 {
		t.Errorf("Expected total 2000, got %d", redisConfig.BenchMark.Total)
	}

	if redisConfig.BenchMark.Parallels != 20 {
		t.Errorf("Expected parallels 20, got %d", redisConfig.BenchMark.Parallels)
	}
}

func TestCommandLineConfigSource(t *testing.T) {
	args := []string{
		"-h", "192.168.1.100",
		"-p", "6380",
		"-n", "5000",
		"-c", "25",
		"-d", "128",
		"-R", "80",
		"-cluster",
	}

	source := NewCommandLineConfigSource(args)

	// 测试CanLoad
	if !source.CanLoad() {
		t.Error("Expected CanLoad() to return true")
	}

	// 测试Load
	config, err := source.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	redisConfig, ok := config.(*RedisConfig)
	if !ok {
		t.Fatal("Expected *RedisConfig")
	}

	// 验证配置值
	if redisConfig.Mode != "cluster" {
		t.Errorf("Expected mode 'cluster', got '%s'", redisConfig.Mode)
	}

	if redisConfig.BenchMark.Total != 5000 {
		t.Errorf("Expected total 5000, got %d", redisConfig.BenchMark.Total)
	}

	if redisConfig.BenchMark.Parallels != 25 {
		t.Errorf("Expected parallels 25, got %d", redisConfig.BenchMark.Parallels)
	}

	if redisConfig.BenchMark.DataSize != 128 {
		t.Errorf("Expected data size 128, got %d", redisConfig.BenchMark.DataSize)
	}

	if redisConfig.BenchMark.ReadPercent != 80 {
		t.Errorf("Expected read percent 80, got %d", redisConfig.BenchMark.ReadPercent)
	}

	expectedAddr := "192.168.1.100:6380"
	if len(redisConfig.Cluster.Addrs) == 0 || redisConfig.Cluster.Addrs[0] != expectedAddr {
		t.Errorf("Expected cluster addr '%s', got %v", expectedAddr, redisConfig.Cluster.Addrs)
	}
}

func TestConfigManager(t *testing.T) {
	manager := NewConfigManager()

	// 创建测试配置源
	args := []string{"-h", "localhost", "-p", "6379", "-n", "1000"}
	cmdSource := NewCommandLineConfigSource(args)

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

func TestCreateDefaultSources(t *testing.T) {
	args := []string{"-h", "localhost"}
	sources := CreateDefaultSources("", args)

	if len(sources) == 0 {
		t.Fatal("Expected at least one source")
	}

	// 验证源的类型和优先级
	cmdLineFound := false
	envFound := false

	for _, source := range sources {
		switch source.(type) {
		case *CommandLineConfigSource:
			cmdLineFound = true
			if source.Priority() != 3 {
				t.Errorf("Expected command line source priority 3, got %d", source.Priority())
			}
		case *EnvironmentConfigSource:
			envFound = true
			if source.Priority() != 2 {
				t.Errorf("Expected environment source priority 2, got %d", source.Priority())
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
