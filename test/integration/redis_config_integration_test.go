package integration_test

import (
	"testing"

	redisadapter "abc-runner/app/adapters/redis"
	redisconfig "abc-runner/app/adapters/redis/config"
	"abc-runner/app/core/interfaces"
)

func TestRedisConfigurationIntegration(t *testing.T) {
	t.Run("Redis Configuration Migration Integration", func(t *testing.T) {
		// 1. 使用新的Redis配置管理器加载配置
		manager := redisconfig.NewRedisConfigManager()
		
		// 从命令行参数加载配置
		args := []string{
			"-h", "localhost",
			"-p", "6379",
			"-n", "1000",
			"-c", "10",
			"-t", "set_get",
		}
		
		err := manager.LoadFromArgs(args)
		if err != nil {
			t.Fatalf("Failed to load configuration: %v", err)
		}
		
		// 2. 获取适配器配置
		adapter := manager.GetAdapter()
		
		// 验证适配器实现了统一接口
		var config interfaces.Config = adapter
		if config.GetProtocol() != "redis" {
			t.Errorf("Expected protocol 'redis', got '%s'", config.GetProtocol())
		}
		
		// 3. 验证配置内容
		benchmark := config.GetBenchmark()
		if benchmark.GetTotal() != 1000 {
			t.Errorf("Expected total 1000, got %d", benchmark.GetTotal())
		}
		
		if benchmark.GetParallels() != 10 {
			t.Errorf("Expected parallels 10, got %d", benchmark.GetParallels())
		}
		
		connection := config.GetConnection()
		addresses := connection.GetAddresses()
		if len(addresses) != 1 || addresses[0] != "localhost:6379" {
			t.Errorf("Expected address 'localhost:6379', got %v", addresses)
		}
	})
	
	t.Run("Redis Adapter with New Configuration", func(t *testing.T) {
		// 创建Redis适配器
		redisAdapter := redisadapter.NewRedisAdapter()
		
		// 使用新配置管理器创建配置
		manager := redisconfig.NewRedisConfigManager()
		args := []string{
			"-h", "127.0.0.1",
			"-p", "6379",
			"-n", "100",
			"-c", "5",
		}
		
		err := manager.LoadFromArgs(args)
		if err != nil {
			t.Fatalf("Failed to load configuration: %v", err)
		}
		
		config := manager.GetAdapter()
		
		// 验证适配器可以接受新配置（不需要真实连接）
		// 这里我们只测试配置验证，不进行实际连接
		err = redisAdapter.ValidateConfig(config)
		if err != nil {
			t.Errorf("Configuration validation failed: %v", err)
		}
		
		// 验证协议名称
		if redisAdapter.GetProtocolName() != "redis" {
			t.Errorf("Expected protocol name 'redis', got '%s'", redisAdapter.GetProtocolName())
		}
	})
	
	t.Run("Configuration Source Priority", func(t *testing.T) {
		// 测试配置源优先级
		defaultSource := redisconfig.NewDefaultConfigSource()
		envSource := redisconfig.NewEnvConfigSource("REDIS_TEST")
		cmdSource := redisconfig.NewCommandLineConfigSource([]string{"-n", "5000"})
		
		loader := redisconfig.NewMultiSourceConfigLoader(defaultSource, envSource, cmdSource)
		
		config, err := loader.Load()
		if err != nil {
			t.Fatalf("Failed to load configuration: %v", err)
		}
		
		// 命令行参数应该有最高优先级
		if config.BenchMark.Total != 5000 {
			t.Errorf("Expected total 5000 from command line, got %d", config.BenchMark.Total)
		}
	})
	
	t.Run("Configuration Validation Chain", func(t *testing.T) {
		// 测试完整的配置验证链
		manager := redisconfig.NewRedisConfigManager()
		
		// 测试有效配置
		validArgs := []string{
			"-h", "localhost",
			"-p", "6379",
			"-n", "1000",
			"-c", "10",
		}
		
		err := manager.LoadFromArgs(validArgs)
		if err != nil {
			t.Errorf("Valid configuration should load successfully: %v", err)
		}
		
		// 测试无效配置
		invalidArgs := []string{
			"-n", "0", // 无效的总请求数
			"-c", "-1", // 无效的并发数
		}
		
		err = manager.LoadFromArgs(invalidArgs)
		if err == nil {
			t.Error("Invalid configuration should fail validation")
		}
	})
	
	t.Run("Different Redis Modes", func(t *testing.T) {
		testCases := []struct {
			name string
			args []string
			mode string
		}{
			{
				name: "Standalone Mode",
				args: []string{"-h", "localhost", "-p", "6379"},
				mode: "standalone",
			},
			{
				name: "Cluster Mode", 
				args: []string{"--cluster", "--cluster-addrs", "node1:6379,node2:6379"},
				mode: "cluster",
			},
			{
				name: "Sentinel Mode",
				args: []string{"--mode", "sentinel", "--sentinel-master", "mymaster", "--sentinel-addrs", "sentinel1:26379"},
				mode: "sentinel",
			},
		}
		
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				manager := redisconfig.NewRedisConfigManager()
				err := manager.LoadFromArgs(tc.args)
				if err != nil {
					t.Fatalf("Failed to load %s configuration: %v", tc.mode, err)
				}
				
				config := manager.GetConfig()
				if config.GetMode() != tc.mode {
					t.Errorf("Expected mode '%s', got '%s'", tc.mode, config.GetMode())
				}
			})
		}
	})
}

func TestBackwardCompatibility(t *testing.T) {
	t.Run("Bridge Functions", func(t *testing.T) {
		// 测试桥接函数的向后兼容性
		args := []string{"-h", "localhost", "-p", "6379", "-n", "2000"}
		
		config, err := redisconfig.LoadRedisConfigFromArgs(args)
		if err != nil {
			t.Fatalf("Bridge function LoadRedisConfigFromArgs failed: %v", err)
		}
		
		// 验证返回的是统一配置接口
		if _, ok := config.(interfaces.Config); !ok {
			t.Error("Bridge function should return interfaces.Config")
		}
		
		benchmark := config.GetBenchmark()
		if benchmark.GetTotal() != 2000 {
			t.Errorf("Expected total 2000, got %d", benchmark.GetTotal())
		}
	})
	
	t.Run("Configuration Source Compatibility", func(t *testing.T) {
		// 测试配置源的兼容性
		sources := redisconfig.CreateRedisConfigSources("", []string{"-n", "3000", "-c", "15"})
		
		if len(sources) == 0 {
			t.Error("CreateRedisConfigSources should return at least one source")
		}
		
		config, err := redisconfig.LoadRedisConfigurationFromSources(sources...)
		if err != nil {
			t.Fatalf("LoadRedisConfigurationFromSources failed: %v", err)
		}
		
		benchmark := config.GetBenchmark()
		if benchmark.GetTotal() != 3000 {
			t.Errorf("Expected total 3000, got %d", benchmark.GetTotal())
		}
		
		if benchmark.GetParallels() != 15 {
			t.Errorf("Expected parallels 15, got %d", benchmark.GetParallels())
		}
	})
}

func BenchmarkConfigurationLoading(b *testing.B) {
	args := []string{"-h", "localhost", "-p", "6379", "-n", "1000", "-c", "10"}
	
	b.Run("New Redis Configuration Manager", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			manager := redisconfig.NewRedisConfigManager()
			_ = manager.LoadFromArgs(args)
		}
	})
	
	b.Run("Bridge Function", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = redisconfig.LoadRedisConfigFromArgs(args)
		}
	})
}