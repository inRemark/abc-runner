package operations

import (
	"testing"
	"time"

	"abc-runner/app/adapters/tcp/config"
)

// MockBenchmarkConfig for testing
type MockBenchmarkConfig struct {
	total     int
	parallels int
	duration  time.Duration
}

func (m *MockBenchmarkConfig) GetTotal() int              { return m.total }
func (m *MockBenchmarkConfig) GetParallels() int          { return m.parallels }
func (m *MockBenchmarkConfig) GetDuration() time.Duration { return m.duration }
func (m *MockBenchmarkConfig) GetTimeout() time.Duration  { return 30 * time.Second }
func (m *MockBenchmarkConfig) GetRampUp() time.Duration   { return 0 }

func TestOperationFactory_CreateOperation(t *testing.T) {
	// 创建测试配置
	cfg := config.NewDefaultTCPConfig()
	factory := NewOperationFactory(cfg)

	testCases := []struct {
		name     string
		testCase string
		jobID    int
		wantType string
	}{
		{
			name:     "echo_test operation",
			testCase: "echo_test",
			jobID:    1,
			wantType: "echo_test",
		},
		{
			name:     "send_only operation",
			testCase: "send_only",
			jobID:    2,
			wantType: "send_only",
		},
		{
			name:     "receive_only operation",
			testCase: "receive_only",
			jobID:    3,
			wantType: "receive_only",
		},
		{
			name:     "bidirectional operation",
			testCase: "bidirectional",
			jobID:    4,
			wantType: "bidirectional",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 更新配置中的测试用例
			cfg.BenchMark.TestCase = tc.testCase
			factory = NewOperationFactory(cfg)

			config := &MockBenchmarkConfig{
				total:     100,
				parallels: 10,
				duration:  60 * time.Second,
			}

			op := factory.CreateOperation(tc.jobID, config)

			// 验证操作类型
			if op.Type != tc.wantType {
				t.Errorf("Expected operation type %s, got %s", tc.wantType, op.Type)
			}

			// 验证键生成
			expectedKeyPrefix := ""
			switch tc.testCase {
			case "echo_test":
				expectedKeyPrefix = "tcp_echo_"
			case "send_only":
				expectedKeyPrefix = "tcp_send_"
			case "receive_only":
				expectedKeyPrefix = "tcp_recv_"
			case "bidirectional":
				expectedKeyPrefix = "tcp_bidi_"
			}

			if len(expectedKeyPrefix) > 0 && len(op.Key) > 0 {
				if op.Key[:len(expectedKeyPrefix)] != expectedKeyPrefix {
					t.Errorf("Expected key to start with %s, got %s", expectedKeyPrefix, op.Key)
				}
			}

			// 验证参数
			if jobID, ok := op.Params["job_id"]; !ok || jobID != tc.jobID {
				t.Errorf("Expected job_id %d, got %v", tc.jobID, jobID)
			}

			if testCase, ok := op.Params["test_case"]; !ok || testCase != tc.testCase {
				t.Errorf("Expected test_case %s, got %v", tc.testCase, testCase)
			}

			// 验证元数据
			if opType, ok := op.Metadata["operation_type"]; !ok || opType != tc.testCase {
				t.Errorf("Expected metadata operation_type %s, got %v", tc.testCase, opType)
			}

			if protocol, ok := op.Metadata["protocol"]; !ok || protocol != "tcp" {
				t.Errorf("Expected metadata protocol tcp, got %v", protocol)
			}
		})
	}
}

func TestOperationFactory_GetOperationType(t *testing.T) {
	cfg := config.NewDefaultTCPConfig()
	cfg.BenchMark.TestCase = "echo_test"
	factory := NewOperationFactory(cfg)

	opType := factory.GetOperationType()
	if opType != "echo_test" {
		t.Errorf("Expected operation type echo_test, got %s", opType)
	}
}

func TestOperationFactory_GetSupportedOperations(t *testing.T) {
	cfg := config.NewDefaultTCPConfig()
	factory := NewOperationFactory(cfg)

	supportedOps := factory.GetSupportedOperations()
	expectedOps := []string{"echo_test", "send_only", "receive_only", "bidirectional"}

	if len(supportedOps) != len(expectedOps) {
		t.Fatalf("Expected %d supported operations, got %d", len(expectedOps), len(supportedOps))
	}

	for i, expected := range expectedOps {
		if supportedOps[i] != expected {
			t.Errorf("Expected operation %s at index %d, got %s", expected, i, supportedOps[i])
		}
	}
}

func TestOperationFactory_generateTestData(t *testing.T) {
	cfg := config.NewDefaultTCPConfig()
	factory := NewOperationFactory(cfg)

	testCases := []struct {
		name     string
		testCase string
		jobID    int
		dataSize int
	}{
		{"echo_test", "echo_test", 1, 100},
		{"send_only", "send_only", 2, 200},
		{"receive_only", "receive_only", 3, 0}, // receive_only returns empty data
		{"bidirectional", "bidirectional", 4, 300},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 设置配置
			cfg.BenchMark.TestCase = tc.testCase
			cfg.BenchMark.DataSize = tc.dataSize
			factory = NewOperationFactory(cfg)

			// 生成测试数据
			data := factory.generateTestData(tc.jobID)

			if tc.testCase == "receive_only" {
				// receive_only should return empty data
				if len(data) != 0 {
					t.Errorf("Expected empty data for receive_only, got %d bytes", len(data))
				}
			} else {
				// Other test cases should return data of expected size
				if len(data) != tc.dataSize {
					t.Errorf("Expected data size %d, got %d", tc.dataSize, len(data))
				}
			}

			// 验证数据模式
			switch tc.testCase {
			case "echo_test":
				// Should contain pattern with job ID
				dataStr := string(data)
				if len(data) > 0 && !contains(dataStr, "ECHO_TEST_") {
					t.Error("Expected echo_test data to contain ECHO_TEST_ pattern")
				}
			case "send_only":
				// Should use incremental pattern
				if len(data) > 0 && data[0] != byte(tc.jobID%256) {
					t.Errorf("Expected first byte to be %d, got %d", tc.jobID%256, data[0])
				}
			}
		})
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}
