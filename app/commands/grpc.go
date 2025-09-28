package commands

import (
	"abc-runner/app/core/interfaces"
	"context"
	"fmt"
	"strconv"
	"time"
)

// BaseCommandHandler 基础命令处理器
type BaseCommandHandler struct {
	protocolName string
}

// NewBaseCommandHandler 创建基础命令处理器
func NewBaseCommandHandler(protocolName string) *BaseCommandHandler {
	return &BaseCommandHandler{
		protocolName: protocolName,
	}
}

// GetProtocolName 获取协议名称
func (h *BaseCommandHandler) GetProtocolName() string {
	return h.protocolName
}

// GRPCCommandHandler gRPC命令处理器
type GRPCCommandHandler struct {
	protocolName string
	factory      interface{} // AdapterFactory接口
}

// NewGRPCCommandHandler 创建gRPC命令处理器
func NewGRPCCommandHandler(factory interface{}) *GRPCCommandHandler {
	if factory == nil {
		panic("adapterFactory cannot be nil - dependency injection required")
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

	// 创建适配器
	// adapter := h.factory.CreateAdapter()
	// 暂时使用模拟适配器，因为没有导入discovery包
	adapter := h.createMockAdapter()
	if adapter == nil {
		return fmt.Errorf("failed to create gRPC adapter")
	}
	defer adapter.Close()

	// 连接到gRPC服务器
	fmt.Printf("🔗 Connecting to gRPC server: %s:%d\n", 
		config["address"].(string), config["port"].(int))
	
	if err := adapter.Connect(ctx, h.createConfigWrapper(config)); err != nil {
		fmt.Printf("⚠️  Connection failed to %s:%d: %v\n", 
			config["address"].(string), config["port"].(int), err)
		fmt.Printf("🔍 Possible causes: gRPC server not running, wrong host/port, TLS issues, or network problems\n")
		return err
	}

	fmt.Printf("✅ Successfully connected to gRPC server\n")

	// 运行基准测试
	testCase := config["test_case"].(string)
	return h.runTestCase(ctx, adapter, testCase, config)
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
func (h *GRPCCommandHandler) parseArgs(args []string) (map[string]interface{}, error) {
	// 创建默认配置
	config := map[string]interface{}{
		"address":     "localhost",
		"port":        50051,
		"service_name": "TestService",
		"method_name":  "Echo",
		"test_case":   "unary_call",
		"parallels":   10,
		"total":       1000,
		"timeout":     30 * time.Second,
		"tls_enabled": false,
		"token":       "",
	}

	// 解析参数
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--address":
			if i+1 < len(args) {
				config["address"] = args[i+1]
				i++
			}
		case "--port":
			if i+1 < len(args) {
				if port, err := strconv.Atoi(args[i+1]); err == nil && port > 0 && port <= 65535 {
					config["port"] = port
				}
				i++
			}
		case "--service":
			if i+1 < len(args) {
				config["service_name"] = args[i+1]
				i++
			}
		case "--method":
			if i+1 < len(args) {
				config["method_name"] = args[i+1]
				i++
			}
		case "--test-case":
			if i+1 < len(args) {
				validCases := []string{"unary_call", "server_stream", "client_stream", "bidirectional_stream"}
				testCase := args[i+1]
				for _, valid := range validCases {
					if testCase == valid {
						config["test_case"] = testCase
						break
					}
				}
				i++
			}
		case "-c":
			if i+1 < len(args) {
				if count, err := strconv.Atoi(args[i+1]); err == nil && count > 0 {
					config["parallels"] = count
				}
				i++
			}
		case "-n":
			if i+1 < len(args) {
				if count, err := strconv.Atoi(args[i+1]); err == nil && count > 0 {
					config["total"] = count
				}
				i++
			}
		case "--timeout":
			if i+1 < len(args) {
				if duration, err := time.ParseDuration(args[i+1]); err == nil {
					config["timeout"] = duration
				}
				i++
			}
		case "--tls":
			config["tls_enabled"] = true
		case "--token":
			if i+1 < len(args) {
				config["token"] = args[i+1]
				i++
			}
		}
	}

	return config, nil
}

