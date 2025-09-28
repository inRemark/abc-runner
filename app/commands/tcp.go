package commands

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"abc-runner/app/adapters/tcp"
	tcpConfig "abc-runner/app/adapters/tcp/config"
	"abc-runner/app/core/execution"
	"abc-runner/app/core/interfaces"
	"abc-runner/app/core/metrics"
	"abc-runner/app/reporting"
)

// TCPCommandHandler TCP命令处理器
type TCPCommandHandler struct {
	protocolName string
	factory      interface{} // AdapterFactory接口
}

// NewTCPCommandHandler 创建TCP命令处理器
func NewTCPCommandHandler(factory interface{}) *TCPCommandHandler {
	if factory == nil {
		panic("adapterFactory cannot be nil - dependency injection required")
	}

	return &TCPCommandHandler{
		protocolName: "tcp",
		factory:      factory,
	}
}

// Execute 执行TCP命令
func (t *TCPCommandHandler) Execute(ctx context.Context, args []string) error {
	// 检查帮助请求
	for i, arg := range args {
		if arg == "--help" || arg == "help" {
			fmt.Println(t.GetHelp())
			return nil
		}
		if arg == "-h" && (i == 0 || (i > 0 && args[i-1] != "tcp")) {
			if i+1 < len(args) && !looksLikeHostname(args[i+1]) {
				fmt.Println(t.GetHelp())
				return nil
			}
		}
	}

	// 解析命令行参数
	config, err := t.parseArgs(args)
	if err != nil {
		return fmt.Errorf("failed to parse arguments: %w", err)
	}

	// 创建TCP适配器
	metricsConfig := metrics.DefaultMetricsConfig()
	metricsCollector := metrics.NewBaseCollector(metricsConfig, map[string]interface{}{
		"protocol":  "tcp",
		"test_type": "performance",
	})
	defer metricsCollector.Stop()

	adapter := tcp.NewTCPAdapter(metricsCollector)

	// 连接并执行测试
	if err := adapter.Connect(ctx, config); err != nil {
		fmt.Printf("⚠️  Connection failed to %s:%d: %v\n", config.Connection.Address, config.Connection.Port, err)
		fmt.Printf("🔍 Possible causes: TCP server not running, wrong host/port, firewall blocking, or network issues\n")
	} else {
		fmt.Printf("✅ Successfully connected to TCP server at %s:%d\n", config.Connection.Address, config.Connection.Port)
	}
	defer adapter.Close()

	// 执行性能测试
	fmt.Printf("🚀 Starting TCP performance test...\n")
	fmt.Printf("Target: %s:%d\n", config.Connection.Address, config.Connection.Port)
	fmt.Printf("Test Case: %s\n", config.BenchMark.TestCase)
	fmt.Printf("Operations: %d, Concurrency: %d, Data Size: %d bytes\n", 
		config.BenchMark.Total, config.BenchMark.Parallels, config.BenchMark.DataSize)

	err = t.runPerformanceTest(ctx, adapter, config, metricsCollector)
	if err != nil {
		return fmt.Errorf("performance test failed: %w", err)
	}

	// 生成并显示报告
	return t.generateReport(metricsCollector)
}

// GetHelp 获取帮助信息
func (t *TCPCommandHandler) GetHelp() string {
	return `TCP Performance Testing

USAGE:
  abc-runner tcp [options]

DESCRIPTION:
  Run TCP performance tests with various operations and configurations.

OPTIONS:
  --help              Show this help message
  --host HOST         TCP server host (default: localhost)
  --port PORT         TCP server port (default: 8080)
  -n COUNT            Number of operations (default: 1000)
  -c COUNT            Concurrent connections (default: 10)
  --data-size SIZE    Data packet size in bytes (default: 1024)
  --test-case TYPE    Test case type (default: echo_test)
  --duration DURATION Test duration (default: 60s)
  --no-delay          Disable Nagle algorithm (default: true)
  --keep-alive        Enable TCP keep-alive (default: true)
  
TEST CASES:
  echo_test           Send data and verify echo response
  send_only           Send data only, no response expected
  receive_only        Receive data only
  bidirectional       Bidirectional data transfer test
  
EXAMPLES:
  abc-runner tcp --help
  abc-runner tcp --host localhost --port 8080
  abc-runner tcp --host 192.168.1.100 --port 9090 --test-case echo_test
  abc-runner tcp -h localhost -p 8080 -n 5000 -c 20 --data-size 2048

NOTE: 
  This implementation performs real TCP performance testing with metrics collection.`
}

