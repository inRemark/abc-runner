package config

import (
	"testing"

	"abc-runner/app/core/interfaces"
)

func TestRedisConfigAdapter(t *testing.T) {
	// 创建测试配置
	redisConfig := NewDefaultRedisConfig()
	redisConfig.Mode = "standalone"
	redisConfig.Standalone.Addr = "localhost:6379"
	redisConfig.Standalone.Password = "test123"
	redisConfig.BenchMark.Total = 5000
	redisConfig.BenchMark.Parallels = 25

	// 创建适配器
	adapter := NewRedisConfigAdapter(redisConfig)

	t.Run("GetProtocol", func(t *testing.T) {
		protocol := adapter.GetProtocol()
		if protocol != "redis" {
			t.Errorf("Expected protocol 'redis', got '%s'", protocol)
		}
	})

	t.Run("GetConnection", func(t *testing.T) {
		conn := adapter.GetConnection()
		addresses := conn.GetAddresses()
		if len(addresses) != 1 || addresses[0] != "localhost:6379" {
			t.Errorf("Expected address 'localhost:6379', got %v", addresses)
		}

		credentials := conn.GetCredentials()
		if credentials["password"] != "test123" {
			t.Errorf("Expected password 'test123', got '%s'", credentials["password"])
		}
	})

	t.Run("GetBenchmark", func(t *testing.T) {
		benchmark := adapter.GetBenchmark()
		if benchmark.GetTotal() != 5000 {
			t.Errorf("Expected total 5000, got %d", benchmark.GetTotal())
		}
		if benchmark.GetParallels() != 25 {
			t.Errorf("Expected parallels 25, got %d", benchmark.GetParallels())
		}
	})

	t.Run("Validate", func(t *testing.T) {
		err := adapter.Validate()
		if err != nil {
			t.Errorf("Validation failed: %v", err)
		}
	})

	t.Run("Clone", func(t *testing.T) {
		cloned := adapter.Clone()
		if cloned == adapter {
			t.Error("Clone should return a different instance")
		}

		if cloned.GetProtocol() != adapter.GetProtocol() {
			t.Error("Cloned config should have same protocol")
		}
	})
}

func TestExtractRedisConfig(t *testing.T) {
	// 创建一个Redis配置适配器
	originalConfig := NewDefaultRedisConfig()
	originalConfig.Mode = "cluster"
	originalConfig.Cluster.Addrs = []string{"node1:6379", "node2:6379", "node3:6379"}
	originalConfig.Cluster.Password = "cluster123"

	adapter := NewRedisConfigAdapter(originalConfig)

	// 提取配置
	extracted, err := ExtractRedisConfig(adapter)
	if err != nil {
		t.Fatalf("Failed to extract Redis config: %v", err)
	}

	// 验证提取的配置
	if extracted.GetMode() != "cluster" {
		t.Errorf("Expected mode 'cluster', got '%s'", extracted.GetMode())
	}

	if len(extracted.Cluster.Addrs) != 3 {
		t.Errorf("Expected 3 cluster addresses, got %d", len(extracted.Cluster.Addrs))
	}

	if extracted.Cluster.Password != "cluster123" {
		t.Errorf("Expected password 'cluster123', got '%s'", extracted.Cluster.Password)
	}
}

func TestAdaptRedisConfig(t *testing.T) {
	// 创建Redis配置
	redisConfig := NewDefaultRedisConfig()
	redisConfig.Mode = "sentinel"
	redisConfig.Sentinel.MasterName = "mymaster"
	redisConfig.Sentinel.Addrs = []string{"sentinel1:26379", "sentinel2:26379"}
	redisConfig.Sentinel.Password = "sentinel123"

	// 适配为统一接口
	config := AdaptRedisConfig(redisConfig)

	// 验证接口
	if _, ok := config.(interfaces.Config); !ok {
		t.Error("Adapted config should implement interfaces.Config")
	}

	// 验证内容
	conn := config.GetConnection()
	addresses := conn.GetAddresses()
	if len(addresses) != 2 {
		t.Errorf("Expected 2 sentinel addresses, got %d", len(addresses))
	}

	credentials := conn.GetCredentials()
	if credentials["master_name"] != "mymaster" {
		t.Errorf("Expected master_name 'mymaster', got '%s'", credentials["master_name"])
	}
}

