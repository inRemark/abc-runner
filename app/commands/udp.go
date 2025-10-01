package commands

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"abc-runner/app/adapters/udp"
	udpConfig "abc-runner/app/adapters/udp/config"
	"abc-runner/app/adapters/udp/operations"
	"abc-runner/app/core/execution"
	"abc-runner/app/core/interfaces"
	"abc-runner/app/core/metrics"
	"abc-runner/app/reporting"
)

// UDPCommandHandler UDP命令处理器
type UDPCommandHandler struct {
	protocolName string
	factory      interface{} // AdapterFactory接口
}

// NewUDPCommandHandler 创建UDP命令处理器
func NewUDPCommandHandler(factory interface{}) *UDPCommandHandler {
	if factory == nil {
		panic("adapterFactory cannot be nil - dependency injection required")
	}

	return &UDPCommandHandler{
		protocolName: "udp",
		factory:      factory,
	}
}

// Execute 执行UDP命令
func (u *UDPCommandHandler) Execute(ctx context.Context, args []string) error {
	// 检查帮助请求
	for i, arg := range args {
		if arg == "--help" || arg == "help" {
			fmt.Println(u.GetHelp())
			return nil
		}
		if arg == "-h" && (i == 0 || (i > 0 && args[i-1] != "udp")) {
			if i+1 < len(args) && !looksLikeHostname(args[i+1]) {
				fmt.Println(u.GetHelp())
				return nil
			}
		}
	}

	// 解析命令行参数
	config, err := u.parseArgs(args)
	if err != nil {
		return fmt.Errorf("failed to parse arguments: %w", err)
	}

	// 创建UDP适配器
	metricsConfig := metrics.DefaultMetricsConfig()
	metricsCollector := metrics.NewBaseCollector(metricsConfig, map[string]interface{}{
		"protocol":  "udp",
		"test_type": "performance",
	})
	defer metricsCollector.Stop()

	adapter := udp.NewUDPAdapter(metricsCollector)

	// 连接并执行测试
	if err := adapter.Connect(ctx, config); err != nil {
		fmt.Printf("⚠️  Connection setup failed for %s:%d: %v\n", config.Connection.Address, config.Connection.Port, err)
		fmt.Printf("🔍 Possible causes: Target unreachable, port blocked, or configuration error\n")
	} else {
		fmt.Printf("✅ Successfully configured UDP connection to %s:%d (%s mode)\n",
			config.Connection.Address, config.Connection.Port, config.UDPSpecific.PacketMode)
	}
	defer adapter.Close()

	// 执行性能测试
	fmt.Printf("🚀 Starting UDP performance test...\n")
	fmt.Printf("Target: %s:%d\n", config.Connection.Address, config.Connection.Port)
	fmt.Printf("Test Case: %s, Packet Mode: %s\n", config.BenchMark.TestCase, config.UDPSpecific.PacketMode)
	fmt.Printf("Packets: %d, Concurrency: %d, Packet Size: %d bytes\n",
		config.BenchMark.Total, config.BenchMark.Parallels, config.BenchMark.DataSize)

	if config.UDPSpecific.PacketMode == "multicast" {
		fmt.Printf("Multicast Group: %s, TTL: %d\n", config.UDPSpecific.MulticastGroup, config.UDPSpecific.TTL)
	}

	err = u.runPerformanceTest(ctx, adapter, config, metricsCollector)
	if err != nil {
		return fmt.Errorf("performance test failed: %w", err)
	}

	// 生成并显示报告
	return u.generateReport(metricsCollector)
}

// GetHelp 获取帮助信息
func (u *UDPCommandHandler) GetHelp() string {
	return `UDP Performance Testing

USAGE:
  abc-runner udp [options]

DESCRIPTION:
  Run UDP performance tests with various packet modes and configurations.

OPTIONS:
  --help              Show this help message
  --host HOST         UDP target host (default: localhost)
  --port PORT         UDP target port (default: 9090)
  -n COUNT            Number of packets (default: 1000)
  -c COUNT            Concurrent senders (default: 20)
  --data-size SIZE    Packet size in bytes (default: 1024, max: 65507)
  --test-case TYPE    Test case type (default: packet_send)
  --packet-mode MODE  Packet mode (default: unicast)
  --multicast-group   Multicast group address (required for multicast)
  --ttl VALUE         Packet TTL (default: 64)
  --duration DURATION Test duration (default: 60s)
  --packet-rate RATE  Packets per second rate (default: 1000)
  
PACKET MODES:
  unicast             Point-to-point communication
  broadcast           Broadcast to network
  multicast           Multicast to specific group
  
TEST CASES:
  packet_send         Send packets only
  packet_receive      Receive packets only
  echo_udp            Send and verify echo response
  multicast           Multicast group test
  
EXAMPLES:
  abc-runner udp --help
  abc-runner udp --host localhost --port 9090
  abc-runner udp --host 192.168.1.100 --packet-mode broadcast
  abc-runner udp --packet-mode multicast --multicast-group 224.0.0.1
  abc-runner udp -h localhost -p 9090 -n 5000 -c 50 --data-size 512

NOTE: 
  UDP testing supports unicast, broadcast, and multicast modes.
  For multicast testing, ensure proper network configuration.`
}