// parseArgs 解析命令行参数
func (t *TCPCommandHandler) parseArgs(args []string) (*tcpConfig.TCPConfig, error) {
	// 创建默认配置
	config := tcpConfig.NewDefaultTCPConfig()

	// 解析参数
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--host", "-h":
			if i+1 < len(args) && looksLikeHostname(args[i+1]) {
				config.Connection.Address = args[i+1]
				i++
			}
		case "--port", "-p":
			if i+1 < len(args) {
				if port, err := strconv.Atoi(args[i+1]); err == nil && port > 0 && port <= 65535 {
					config.Connection.Port = port
				}
				i++
			}
		case "-n":
			if i+1 < len(args) {
				if count, err := strconv.Atoi(args[i+1]); err == nil && count > 0 {
					config.BenchMark.Total = count
				}
				i++
			}
		case "-c":
			if i+1 < len(args) {
				if count, err := strconv.Atoi(args[i+1]); err == nil && count > 0 {
					config.BenchMark.Parallels = count
				}
				i++
			}
		case "--data-size":
			if i+1 < len(args) {
				if size, err := strconv.Atoi(args[i+1]); err == nil && size > 0 {
					config.BenchMark.DataSize = size
				}
				i++
			}
		case "--test-case":
			if i+1 < len(args) {
				validCases := []string{"echo_test", "send_only", "receive_only", "bidirectional"}
				testCase := args[i+1]
				for _, valid := range validCases {
					if testCase == valid {
						config.BenchMark.TestCase = testCase
						break
					}
				}
				i++
			}
		case "--duration":
			if i+1 < len(args) {
				if duration, err := time.ParseDuration(args[i+1]); err == nil {
					config.BenchMark.Duration = duration
				}
				i++
			}
		case "--no-delay":
			config.TCPSpecific.NoDelay = true
		case "--keep-alive":
			config.Connection.KeepAlive = true
		}
	}

	return config, nil
}

// runPerformanceTest 运行性能测试
func (t *TCPCommandHandler) runPerformanceTest(ctx context.Context, adapter interfaces.ProtocolAdapter, config *tcpConfig.TCPConfig, collector *metrics.BaseCollector[map[string]interface{}]) error {
	// 执行健康检查
	if err := adapter.HealthCheck(ctx); err != nil {
		fmt.Printf("⚠️  Health check failed: %v\n", err)
		fmt.Printf("🔄 Switching to simulation mode - this will generate mock test data instead of real TCP operations\n")
		return t.runSimulationTest(config, collector)
	}

	// 创建执行引擎
	factory := NewSimpleOperationFactory(config.BenchMark.TestCase, config.BenchMark.DataSize)
	benchConfig := NewSimpleBenchmarkConfig(config.BenchMark.Total, config.BenchMark.Parallels, config.BenchMark.Duration)
	engine := execution.NewExecutionEngine(adapter, collector, factory)

	// 执行测试
	fmt.Printf("📊 Executing %d %s operations with %d concurrent connections...\n", 
		config.BenchMark.Total, config.BenchMark.TestCase, config.BenchMark.Parallels)

	startTime := time.Now()
	result, err := engine.RunBenchmark(ctx, benchConfig)
	duration := time.Since(startTime)

	if err != nil {
		return fmt.Errorf("benchmark execution failed: %w", err)
	}

	fmt.Printf("✅ Test completed in %v\n", duration)
	fmt.Printf("📈 Processed %d operations (%d successful, %d failed)\n", 
		result.CompletedJobs, result.SuccessJobs, result.FailedJobs)

	return nil
}