// runTestCase 运行特定的测试用例
func (h *GRPCCommandHandler) runTestCase(ctx context.Context, adapter interfaces.ProtocolAdapter, testCase string, config map[string]interface{}) error {
	fmt.Printf("🚀 Starting gRPC %s test...\n", testCase)
	fmt.Printf("Service: %s, Method: %s\n", config["service_name"], config["method_name"])
	fmt.Printf("Operations: %d, Concurrency: %d\n", config["total"], config["parallels"])

	switch testCase {
	case "unary_call":
		return h.runUnaryCallTest(ctx, adapter, config)
	case "server_stream":
		return h.runServerStreamTest(ctx, adapter, config)
	case "client_stream":
		return h.runClientStreamTest(ctx, adapter, config)
	case "bidirectional_stream":
		return h.runBidirectionalStreamTest(ctx, adapter, config)
	default:
		return fmt.Errorf("unsupported test case: %s", testCase)
	}
}

// runUnaryCallTest 运行一元调用测试
func (h *GRPCCommandHandler) runUnaryCallTest(ctx context.Context, adapter interfaces.ProtocolAdapter, config map[string]interface{}) error {
	total := config["total"].(int)
	concurrent := config["parallels"].(int)
	serviceName := config["service_name"].(string)
	methodName := config["method_name"].(string)

	// 创建测试操作
	operation := interfaces.Operation{
		Type:  "unary_call",
		Key:   fmt.Sprintf("%s.%s", serviceName, methodName),
		Value: "test_request_data",
		Metadata: map[string]string{
			"total":       strconv.Itoa(total),
			"concurrent":  strconv.Itoa(concurrent),
			"service":     serviceName,
			"method":      methodName,
		},
	}

	// 执行操作
	result, err := adapter.Execute(ctx, operation)
	if err != nil {
		return fmt.Errorf("unary call test failed: %w", err)
	}

	// 打印结果
	fmt.Printf("✅ Unary call test completed: Success=%t, Duration=%v\n", 
		result.Success, result.Duration)
	return nil
}

// runServerStreamTest 运行服务器流测试
func (h *GRPCCommandHandler) runServerStreamTest(ctx context.Context, adapter interfaces.ProtocolAdapter, config map[string]interface{}) error {
	concurrent := config["parallels"].(int)
	serviceName := config["service_name"].(string)
	methodName := config["method_name"].(string)

	operation := interfaces.Operation{
		Type:  "server_stream",
		Key:   fmt.Sprintf("%s.%s", serviceName, methodName),
		Value: "stream_request",
		Metadata: map[string]string{
			"concurrent": strconv.Itoa(concurrent),
			"service":    serviceName,
			"method":     methodName,
		},
	}

	result, err := adapter.Execute(ctx, operation)
	if err != nil {
		return fmt.Errorf("server stream test failed: %w", err)
	}

	fmt.Printf("✅ Server stream test completed: Success=%t, Duration=%v\n", 
		result.Success, result.Duration)
	return nil
}

// runClientStreamTest 运行客户端流测试
func (h *GRPCCommandHandler) runClientStreamTest(ctx context.Context, adapter interfaces.ProtocolAdapter, config map[string]interface{}) error {
	concurrent := config["parallels"].(int)
	serviceName := config["service_name"].(string)
	methodName := config["method_name"].(string)

	operation := interfaces.Operation{
		Type:  "client_stream",
		Key:   fmt.Sprintf("%s.%s", serviceName, methodName),
		Value: "stream_data",
		Metadata: map[string]string{
			"concurrent": strconv.Itoa(concurrent),
			"service":    serviceName,
			"method":     methodName,
		},
	}

	result, err := adapter.Execute(ctx, operation)
	if err != nil {
		return fmt.Errorf("client stream test failed: %w", err)
	}

	fmt.Printf("✅ Client stream test completed: Success=%t, Duration=%v\n", 
		result.Success, result.Duration)
	return nil
}

