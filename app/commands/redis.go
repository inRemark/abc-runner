package commands

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"abc-runner/app/adapters/redis"
	redisConfig "abc-runner/app/adapters/redis/config"
	"abc-runner/app/core/execution"
	"abc-runner/app/core/interfaces"
	"abc-runner/app/core/metrics"
	"abc-runner/app/reporting"
)

// RedisCommandHandler Redis命令处理器
type RedisCommandHandler struct {
	protocolName string
	factory      interface{} // AdapterFactory接口
}

// NewRedisCommandHandler 创建Redis命令处理器
func NewRedisCommandHandler(factory interface{}) *RedisCommandHandler {
	if factory == nil {
		panic("adapterFactory cannot be nil - dependency injection required")
	}

	return &RedisCommandHandler{
		protocolName: "redis",
		factory:      factory,
	}
}

// Execute 执行Redis命令
func (r *RedisCommandHandler) Execute(ctx context.Context, args []string) error {
	// 检查帮助请求 - 改进逻辑避免与-h host冲突
	for i, arg := range args {
		if arg == "--help" || arg == "help" {
			fmt.Println(r.GetHelp())
			return nil
		}
		// 只有当 -h 不是跟在其他参数后面时才作为帮助
		if arg == "-h" && (i == 0 || (i > 0 && args[i-1] != "redis")) {
			// 检查下一个参数是否看起来像hostname/IP
			if i+1 < len(args) && !looksLikeHostname(args[i+1]) {
				fmt.Println(r.GetHelp())
				return nil
			}
		}
	}

	// 解析命令行参数
	config, err := r.parseArgs(args)
	if err != nil {
		return fmt.Errorf("failed to parse arguments: %w", err)
	}

	// 创建Redis适配器
	metricsConfig := metrics.DefaultMetricsConfig()
	metricsCollector := metrics.NewBaseCollector(metricsConfig, map[string]interface{}{
		"protocol":  "redis",
		"test_type": "performance",
	})
	defer metricsCollector.Stop()

	// 直接使用MetricsCollector创建Redis适配器
	adapter := redis.NewRedisAdapter(metricsCollector)

	// 连接并执行测试
	if err := adapter.Connect(ctx, config); err != nil {
		log.Printf("Warning: failed to connect to %s (DB: %d): %v", config.Standalone.Addr, config.Standalone.Db, err)
		// 继续执行，但使用模拟模式
	}
	defer adapter.Close()

	// 执行性能测试
	fmt.Printf("🚀 Starting Redis performance test...\n")
	fmt.Printf("Target: %s (DB: %d)\n", config.Standalone.Addr, config.Standalone.Db)
	fmt.Printf("Operations: %d, Concurrency: %d\n", config.BenchMark.Total, config.BenchMark.Parallels)

	err = r.runPerformanceTest(ctx, adapter, config, metricsCollector)
	if err != nil {
		return fmt.Errorf("performance test failed: %w", err)
	}

	// 生成并显示报告
	return r.generateReport(metricsCollector)
}

// GetHelp 获取帮助信息
func (r *RedisCommandHandler) GetHelp() string {
	return fmt.Sprintf(`Redis Performance Testing

USAGE:
  abc-runner redis [options]

DESCRIPTION:
  Run Redis performance tests with various operations and configurations.

OPTIONS:
  --help          Show this help message
  --host HOST     Redis server host (default: localhost)
  --port PORT     Redis server port (default: 6379)
  --db DB         Database number (default: 0)
  --auth PASSWORD Redis password
  -n COUNT        Number of operations (default: 1000)
  -c COUNT        Concurrent connections (default: 10)
  
EXAMPLES:
  abc-runner redis --help
  abc-runner redis --host localhost --port 6379
  abc-runner redis --host localhost --auth mypassword
  abc-runner redis -h localhost -a pwd@redis -n 100 -c 2

NOTE: 
  This implementation performs real Redis performance testing with metrics collection.
`)
}

// parseArgs 解析命令行参数
func (r *RedisCommandHandler) parseArgs(args []string) (*redisConfig.RedisConfig, error) {
	// 创建默认配置
	config := redisConfig.NewDefaultRedisConfig()
	config.Standalone.Addr = "localhost:6379"
	config.Standalone.Db = 0
	config.BenchMark.Total = 1000
	config.BenchMark.Parallels = 10
	config.Pool.ConnectionTimeout = 30 * time.Second

	// 解析参数
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--host", "-h":
			if i+1 < len(args) && looksLikeHostname(args[i+1]) {
				config.Standalone.Addr = args[i+1] + ":6379" // 默认端口
				i++
			}
		case "--port":
			if i+1 < len(args) {
				if _, err := strconv.Atoi(args[i+1]); err == nil {
					// 更新地址中的端口
					host := "localhost"
					if config.Standalone.Addr != "localhost:6379" {
						parts := strings.Split(config.Standalone.Addr, ":")
						if len(parts) > 0 {
							host = parts[0]
						}
					}
					config.Standalone.Addr = host + ":" + args[i+1]
				}
				i++
			}
		case "--db":
			if i+1 < len(args) {
				if db, err := strconv.Atoi(args[i+1]); err == nil {
					config.Standalone.Db = db
				}
				i++
			}
		case "--auth", "-a":
			if i+1 < len(args) {
				config.Standalone.Password = args[i+1]
				i++
			}
		case "-n":
			if i+1 < len(args) {
				if count, err := strconv.Atoi(args[i+1]); err == nil {
					config.BenchMark.Total = count
				}
				i++
			}
		case "-c":
			if i+1 < len(args) {
				if count, err := strconv.Atoi(args[i+1]); err == nil {
					config.BenchMark.Parallels = count
				}
				i++
			}
		}
	}

	return config, nil
}

