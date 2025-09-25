package config

import (
	"fmt"
	"testing"
)

func TestConfigValidator(t *testing.T) {
	validator := NewConfigValidator()

	t.Run("Valid Default Config", func(t *testing.T) {
		config := NewDefaultRedisConfig()
		err := validator.Validate(config)
		if err != nil {
			t.Errorf("Default config should be valid: %v", err)
		}
	})

	t.Run("Add Custom Rule", func(t *testing.T) {
		customRule := func(config *RedisConfig) error {
			if config.BenchMark.Total > 1000000 {
				return fmt.Errorf("total requests too high: %d", config.BenchMark.Total)
			}
			return nil
		}

		validator.AddRule(customRule)

		config := NewDefaultRedisConfig()
		config.BenchMark.Total = 2000000

		err := validator.Validate(config)
		if err == nil {
			t.Error("Validation should fail for high total requests")
		}
	})
}

func TestValidateProtocol(t *testing.T) {
	t.Run("Valid Protocol", func(t *testing.T) {
		config := NewDefaultRedisConfig()
		config.Protocol = "redis"

		err := validateProtocol(config)
		if err != nil {
			t.Errorf("Protocol validation should pass: %v", err)
		}
	})

	t.Run("Empty Protocol", func(t *testing.T) {
		config := NewDefaultRedisConfig()
		config.Protocol = ""

		err := validateProtocol(config)
		if err == nil {
			t.Error("Empty protocol should fail validation")
		}
	})

	t.Run("Invalid Protocol", func(t *testing.T) {
		config := NewDefaultRedisConfig()
		config.Protocol = "invalid"

		err := validateProtocol(config)
		if err == nil {
			t.Error("Invalid protocol should fail validation")
		}
	})
}

func TestValidateMode(t *testing.T) {
	validModes := []string{"standalone", "sentinel", "cluster"}

	for _, mode := range validModes {
		t.Run("Valid Mode: "+mode, func(t *testing.T) {
			config := NewDefaultRedisConfig()
			config.Mode = mode

			err := validateMode(config)
			if err != nil {
				t.Errorf("Mode '%s' should be valid: %v", mode, err)
			}
		})
	}

	t.Run("Invalid Mode", func(t *testing.T) {
		config := NewDefaultRedisConfig()
		config.Mode = "invalid"

		err := validateMode(config)
		if err == nil {
			t.Error("Invalid mode should fail validation")
		}
	})

	t.Run("Empty Mode", func(t *testing.T) {
		config := NewDefaultRedisConfig()
		config.Mode = ""

		// 空模式应该通过验证，因为GetMode()会返回默认值
		err := validateMode(config)
		if err != nil {
			t.Errorf("Empty mode should be valid (defaults to standalone): %v", err)
		}
	})
}

func TestValidateConnection(t *testing.T) {
	t.Run("Valid Standalone", func(t *testing.T) {
		config := NewDefaultRedisConfig()
		config.Mode = "standalone"
		config.Standalone.Addr = "localhost:6379"

		err := validateConnection(config)
		if err != nil {
			t.Errorf("Valid standalone config should pass: %v", err)
		}
	})

	t.Run("Invalid Standalone - Empty Addr", func(t *testing.T) {
		config := NewDefaultRedisConfig()
		config.Mode = "standalone"
		config.Standalone.Addr = ""

		err := validateConnection(config)
		if err == nil {
			t.Error("Empty standalone addr should fail validation")
		}
	})

	t.Run("Invalid Standalone - Bad Format", func(t *testing.T) {
		config := NewDefaultRedisConfig()
		config.Mode = "standalone"
		config.Standalone.Addr = "invalid-address"

		err := validateConnection(config)
		if err == nil {
			t.Error("Invalid address format should fail validation")
		}
	})

	t.Run("Valid Cluster", func(t *testing.T) {
		config := NewDefaultRedisConfig()
		config.Mode = "cluster"
		config.Cluster.Addrs = []string{"node1:6379", "node2:6379", "node3:6379"}

		err := validateConnection(config)
		if err != nil {
			t.Errorf("Valid cluster config should pass: %v", err)
		}
	})

	t.Run("Invalid Cluster - Empty Addrs", func(t *testing.T) {
		config := NewDefaultRedisConfig()
		config.Mode = "cluster"
		config.Cluster.Addrs = []string{}

		err := validateConnection(config)
		if err == nil {
			t.Error("Empty cluster addrs should fail validation")
		}
	})

	t.Run("Invalid Cluster - Bad Address Format", func(t *testing.T) {
		config := NewDefaultRedisConfig()
		config.Mode = "cluster"
		config.Cluster.Addrs = []string{"node1:6379", "invalid-node"}

		err := validateConnection(config)
		if err == nil {
			t.Error("Invalid cluster address format should fail validation")
		}
	})

	t.Run("Valid Sentinel", func(t *testing.T) {
		config := NewDefaultRedisConfig()
		config.Mode = "sentinel"
		config.Sentinel.Addrs = []string{"sentinel1:26379", "sentinel2:26379"}
		config.Sentinel.MasterName = "mymaster"

		err := validateConnection(config)
		if err != nil {
			t.Errorf("Valid sentinel config should pass: %v", err)
		}
	})

	t.Run("Invalid Sentinel - Empty Addrs", func(t *testing.T) {
		config := NewDefaultRedisConfig()
		config.Mode = "sentinel"
		config.Sentinel.Addrs = []string{}
		config.Sentinel.MasterName = "mymaster"

		err := validateConnection(config)
		if err == nil {
			t.Error("Empty sentinel addrs should fail validation")
		}
	})

	t.Run("Invalid Sentinel - Empty MasterName", func(t *testing.T) {
		config := NewDefaultRedisConfig()
		config.Mode = "sentinel"
		config.Sentinel.Addrs = []string{"sentinel1:26379"}
		config.Sentinel.MasterName = ""

		err := validateConnection(config)
		if err == nil {
			t.Error("Empty sentinel master name should fail validation")
		}
	})
}

