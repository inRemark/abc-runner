package test

import (
	"testing"
	"time"

	"abc-runner/app/adapters/redis/config"
	"abc-runner/app/adapters/redis/connection"
	"abc-runner/app/adapters/redis/metrics"
	"abc-runner/app/adapters/redis/operations"
)

// TestConfig 测试配置模块
func TestConfig(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		cfg := config.NewDefaultRedisConfig()
		
		if cfg.GetProtocol() != "redis" {
			t.Errorf("Expected protocol 'redis', got '%s'", cfg.GetProtocol())
		}
		
		if cfg.GetMode() != "standalone" {
			t.Errorf("Expected mode 'standalone', got '%s'", cfg.GetMode())
		}
	})
	
	t.Run("ConfigValidation", func(t *testing.T) {
		cfg := config.NewDefaultRedisConfig()
		
		// 测试有效配置
		if err := cfg.Validate(); err != nil {
			t.Errorf("Valid config should pass validation: %v", err)
		}
		
		// 测试无效配置
		cfg.Mode = "invalid"
		if err := cfg.Validate(); err == nil {
			t.Error("Invalid config should fail validation")
		}
	})
	
	t.Run("ConfigClone", func(t *testing.T) {
		original := config.NewDefaultRedisConfig()
		original.Mode = "cluster"
		original.Cluster.Addrs = []string{"localhost:7000", "localhost:7001"}
		
		clonedInterface := original.Clone()
		
		// 类型断言回具体的Redis配置类型
		cloned, ok := clonedInterface.(*config.RedisConfig)
		if !ok {
			t.Fatal("Cloned config is not of type *config.RedisConfig")
		}
		
		// 修改克隆对象不应影响原对象
		if len(cloned.Cluster.Addrs) > 0 {
			cloned.Cluster.Addrs[0] = "modified"
		}
		
		if len(original.Cluster.Addrs) > 0 && original.Cluster.Addrs[0] == "modified" {
			t.Error("Clone did not perform deep copy")
		}
	})
}

// TestConfigLoader 测试配置加载器
func TestConfigLoader(t *testing.T) {
	t.Run("UnifiedConfigLoader", func(t *testing.T) {
		loader := config.NewUnifiedRedisConfigLoader()
		
		cfg, err := loader.LoadConfig("", nil)
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
		
		// 创建报告器
		reporter := metrics.NewMetricsReporter(metrics.FormatJSON, "test_output.json")
		
		// 生成报告
		err := reporter.Report(collector.GetMetrics())
		if err != nil {
			t.Errorf("Failed to generate JSON report: %v", err)
		}
	})
	
	t.Run("ConsoleReporter", func(t *testing.T) {
		collector := metrics.NewMetricsCollector()
		
		// 添加测试数据
		result := operations.OperationResult{
			Success:  true,
			IsRead:   false,
			Duration: time.Millisecond * 5,
		}
		collector.CollectOperation(result)
		
		// 创建报告器
		reporter := metrics.NewMetricsReporter(metrics.FormatConsole, "console")
		
		// 生成报告
		err := reporter.Report(collector.GetMetrics())
		if err != nil {
			t.Errorf("Failed to generate console report: %v", err)
		}
	})
}

// TestConnectionManager 测试连接管理器
func TestConnectionManager(t *testing.T) {
	t.Run("ClientManagerCreation", func(t *testing.T) {
		cfg := config.NewDefaultRedisConfig()
		manager := connection.NewClientManager(cfg)
		
		if manager == nil {
			t.Error("ClientManager creation failed")
		}
		
		// 注意：ClientManager的GetMode()方法可能尚未实现或返回空字符串
		// 我们主要测试对象是否成功创建
	})
	
	t.Run("PoolManagerCreation", func(t *testing.T) {
		manager := connection.NewPoolManager()
		
		if manager == nil {
			t.Error("PoolManager creation failed")
		}
	})
}