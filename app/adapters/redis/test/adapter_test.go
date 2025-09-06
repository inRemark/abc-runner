package test

import (
	"testing"
	"time"

	"redis-runner/app/adapters/redis/config"
	"redis-runner/app/adapters/redis/connection"
	"redis-runner/app/adapters/redis/metrics"
	"redis-runner/app/adapters/redis/operations"
)

// TestRedisConfig 测试Redis配置
func TestRedisConfig(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		cfg := config.NewDefaultRedisConfig()
		
		if cfg.GetProtocol() != "redis" {
			t.Errorf("Expected protocol 'redis', got '%s'", cfg.GetProtocol())
		}
		
		if cfg.GetMode() != "standalone" {
			t.Errorf("Expected mode 'standalone', got '%s'", cfg.GetMode())
		}
		
		if err := cfg.Validate(); err != nil {
			t.Errorf("Default config validation failed: %v", err)
		}
	})
	
	t.Run("ConfigValidation", func(t *testing.T) {
		cfg := config.NewDefaultRedisConfig()
		
		// 测试无效模式
		cfg.Mode = "invalid_mode"
		if err := cfg.Validate(); err == nil {
			t.Error("Expected validation error for invalid mode")
		}
		
		// 测试空地址
		cfg.Mode = "standalone"
		cfg.Standalone.Addr = ""
		if err := cfg.Validate(); err == nil {
			t.Error("Expected validation error for empty address")
		}
	})
	
	t.Run("ConfigClone", func(t *testing.T) {
		original := config.NewDefaultRedisConfig()
		original.Cluster.Addrs = []string{"addr1", "addr2"}
		
		cloned := original.Clone()
		
		// 修改克隆对象不应影响原对象
		cloned.Cluster.Addrs[0] = "modified"
		
		if original.Cluster.Addrs[0] == "modified" {
			t.Error("Clone did not perform deep copy")
		}
	})
}

// TestConfigLoader 测试配置加载器
func TestConfigLoader(t *testing.T) {
	t.Run("DefaultConfigSource", func(t *testing.T) {
		source := config.NewDefaultConfigSource()
		
		if !source.CanLoad() {
			t.Error("Default config source should always be able to load")
		}
		
		if source.Priority() != 1 {
			t.Errorf("Expected priority 1, got %d", source.Priority())
		}
		
		cfg, err := source.Load()
		if err != nil {
			t.Errorf("Failed to load default config: %v", err)
		}
		
		if cfg == nil {
			t.Error("Loaded config is nil")
		}
	})
	
	t.Run("MultiSourceLoader", func(t *testing.T) {
		loader := config.NewMultiSourceConfigLoader()
		loader.AddSource(config.NewDefaultConfigSource())
		
		cfg, err := loader.Load()
		if err != nil {
			t.Errorf("Failed to load config: %v", err)
		}
		
		if cfg == nil {
			t.Error("Loaded config is nil")
		}
	})
}

// TestOperations 测试操作模块
func TestOperations(t *testing.T) {
	t.Run("OperationFactory", func(t *testing.T) {
		factory := operations.NewOperationFactory()
		
		// 测试创建Get操作
		getOp, err := factory.Create(operations.OperationGet)
		if err != nil {
			t.Errorf("Failed to create Get operation: %v", err)
		}
		
		if getOp.GetType() != operations.OperationGet {
			t.Errorf("Expected operation type %s, got %s", operations.OperationGet, getOp.GetType())
		}
		
		// 测试创建不支持的操作
		_, err = factory.Create("unsupported")
		if err == nil {
			t.Error("Expected error for unsupported operation type")
		}
		
		// 测试列出支持的操作
		supportedOps := factory.ListSupportedOperations()
		if len(supportedOps) == 0 {
			t.Error("No supported operations found")
		}
	})
	
	t.Run("OperationBuilder", func(t *testing.T) {
		builder := operations.NewOperationBuilder()
		factory := builder.WithPublishChannel("test_channel").Build()
		
		pubOp, err := factory.Create(operations.OperationPublish)
		if err != nil {
			t.Errorf("Failed to create Publish operation: %v", err)
		}
		
		if pubOp == nil {
			t.Error("Created operation is nil")
		}
	})
	
	t.Run("OperationValidation", func(t *testing.T) {
		setOp := operations.NewSetOperation()
		
		// 测试有效参数
		validParams := operations.OperationParams{
			DataSize: 100,
			TTL:      time.Minute,
		}
		
		if err := setOp.Validate(validParams); err != nil {
			t.Errorf("Validation failed for valid params: %v", err)
		}
		
		// 测试无效参数
		invalidParams := operations.OperationParams{
			DataSize: 0, // 无效的数据大小
		}
		
		if err := setOp.Validate(invalidParams); err == nil {
			t.Error("Expected validation error for invalid params")
		}
	})
}

