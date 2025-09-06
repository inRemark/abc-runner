package commands

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"redis-runner/app/adapters/kafka"
	kafkaConfig "redis-runner/app/adapters/kafka/config"
	"redis-runner/app/core/command"
	"redis-runner/app/core/config"
	"redis-runner/app/core/interfaces"
	"redis-runner/app/core/runner"
	"redis-runner/app/core/utils"
)

// KafkaCommandHandler Kafka增强版命令处理器
type KafkaCommandHandler struct {
	*command.BaseCommandHandler
	adapter           *kafka.KafkaAdapter
	operationRegistry *utils.OperationRegistry
	keyGenerator      *utils.DefaultKeyGenerator
	metricsCollector  interfaces.MetricsCollector
	runner            *runner.EnhancedRunner
}

// NewKafkaCommandHandler 创建Kafka命令处理器
func NewKafkaCommandHandler() *KafkaCommandHandler {
	configManager := config.NewConfigManager()
	adapter := kafka.NewKafkaAdapter()
	
	baseHandler := command.NewBaseCommandHandler(
		"kafka-enhanced",
		"Kafka performance testing with connection pooling",
		command.Enhanced,
		false, // 不是弃用的
		adapter,
		configManager,
	)

	return &KafkaCommandHandler{
		BaseCommandHandler: baseHandler,
		adapter:            adapter,
		operationRegistry:  utils.NewOperationRegistry(),
		keyGenerator:       utils.NewDefaultKeyGenerator(),
		metricsCollector:   adapter.GetMetricsCollector(),
	}
}

// ExecuteCommand 执行Kafka命令
func (k *KafkaCommandHandler) ExecuteCommand(ctx context.Context, args []string) error {
	log.Println("Starting Kafka Enhanced benchmark...")

	// 1. 加载配置
	if err := k.loadConfiguration(args); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// 2. 连接Kafka
	if err := k.connectKafka(ctx); err != nil {
		return fmt.Errorf("failed to connect to Kafka: %w", err)
	}
	defer k.adapter.Close()

	// 3. 注册操作
	k.registerOperations()

	// 4. 创建运行引擎
	k.runner = runner.NewEnhancedRunner(
		k.adapter,
		k.GetConfigManager().GetConfig(),
		k.metricsCollector,
		k.keyGenerator,
		k.operationRegistry,
	)

	// 5. 执行基准测试
	metrics, err := k.runner.RunBenchmark(ctx)
	if err != nil {
		return fmt.Errorf("benchmark execution failed: %w", err)
	}

	// 6. 输出结果
	k.printResults(metrics)

	return nil
}

// loadConfiguration 加载配置
func (k *KafkaCommandHandler) loadConfiguration(args []string) error {
	configManager := k.GetConfigManager()

	// 检查是否使用配置文件
	if k.hasConfigFlag(args) {
		log.Println("Loading Kafka configuration from file...")
		// TODO: 实现Kafka配置从文件加载
		// 这里可以实现Kafka配置适配器，类似于Redis
		sources := config.CreateKafkaConfigSources("conf/kafka.yaml", nil)
		return configManager.LoadConfiguration(sources...)
	}

	// 使用命令行参数创建配置
	log.Println("Loading Kafka configuration from command line...")
	kafkaConfig := k.createConfigFromArgs(args)
	// 这里需要将Kafka配置适配为统一接口
	// TODO: 实现KafkaConfigAdapter
	configManager.SetConfig(kafkaConfig)
	return nil
}

// hasConfigFlag 检查是否有config标志
func (k *KafkaCommandHandler) hasConfigFlag(args []string) bool {
	for _, arg := range args {
		if arg == "-config" || arg == "--config" {
			return true
		}
	}
	return false
}