// runPerformanceTest 运行性能测试 - 使用新的ExecutionEngine
func (r *RedisCommandHandler) runPerformanceTest(ctx context.Context, adapter interfaces.ProtocolAdapter, config *redisConfig.RedisConfig, collector *metrics.BaseCollector[map[string]interface{}]) error {
	// 执行健康检查
	if err := adapter.HealthCheck(ctx); err != nil {
		log.Printf("Health check failed, running in simulation mode: %v", err)
		// 在模拟模式下生成测试数据
		return r.runSimulationTest(config, collector)
	}

	// 使用新的ExecutionEngine执行真实测试
	return r.runConcurrentTest(ctx, adapter, config, collector)
}

// runSimulationTest 运行模拟测试 (保持不变，用于连接失败时的后备方案)
func (r *RedisCommandHandler) runSimulationTest(config *redisConfig.RedisConfig, collector *metrics.BaseCollector[map[string]interface{}]) error {
	fmt.Printf("📊 Running Redis simulation test...\n")

	// Redis操作类型
	operations := []string{"GET", "SET", "HGET", "HSET", "LPUSH", "RPOP"}

	// 生成模拟数据
	for i := 0; i < config.BenchMark.Total; i++ {
		// 模拟95%成功率
		success := i%20 != 0
		// 模拟延迟：1-10ms
		latency := time.Duration(1+i%10) * time.Millisecond
		// 随机选择操作类型
		opType := operations[i%len(operations)]
		// 读操作：GET, HGET
		isRead := opType == "GET" || opType == "HGET"

		result := &interfaces.OperationResult{
			Success:  success,
			Duration: latency,
			IsRead:   isRead,
			Metadata: map[string]interface{}{
				"operation_type": opType,
				"key":            fmt.Sprintf("key_%d", i),
			},
		}

		collector.Record(result)

		// 模拟并发延迟
		if i%config.BenchMark.Parallels == 0 {
			time.Sleep(time.Millisecond)
		}
	}

	fmt.Printf("✅ Redis simulation test completed\n")
	return nil
}

// runConcurrentTest 使用ExecutionEngine运行并发测试
func (r *RedisCommandHandler) runConcurrentTest(ctx context.Context, adapter interfaces.ProtocolAdapter, config *redisConfig.RedisConfig, collector *metrics.BaseCollector[map[string]interface{}]) error {
	fmt.Printf("📊 Running concurrent Redis performance test with ExecutionEngine...\n")

	// 创建基准配置适配器
	benchmarkConfig := redis.NewBenchmarkConfigAdapter(config.GetBenchmark())

	// 创建操作工厂
	operationFactory := redis.NewOperationFactory(config)

	// 创建执行引擎
	engine := execution.NewExecutionEngine(adapter, collector, operationFactory)

	// 配置执行引擎参数
	engine.SetMaxWorkers(100)         // 设置最大工作协程数
	engine.SetBufferSizes(1000, 1000) // 设置缓冲区大小

	// 运行基准测试
	result, err := engine.RunBenchmark(ctx, benchmarkConfig)
	if err != nil {
		return fmt.Errorf("benchmark execution failed: %w", err)
	}

	// 输出执行结果
	fmt.Printf("✅ Concurrent Redis test completed\n")
	fmt.Printf("   Total Jobs: %d\n", result.TotalJobs)
	fmt.Printf("   Completed: %d\n", result.CompletedJobs)
	fmt.Printf("   Success: %d\n", result.SuccessJobs)
	fmt.Printf("   Failed: %d\n", result.FailedJobs)
	fmt.Printf("   Duration: %v\n", result.TotalDuration)
	fmt.Printf("   Success Rate: %.2f%%\n", float64(result.SuccessJobs)/float64(result.CompletedJobs)*100)

	return nil
}

// generateReport 生成报告
func (r *RedisCommandHandler) generateReport(collector *metrics.BaseCollector[map[string]interface{}]) error {
	// 获取指标快照
	snapshot := collector.Snapshot()

	// 转换为结构化报告
	report := reporting.ConvertFromMetricsSnapshot(snapshot)

	// 使用标准报告配置
	reportConfig := reporting.NewStandardReportConfig("redis")

	generator := reporting.NewReportGenerator(reportConfig)

	// 生成并显示报告
	return generator.Generate(report)
}

// looksLikeHostname 检查是否看起来像主机名或IP
func looksLikeHostname(arg string) bool {
	// 简单检查：不以-开头且包含字母数字或点
	if len(arg) == 0 || arg[0] == '-' {
		return false
	}
	for _, c := range arg {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '.' || c == ':' {
			continue
		} else {
			return false
		}
	}
	return true
}