// parseArgs 解析命令行参数
func (u *UDPCommandHandler) parseArgs(args []string) (*udpConfig.UDPConfig, error) {
	// 创建默认配置
	config := udpConfig.NewDefaultUDPConfig()

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
				if size, err := strconv.Atoi(args[i+1]); err == nil && size > 0 && size <= 65507 {
					config.BenchMark.DataSize = size
				}
				i++
			}
		case "--test-case":
			if i+1 < len(args) {
				validCases := []string{"packet_send", "packet_receive", "echo_udp", "multicast"}
				testCase := args[i+1]
				for _, valid := range validCases {
					if testCase == valid {
						config.BenchMark.TestCase = testCase
						break
					}
				}
				i++
			}
		case "--packet-mode":
			if i+1 < len(args) {
				validModes := []string{"unicast", "broadcast", "multicast"}
				mode := args[i+1]
				for _, valid := range validModes {
					if mode == valid {
						config.UDPSpecific.PacketMode = mode
						break
					}
				}
				i++
			}
		case "--multicast-group":
			if i+1 < len(args) {
				if ip := net.ParseIP(args[i+1]); ip != nil && ip.IsMulticast() {
					config.UDPSpecific.MulticastGroup = args[i+1]
					config.UDPSpecific.PacketMode = "multicast" // 自动设置为组播模式
				}
				i++
			}
		case "--ttl":
			if i+1 < len(args) {
				if ttl, err := strconv.Atoi(args[i+1]); err == nil && ttl > 0 && ttl <= 255 {
					config.UDPSpecific.TTL = ttl
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
		case "--packet-rate":
			if i+1 < len(args) {
				if rate, err := strconv.Atoi(args[i+1]); err == nil && rate > 0 {
					config.BenchMark.PacketRate = rate
				}
				i++
			}
		}
	}

	return config, nil
}

// runPerformanceTest 运行性能测试
// runPerformanceTest 运行UDP性能测试
func (u *UDPCommandHandler) runPerformanceTest(ctx context.Context, adapter interfaces.ProtocolAdapter, config *udpConfig.UDPConfig, collector *metrics.BaseCollector[map[string]interface{}]) error {
	// 执行健康检查
	if err := adapter.HealthCheck(ctx); err != nil {
		fmt.Printf("⚠️  Health check failed: %v\n", err)
		fmt.Printf("🔄 Switching to simulation mode - this will generate mock UDP test data\n")
		return u.runSimulationTest(config, collector)
	}

	// 创建执行引擎
	factory := operations.NewSimpleOperationFactory(config.BenchMark.TestCase, config.BenchMark.DataSize)
	benchConfig := udpConfig.NewSimpleBenchmarkConfig(config.BenchMark.Total, config.BenchMark.Parallels, config.BenchMark.Duration)
	engine := execution.NewExecutionEngine(adapter, collector, factory)

	// 执行测试
	fmt.Printf("📊 Sending %d packets with %d concurrent workers...\n",
		config.BenchMark.Total, config.BenchMark.Parallels)

	// 记录测试开始时间
	testStartTime := time.Now()
	result, err := engine.RunBenchmark(ctx, benchConfig)
	actualTestDuration := time.Since(testStartTime)

	if err != nil {
		return fmt.Errorf("benchmark execution failed: %w", err)
	}

	fmt.Printf("✅ Test completed in %v (Actual: %v)\n", result.TotalDuration, actualTestDuration)
	fmt.Printf("📈 Processed %d packets (%d successful, %d failed)\n",
		result.CompletedJobs, result.SuccessJobs, result.FailedJobs)

	if result.CompletedJobs > 0 {
		// 计算正确的PPS（Packets Per Second）
		actualPPS := float64(result.CompletedJobs) / actualTestDuration.Seconds()
		fmt.Printf("📈 Actual PPS: %.2f packets/sec\n", actualPPS)
	}

	// 更新收集器的协议数据，包含实际测试时间
	collector.UpdateProtocolMetrics(map[string]interface{}{
		"protocol":         "udp",
		"test_type":        "performance",
		"actual_duration":  actualTestDuration,
		"execution_result": result,
		"test_case":        config.BenchMark.TestCase,
	})

	return nil
}

// runSimulationTest 运行模拟测试
func (u *UDPCommandHandler) runSimulationTest(config *udpConfig.UDPConfig, collector *metrics.BaseCollector[map[string]interface{}]) error {
	fmt.Printf("🎭 Running UDP simulation test...\n")

	// 模拟数据包发送
	for i := 0; i < config.BenchMark.Total; i++ {
		// 模拟网络延迟
		time.Sleep(time.Microsecond * time.Duration(100+i%500))

		// 创建模拟结果
		result := &interfaces.OperationResult{
			Success:  true,
			Duration: time.Microsecond * time.Duration(100+i%1000), // 模拟UDP的低延迟
			IsRead:   config.BenchMark.TestCase == "packet_receive" || config.BenchMark.TestCase == "echo_udp",
			Error:    nil,
			Value:    u.generatePacketData(config.BenchMark.DataSize, i),
			Metadata: map[string]interface{}{
				"simulated":      true,
				"test_case":      config.BenchMark.TestCase,
				"packet_size":    config.BenchMark.DataSize,
				"packet_id":      i,
				"packet_mode":    config.UDPSpecific.PacketMode,
				"loss_simulated": false,
			},
		}

		// 模拟UDP丢包（约5%的丢包率）
		if i%20 == 0 {
			result.Success = false
			result.Error = fmt.Errorf("simulated packet loss for packet %d", i)
			result.Metadata["loss_simulated"] = true
		}

		collector.Record(result)
	}

	fmt.Printf("✅ Simulation completed with %d packets\n", config.BenchMark.Total)
	return nil
}

