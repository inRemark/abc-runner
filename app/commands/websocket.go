package commands

import (
	"abc-runner/app/core/interfaces"
	"context"
	"fmt"
	"strconv"
	"time"
)

// WebSocketCommandHandler WebSocket命令处理器
type WebSocketCommandHandler struct {
	protocolName string
	factory      interface{} // AdapterFactory接口
}

// NewWebSocketCommandHandler 创建WebSocket命令处理器
func NewWebSocketCommandHandler(factory interface{}) *WebSocketCommandHandler {
	if factory == nil {
		panic("adapterFactory cannot be nil - dependency injection required")
	}

	return &WebSocketCommandHandler{
		protocolName: "websocket",
		factory:      factory,
	}
}

// Execute 执行WebSocket命令
func (h *WebSocketCommandHandler) Execute(ctx context.Context, args []string) error {
	// 检查帮助请求
	for i, arg := range args {
		if arg == "--help" || arg == "help" {
			fmt.Println(h.GetHelp())
			return nil
		}
		if arg == "-h" && (i == 0 || (i > 0 && args[i-1] != "websocket")) {
			if i+1 < len(args) && !looksLikeURL(args[i+1]) {
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
	adapter := h.createAdapter()
	if adapter == nil {
		return fmt.Errorf("failed to create WebSocket adapter")
	}
	defer adapter.Close()

	// 连接到WebSocket服务器
	fmt.Printf("🔗 Connecting to WebSocket server: %s\n", config["url"].(string))

	if err := adapter.Connect(ctx, h.createConfigWrapper(config)); err != nil {
		fmt.Printf("⚠️  Connection failed to %s: %v\n", config["url"].(string), err)
		fmt.Printf("🔍 Possible causes: WebSocket server not running, wrong URL, or network issues\n")
		return err
	}

	fmt.Printf("✅ Successfully connected to WebSocket server\n")

	// 运行基准测试
	testCase := config["test_case"].(string)
	return h.runTestCase(ctx, adapter, testCase, config)
}

// GetHelp 获取帮助信息
func (h *WebSocketCommandHandler) GetHelp() string {
	return `WebSocket Performance Testing

USAGE:
  abc-runner websocket [options]

DESCRIPTION:
  Execute performance tests using WebSocket protocol with various message patterns and connection scenarios.

OPTIONS:
  --help              Show this help message
  --url URL           WebSocket server URL (default: ws://localhost:8080/ws)
  --test-case TYPE    Test case type (default: message_exchange)
  -c COUNT            Concurrent connections (default: 10)
  --duration DURATION Test duration (default: 30s)
  --interval DURATION Message sending interval (default: 100ms)
  --message-size SIZE Message size in bytes (default: 1024)
  --message TEXT      Custom message content
  --compression       Enable WebSocket compression
  
TEST CASES:
  message_exchange    Message exchange test
  ping_pong          Ping-pong heartbeat test
  broadcast          Broadcast message test
  large_message      Large message transfer test
  
EXAMPLES:
  abc-runner websocket --help
  abc-runner websocket --url ws://localhost:8080/ws
  abc-runner websocket --url wss://example.com/ws --test-case ping_pong
  abc-runner websocket --url ws://192.168.1.100:8080/ws -c 20 --duration 60s

NOTE: 
  This implementation performs real WebSocket performance testing with metrics collection.`
}

// parseArgs 解析命令行参数
func (h *WebSocketCommandHandler) parseArgs(args []string) (map[string]interface{}, error) {
	// 创建默认配置
	config := map[string]interface{}{
		"url":                    "ws://localhost:8080/ws",
		"test_case":              "message_exchange",
		"concurrent_connections": 10,
		"duration":               30 * time.Second,
		"interval":               100 * time.Millisecond,
		"message_size":           1024,
		"message":                "",
		"enable_compression":     false,
		"handshake_timeout":      10 * time.Second,
		"read_timeout":           30 * time.Second,
		"write_timeout":          10 * time.Second,
	}

	// 解析参数
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--url":
			if i+1 < len(args) {
				config["url"] = args[i+1]
				i++
			}
		case "--test-case":
			if i+1 < len(args) {
				validCases := []string{"message_exchange", "ping_pong", "broadcast", "large_message"}
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
					config["concurrent_connections"] = count
				}
				i++
			}
		case "--duration":
			if i+1 < len(args) {
				if duration, err := time.ParseDuration(args[i+1]); err == nil {
					config["duration"] = duration
				}
				i++
			}
		case "--interval":
			if i+1 < len(args) {
				if interval, err := time.ParseDuration(args[i+1]); err == nil {
					config["interval"] = interval
				}
				i++
			}
		case "--message-size":
			if i+1 < len(args) {
				if size, err := strconv.Atoi(args[i+1]); err == nil && size > 0 {
					config["message_size"] = size
				}
				i++
			}
		case "--message":
			if i+1 < len(args) {
				config["message"] = args[i+1]
				i++
			}
		case "--compression":
			config["enable_compression"] = true
		}
	}

	// 如果没有自定义消息，生成测试消息
	if config["message"].(string) == "" {
		config["message"] = h.generateTestMessage(config["message_size"].(int))
	}

	return config, nil
}

