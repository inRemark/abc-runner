package commands

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"abc-runner/app/adapters/grpc/config"
	"abc-runner/app/adapters/grpc/operations"
	"abc-runner/app/core/execution"
	"abc-runner/app/core/interfaces"
	"abc-runner/app/core/metrics"
)

// GRPCCommandHandler gRPC命令处理器
type GRPCCommandHandler struct {
	protocolName string
	factory      interfaces.GRPCAdapterFactory // 使用gRPC专用工厂接口
}

// NewGRPCCommandHandler 创建gRPC命令处理器
func NewGRPCCommandHandler(factory interfaces.GRPCAdapterFactory) *GRPCCommandHandler {
	if factory == nil {
		panic("grpcAdapterFactory cannot be nil - dependency injection required")
	}

	return &GRPCCommandHandler{
		protocolName: "grpc",
		factory:      factory,
	}
}

// Execute 执行gRPC命令
func (h *GRPCCommandHandler) Execute(ctx context.Context, args []string) error {
	// 检查帮助请求
	for i, arg := range args {
		if arg == "--help" || arg == "help" {
			fmt.Println(h.GetHelp())
			return nil
		}
		if arg == "-h" && (i == 0 || (i > 0 && args[i-1] != "grpc")) {
			if i+1 < len(args) && !looksLikeHostname(args[i+1]) {
				fmt.Println(h.GetHelp())
				return nil
			}
		}
	}

	// 解析命令行参数
	config, err := h.parseArgs(args)
	if err != nil {
		return fmt.Errorf("failed to parse arguments: %w", err)
	}

	// 创建指标收集器
	metricsConfig := metrics.DefaultMetricsConfig()
	metricsCollector := metrics.NewBaseCollector(metricsConfig, map[string]interface{}{
		"protocol":  "grpc",
		"test_type": "performance",
	})
	defer metricsCollector.Stop()

	// 创建适配器
	adapter := h.factory.CreateGRPCAdapter()
	if adapter == nil {
		return fmt.Errorf("failed to create gRPC adapter")
	}
	defer adapter.Close()

	// 连接到gRPC服务器
	fmt.Printf("🔗 Connecting to gRPC server: %s:%d\n",
		config.Connection.Address, config.Connection.Port)

	if err := adapter.Connect(ctx, config); err != nil {
		fmt.Printf("⚠️  Connection failed to %s:%d: %v\n",
			config.Connection.Address, config.Connection.Port, err)
		fmt.Printf("🔍 Possible causes: gRPC server not running, wrong host/port, TLS issues, or network problems\n")
	} else {
		fmt.Printf("✅ Successfully connected to gRPC server\n")
	}

	// 运行性能测试
	fmt.Printf("🚀 Starting gRPC performance test...\n")
	fmt.Printf("Target: %s:%d\n", config.Connection.Address, config.Connection.Port)
	fmt.Printf("Test Case: %s\n", config.BenchMark.TestCase)
	fmt.Printf("Operations: %d, Concurrency: %d, Data Size: %d bytes\n",
		config.BenchMark.Total, config.BenchMark.Parallels, config.BenchMark.DataSize)

	err = h.runPerformanceTest(ctx, adapter, config, metricsCollector)
	if err != nil {
		return fmt.Errorf("performance test failed: %w", err)
	}

	// 生成并显示报告
	return h.generateReport(metricsCollector)
}

// GetHelp 获取帮助信息
func (h *GRPCCommandHandler) GetHelp() string {
	return `gRPC Performance Testing

USAGE:
  abc-runner grpc [options]

DESCRIPTION:
  Run gRPC performance tests with various call patterns including unary, streaming, and bidirectional calls.

OPTIONS:
  --help              Show this help message
  --address HOST      gRPC server address (default: localhost)
  --port PORT         gRPC server port (default: 50051)
  --service NAME      gRPC service name (default: TestService)
  --method NAME       gRPC method name (default: Echo)
  --test-case TYPE    Test case type (default: unary_call)
  -c COUNT            Concurrent connections (default: 10)
  -n COUNT            Total operations (default: 1000)
  --timeout DURATION  Operation timeout (default: 30s)
  --tls               Enable TLS (default: false)
  --token TOKEN       Authentication token
  
TEST CASES:
  unary_call          Standard unary gRPC call
  server_stream       Server streaming call
  client_stream       Client streaming call
  bidirectional_stream Bidirectional streaming call
  
EXAMPLES:
  abc-runner grpc --help
  abc-runner grpc --address localhost --port 50051
  abc-runner grpc --service MyService --method GetData --test-case unary_call
  abc-runner grpc --address 192.168.1.100 --port 9090 -c 20 -n 5000

NOTE: 
  This implementation performs real gRPC performance testing with metrics collection.`
}

