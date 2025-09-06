package unified

import (
	"context"
	"fmt"
	"time"

	"redis-runner/app/core/interfaces"
)

// MockRedisAdapter Redis适配器的临时实现
type MockRedisAdapter struct {
	config interfaces.Config
	connected bool
}

func (m *MockRedisAdapter) Connect(ctx context.Context, config interfaces.Config) error {
	m.config = config
	m.connected = true
	return nil
}

func (m *MockRedisAdapter) Execute(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	return &interfaces.OperationResult{
		Success:  true,
		Duration: 1 * time.Millisecond,
		IsRead:   operation.Type == "get",
		Value:    "mock_value",
	}, nil
}

func (m *MockRedisAdapter) Close() error {
	m.connected = false
	return nil
}

func (m *MockRedisAdapter) GetProtocolMetrics() map[string]interface{} {
	return map[string]interface{}{
		"protocol": "redis",
		"connected": m.connected,
		"operations_executed": 1,
	}
}

func (m *MockRedisAdapter) HealthCheck(ctx context.Context) error {
	if !m.connected {
		return fmt.Errorf("not connected")
	}
	return nil
}

func (m *MockRedisAdapter) GetProtocolName() string {
	return "redis"
}

// MockHttpAdapter HTTP适配器的临时实现
type MockHttpAdapter struct {
	config interfaces.Config
	connected bool
}

func (m *MockHttpAdapter) Connect(ctx context.Context, config interfaces.Config) error {
	m.config = config
	m.connected = true
	return nil
}

func (m *MockHttpAdapter) Execute(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	return &interfaces.OperationResult{
		Success:  true,
		Duration: 5 * time.Millisecond,
		IsRead:   operation.Type == "GET",
		Value:    "200 OK",
	}, nil
}

func (m *MockHttpAdapter) Close() error {
	m.connected = false
	return nil
}

func (m *MockHttpAdapter) GetProtocolMetrics() map[string]interface{} {
	return map[string]interface{}{
		"protocol": "http",
		"connected": m.connected,
		"requests_sent": 1,
	}
}

func (m *MockHttpAdapter) HealthCheck(ctx context.Context) error {
	if !m.connected {
		return fmt.Errorf("not connected")
	}
	return nil
}

func (m *MockHttpAdapter) GetProtocolName() string {
	return "http"
}

// MockKafkaAdapter Kafka适配器的临时实现
type MockKafkaAdapter struct {
	config interfaces.Config
	connected bool
}

func (m *MockKafkaAdapter) Connect(ctx context.Context, config interfaces.Config) error {
	m.config = config
	m.connected = true
	return nil
}

func (m *MockKafkaAdapter) Execute(ctx context.Context, operation interfaces.Operation) (*interfaces.OperationResult, error) {
	return &interfaces.OperationResult{
		Success:  true,
		Duration: 2 * time.Millisecond,
		IsRead:   operation.Type == "consume",
		Value:    "message_received",
	}, nil
}

func (m *MockKafkaAdapter) Close() error {
	m.connected = false
	return nil
}

func (m *MockKafkaAdapter) GetProtocolMetrics() map[string]interface{} {
	return map[string]interface{}{
		"protocol": "kafka",
		"connected": m.connected,
		"messages_processed": 1,
	}
}

func (m *MockKafkaAdapter) HealthCheck(ctx context.Context) error {
	if !m.connected {
		return fmt.Errorf("not connected")
	}
	return nil
}

func (m *MockKafkaAdapter) GetProtocolName() string {
	return "kafka"
}

// MockConfig 临时配置实现
type MockConfig struct {
	protocol string
	addresses []string
}

func NewMockConfig(protocol string, addresses []string) *MockConfig {
	return &MockConfig{
		protocol: protocol,
		addresses: addresses,
	}
}

func (m *MockConfig) GetProtocol() string {
	return m.protocol
}

func (m *MockConfig) GetConnection() interfaces.ConnectionConfig {
	return &MockConnectionConfig{addresses: m.addresses}
}

func (m *MockConfig) GetBenchmark() interfaces.BenchmarkConfig {
	return &MockBenchmarkConfig{}
}

func (m *MockConfig) Validate() error {
	if m.protocol == "" {
		return fmt.Errorf("protocol cannot be empty")
	}
	return nil
}

func (m *MockConfig) Clone() interfaces.Config {
	return &MockConfig{
		protocol: m.protocol,
		addresses: append([]string{}, m.addresses...),
	}
}

// MockConnectionConfig 临时连接配置实现
type MockConnectionConfig struct {
	addresses []string
}

func (m *MockConnectionConfig) GetAddresses() []string {
	return m.addresses
}

func (m *MockConnectionConfig) GetCredentials() map[string]string {
	return map[string]string{}
}

func (m *MockConnectionConfig) GetPoolConfig() interfaces.PoolConfig {
	return &MockPoolConfig{}
}

func (m *MockConnectionConfig) GetTimeout() time.Duration {
	return 30 * time.Second
}

// MockBenchmarkConfig 临时基准测试配置实现
type MockBenchmarkConfig struct{}

func (m *MockBenchmarkConfig) GetTotal() int { return 1000 }
func (m *MockBenchmarkConfig) GetParallels() int { return 10 }
func (m *MockBenchmarkConfig) GetDataSize() int { return 1024 }
func (m *MockBenchmarkConfig) GetTTL() time.Duration { return 0 }
func (m *MockBenchmarkConfig) GetReadPercent() int { return 50 }
func (m *MockBenchmarkConfig) GetRandomKeys() int { return 1000 }
func (m *MockBenchmarkConfig) GetTestCase() string { return "default" }

// MockPoolConfig 临时连接池配置实现
type MockPoolConfig struct{}

func (m *MockPoolConfig) GetPoolSize() int { return 10 }
func (m *MockPoolConfig) GetMinIdle() int { return 1 }
func (m *MockPoolConfig) GetMaxIdle() int { return 5 }
func (m *MockPoolConfig) GetIdleTimeout() time.Duration { return 5 * time.Minute }
func (m *MockPoolConfig) GetConnectionTimeout() time.Duration { return 30 * time.Second }