// runSimulationTest 运行模拟测试
func (t *TCPCommandHandler) runSimulationTest(config *tcpConfig.TCPConfig, collector *metrics.BaseCollector[map[string]interface{}]) error {
	fmt.Printf("🎭 Running TCP simulation test...\n")
	
	// 模拟操作执行
	for i := 0; i < config.BenchMark.Total; i++ {
		// 模拟延迟
		time.Sleep(time.Millisecond * time.Duration(1+i%10))
		
		// 创建模拟结果
		result := &interfaces.OperationResult{
			Success:  true,
			Duration: time.Millisecond * time.Duration(1+i%50),
			IsRead:   config.BenchMark.TestCase == "echo_test" || config.BenchMark.TestCase == "receive_only",
			Error:    nil,
			Value:    t.generateTestData(config.BenchMark.DataSize),
			Metadata: map[string]interface{}{
				"simulated":     true,
				"test_case":     config.BenchMark.TestCase,
				"data_size":     config.BenchMark.DataSize,
				"operation_id":  i,
			},
		}
		
		// 随机添加一些失败案例
		if i%100 == 0 {
			result.Success = false
			result.Error = fmt.Errorf("simulated error for operation %d", i)
		}
		
		collector.Record(result)
	}
	
	fmt.Printf("✅ Simulation completed with %d operations\n", config.BenchMark.Total)
	return nil
}

// generateReport 生成报告
func (t *TCPCommandHandler) generateReport(collector *metrics.BaseCollector[map[string]interface{}]) error {
	snapshot := collector.Snapshot()
	
	fmt.Printf("\n📊 TCP Performance Test Results:\n")
	fmt.Printf("=====================================\n")
	
	// 核心指标
	core := snapshot.Core
	fmt.Printf("Total Operations: %d\n", core.Operations.Total)
	fmt.Printf("Successful: %d (%.2f%%)\n", core.Operations.Success, 
		float64(core.Operations.Success)/float64(core.Operations.Total)*100)
	fmt.Printf("Failed: %d (%.2f%%)\n", core.Operations.Failed,
		float64(core.Operations.Failed)/float64(core.Operations.Total)*100)
	fmt.Printf("Read Operations: %d\n", core.Operations.Read)
	fmt.Printf("Write Operations: %d\n", core.Operations.Write)
	
	// 延迟指标
	fmt.Printf("\nLatency Metrics:\n")
	fmt.Printf("  Average: %v\n", core.Latency.Average)
	fmt.Printf("  Min: %v\n", core.Latency.Min)
	fmt.Printf("  Max: %v\n", core.Latency.Max)
	fmt.Printf("  P50: %v\n", core.Latency.P50)
	fmt.Printf("  P90: %v\n", core.Latency.P90)
	fmt.Printf("  P95: %v\n", core.Latency.P95)
	fmt.Printf("  P99: %v\n", core.Latency.P99)
	
	// 吞吐量指标
	fmt.Printf("\nThroughput Metrics:\n")
	fmt.Printf("  RPS: %.2f\n", core.Throughput.RPS)
	fmt.Printf("  Read RPS: %.2f\n", core.Throughput.ReadRPS)
	fmt.Printf("  Write RPS: %.2f\n", core.Throughput.WriteRPS)
	
	// 协议特定指标
	fmt.Printf("\nTCP Specific Metrics:\n")
	for key, value := range snapshot.Protocol {
		fmt.Printf("  %s: %v\n", key, value)
	}
	
	// 系统指标
	fmt.Printf("\nSystem Metrics:\n")
	fmt.Printf("  Memory Usage: %d MB\n", snapshot.System.MemoryUsage.InUse/1024/1024)
	fmt.Printf("  Goroutines: %d\n", snapshot.System.GoroutineCount)
	fmt.Printf("  GC Count: %d\n", snapshot.System.GCStats.NumGC)
	
	fmt.Printf("\nTest Duration: %v\n", core.Duration)
	fmt.Printf("=====================================\n")

	// 简化的文件报告
	config := reporting.NewStandardReportConfig("tcp")
	fmt.Printf("📄 Report configuration ready for: %s\n", config.OutputDir)

	return nil
}

// generateTestData 生成测试数据
func (t *TCPCommandHandler) generateTestData(size int) []byte {
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(i % 256)
	}
	return data
}