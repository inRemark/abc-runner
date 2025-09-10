package main

import (
	"fmt"
	"log"

	"abc-runner/app/core/config"
)

func main() {
	fmt.Println("=== 核心配置示例 ===")

	// 1. 创建核心配置加载器
	fmt.Println("\n1. 加载核心配置")
	loader := config.NewCoreConfigLoader()

	// 2. 加载默认核心配置
	defaultConfig := loader.GetDefaultConfig()
	fmt.Printf("默认日志级别: %s\n", defaultConfig.Core.Logging.Level)
	fmt.Printf("默认报告输出目录: %s\n", defaultConfig.Core.Reports.OutputDir)
	fmt.Printf("默认监控启用状态: %t\n", defaultConfig.Core.Monitoring.Enabled)

	// 3. 从文件加载核心配置（如果文件存在）
	coreConfig, err := loader.LoadFromFile("config/examples/core.yaml")
	if err != nil {
		log.Printf("警告: 无法加载核心配置文件: %v", err)
		fmt.Println("使用默认核心配置")
		coreConfig = defaultConfig
	} else {
		fmt.Println("成功加载核心配置文件")
		fmt.Printf("配置的日志级别: %s\n", coreConfig.Core.Logging.Level)
		fmt.Printf("配置的报告输出目录: %s\n", coreConfig.Core.Reports.OutputDir)
	}

	// 4. 使用配置管理器
	fmt.Println("\n2. 使用配置管理器")
	manager := config.NewConfigManager()

	// 加载核心配置
	err = manager.LoadCoreConfiguration("config/examples/core.yaml")
	if err != nil {
		log.Printf("警告: 无法加载核心配置: %v", err)
		fmt.Println("使用默认核心配置")
	}

	// 获取核心配置
	loadedCoreConfig := manager.GetCoreConfig()
	fmt.Printf("管理器中的日志级别: %s\n", loadedCoreConfig.Core.Logging.Level)
	fmt.Printf("管理器中的报告格式: %v\n", loadedCoreConfig.Core.Reports.Formats)

	// 5. 保存配置到文件
	fmt.Println("\n3. 保存配置到文件")
	newConfig := &config.CoreConfig{
		Core: config.CoreConfigSection{
			Logging: config.LoggingConfig{
				Level:    "debug",
				Format:   "text",
				Output:   "file",
				FilePath: "./logs",
			},
			Reports: config.ReportsConfig{
				Enabled:   true,
				Formats:   []string{"console", "json"},
				OutputDir: "./my-reports",
			},
			Monitoring: config.MonitoringConfig{
				Enabled:         true,
				MetricsInterval: 10000000000, // 10秒
				Prometheus: config.PrometheusConfig{
					Enabled: true,
					Port:    9090,
				},
			},
			Connection: config.ConnectionConfig{
				Timeout:         30000000000, // 30秒
				KeepAlive:       30000000000, // 30秒
				MaxIdleConns:    100,
				IdleConnTimeout: 90000000000, // 90秒
			},
		},
	}

	err = loader.SaveToFile(newConfig, "example_core_config.yaml")
	if err != nil {
		log.Printf("保存配置文件失败: %v", err)
	} else {
		fmt.Println("配置已保存到 example_core_config.yaml")
	}

	fmt.Println("\n=== 示例完成 ===")
}