// createConfigFromArgs 从命令行参数创建配置
func (k *KafkaCommandHandler) createConfigFromArgs(args []string) *kafkaConfig.KafkaAdapterConfig {
	// 默认配置
	cfg := &kafkaConfig.KafkaAdapterConfig{
		Protocol: "kafka",
		Brokers:  []string{"localhost:9092"},
		TopicConfigs: []kafkaConfig.TopicConfig{
			{
				Name:       "test-topic",
				Partitions: 1,
			},
		},
		Producer: kafkaConfig.ProducerConfig{
			BatchSize:    16384,
			BatchTimeout: time.Millisecond * 100,
			RetryMax:     3,
			RequiredAcks: 1,
			Compression:  "snappy",
		},
		Consumer: kafkaConfig.ConsumerConfig{
			GroupID:          "test-group",
			AutoOffsetReset:  "earliest",
			CommitInterval:   time.Second * 1,
			SessionTimeout:   time.Second * 30,
			HeartbeatTimeout: time.Second * 3,
		},
		Benchmark: kafkaConfig.KafkaBenchmarkConfig{
			Total:       1000,
			Parallels:   3,
			MessageSize: 1024,
			TestType:    "produce",
		},
		Performance: kafkaConfig.PerformanceConfig{
			ConnectionPoolSize: 10,
			ProducerPoolSize:   5,
			ConsumerPoolSize:   5,
		},
	}

	// 解析命令行参数
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--broker", "--brokers":
			if i+1 < len(args) {
				cfg.Brokers = []string{args[i+1]}
				i++
			}
		case "--topic":
			if i+1 < len(args) {
				cfg.TopicConfigs[0].Name = args[i+1]
				i++
			}
		case "-n":
			if i+1 < len(args) {
				if total, err := strconv.Atoi(args[i+1]); err == nil {
					cfg.Benchmark.Total = total
				}
				i++
			}
		case "-c":
			if i+1 < len(args) {
				if parallels, err := strconv.Atoi(args[i+1]); err == nil {
					cfg.Benchmark.Parallels = parallels
				}
				i++
			}
		case "--message-size":
			if i+1 < len(args) {
				if size, err := strconv.Atoi(args[i+1]); err == nil {
					cfg.Benchmark.MessageSize = size
				}
				i++
			}
		case "--test-type":
			if i+1 < len(args) {
				cfg.Benchmark.TestType = args[i+1]
				i++
			}
		case "--group-id":
			if i+1 < len(args) {
				cfg.Consumer.GroupID = args[i+1]
				i++
			}
		case "--partitions":
			if i+1 < len(args) {
				if partitions, err := strconv.Atoi(args[i+1]); err == nil {
					cfg.TopicConfigs[0].Partitions = partitions
				}
				i++
			}
		}
	}

	return cfg
}

// connectKafka 连接Kafka
func (k *KafkaCommandHandler) connectKafka(ctx context.Context) error {
	cfg := k.GetConfigManager().GetConfig()

	log.Printf("Connecting to Kafka brokers: %v", cfg.(*kafkaConfig.KafkaAdapterConfig).Brokers)

	if err := k.adapter.Connect(ctx, cfg); err != nil {
		return err
	}

	log.Println("Kafka connection established successfully")
	return nil
}

// registerOperations 注册操作
func (k *KafkaCommandHandler) registerOperations() {
	kafka.RegisterKafkaOperations(k.operationRegistry)
}

// printResults 打印结果
func (k *KafkaCommandHandler) printResults(metrics *interfaces.Metrics) {
	cfg := k.GetConfigManager().GetConfig().(*kafkaConfig.KafkaAdapterConfig)

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("KAFKA BENCHMARK RESULTS")
	fmt.Println(strings.Repeat("=", 60))

	// 基本信息
	fmt.Printf("Kafka Brokers: %v\n", cfg.Brokers)
	fmt.Printf("Test Type: %s\n", cfg.Benchmark.TestType)
	fmt.Printf("Topic: %s\n", cfg.TopicConfigs[0].Name)
	fmt.Printf("Total Messages: %d\n", cfg.Benchmark.Total)
	fmt.Printf("Parallel Connections: %d\n", cfg.Benchmark.Parallels)
	fmt.Printf("Message Size: %d bytes\n", cfg.Benchmark.MessageSize)
	fmt.Printf("Messages/sec: %d\n", metrics.RPS)
	fmt.Printf("Success Rate: %.2f%%\n", 100.0-metrics.ErrorRate)

	fmt.Println(strings.Repeat("-", 60))

	// 延迟统计
	fmt.Printf("Avg Latency: %.3f ms\n", float64(metrics.AvgLatency)/float64(time.Millisecond))
	fmt.Printf("Min Latency: %.3f ms\n", float64(metrics.MinLatency)/float64(time.Millisecond))
	fmt.Printf("Max Latency: %.3f ms\n", float64(metrics.MaxLatency)/float64(time.Millisecond))
	fmt.Printf("P90 Latency: %.3f ms\n", float64(metrics.P90Latency)/float64(time.Millisecond))
	fmt.Printf("P95 Latency: %.3f ms\n", float64(metrics.P95Latency)/float64(time.Millisecond))
	fmt.Printf("P99 Latency: %.3f ms\n", float64(metrics.P99Latency)/float64(time.Millisecond))

	fmt.Println(strings.Repeat("-", 60))

	// Kafka特定指标
	kafkaMetrics := k.adapter.GetProtocolMetrics()
	if totalBytes, exists := kafkaMetrics["total_bytes"]; exists {
		fmt.Printf("Total Bytes: %v\n", totalBytes)
	}
	if avgMsgSize, exists := kafkaMetrics["avg_message_size"]; exists {
		fmt.Printf("Avg Message Size: %v bytes\n", avgMsgSize)
	}
	if throughputMbps, exists := kafkaMetrics["throughput_mbps"]; exists {
		fmt.Printf("Throughput: %.2f MB/s\n", throughputMbps)
	}

	// 生产者/消费者特定信息
	if cfg.Benchmark.TestType == "produce" {
		fmt.Println("\nProducer Configuration:")
		fmt.Printf("  Batch Size: %d\n", cfg.Producer.BatchSize)
		fmt.Printf("  Batch Timeout: %v\n", cfg.Producer.BatchTimeout)
		fmt.Printf("  Required Acks: %v\n", cfg.Producer.RequiredAcks)
		fmt.Printf("  Compression: %v\n", cfg.Producer.Compression)
	} else if cfg.Benchmark.TestType == "consume" {
		fmt.Println("\nConsumer Configuration:")
		fmt.Printf("  Group ID: %s\n", cfg.Consumer.GroupID)
		fmt.Printf("  Auto Offset Reset: %s\n", cfg.Consumer.AutoOffsetReset)
		fmt.Printf("  Session Timeout: %v\n", cfg.Consumer.SessionTimeout)
	}

	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("KAFKA BENCHMARK COMPLETED")
	fmt.Println(strings.Repeat("=", 60))
}