// parseArgs 解析命令行参数
func (h *GRPCCommandHandler) parseArgs(args []string) (*config.GRPCConfig, error) {
	// 创建默认配置
	gRPCConfig := config.NewDefaultGRPCConfig()

	// 解析参数
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--address":
			if i+1 < len(args) {
				gRPCConfig.Connection.Address = args[i+1]
				i++
			}
		case "--port":
			if i+1 < len(args) {
				if port, err := strconv.Atoi(args[i+1]); err == nil && port > 0 && port <= 65535 {
					gRPCConfig.Connection.Port = port
				}
				i++
			}
		case "--service":
			if i+1 < len(args) {
				gRPCConfig.GRPCSpecific.ServiceName = args[i+1]
				i++
			}
		case "--method":
			if i+1 < len(args) {
				gRPCConfig.GRPCSpecific.MethodName = args[i+1]
				i++
			}
		case "--test-case":
			if i+1 < len(args) {
				validCases := []string{"unary_call", "server_stream", "client_stream", "bidirectional_stream"}
				testCase := args[i+1]
				for _, valid := range validCases {
					if testCase == valid {
						gRPCConfig.BenchMark.TestCase = testCase
						break
					}
				}
				i++
			}
		case "-c":
			if i+1 < len(args) {
				if count, err := strconv.Atoi(args[i+1]); err == nil && count > 0 {
					gRPCConfig.BenchMark.Parallels = count
					gRPCConfig.Connection.Pool.PoolSize = count
				}
				i++
			}
		case "-n":
			if i+1 < len(args) {
				if count, err := strconv.Atoi(args[i+1]); err == nil && count > 0 {
					gRPCConfig.BenchMark.Total = count
				}
				i++
			}
		case "--timeout":
			if i+1 < len(args) {
				if duration, err := time.ParseDuration(args[i+1]); err == nil {
					gRPCConfig.Connection.Timeout = duration
					gRPCConfig.BenchMark.Timeout = duration
				}
				i++
			}
		case "--tls":
			gRPCConfig.GRPCSpecific.TLS.Enabled = true
		case "--token":
			if i+1 < len(args) {
				gRPCConfig.GRPCSpecific.Auth.Enabled = true
				gRPCConfig.GRPCSpecific.Auth.Method = "token"
				gRPCConfig.GRPCSpecific.Auth.Token = args[i+1]
				i++
			}
		}
	}

	return gRPCConfig, nil
}

// runPerformanceTest 运行性能测试
// runPerformanceTest 运行gRPC性能测试
func (h *GRPCCommandHandler) runPerformanceTest(
	ctx context.Context,
	adapter interfaces.ProtocolAdapter,
	config *config.GRPCConfig,
	metricsCollector interfaces.DefaultMetricsCollector,
) error {
	// 创建操作工厂
	operationFactory := operations.NewOperationFactory(config)

	// 创建执行引擎
	engine := execution.NewExecutionEngine(adapter, metricsCollector, operationFactory)

	// 配置执行引擎
	engine.SetMaxWorkers(config.BenchMark.Parallels * 3) // 适度超配以提高并发性能
	engine.SetBufferSizes(
		config.BenchMark.Parallels*10, // job buffer
		config.BenchMark.Parallels*10, // result buffer
	)

	// 记录测试开始时间
	testStartTime := time.Now()

	// 运行基准测试
	result, err := engine.RunBenchmark(ctx, &config.BenchMark)
	if err != nil {
		return fmt.Errorf("benchmark execution failed: %w", err)
	}

	// 计算实际测试时间
	actualTestDuration := time.Since(testStartTime)

	// 输出执行结果
	fmt.Printf("\n📊 Execution Results:\n")
	fmt.Printf("Total Jobs: %d\n", result.TotalJobs)
	fmt.Printf("Completed Jobs: %d\n", result.CompletedJobs)
	fmt.Printf("Success Jobs: %d\n", result.SuccessJobs)
	fmt.Printf("Failed Jobs: %d\n", result.FailedJobs)
	fmt.Printf("Total Duration: %v\n", result.TotalDuration)
	fmt.Printf("Actual Test Duration: %v\n", actualTestDuration)
	if result.TotalJobs > 0 {
		fmt.Printf("Success Rate: %.2f%%\n", float64(result.SuccessJobs)/float64(result.TotalJobs)*100)
		// 计算正确的RPS（基于实际测试时间）
		actualRPS := float64(result.CompletedJobs) / actualTestDuration.Seconds()
		fmt.Printf("Actual RPS: %.2f calls/sec\n", actualRPS)
	}

	// 更新收集器的协议数据，包含实际测试时间
	if baseCollector, ok := metricsCollector.(*metrics.BaseCollector[map[string]interface{}]); ok {
		baseCollector.UpdateProtocolMetrics(map[string]interface{}{
			"protocol":         "grpc",
			"test_type":        "performance",
			"actual_duration":  actualTestDuration,
			"execution_result": result,
			"service":          config.GRPCSpecific.ServiceName,
			"method":           config.GRPCSpecific.MethodName,
		})
	}

	return nil
}

