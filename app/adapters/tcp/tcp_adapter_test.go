package tcp

import (
	"testing"

	"abc-runner/app/adapters/tcp/config"
	"abc-runner/app/core/metrics"
)

func TestNewTCPAdapter(t *testing.T) {
	// 创建指标收集器
	metricsConfig := metrics.DefaultMetricsConfig()
	collector := metrics.NewBaseCollector(metricsConfig, map[string]interface{}{
		"protocol": "tcp",
	})

	// 创建TCP适配器
	adapter := NewTCPAdapter(collector)

	// 验证基本属性
	if adapter == nil {
		t.Fatal("NewTCPAdapter returned nil")
	}

	if adapter.GetProtocolName() != "tcp" {
		t.Errorf("Expected protocol name 'tcp', got '%s'", adapter.GetProtocolName())
	}

	if adapter.GetMetricsCollector() != collector {
		t.Error("GetMetricsCollector returned wrong collector")
	}

	if adapter.isConnected {
		t.Error("New adapter should not be connected")
	}
}

func TestTCPConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		modifyConfig func(*config.TCPConfig)
		expectError bool
	}{
		{
			name:        "default config should be valid",
			modifyConfig: func(c *config.TCPConfig) {},
			expectError: false,
		},
		{
			name: "empty address should fail",
			modifyConfig: func(c *config.TCPConfig) {
				c.Connection.Address = ""
			},
			expectError: true,
		},
		{
			name: "invalid port should fail",
			modifyConfig: func(c *config.TCPConfig) {
				c.Connection.Port = 0
			},
			expectError: true,
		},
		{
			name: "zero total operations should fail",
			modifyConfig: func(c *config.TCPConfig) {
				c.BenchMark.Total = 0
			},
			expectError: true,
		},
		{
			name: "invalid test case should fail",
			modifyConfig: func(c *config.TCPConfig) {
				c.BenchMark.TestCase = "invalid_test"
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.NewDefaultTCPConfig()
			tt.modifyConfig(config)

			err := config.Validate()
			if tt.expectError && err == nil {
				t.Error("Expected validation error, but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no validation error, but got: %v", err)
			}
		})
	}
}

func TestTCPConfigClone(t *testing.T) {
	original := config.NewDefaultTCPConfig()
	original.Connection.Address = "test.example.com"
	original.Connection.Port = 12345

	cloned := original.Clone().(*config.TCPConfig)

	if cloned.Connection.Address != original.Connection.Address {
		t.Error("Cloned config address doesn't match original")
	}

	if cloned.Connection.Port != original.Connection.Port {
		t.Error("Cloned config port doesn't match original")
	}

	// 修改克隆应该不影响原始
	cloned.Connection.Address = "modified.example.com"
	if original.Connection.Address == "modified.example.com" {
		t.Error("Modifying cloned config affected original")
	}
}

func TestTCPConfigInterfaces(t *testing.T) {
	config := config.NewDefaultTCPConfig()

	// 测试Config接口
	if config.GetProtocol() != "tcp" {
		t.Errorf("Expected protocol 'tcp', got '%s'", config.GetProtocol())
	}

	// 测试ConnectionConfig接口
	connConfig := config.GetConnection()
	addresses := connConfig.GetAddresses()
	if len(addresses) != 1 {
		t.Errorf("Expected 1 address, got %d", len(addresses))
	}

	expectedAddr := "localhost:8080"
	if addresses[0] != expectedAddr {
		t.Errorf("Expected address '%s', got '%s'", expectedAddr, addresses[0])
	}

	// 测试BenchmarkConfig接口
	benchConfig := config.GetBenchmark()
	if benchConfig.GetTotal() != 1000 {
		t.Errorf("Expected total 1000, got %d", benchConfig.GetTotal())
	}

	if benchConfig.GetParallels() != 10 {
		t.Errorf("Expected parallels 10, got %d", benchConfig.GetParallels())
	}

	if benchConfig.GetTestCase() != "echo_test" {
		t.Errorf("Expected test case 'echo_test', got '%s'", benchConfig.GetTestCase())
	}
}