// TestMetricsCollector 测试指标收集器
func TestMetricsCollector(t *testing.T) {
	t.Run("BasicCollection", func(t *testing.T) {
		collector := metrics.NewMetricsCollector()
		
		// 模拟操作结果
		result := operations.OperationResult{
			Success:  true,
			IsRead:   true,
			Duration: time.Millisecond * 10,
		}
		
		collector.CollectOperation(result)
		
		metricsData := collector.GetMetrics()
		
		if metricsData["basic_metrics"] == nil {
			t.Error("Basic metrics not found")
		}
	})
	
	t.Run("MetricsReset", func(t *testing.T) {
		collector := metrics.NewMetricsCollector()
		
		// 添加一些数据
		result := operations.OperationResult{
			Success:  true,
			IsRead:   false,
			Duration: time.Millisecond * 5,
		}
		collector.CollectOperation(result)
		
		// 重置
		collector.Reset()
		
		summary := collector.GetSummary()
		if summary.BasicMetrics.TotalOperations != 0 {
			t.Error("Metrics not reset properly")
		}
	})
	
	t.Run("LatencyCalculation", func(t *testing.T) {
		collector := metrics.NewMetricsCollector()
		
		// 添加多个延迟样本
		latencies := []time.Duration{
			time.Millisecond * 1,
			time.Millisecond * 2,
			time.Millisecond * 3,
			time.Millisecond * 4,
			time.Millisecond * 5,
		}
		
		for _, latency := range latencies {
			result := operations.OperationResult{
				Success:  true,
				IsRead:   true,
				Duration: latency,
			}
			collector.CollectOperation(result)
		}
		
		summary := collector.GetSummary()
		
		if summary.LatencyMetrics.MinLatency != time.Millisecond {
			t.Errorf("Expected min latency %v, got %v", time.Millisecond, summary.LatencyMetrics.MinLatency)
		}
		
		if summary.LatencyMetrics.MaxLatency != time.Millisecond*5 {
			t.Errorf("Expected max latency %v, got %v", time.Millisecond*5, summary.LatencyMetrics.MaxLatency)
		}
	})
}