func TestValidateBenchmark(t *testing.T) {
	t.Run("Valid Benchmark", func(t *testing.T) {
		config := NewDefaultRedisConfig()

		err := validateBenchmark(config)
		if err != nil {
			t.Errorf("Default benchmark config should be valid: %v", err)
		}
	})

	t.Run("Invalid Total", func(t *testing.T) {
		config := NewDefaultRedisConfig()
		config.BenchMark.Total = 0

		err := validateBenchmark(config)
		if err == nil {
			t.Error("Zero total should fail validation")
		}
	})

	t.Run("Invalid Parallels", func(t *testing.T) {
		config := NewDefaultRedisConfig()
		config.BenchMark.Parallels = -1

		err := validateBenchmark(config)
		if err == nil {
			t.Error("Negative parallels should fail validation")
		}
	})

	t.Run("Invalid DataSize", func(t *testing.T) {
		config := NewDefaultRedisConfig()
		config.BenchMark.DataSize = 0

		err := validateBenchmark(config)
		if err == nil {
			t.Error("Zero data size should fail validation")
		}
	})

	t.Run("Invalid ReadPercent - Too High", func(t *testing.T) {
		config := NewDefaultRedisConfig()
		config.BenchMark.ReadPercent = 150

		err := validateBenchmark(config)
		if err == nil {
			t.Error("Read percent > 100 should fail validation")
		}
	})

	t.Run("Invalid ReadPercent - Negative", func(t *testing.T) {
		config := NewDefaultRedisConfig()
		config.BenchMark.ReadPercent = -10

		err := validateBenchmark(config)
		if err == nil {
			t.Error("Negative read percent should fail validation")
		}
	})

	t.Run("Invalid RandomKeys", func(t *testing.T) {
		config := NewDefaultRedisConfig()
		config.BenchMark.RandomKeys = -1

		err := validateBenchmark(config)
		if err == nil {
			t.Error("Negative random keys should fail validation")
		}
	})

	t.Run("Invalid TTL", func(t *testing.T) {
		config := NewDefaultRedisConfig()
		config.BenchMark.TTL = -1

		err := validateBenchmark(config)
		if err == nil {
			t.Error("Negative TTL should fail validation")
		}
	})

	t.Run("Valid Test Cases", func(t *testing.T) {
		validCases := []string{"get", "set", "set_get", "set_get_random", "pub", "sub"}

		for _, testCase := range validCases {
			config := NewDefaultRedisConfig()
			config.BenchMark.Case = testCase

			err := validateBenchmark(config)
			if err != nil {
				t.Errorf("Test case '%s' should be valid: %v", testCase, err)
			}
		}
	})

	t.Run("Invalid Test Case", func(t *testing.T) {
		config := NewDefaultRedisConfig()
		config.BenchMark.Case = "invalid_case"

		err := validateBenchmark(config)
		if err == nil {
			t.Error("Invalid test case should fail validation")
		}
	})

	t.Run("Empty Test Case", func(t *testing.T) {
		config := NewDefaultRedisConfig()
		config.BenchMark.Case = ""

		err := validateBenchmark(config)
		if err != nil {
			t.Errorf("Empty test case should be valid: %v", err)
		}
	})
}

func TestIsValidAddr(t *testing.T) {
	testCases := []struct {
		addr     string
		expected bool
		desc     string
	}{
		{"localhost:6379", true, "localhost with port"},
		{"127.0.0.1:6379", true, "IP with port"},
		{"redis.example.com:6379", true, "domain with port"},
		{"localhost", false, "missing port"},
		{":6379", false, "missing host"},
		{"localhost:", false, "missing port number"},
		{"", false, "empty address"},
		{"localhost:abc", false, "invalid port"},
		{"localhost:6379:extra", false, "too many colons"},
		{"redis-server:26379", true, "hyphenated hostname"},
		{"192.168.1.100:6379", true, "IP address"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result := isValidAddr(tc.addr)
			if result != tc.expected {
				t.Errorf("isValidAddr('%s') = %v, expected %v", tc.addr, result, tc.expected)
			}
		})
	}
}

// 基准测试
func BenchmarkValidation(b *testing.B) {
	config := NewDefaultRedisConfig()
	validator := NewConfigValidator()

	b.Run("Protocol Validation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = validateProtocol(config)
		}
	})

	b.Run("Mode Validation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = validateMode(config)
		}
	})

	b.Run("Connection Validation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = validateConnection(config)
		}
	})

	b.Run("Benchmark Validation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = validateBenchmark(config)
		}
	})

	b.Run("Full Validation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = validator.Validate(config)
		}
	})
}