// GetUsage 获取使用说明
func (k *KafkaCommandHandler) GetUsage() string {
	return `Usage: redis-runner kafka-enhanced [options]

Enhanced Kafka Performance Testing Tool

Options:
  --broker <address>        Kafka broker address (default: localhost:9092)
  --topic <name>            Topic name (default: test-topic)
  --test-type <type>        Test type: produce/consume (default: produce)
  -n <messages>             Total number of messages (default: 1000)
  -c <connections>          Number of parallel connections (default: 3)
  --message-size <bytes>    Message size in bytes (default: 1024)
  --group-id <id>           Consumer group ID (default: test-group)
  --partitions <num>        Number of partitions (default: 1)
  --config <file>           Configuration file path

Configuration File:
  --config conf/kafka.yaml

Examples:
  # Basic producer test
  redis-runner kafka-enhanced --broker localhost:9092 --topic test -n 10000 -c 5

  # Consumer test
  redis-runner kafka-enhanced --test-type consume --topic test --group-id my-group

  # High throughput test with configuration file
  redis-runner kafka-enhanced --config conf/kafka.yaml

  # Custom message size
  redis-runner kafka-enhanced --topic test --message-size 4096 -n 50000

For more information: https://docs.redis-runner.com/kafka-enhanced`
}

// ValidateArgs 验证参数
func (k *KafkaCommandHandler) ValidateArgs(args []string) error {
	// 基本参数验证
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--broker", "--brokers":
			if i+1 >= len(args) {
				return fmt.Errorf("--broker requires a value")
			}
			broker := args[i+1]
			if !strings.Contains(broker, ":") {
				return fmt.Errorf("broker address must include port (e.g., localhost:9092)")
			}
			i++
		case "--topic":
			if i+1 >= len(args) {
				return fmt.Errorf("--topic requires a value")
			}
			topic := args[i+1]
			if topic == "" {
				return fmt.Errorf("topic name cannot be empty")
			}
			i++
		case "-n":
			if i+1 >= len(args) {
				return fmt.Errorf("-n requires a value")
			}
			if _, err := strconv.Atoi(args[i+1]); err != nil {
				return fmt.Errorf("invalid value for -n: %s", args[i+1])
			}
			i++
		case "-c":
			if i+1 >= len(args) {
				return fmt.Errorf("-c requires a value")
			}
			if parallels, err := strconv.Atoi(args[i+1]); err != nil {
				return fmt.Errorf("invalid value for -c: %s", args[i+1])
			} else if parallels <= 0 {
				return fmt.Errorf("-c must be greater than 0")
			}
			i++
		case "--test-type":
			if i+1 >= len(args) {
				return fmt.Errorf("--test-type requires a value")
			}
			testType := args[i+1]
			validTypes := []string{"produce", "consume", "both"}
			found := false
			for _, valid := range validTypes {
				if testType == valid {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("invalid test type: %s (valid: produce, consume, both)", testType)
			}
			i++
		case "--message-size":
			if i+1 >= len(args) {
				return fmt.Errorf("--message-size requires a value")
			}
			if size, err := strconv.Atoi(args[i+1]); err != nil {
				return fmt.Errorf("invalid message size: %s", args[i+1])
			} else if size <= 0 {
				return fmt.Errorf("message size must be greater than 0")
			}
			i++
		case "--partitions":
			if i+1 >= len(args) {
				return fmt.Errorf("--partitions requires a value")
			}
			if partitions, err := strconv.Atoi(args[i+1]); err != nil {
				return fmt.Errorf("invalid partitions value: %s", args[i+1])
			} else if partitions <= 0 {
				return fmt.Errorf("partitions must be greater than 0")
			}
			i++
		}
	}

	return nil
}