// TestMetricsReporter 测试指标报告器
func TestMetricsReporter(t *testing.T) {
	t.Run("JSONReporter", func(t *testing.T) {
		collector := metrics.NewMetricsCollector()
		
		// 添加测试数据
		result := operations.OperationResult{
			Success:  true,
			IsRead:   true,
			Duration: time.Millisecond * 10,
		}
		collector.CollectOperation(result)
		
		reporter := metrics.NewMetricsReporter(metrics.FormatJSON, "/tmp/test_metrics.json")
		
		metricsData := collector.GetMetrics()
		if err := reporter.Report(metricsData); err != nil {
			t.Errorf("Failed to generate JSON report: %v", err)
		}
	})
	
	t.Run("ConsoleReporter", func(t *testing.T) {
		collector := metrics.NewMetricsCollector()
		
		// 添加测试数据
		result := operations.OperationResult{
			Success:  true,
			IsRead:   false,
			Duration: time.Millisecond * 15,
		}
		collector.CollectOperation(result)
		
		reporter := metrics.NewMetricsReporter(metrics.FormatConsole, "console")
		
		metricsData := collector.GetMetrics()
		if err := reporter.Report(metricsData); err != nil {
			t.Errorf("Failed to generate console report: %v", err)
		}
	})
	
	t.Run("ReportBuilder", func(t *testing.T) {
		collector := metrics.NewMetricsCollector()
		
		// 添加测试数据
		result := operations.OperationResult{
			Success:  true,
			IsRead:   true,
			Duration: time.Millisecond * 20,
		}
		collector.CollectOperation(result)
		
		builder := metrics.NewReportBuilder(collector)
		builder.WithConsole()
		
		if err := builder.Generate(); err != nil {
			t.Errorf("Failed to generate report: %v", err)
		}
	})
}

// TestConnectionManager 测试连接管理器
func TestConnectionManager(t *testing.T) {
	t.Run("ConfigValidation", func(t *testing.T) {
		cfg := config.NewDefaultRedisConfig()
		cfg.Standalone.Addr = "invalid:addr:port"
		
		manager := connection.NewClientManager(cfg)
		
		// 连接应该失败
		err := manager.Connect()
		if err == nil {
			t.Error("Expected connection error for invalid address")
		}
	})
	
	t.Run("ClientFactory", func(t *testing.T) {
		factory := connection.NewClientFactory()
		
		// 测试创建单机客户端（会失败，因为没有真实的Redis服务器）
		standaloneConfig := config.StandAloneInfo{
			Addr:     "localhost:6379",
			Password: "",
			Db:       0,
		}
		
		_, err := factory.CreateStandaloneClient(standaloneConfig)
		// 预期会失败，因为没有Redis服务器运行
		if err == nil {
			t.Log("Connection succeeded (Redis server is running)")
		} else {
			t.Log("Connection failed as expected (no Redis server)")
		}
	})
}

// BenchmarkOperations 基准测试操作
func BenchmarkOperations(b *testing.B) {
	b.Run("OperationFactory", func(b *testing.B) {
		factory := operations.NewOperationFactory()
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := factory.Create(operations.OperationGet)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	
	b.Run("MetricsCollection", func(b *testing.B) {
		collector := metrics.NewMetricsCollector()
		result := operations.OperationResult{
			Success:  true,
			IsRead:   true,
			Duration: time.Millisecond,
		}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			collector.CollectOperation(result)
		}
	})
}

// TestIntegration 集成测试
func TestIntegration(t *testing.T) {
	t.Run("FullWorkflow", func(t *testing.T) {
		// 创建配置
		_ = config.NewDefaultRedisConfig()
		
		// 创建操作工厂
		factory := operations.NewOperationFactory()
		
		// 创建指标收集器
		collector := metrics.NewMetricsCollector()
		
		// 创建操作
		getOp, err := factory.Create(operations.OperationGet)
		if err != nil {
			t.Fatalf("Failed to create operation: %v", err)
		}
		
		// 模拟操作执行和指标收集
		result := operations.OperationResult{
			Success:  true,
			IsRead:   true,
			Duration: time.Millisecond * 5,
			ExtraData: map[string]interface{}{
				"operation_type": string(getOp.GetType()),
			},
		}
		
		collector.CollectOperation(result)
		
		// 生成报告
		builder := metrics.NewReportBuilder(collector)
		builder.WithConsole()
		
		if err := builder.Generate(); err != nil {
			t.Errorf("Failed to generate integration report: %v", err)
		}
		
		// 验证指标
		summary := collector.GetSummary()
		if summary.BasicMetrics.TotalOperations != 1 {
			t.Errorf("Expected 1 operation, got %d", summary.BasicMetrics.TotalOperations)
		}
	})
}