// generateReport 生成报告
// generateReport 生成UDP性能测试报告
func (u *UDPCommandHandler) generateReport(collector *metrics.BaseCollector[map[string]interface{}]) error {
	snapshot := collector.Snapshot()

	// 从协议数据中获取实际测试时间
	var actualDuration time.Duration
	if protocolData, ok := snapshot.Protocol["actual_duration"]; ok {
		if duration, ok := protocolData.(time.Duration); ok {
			actualDuration = duration
		}
	}

	// 如果没有实际时间，使用默认时间
	if actualDuration == 0 {
		actualDuration = snapshot.Core.Duration
	}

	// 更新快照中的测试时间和吸吐量指标
	snapshot.Core.Duration = actualDuration
	if actualDuration > 0 {
		// 重新计算吸吐量（基于实际测试时间）
		total := snapshot.Core.Operations.Read + snapshot.Core.Operations.Write
		seconds := actualDuration.Seconds()
		snapshot.Core.Throughput.RPS = float64(total) / seconds
		snapshot.Core.Throughput.ReadRPS = float64(snapshot.Core.Operations.Read) / seconds
		snapshot.Core.Throughput.WriteRPS = float64(snapshot.Core.Operations.Write) / seconds
	}

	fmt.Printf("\n📊 UDP Performance Test Results:\n")
	fmt.Printf("=====================================\n")

	// 核心指标
	core := snapshot.Core
	fmt.Printf("Total Packets: %d\n", core.Operations.Total)
	fmt.Printf("Successful: %d (%.2f%%)\n", core.Operations.Success,
		float64(core.Operations.Success)/float64(core.Operations.Total)*100)
	fmt.Printf("Failed/Lost: %d (%.2f%%)\n", core.Operations.Failed,
		float64(core.Operations.Failed)/float64(core.Operations.Total)*100)
	fmt.Printf("Sent Packets: %d\n", core.Operations.Write)
	fmt.Printf("Received Packets: %d\n", core.Operations.Read)

	// 延迟指标（UDP应该有很低的延迟）
	fmt.Printf("\nLatency Metrics:\n")
	fmt.Printf("  Average: %v\n", core.Latency.Average)
	fmt.Printf("  Min: %v\n", core.Latency.Min)
	fmt.Printf("  Max: %v\n", core.Latency.Max)
	fmt.Printf("  P50: %v\n", core.Latency.P50)
	fmt.Printf("  P90: %v\n", core.Latency.P90)
	fmt.Printf("  P95: %v\n", core.Latency.P95)
	fmt.Printf("  P99: %v\n", core.Latency.P99)

	// 吸吐量指标（使用修正后的数值）
	fmt.Printf("\nThroughput Metrics (Corrected):\n")
	fmt.Printf("  Packets Per Second: %.2f\n", core.Throughput.RPS)
	fmt.Printf("  Send PPS: %.2f\n", core.Throughput.WriteRPS)
	fmt.Printf("  Receive PPS: %.2f\n", core.Throughput.ReadRPS)

	// UDP特定指标
	fmt.Printf("\nUDP Specific Metrics:\n")
	for key, value := range snapshot.Protocol {
		fmt.Printf("  %s: %v\n", key, value)
	}

	// 系统指标
	fmt.Printf("\nSystem Metrics:\n")
	fmt.Printf("  Memory Usage: %d MB\n", snapshot.System.MemoryUsage.InUse/1024/1024)
	fmt.Printf("  Goroutines: %d\n", snapshot.System.GoroutineCount)
	fmt.Printf("  GC Count: %d\n", snapshot.System.GCStats.NumGC)

	fmt.Printf("\nTest Duration: %v (Actual: %v)\n", core.Duration, actualDuration)
	fmt.Printf("=====================================\n")

	// 简化的文件报告
	config := reporting.NewStandardReportConfig("udp")
	fmt.Printf("📄 Report configuration ready for: %s\n", config.OutputDir)

	return nil
}

// generatePacketData 生成数据包数据
func (u *UDPCommandHandler) generatePacketData(size, packetId int) []byte {
	data := make([]byte, size)

	// 前8字节作为包ID和序列号
	for i := 0; i < 8 && i < size; i++ {
		data[i] = byte((packetId >> (8 * (7 - i))) & 0xFF)
	}

	// 其余字节为模式数据
	for i := 8; i < size; i++ {
		data[i] = byte(i % 256)
	}

	return data
}
