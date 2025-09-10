package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"abc-runner/app/adapters/kafka"
	"abc-runner/app/adapters/kafka/config"
	"abc-runner/app/adapters/kafka/operations"
	"abc-runner/app/core/interfaces"
)

func main() {
	fmt.Println("=== Kafka模块重构演示 ===")

	// 创建Kafka适配器（注释掉以避免未使用警告）
	// adapter := kafka.NewKafkaAdapter()

	// 创建配置
	cfg := &config.KafkaAdapterConfig{
		Brokers:  []string{"localhost:9092"},
		ClientID: "abc-runner-demo",

		Producer: config.ProducerConfig{
			Acks:        "1",
			Retries:     3,
			BatchSize:   1024,
			Compression: "none",
		},

		Consumer: config.ConsumerConfig{
			GroupID:         "demo-group",
			AutoOffsetReset: "latest",
			FetchMinBytes:   1,
			FetchMaxBytes:   1024,
		},

		Performance: config.PerformanceConfig{
			ConnectionPoolSize: 2,
			ProducerPoolSize:   1,
			ConsumerPoolSize:   1,
		},

		Benchmark: config.KafkaBenchmarkConfig{
			DefaultTopic: "demo-topic",
			Total:        10,
			Parallels:    1,
			DataSize:     100,
			Timeout:      30 * time.Second,
		},
	}

	// ctx := context.Background() // 在需要时可以取消注释

	// 演示配置功能
	demonstrateConfig(cfg)

	// 演示适配器功能（注释掉，因为需要真实的Kafka集群）
	// 在生产环境中，可以取消注释来测试实际功能
	// demonstrateAdapter(ctx, adapter, cfg)

	fmt.Println("\\n=== 演示完成 ===")
}

func demonstrateConfig(cfg *config.KafkaAdapterConfig) {
	fmt.Println("\\n--- 配置管理演示 ---")

	// 验证配置
	if err := cfg.Validate(); err != nil {
		log.Printf("配置验证失败: %v", err)
		return
	}
	fmt.Println("✓ 配置验证通过")

	// 显示配置信息
	fmt.Printf("Brokers: %v\\n", cfg.Brokers)
	fmt.Printf("Client ID: %s\\n", cfg.ClientID)
	fmt.Printf("Producer Acks: %s\\n", cfg.Producer.Acks)
	fmt.Printf("Consumer Group: %s\\n", cfg.Consumer.GroupID)
	fmt.Printf("Connection Pool Size: %d\\n", cfg.Performance.ConnectionPoolSize)

	// 克隆配置
	cloned := cfg.Clone().(*config.KafkaAdapterConfig)
	cloned.ClientID = "cloned-client"

	fmt.Printf("原始配置 Client ID: %s\\n", cfg.ClientID)
	fmt.Printf("克隆配置 Client ID: %s\\n", cloned.ClientID)
	fmt.Println("✓ 配置克隆功能正常")

	// 配置加载器演示
	loader := config.NewConfigLoader()
	defaultCfg := loader.LoadFromDefault()
	fmt.Printf("默认配置 Brokers: %v\\n", defaultCfg.Brokers)
	fmt.Println("✓ 配置加载器功能正常")
}

func demonstrateAdapter(ctx context.Context, adapter *kafka.KafkaAdapter, cfg *config.KafkaAdapterConfig) {
	fmt.Println("\\n--- 适配器功能演示 ---")

	// 连接
	err := adapter.Connect(ctx, cfg)
	if err != nil {
		log.Printf("连接失败: %v", err)
		return
	}
	defer adapter.Close()
	fmt.Println("✓ 连接成功")

	// 健康检查
	err = adapter.HealthCheck(ctx)
	if err != nil {
		log.Printf("健康检查失败: %v", err)
		return
	}
	fmt.Println("✓ 健康检查通过")

	// 单条消息生产
	operation := interfaces.Operation{
		Type:  "produce",
		Key:   "demo-key",
		Value: "demo-value",
		Params: map[string]interface{}{
			"topic": "demo-topic",
		},
	}

	result, err := adapter.Execute(ctx, operation)
	if err != nil {
		log.Printf("生产消息失败: %v", err)
		return
	}

	if result.Success {
		fmt.Println("✓ 单条消息生产成功")
		fmt.Printf("  执行时间: %v\\n", result.Duration)
	}

	// 批量消息生产
	messages := []*operations.Message{
		{Key: "batch-key-1", Value: "batch-value-1"},
		{Key: "batch-key-2", Value: "batch-value-2"},
		{Key: "batch-key-3", Value: "batch-value-3"},
	}

	batchOp := interfaces.Operation{
		Type: "produce_batch",
		Params: map[string]interface{}{
			"topic":    "demo-topic",
			"messages": messages,
		},
	}

	batchResult, err := adapter.Execute(ctx, batchOp)
	if err != nil {
		log.Printf("批量生产消息失败: %v", err)
		return
	}

	if batchResult.Success {
		fmt.Println("✓ 批量消息生产成功")
		if batch, ok := batchResult.Value.(*operations.BatchResult); ok {
			fmt.Printf("  成功数量: %d\\n", batch.SuccessCount)
			fmt.Printf("  总耗时: %v\\n", batch.TotalDuration)
		}
	}

	// 获取指标
	metrics := adapter.GetProtocolMetrics()
	fmt.Println("\\n--- 性能指标 ---")
	if produced, ok := metrics["produced_messages"]; ok {
		fmt.Printf("已生产消息数: %v\\n", produced)
	}
	if produceRate, ok := metrics["produce_rate"]; ok {
		fmt.Printf("生产速率: %v msg/s\\n", produceRate)
	}
	if connections, ok := metrics["active_connections"]; ok {
		fmt.Printf("活跃连接数: %v\\n", connections)
	}
}