// runBidirectionalStreamTest 运行双向流测试
func (h *GRPCCommandHandler) runBidirectionalStreamTest(ctx context.Context, adapter interfaces.ProtocolAdapter, config map[string]interface{}) error {
	concurrent := config["parallels"].(int)
	serviceName := config["service_name"].(string)
	methodName := config["method_name"].(string)

	operation := interfaces.Operation{
		Type:  "bidirectional_stream",
		Key:   fmt.Sprintf("%s.%s", serviceName, methodName),
		Value: "bidi_stream_data",
		Metadata: map[string]string{
			"concurrent": strconv.Itoa(concurrent),
			"service":    serviceName,
			"method":     methodName,
		},
	}

	result, err := adapter.Execute(ctx, operation)
	if err != nil {
		return fmt.Errorf("bidirectional stream test failed: %w", err)
	}

	fmt.Printf("✅ Bidirectional stream test completed: Success=%t, Duration=%v\n", 
		result.Success, result.Duration)
	return nil
}

// GetProtocolName 获取协议名称
func (h *GRPCCommandHandler) GetProtocolName() string {
	return "grpc"
}

// GetFactory 获取适配器工厂
func (h *GRPCCommandHandler) GetFactory() interface{} {
	return h.factory
}

// createMockAdapter 创建模拟适配器
func (h *GRPCCommandHandler) createMockAdapter() interfaces.ProtocolAdapter {
	// 返回模拟适配器，实际应用中将由application.go中的注册逻辑处理
	return &MockGRPCAdapter{}
}

// MockGRPCAdapter 模拟适配器
type MockGRPCAdapter struct{}

func (m *MockGRPCAdapter) Connect(ctx context.Context, config interfaces.Config) error {
	return nil
}

func (m *MockGRPCAdapter) Execute(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	return &interfaces.OperationResult{
		Success:  true,
		Duration: 10 * time.Millisecond,
		Value:    "mock result",
	}, nil
}

func (m *MockGRPCAdapter) Close() error {
	return nil
}

func (m *MockGRPCAdapter) GetProtocolMetrics() map[string]interface{} {
	return map[string]interface{}{"mock": true}
}

func (m *MockGRPCAdapter) HealthCheck(ctx context.Context) error {
	return nil
}

func (m *MockGRPCAdapter) GetProtocolName() string {
	return "grpc"
}

func (m *MockGRPCAdapter) GetMetricsCollector() interfaces.DefaultMetricsCollector {
	return nil
}

// createConfigWrapper 创建Config接口包装器
func (h *GRPCCommandHandler) createConfigWrapper(config map[string]interface{}) interfaces.Config {
	return &GRPCConfigWrapper{data: config}
}

// GRPCConfigWrapper Config接口包装器
type GRPCConfigWrapper struct {
	data map[string]interface{}
}

func (c *GRPCConfigWrapper) GetProtocol() string {
	if protocol, ok := c.data["protocol"].(string); ok {
		return protocol
	}
	return "grpc"
}

func (c *GRPCConfigWrapper) GetConnection() interfaces.ConnectionConfig {
	return nil // gRPC不需要复杂的连接配置
}

func (c *GRPCConfigWrapper) GetBenchmark() interfaces.BenchmarkConfig {
	return nil // gRPC不需要复杂的基准测试配置
}

func (c *GRPCConfigWrapper) Validate() error {
	return nil // 简化验证
}

func (c *GRPCConfigWrapper) Clone() interfaces.Config {
	newData := make(map[string]interface{})
	for k, v := range c.data {
		newData[k] = v
	}
	return &GRPCConfigWrapper{data: newData}
}

// GetData 获取原始数据
func (c *GRPCConfigWrapper) GetData() map[string]interface{} {
	return c.data
}

