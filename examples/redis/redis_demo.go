package main

import (
	"fmt"
	"log"
	"time"

	"redis-runner/app/adapters/redis/config"
	"redis-runner/app/adapters/redis/connection"
	"redis-runner/app/adapters/redis/metrics"
	"redis-runner/app/adapters/redis/operations"
)

// Redis适配器新架构示例
func main() {
	fmt.Println("=== Redis 新架构示例 ===")

	// 1. 创建配置
	fmt.Println("\n1. 创建和加载配置")
	cfg := config.NewDefaultRedisConfig()
	cfg.Standalone.Addr = "localhost:6379"
	fmt.Printf("配置协议: %s\n", cfg.GetProtocol())
	fmt.Printf("配置模式: %s\n", cfg.GetMode())

	// 使用配置加载器
	loader := config.NewMultiSourceConfigLoader()
	loader.AddSource(config.NewDefaultConfigSource())
	loadedCfg, err := loader.Load()
	if err != nil {
		log.Printf("配置加载失败: %v", err)
	} else {
		fmt.Printf("加载的配置协议: %s\n", loadedCfg.GetProtocol())
	}

	// 2. 创建连接管理器
	fmt.Println("\n2. 连接管理")
	clientManager := connection.NewClientManager(cfg)
	fmt.Printf("连接模式: %s\n", clientManager.GetMode())

	// 演示连接（会失败，因为没有Redis服务器运行）
	err = clientManager.Connect()
	if err != nil {
		fmt.Printf("连接失败（预期）: %v\n", err)
	}

	// 3. 操作工厂和操作管理
	fmt.Println("\n3. 操作管理")
	factory := operations.NewOperationFactory()
	
	// 列出支持的操作
	supportedOps := factory.ListSupportedOperations()
	fmt.Printf("支持的操作类型: %v\n", supportedOps)

	// 创建操作实例
	getOp, err := factory.Create(operations.OperationGet)
	if err != nil {
		log.Printf("创建Get操作失败: %v", err)
	} else {
		fmt.Printf("创建的操作类型: %s\n", getOp.GetType())
	}

	// 使用操作构建器
	builder := operations.NewOperationBuilder()
	customFactory := builder.WithPublishChannel("custom_channel").Build()
	pubOp, err := customFactory.Create(operations.OperationPublish)
	if err != nil {
		log.Printf("创建Publish操作失败: %v", err)
	} else {
		fmt.Printf("自定义Publish操作类型: %s\n", pubOp.GetType())
	}

	// 4. 指标收集和报告
	fmt.Println("\n4. 指标收集和报告")
	collector := metrics.NewMetricsCollector()

	// 模拟一些操作结果
	results := []operations.OperationResult{
		{
			Success:  true,
			IsRead:   true,
			Duration: time.Millisecond * 5,
			ExtraData: map[string]interface{}{
				"operation_type": string(operations.OperationGet),
			},
		},
		{
			Success:  true,
			IsRead:   false,
			Duration: time.Millisecond * 3,
			ExtraData: map[string]interface{}{
				"operation_type": string(operations.OperationSet),
			},
		},
		{
			Success:  false,
			IsRead:   true,
			Duration: time.Millisecond * 10,
			Error:    fmt.Errorf("模拟错误"),
			ExtraData: map[string]interface{}{
				"operation_type": string(operations.OperationGet),
			},
		},
	}

	// 收集指标
	for _, result := range results {
		collector.CollectOperation(result)
	}

	// 打印指标摘要
	collector.PrintSummary()

	// 5. 生成多格式报告
	fmt.Println("\n5. 生成报告")
	reportBuilder := metrics.NewReportBuilder(collector)
	reportBuilder.WithConsole()

	err = reportBuilder.Generate()
	if err != nil {
		log.Printf("生成报告失败: %v", err)
	}

	// 6. 连接池管理（示例）
	fmt.Println("\n6. 连接池管理")
	poolManager := connection.NewPoolManager()
	
	err = poolManager.CreatePool("default", cfg)
	if err != nil {
		fmt.Printf("创建连接池失败（预期）: %v\n", err)
	} else {
		fmt.Println("连接池创建成功")
	}

	// 7. 配置验证
	fmt.Println("\n7. 配置验证")
	if err := cfg.Validate(); err != nil {
		fmt.Printf("配置验证失败: %v\n", err)
	} else {
		fmt.Println("配置验证通过")
	}

	// 8. 演示配置克隆
	fmt.Println("\n8. 配置克隆")
	clonedCfg := cfg.Clone()
	clonedCfg.Standalone.Addr = "localhost:6380"
	fmt.Printf("原始配置地址: %s\n", cfg.Standalone.Addr)
	fmt.Printf("克隆配置地址: %s\n", clonedCfg.Standalone.Addr)

	fmt.Println("\n=== 示例完成 ===")
}