// generateReport 生成报告
// generateReport 生成gRPC性能测试报告
func (h *GRPCCommandHandler) generateReport(metricsCollector interfaces.DefaultMetricsCollector) error {
	snapshot := metricsCollector.Snapshot()
	if snapshot == nil {
		return fmt.Errorf("failed to get metrics snapshot")
	}

	// 尝试获取实际测试时间（仅当metricsCollector是BaseCollector时）
	var actualDuration time.Duration
	if baseCollector, ok := metricsCollector.(*metrics.BaseCollector[map[string]interface{}]); ok {
		baseSnapshot := baseCollector.Snapshot()
		if protocolData, ok := baseSnapshot.Protocol["actual_duration"]; ok {
			if duration, ok := protocolData.(time.Duration); ok {
				actualDuration = duration
			}
		}

		// 如果没有实际时间，使用默认时间
		if actualDuration == 0 {
			actualDuration = baseSnapshot.Core.Duration
		}

		// 更新快照中的测试时间和吸吐量指标
		baseSnapshot.Core.Duration = actualDuration
		if actualDuration > 0 {
			// 重新计算吸吐量（基于实际测试时间）
			total := baseSnapshot.Core.Operations.Read + baseSnapshot.Core.Operations.Write
			seconds := actualDuration.Seconds()
			baseSnapshot.Core.Throughput.RPS = float64(total) / seconds
			baseSnapshot.Core.Throughput.ReadRPS = float64(baseSnapshot.Core.Operations.Read) / seconds
			baseSnapshot.Core.Throughput.WriteRPS = float64(baseSnapshot.Core.Operations.Write) / seconds
		}

		// 使用更新后的数据
		snapshot = baseSnapshot
	}

	// 输出简单报告
	fmt.Printf("\n📊 gRPC Performance Metrics:\n")
	fmt.Printf("=====================================\n")
	fmt.Printf("Core Metrics:\n")
	fmt.Printf("  Total Operations: %d\n", snapshot.Core.Operations.Total)
	fmt.Printf("  Successful Operations: %d\n", snapshot.Core.Operations.Success)
	fmt.Printf("  Failed Operations: %d\n", snapshot.Core.Operations.Failed)
	fmt.Printf("  Success Rate: %.2f%%\n", snapshot.Core.Operations.Rate)
	fmt.Printf("Latency Metrics:\n")
	fmt.Printf("  Average: %v\n", snapshot.Core.Latency.Average)
	fmt.Printf("  P95: %v\n", snapshot.Core.Latency.P95)
	fmt.Printf("  P99: %v\n", snapshot.Core.Latency.P99)
	fmt.Printf("Throughput (Corrected): %.2f RPS\n", snapshot.Core.Throughput.RPS)

	if actualDuration > 0 {
		fmt.Printf("Test Duration: %v (Actual: %v)\n", snapshot.Core.Duration, actualDuration)
	} else {
		fmt.Printf("Test Duration: %v\n", snapshot.Core.Duration)
	}
	fmt.Printf("=====================================\n")

	return nil
}

// GetProtocolName 获取协议名称
func (h *GRPCCommandHandler) GetProtocolName() string {
	return "grpc"
}

// GetFactory 获取适配器工厂
func (h *GRPCCommandHandler) GetFactory() interfaces.GRPCAdapterFactory {
	return h.factory
}