func TestConfigAdapterInterfaces(t *testing.T) {
	redisConfig := NewDefaultRedisConfig()
	adapter := NewRedisConfigAdapter(redisConfig)

	t.Run("Config Interface", func(t *testing.T) {
		var config interfaces.Config = adapter
		if config.GetProtocol() != "redis" {
			t.Error("Config interface not working correctly")
		}
	})

	t.Run("Connection Interface", func(t *testing.T) {
		conn := adapter.GetConnection()
		var connectionConfig interfaces.ConnectionConfig = conn
		if len(connectionConfig.GetAddresses()) == 0 {
			t.Error("Connection interface not working correctly")
		}
	})

	t.Run("Benchmark Interface", func(t *testing.T) {
		benchmark := adapter.GetBenchmark()
		var benchmarkConfig interfaces.BenchmarkConfig = benchmark
		if benchmarkConfig.GetTotal() <= 0 {
			t.Error("Benchmark interface not working correctly")
		}
	})

	t.Run("Pool Interface", func(t *testing.T) {
		conn := adapter.GetConnection()
		pool := conn.GetPoolConfig()
		var poolConfig interfaces.PoolConfig = pool
		if poolConfig.GetPoolSize() <= 0 {
			t.Error("Pool interface not working correctly")
		}
	})
}

func TestRedisConfigValidation(t *testing.T) {
	t.Run("Valid Standalone Config", func(t *testing.T) {
		config := NewDefaultRedisConfig()
		config.Mode = "standalone"
		config.Standalone.Addr = "localhost:6379"
		
		adapter := NewRedisConfigAdapter(config)
		err := adapter.Validate()
		if err != nil {
			t.Errorf("Valid config should not fail validation: %v", err)
		}
	})

	t.Run("Invalid Standalone Config", func(t *testing.T) {
		config := NewDefaultRedisConfig()
		config.Mode = "standalone"
		config.Standalone.Addr = "" // 空地址应该失败
		
		adapter := NewRedisConfigAdapter(config)
		err := adapter.Validate()
		if err == nil {
			t.Error("Invalid config should fail validation")
		}
	})

	t.Run("Valid Cluster Config", func(t *testing.T) {
		config := NewDefaultRedisConfig()
		config.Mode = "cluster"
		config.Cluster.Addrs = []string{"node1:6379", "node2:6379"}
		
		adapter := NewRedisConfigAdapter(config)
		err := adapter.Validate()
		if err != nil {
			t.Errorf("Valid cluster config should not fail validation: %v", err)
		}
	})

	t.Run("Invalid Cluster Config", func(t *testing.T) {
		config := NewDefaultRedisConfig()
		config.Mode = "cluster"
		config.Cluster.Addrs = []string{} // 空地址列表应该失败
		
		adapter := NewRedisConfigAdapter(config)
		err := adapter.Validate()
		if err == nil {
			t.Error("Invalid cluster config should fail validation")
		}
	})
}

func TestDefaultValues(t *testing.T) {
	config := NewDefaultRedisConfig()
	adapter := NewRedisConfigAdapter(config)

	t.Run("Default Protocol", func(t *testing.T) {
		if adapter.GetProtocol() != "redis" {
			t.Errorf("Expected default protocol 'redis', got '%s'", adapter.GetProtocol())
		}
	})

	t.Run("Default Benchmark Values", func(t *testing.T) {
		benchmark := adapter.GetBenchmark()
		if benchmark.GetTotal() <= 0 {
			t.Error("Default total should be positive")
		}
		if benchmark.GetParallels() <= 0 {
			t.Error("Default parallels should be positive")
		}
		if benchmark.GetDataSize() <= 0 {
			t.Error("Default data size should be positive")
		}
		if benchmark.GetTTL() <= 0 {
			t.Error("Default TTL should be positive")
		}
	})

	t.Run("Default Pool Values", func(t *testing.T) {
		conn := adapter.GetConnection()
		pool := conn.GetPoolConfig()
		if pool.GetPoolSize() <= 0 {
			t.Error("Default pool size should be positive")
		}
		if pool.GetMinIdle() < 0 {
			t.Error("Default min idle should be non-negative")
		}
		if pool.GetConnectionTimeout() <= 0 {
			t.Error("Default connection timeout should be positive")
		}
	})
}

// 基准测试
func BenchmarkRedisConfigAdapter(b *testing.B) {
	config := NewDefaultRedisConfig()
	
	b.Run("Create Adapter", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NewRedisConfigAdapter(config)
		}
	})

	adapter := NewRedisConfigAdapter(config)
	
	b.Run("Get Protocol", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = adapter.GetProtocol()
		}
	})

	b.Run("Get Connection", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = adapter.GetConnection()
		}
	})

	b.Run("Get Benchmark", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = adapter.GetBenchmark()
		}
	})

	b.Run("Validate", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = adapter.Validate()
		}
	})
}