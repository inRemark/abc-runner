package config

import (
	"os"
	"testing"
)

func TestMultiSourceConfigLoader(t *testing.T) {
	t.Run("Load From Default Source", func(t *testing.T) {
		loader := NewMultiSourceConfigLoader(NewDefaultConfigSource())
		config, err := loader.Load()
		if err != nil {
			t.Fatalf("Failed to load from default source: %v", err)
		}
		if config == nil {
			t.Error("Config should not be nil")
		}
		if config.GetProtocol() != "redis" {
			t.Errorf("Expected protocol 'redis', got '%s'", config.GetProtocol())
		}
	})

	t.Run("Priority Order", func(t *testing.T) {
		// 创建多个配置源
		defaultSource := NewDefaultConfigSource()
		envSource := NewEnvConfigSource("REDIS_TEST")
		
		loader := NewMultiSourceConfigLoader(defaultSource, envSource)
		sources := loader.GetSources()
		
		if len(sources) != 2 {
			t.Errorf("Expected 2 sources, got %d", len(sources))
		}
	})
}

func TestYAMLConfigSource(t *testing.T) {
	// 创建临时YAML文件
	yamlContent := `
redis:
  protocol: redis
  mode: standalone
  standalone:
    addr: "localhost:6379"
    password: "test123"
    db: 1
  benchmark:
    total: 2000
    parallels: 20
    data_size: 64
    ttl: 300
    read_percent: 60
    random_keys: 100
    case: "set_get"
  pool:
    pool_size: 15
    min_idle: 3
    max_idle: 10
`
	
	tmpFile, err := os.CreateTemp("", "redis_config_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	
	if _, err := tmpFile.WriteString(yamlContent); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	t.Run("Load Valid YAML", func(t *testing.T) {
		source := NewYAMLConfigSource(tmpFile.Name())
		
		if !source.CanLoad() {
			t.Error("Source should be able to load")
		}
		
		config, err := source.Load()
		if err != nil {
			t.Fatalf("Failed to load YAML config: %v", err)
		}
		
		if config.GetMode() != "standalone" {
			t.Errorf("Expected mode 'standalone', got '%s'", config.GetMode())
		}
		
		if config.Standalone.Addr != "localhost:6379" {
			t.Errorf("Expected addr 'localhost:6379', got '%s'", config.Standalone.Addr)
		}
		
		if config.Standalone.Password != "test123" {
			t.Errorf("Expected password 'test123', got '%s'", config.Standalone.Password)
		}
		
		if config.BenchMark.Total != 2000 {
			t.Errorf("Expected total 2000, got %d", config.BenchMark.Total)
		}
	})

	t.Run("Non-existent File", func(t *testing.T) {
		source := NewYAMLConfigSource("non-existent-file.yaml")
		
		if source.CanLoad() {
			t.Error("Source should not be able to load non-existent file")
		}
		
		_, err := source.Load()
		if err == nil {
			t.Error("Loading non-existent file should return error")
		}
	})
}

func TestEnvConfigSource(t *testing.T) {
	// 设置测试环境变量
	testEnvVars := map[string]string{
		"REDIS_TEST_PROTOCOL":     "redis",
		"REDIS_TEST_MODE":         "cluster",
		"REDIS_TEST_TOTAL":        "5000",
		"REDIS_TEST_PARALLELS":    "30",
		"REDIS_TEST_DATA_SIZE":    "128",
		"REDIS_TEST_TTL":          "600",
		"REDIS_TEST_READ_PERCENT": "70",
		"REDIS_TEST_RANDOM_KEYS":  "200",
		"REDIS_TEST_CASE":         "get",
		"REDIS_TEST_CLUSTER_ADDRS": "node1:6379,node2:6379,node3:6379",
		"REDIS_TEST_CLUSTER_PASSWORD": "cluster123",
	}

	// 设置环境变量
	for key, value := range testEnvVars {
		os.Setenv(key, value)
	}
	
	// 清理函数
	defer func() {
		for key := range testEnvVars {
			os.Unsetenv(key)
		}
	}()

	t.Run("Load From Environment", func(t *testing.T) {
		source := NewEnvConfigSource("REDIS_TEST")
		
		if !source.CanLoad() {
			t.Error("Source should be able to load from environment")
		}
		
		config, err := source.Load()
		if err != nil {
			t.Fatalf("Failed to load env config: %v", err)
		}
		
		if config.GetProtocol() != "redis" {
			t.Errorf("Expected protocol 'redis', got '%s'", config.GetProtocol())
		}
		
		if config.GetMode() != "cluster" {
			t.Errorf("Expected mode 'cluster', got '%s'", config.GetMode())
		}
		
		if config.BenchMark.Total != 5000 {
			t.Errorf("Expected total 5000, got %d", config.BenchMark.Total)
		}
		
		if len(config.Cluster.Addrs) != 3 {
			t.Errorf("Expected 3 cluster addresses, got %d", len(config.Cluster.Addrs))
		}
	})

	t.Run("No Environment Variables", func(t *testing.T) {
		source := NewEnvConfigSource("REDIS_NONEXISTENT")
		
		if source.CanLoad() {
			t.Error("Source should not be able to load without env vars")
		}
	})
}

func TestCommandLineConfigSource(t *testing.T) {
	t.Run("Parse Basic Arguments", func(t *testing.T) {
		args := []string{
			"-h", "redis.example.com",
			"-p", "6380",
			"-a", "mypassword",
			"-n", "10000",
			"-c", "50",
			"-d", "256",
			"-ttl", "7200",
			"-R", "80",
			"-r", "1000",
			"-t", "set_get_random",
		}
		
		source := NewCommandLineConfigSource(args)
		
		if !source.CanLoad() {
			t.Error("Source should be able to load from command line")
		}
		
		config, err := source.Load()
		if err != nil {
			t.Fatalf("Failed to load command line config: %v", err)
		}
		
		if config.Standalone.Addr != "redis.example.com:6380" {
			t.Errorf("Expected addr 'redis.example.com:6380', got '%s'", config.Standalone.Addr)
		}
		
		if config.Standalone.Password != "mypassword" {
			t.Errorf("Expected password 'mypassword', got '%s'", config.Standalone.Password)
		}
		
		if config.BenchMark.Total != 10000 {
			t.Errorf("Expected total 10000, got %d", config.BenchMark.Total)
		}
		
		if config.BenchMark.Parallels != 50 {
			t.Errorf("Expected parallels 50, got %d", config.BenchMark.Parallels)
		}
	})

	t.Run("Parse Cluster Arguments", func(t *testing.T) {
		args := []string{
			"--cluster",
			"--cluster-addrs", "node1:6379,node2:6379,node3:6379",
			"-a", "clusterpass",
		}
		
		source := NewCommandLineConfigSource(args)
		config, err := source.Load()
		if err != nil {
			t.Fatalf("Failed to load cluster config: %v", err)
		}
		
		if config.GetMode() != "cluster" {
			t.Errorf("Expected mode 'cluster', got '%s'", config.GetMode())
		}
		
		// 正确的集群地址数量应该是3个
		if len(config.Cluster.Addrs) != 3 {
			t.Errorf("Expected 3 cluster addresses, got %d", len(config.Cluster.Addrs))
		}
		
		if config.Cluster.Password != "clusterpass" {
			t.Errorf("Expected password 'clusterpass', got '%s'", config.Cluster.Password)
		}
	})

	t.Run("Empty Arguments", func(t *testing.T) {
		source := NewCommandLineConfigSource([]string{})
		
		if source.CanLoad() {
			t.Error("Source should not be able to load with empty args")
		}
	})
}

func TestCreateStandardLoader(t *testing.T) {
	t.Run("With Config Path", func(t *testing.T) {
		args := []string{"-n", "1000", "-c", "10"}
		loader := CreateStandardLoader("test-config.yaml", args)
		
		sources := loader.GetSources()
		if len(sources) < 2 {
			t.Errorf("Expected at least 2 sources, got %d", len(sources))
		}
	})

	t.Run("Without Config Path", func(t *testing.T) {
		args := []string{"-n", "2000", "-c", "20"}
		loader := CreateStandardLoader("", args)
		
		sources := loader.GetSources()
		if len(sources) < 3 {
			t.Errorf("Expected at least 3 sources, got %d", len(sources))
		}
	})

	t.Run("Load Configuration", func(t *testing.T) {
		args := []string{"-h", "localhost", "-p", "6379", "-n", "1000"}
		loader := CreateStandardLoader("", args)
		
		config, err := loader.Load()
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}
		
		if config == nil {
			t.Error("Config should not be nil")
		}
		
		if config.BenchMark.Total != 1000 {
			t.Errorf("Expected total 1000, got %d", config.BenchMark.Total)
		}
	})
}