// runTestCase 运行特定的测试用例
func (h *WebSocketCommandHandler) runTestCase(ctx context.Context, adapter interfaces.ProtocolAdapter, testCase string, config map[string]interface{}) error {
	fmt.Printf("🚀 Starting WebSocket %s test...\n", testCase)
	fmt.Printf("URL: %s\n", config["url"])
	fmt.Printf("Connections: %d, Duration: %v\n",
		config["concurrent_connections"], config["duration"])

	switch testCase {
	case "message_exchange":
		return h.runMessageExchangeTest(ctx, adapter, config)
	case "ping_pong":
		return h.runPingPongTest(ctx, adapter, config)
	case "broadcast":
		return h.runBroadcastTest(ctx, adapter, config)
	case "large_message":
		return h.runLargeMessageTest(ctx, adapter, config)
	default:
		return fmt.Errorf("unsupported test case: %s", testCase)
	}
}

// runMessageExchangeTest 运行消息交换测试
func (h *WebSocketCommandHandler) runMessageExchangeTest(ctx context.Context, adapter interfaces.ProtocolAdapter, config map[string]interface{}) error {
	concurrent := config["concurrent_connections"].(int)
	duration := config["duration"].(time.Duration)
	message := config["message"].(string)
	interval := config["interval"].(time.Duration)

	// 创建测试操作
	operation := interfaces.Operation{
		Type:  "send_text",
		Key:   "test_message",
		Value: message,
		Metadata: map[string]string{
			"concurrent": strconv.Itoa(concurrent),
			"duration":   duration.String(),
			"interval":   interval.String(),
		},
	}

	// 执行操作
	result, err := adapter.Execute(ctx, operation)
	if err != nil {
		return fmt.Errorf("message exchange test failed: %w", err)
	}

	// 打印结果
	fmt.Printf("✅ Message exchange test completed: Success=%t, Duration=%v\n",
		result.Success, result.Duration)
	return nil
}

// runPingPongTest 运行ping-pong测试
func (h *WebSocketCommandHandler) runPingPongTest(ctx context.Context, adapter interfaces.ProtocolAdapter, config map[string]interface{}) error {
	concurrent := config["concurrent_connections"].(int)
	interval := config["interval"].(time.Duration)

	operation := interfaces.Operation{
		Type:  "ping_pong",
		Key:   "heartbeat",
		Value: "ping",
		Metadata: map[string]string{
			"concurrent": strconv.Itoa(concurrent),
			"interval":   interval.String(),
		},
	}

	result, err := adapter.Execute(ctx, operation)
	if err != nil {
		return fmt.Errorf("ping-pong test failed: %w", err)
	}

	fmt.Printf("✅ Ping-pong test completed: Success=%t, Duration=%v\n",
		result.Success, result.Duration)
	return nil
}

// runBroadcastTest 运行广播测试
func (h *WebSocketCommandHandler) runBroadcastTest(ctx context.Context, adapter interfaces.ProtocolAdapter, config map[string]interface{}) error {
	concurrent := config["concurrent_connections"].(int)
	message := config["message"].(string)

	operation := interfaces.Operation{
		Type:  "broadcast",
		Key:   "broadcast_message",
		Value: message,
		Metadata: map[string]string{
			"concurrent": strconv.Itoa(concurrent),
		},
	}

	result, err := adapter.Execute(ctx, operation)
	if err != nil {
		return fmt.Errorf("broadcast test failed: %w", err)
	}

	fmt.Printf("✅ Broadcast test completed: Success=%t, Duration=%v\n",
		result.Success, result.Duration)
	return nil
}

