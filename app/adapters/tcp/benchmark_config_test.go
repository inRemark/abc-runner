package tcp

import (
	"testing"
	"time"

	"abc-runner/app/adapters/tcp/config"
	"abc-runner/app/core/execution"
)

func TestBenchmarkConfigAdapter_Basic(t *testing.T) {
	// 创建TCP配置
	tcpConfig := config.NewDefaultTCPConfig()
	
	// 修改一些值进行测试
	tcpConfig.BenchMark.Total = 500
	tcpConfig.BenchMark.Parallels = 20
	tcpConfig.BenchMark.DataSize = 2048
	tcpConfig.BenchMark.Duration = 120 * time.Second
	tcpConfig.BenchMark.TestCase = "bidirectional"
	tcpConfig.BenchMark.ReadPercent = 70

	// 创建适配器
	adapter := NewBenchmarkConfigAdapter(&tcpConfig.BenchMark)

	// 测试基本方法
	if adapter.GetTotal() != 500 {
		t.Errorf("Expected total 500, got %d", adapter.GetTotal())
	}

	if adapter.GetParallels() != 20 {
		t.Errorf("Expected parallels 20, got %d", adapter.GetParallels())
	}

	if adapter.GetDataSize() != 2048 {
		t.Errorf("Expected data size 2048, got %d", adapter.GetDataSize())
	}

	if adapter.GetDuration() != 120*time.Second {
		t.Errorf("Expected duration 120s, got %v", adapter.GetDuration())
	}

	if adapter.GetTestCase() != "bidirectional" {
		t.Errorf("Expected test case bidirectional, got %s", adapter.GetTestCase())
	}

	if adapter.GetReadPercent() != 70 {
		t.Errorf("Expected read percent 70, got %d", adapter.GetReadPercent())
	}
}

func TestBenchmarkConfigAdapter_DefaultValues(t *testing.T) {
	// 创建默认配置
	tcpConfig := config.NewDefaultTCPConfig()
	adapter := NewBenchmarkConfigAdapter(&tcpConfig.BenchMark)

	// 测试默认值
	if adapter.GetTotal() != 1000 {
		t.Errorf("Expected default total 1000, got %d", adapter.GetTotal())
	}

	if adapter.GetParallels() != 10 {
		t.Errorf("Expected default parallels 10, got %d", adapter.GetParallels())
	}

	if adapter.GetDataSize() != 1024 {
		t.Errorf("Expected default data size 1024, got %d", adapter.GetDataSize())
	}

	if adapter.GetTestCase() != "echo_test" {
		t.Errorf("Expected default test case echo_test, got %s", adapter.GetTestCase())
	}

	if adapter.GetReadPercent() != 80 {
		t.Errorf("Expected default read percent 80, got %d", adapter.GetReadPercent())
	}

	// 测试TCP特定的默认值
	if adapter.GetTimeout() != 30*time.Second {
		t.Errorf("Expected default timeout 30s, got %v", adapter.GetTimeout())
	}

	if adapter.GetRampUp() != 5*time.Second {
		t.Errorf("Expected default ramp up 5s, got %v", adapter.GetRampUp())
	}
}

func TestBenchmarkConfigAdapter_Duration(t *testing.T) {
	testCases := []struct {
		name            string
		inputDuration   time.Duration
		expectedDuration time.Duration
	}{
		{
			name:            "custom duration",
			inputDuration:   45 * time.Second,
			expectedDuration: 45 * time.Second,
		},
		{
			name:            "zero duration",
			inputDuration:   0,
			expectedDuration: 0,
		},
		{
			name:            "long duration",
			inputDuration:   10 * time.Minute,
			expectedDuration: 10 * time.Minute,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tcpConfig := config.NewDefaultTCPConfig()
			tcpConfig.BenchMark.Duration = tc.inputDuration
			
			adapter := NewBenchmarkConfigAdapter(&tcpConfig.BenchMark)
			
			if adapter.GetDuration() != tc.expectedDuration {
				t.Errorf("Expected duration %v, got %v", tc.expectedDuration, adapter.GetDuration())
			}
		})
	}
}

func TestBenchmarkConfigAdapter_TestCases(t *testing.T) {
	testCases := []string{"echo_test", "send_only", "receive_only", "bidirectional"}

	for _, testCase := range testCases {
		t.Run(testCase, func(t *testing.T) {
			tcpConfig := config.NewDefaultTCPConfig()
			tcpConfig.BenchMark.TestCase = testCase
			
			adapter := NewBenchmarkConfigAdapter(&tcpConfig.BenchMark)
			
			if adapter.GetTestCase() != testCase {
				t.Errorf("Expected test case %s, got %s", testCase, adapter.GetTestCase())
			}
		})
	}
}

func TestBenchmarkConfigAdapter_ExecutionBenchmarkConfigInterface(t *testing.T) {
	// 验证适配器实现了execution.BenchmarkConfig接口
	tcpConfig := config.NewDefaultTCPConfig()
	adapter := NewBenchmarkConfigAdapter(&tcpConfig.BenchMark)

	// 这个测试主要是编译时检查，确保接口实现正确
	var _ execution.BenchmarkConfig = adapter

	// 测试接口方法
	if adapter.GetTimeout() <= 0 {
		t.Error("Expected positive timeout value")
	}

	if adapter.GetRampUp() < 0 {
		t.Error("Expected non-negative ramp up value")
	}
}

func TestBenchmarkConfigAdapter_EdgeCases(t *testing.T) {
	// 测试边界情况
	tcpConfig := config.NewDefaultTCPConfig()
	
	// 测试极小值
	tcpConfig.BenchMark.Total = 1
	tcpConfig.BenchMark.Parallels = 1
	tcpConfig.BenchMark.DataSize = 1
	
	adapter := NewBenchmarkConfigAdapter(&tcpConfig.BenchMark)
	
	if adapter.GetTotal() != 1 {
		t.Errorf("Expected total 1, got %d", adapter.GetTotal())
	}
	
	if adapter.GetParallels() != 1 {
		t.Errorf("Expected parallels 1, got %d", adapter.GetParallels())
	}
	
	if adapter.GetDataSize() != 1 {
		t.Errorf("Expected data size 1, got %d", adapter.GetDataSize())
	}
	
	// 测试大值
	tcpConfig.BenchMark.Total = 1000000
	tcpConfig.BenchMark.Parallels = 1000
	tcpConfig.BenchMark.DataSize = 1048576 // 1MB
	
	adapter = NewBenchmarkConfigAdapter(&tcpConfig.BenchMark)
	
	if adapter.GetTotal() != 1000000 {
		t.Errorf("Expected total 1000000, got %d", adapter.GetTotal())
	}
	
	if adapter.GetParallels() != 1000 {
		t.Errorf("Expected parallels 1000, got %d", adapter.GetParallels())
	}
	
	if adapter.GetDataSize() != 1048576 {
		t.Errorf("Expected data size 1048576, got %d", adapter.GetDataSize())
	}
}