func TestPriorityOrdering(t *testing.T) {
	defaultSource := NewDefaultConfigSource()
	envSource := NewEnvConfigSource("TEST")
	yamlSource := NewYAMLConfigSource("test.yaml")
	cmdSource := NewCommandLineConfigSource([]string{"-n", "9999"})

	priorities := []int{
		defaultSource.Priority(),
		envSource.Priority(),
		yamlSource.Priority(),
		cmdSource.Priority(),
	}

	// 验证命令行参数优先级最高
	if cmdSource.Priority() <= envSource.Priority() {
		t.Error("Command line should have higher priority than environment")
	}

	if envSource.Priority() <= yamlSource.Priority() {
		t.Error("Environment should have higher priority than YAML")
	}

	if yamlSource.Priority() <= defaultSource.Priority() {
		t.Error("YAML should have higher priority than default")
	}

	t.Logf("Priorities: default=%d, yaml=%d, env=%d, cmd=%d", 
		priorities[0], priorities[2], priorities[1], priorities[3])
}

// 基准测试
func BenchmarkConfigLoaders(b *testing.B) {
	b.Run("DefaultSource", func(b *testing.B) {
		source := NewDefaultConfigSource()
		for i := 0; i < b.N; i++ {
			_, _ = source.Load()
		}
	})

	b.Run("CommandLineSource", func(b *testing.B) {
		args := []string{"-h", "localhost", "-p", "6379", "-n", "1000", "-c", "10"}
		source := NewCommandLineConfigSource(args)
		for i := 0; i < b.N; i++ {
			_, _ = source.Load()
		}
	})

	b.Run("MultiSourceLoader", func(b *testing.B) {
		loader := NewMultiSourceConfigLoader(
			NewDefaultConfigSource(),
			NewCommandLineConfigSource([]string{"-n", "1000"}),
		)
		for i := 0; i < b.N; i++ {
			_, _ = loader.Load()
		}
	})
}