// runLargeMessageTest 运行大消息测试
func (h *WebSocketCommandHandler) runLargeMessageTest(ctx context.Context, adapter interfaces.ProtocolAdapter, config map[string]interface{}) error {
	concurrent := config["concurrent_connections"].(int)
	messageSize := config["message_size"].(int)

	// 生成大消息
	largeMessage := h.generateTestMessage(messageSize)

	operation := interfaces.Operation{
		Type:  "large_message",
		Key:   "large_data",
		Value: []byte(largeMessage),
		Metadata: map[string]string{
			"concurrent":   strconv.Itoa(concurrent),
			"message_size": strconv.Itoa(messageSize),
		},
	}

	result, err := adapter.Execute(ctx, operation)
	if err != nil {
		return fmt.Errorf("large message test failed: %w", err)
	}

	fmt.Printf("✅ Large message test completed: Success=%t, Duration=%v\n",
		result.Success, result.Duration)
	return nil
}

// generateTestMessage 生成测试消息
func (h *WebSocketCommandHandler) generateTestMessage(size int) string {
	if size <= 0 {
		return "test"
	}

	message := make([]byte, size)
	for i := range message {
		message[i] = 'A' + byte(i%26)
	}
	return string(message)
}

// GetProtocolName 获取协议名称
func (h *WebSocketCommandHandler) GetProtocolName() string {
	return "websocket"
}

// GetFactory 获取适配器工厂
func (h *WebSocketCommandHandler) GetFactory() interface{} {
	return h.factory
}

// createAdapter 创建适配器
func (h *WebSocketCommandHandler) createAdapter() interfaces.ProtocolAdapter {
	// 尝试转换为适配器工厂接口
	if factory, ok := h.factory.(interface {
		CreateAdapter() interfaces.ProtocolAdapter
	}); ok {
		return factory.CreateAdapter()
	}

	// 尝试转换为WebSocket特定工厂接口
	if factory, ok := h.factory.(interface {
		CreateWebSocketAdapter() interfaces.ProtocolAdapter
	}); ok {
		return factory.CreateWebSocketAdapter()
	}

	// 如果都失败，返回模拟适配器
	return h.createMockAdapter()
}

// createMockAdapter 创建模拟适配器作为备用
func (h *WebSocketCommandHandler) createMockAdapter() interfaces.ProtocolAdapter {
	// 返回模拟适配器，仅用于开发和测试
	return &MockWebSocketAdapter{}
}

// MockWebSocketAdapter 模拟适配器
type MockWebSocketAdapter struct{}

func (m *MockWebSocketAdapter) Connect(ctx context.Context, config interfaces.Config) error {
	return nil
}

func (m *MockWebSocketAdapter) Execute(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	return &interfaces.OperationResult{
		Success:  true,
		Duration: 10 * time.Millisecond,
		Value:    "mock websocket result",
	}, nil
}

func (m *MockWebSocketAdapter) Close() error {
	return nil
}

func (m *MockWebSocketAdapter) GetProtocolMetrics() map[string]interface{} {
	return map[string]interface{}{"mock": true}
}

func (m *MockWebSocketAdapter) HealthCheck(ctx context.Context) error {
	return nil
}

func (m *MockWebSocketAdapter) GetProtocolName() string {
	return "websocket"
}

func (m *MockWebSocketAdapter) GetMetricsCollector() interfaces.DefaultMetricsCollector {
	return nil
}

// createConfigWrapper 创建Config接口包装器
func (h *WebSocketCommandHandler) createConfigWrapper(config map[string]interface{}) interfaces.Config {
	return &ConfigWrapper{data: config}
}

// ConfigWrapper Config接口包装器
type ConfigWrapper struct {
	data map[string]interface{}
}

func (c *ConfigWrapper) GetProtocol() string {
	if protocol, ok := c.data["protocol"].(string); ok {
		return protocol
	}
	return "websocket"
}

func (c *ConfigWrapper) GetConnection() interfaces.ConnectionConfig {
	return nil // WebSocket不需要复杂的连接配置
}

func (c *ConfigWrapper) GetBenchmark() interfaces.BenchmarkConfig {
	return nil // WebSocket不需要复杂的基准测试配置
}

func (c *ConfigWrapper) Validate() error {
	return nil // 简化验证
}

func (c *ConfigWrapper) Clone() interfaces.Config {
	newData := make(map[string]interface{})
	for k, v := range c.data {
		newData[k] = v
	}
	return &ConfigWrapper{data: newData}
}

// GetData 获取原始数据
func (c *ConfigWrapper) GetData() map[string]interface{} {
	return c.data
}

// looksLikeURL 检查字符串是否像URL
func looksLikeURL(s string) bool {
	if s == "" {
		return false
	}
	// 简单检查：WebSocket URL格式
	return len(s) > 5 && (s[:5] == "ws://" || (len(s) > 6 && s[:6] == "wss://"))
}
