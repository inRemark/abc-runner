package main

import (
	"fmt"
	"log"

	"abc-runner/app/core/config"
	"abc-runner/app/core/interfaces"
)

// mockConfigSourceFactory 模拟配置源工厂
type mockConfigSourceFactory struct{}

func (m *mockConfigSourceFactory) CreateRedisConfigSource() interfaces.ConfigSource {
	return nil
}

func (m *mockConfigSourceFactory) CreateHttpConfigSource() interfaces.ConfigSource {
	return nil
}

func (m *mockConfigSourceFactory) CreateKafkaConfigSource() interfaces.ConfigSource {
	return nil
}

func main() {
	// 创建配置管理器
	manager := config.NewConfigManager(&mockConfigSourceFactory{})
	
	// 加载核心配置
	err := manager.LoadCoreConfiguration("")
	if err != nil {
		log.Printf("Warning: Failed to load core config: %v", err)
		// 使用默认配置继续
	}
	
	coreConfig := manager.GetCoreConfig()
	
	// 显示核心配置信息
	fmt.Println("=== Core Configuration ===")
	fmt.Printf("Logging Level: %s\n", coreConfig.Core.Logging.Level)
	fmt.Printf("Reports Enabled: %t\n", coreConfig.Core.Reports.Enabled)
	fmt.Printf("Monitoring Enabled: %t\n", coreConfig.Core.Monitoring.Enabled)
	
	// 显示报告配置
	fmt.Println("\n=== Reports Configuration ===")
	fmt.Printf("Formats: %v\n", coreConfig.Core.Reports.Formats)
	fmt.Printf("Output Directory: %s\n", coreConfig.Core.Reports.OutputDir)
	fmt.Printf("File Prefix: %s\n", coreConfig.Core.Reports.FilePrefix)
	
	// 显示监控配置
	fmt.Println("\n=== Monitoring Configuration ===")
	fmt.Printf("Metrics Interval: %v\n", coreConfig.Core.Monitoring.MetricsInterval)
	fmt.Printf("Prometheus Enabled: %t\n", coreConfig.Core.Monitoring.Prometheus.Enabled)
	if coreConfig.Core.Monitoring.Prometheus.Enabled {
		fmt.Printf("Prometheus Port: %d\n", coreConfig.Core.Monitoring.Prometheus.Port)
	}
	
	fmt.Printf("StatsD Enabled: %t\n", coreConfig.Core.Monitoring.Statsd.Enabled)
	if coreConfig.Core.Monitoring.Statsd.Enabled {
		fmt.Printf("StatsD Host: %s\n", coreConfig.Core.Monitoring.Statsd.Host)
	}
}