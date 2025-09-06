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
	"redis-runner/app/core/config"
	"redis-runner/app/core/interfaces"
	"redis-runner/app/core/runner"
	"redis-runner/app/core/utils"
)

// KafkaSimpleHandler 简化的Kafka命令处理器
type KafkaSimpleHandler struct {
	adapter           *kafka.KafkaAdapter
	configManager     *config.ConfigManager
	operationRegistry *utils.OperationRegistry
	keyGenerator      *utils.DefaultKeyGenerator
	metricsCollector  interfaces.MetricsCollector
	runner            *runner.EnhancedRunner
}

// NewKafkaCommandHandler 创建Kafka命令处理器（统一接口）
func NewKafkaCommandHandler() *KafkaSimpleHandler {
	return &KafkaSimpleHandler{
		adapter:           kafka.NewKafkaAdapter(),
		configManager:     config.NewConfigManager(),
		operationRegistry: utils.NewOperationRegistry(),
		keyGenerator:      utils.NewDefaultKeyGenerator(),
	}
}

// Execute 执行Kafka命令
func (k *KafkaSimpleHandler) Execute(ctx context.Context, args []string) error {
	log.Println("Starting Kafka performance test...")

	// 1. 加载配置
	if err := k.loadConfiguration(args); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// 2. 连接Kafka
	if err := k.connectKafka(ctx); err != nil {
		return fmt.Errorf("failed to connect to Kafka: %w", err)
	}
	defer k.adapter.Close()

	// 3. 设置指标收集器
	k.metricsCollector = k.adapter.GetMetricsCollector()

	// 4. 注册操作
	k.registerOperations()

	// 5. 创建运行引擎
	k.runner = runner.NewEnhancedRunner(
		k.adapter,
		k.configManager.GetConfig(),
		k.metricsCollector,
		k.keyGenerator,
		k.operationRegistry,
	)

	// 6. 执行基准测试
	metrics, err := k.runner.RunBenchmark(ctx)
	if err != nil {
		return fmt.Errorf("benchmark execution failed: %w", err)
	}

	// 7. 输出结果
	k.printResults(metrics)

	log.Println("Kafka performance test completed successfully")
	return nil
}

// GetHelp 获取帮助信息
func (k *KafkaSimpleHandler) GetHelp() string {
	return `Usage: redis-runner kafka [options]

Kafka Performance Testing Tool

Options:
  --broker <address>       Kafka broker address (default: localhost:9092)
  --brokers <addresses>    Multiple Kafka brokers (comma-separated)
  --topic <name>           Kafka topic name (default: test-topic)
  --partitions <num>       Number of partitions (default: 1)
  -n <messages>            Total number of messages (default: 1000)
  -c <producers>           Number of parallel producers/consumers (default: 3)
  --message-size <bytes>   Message size in bytes (default: 1024)
  --test-type <type>       Test type: produce, consume, produce_consume (default: produce)
  --duration <time>        Test duration (e.g. 30s, 5m) - overrides -n
  --config <file>          Configuration file path

Producer Options:
  --batch-size <size>      Producer batch size (default: 16384)
  --batch-timeout <time>   Producer batch timeout (default: 100ms)
  --compression <type>     Compression type: none, gzip, snappy, lz4, zstd (default: snappy)
  --acks <level>           Required acknowledgments: 0, 1, all (default: 1)
  --retries <num>          Max retries (default: 3)

Consumer Options:
  --group-id <id>          Consumer group ID (default: test-group)
  --offset-reset <policy>  Auto offset reset: earliest, latest (default: earliest)
  --commit-interval <time> Commit interval (default: 1s)
  --session-timeout <time> Session timeout (default: 30s)

Performance Options:
  --connection-pool <size> Connection pool size (default: 10)
  --producer-pool <size>   Producer pool size (default: 5)
  --consumer-pool <size>   Consumer pool size (default: 5)

Configuration File:
  --config conf/kafka.yaml

Test Types:
  produce           Only produce messages
  consume           Only consume messages
  produce_consume   Mixed produce and consume operations

Examples:
  # Basic producer test
  redis-runner kafka --broker localhost:9092 --topic test-topic -n 10000 -c 5

  # Consumer test with specific group
  redis-runner kafka --broker localhost:9092 --topic test-topic \\
    --test-type consume --group-id my-group -n 1000

  # High-throughput test with larger messages
  redis-runner kafka --brokers localhost:9092,localhost:9093 \\
    --topic high-throughput --message-size 4096 \\
    --batch-size 65536 -n 100000 -c 10

  # Duration-based mixed workload
  redis-runner kafka --broker localhost:9092 --topic mixed-workload \\
    --test-type produce_consume --duration 60s -c 8

  # Load test with configuration file
  redis-runner kafka --config conf/kafka.yaml

  # Performance test with compression
  redis-runner kafka --broker localhost:9092 --topic perf-test \\
    --compression lz4 --acks all --batch-size 32768 -n 50000

For more information: https://docs.redis-runner.com/kafka`
}

// loadConfiguration 加载配置
func (k *KafkaSimpleHandler) loadConfiguration(args []string) error {
	// 检查是否使用配置文件
	if k.hasConfigFlag(args) {
		log.Println("Loading Kafka configuration from file...")
		// 使用多源配置加载器
		sources := config.CreateKafkaConfigSources("conf/kafka.yaml", nil)
		return k.configManager.LoadConfiguration(sources...)
	}

	// 使用命令行参数创建配置
	log.Println("Loading Kafka configuration from command line...")
	kafkaCfg := k.createConfigFromArgs(args)
	k.configManager.SetConfig(kafkaCfg)
	return nil
}

// hasConfigFlag 检查是否有config标志
func (k *KafkaSimpleHandler) hasConfigFlag(args []string) bool {
	for _, arg := range args {
		if arg == "-config" || arg == "--config" {
			return true
		}
	}
	return false
}

// createConfigFromArgs 从命令行参数创建配置
func (k *KafkaSimpleHandler) createConfigFromArgs(args []string) *kafkaConfig.KafkaAdapterConfig {
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
		case "--broker":
			if i+1 < len(args) {
				cfg.Brokers = []string{args[i+1]}
				i++
			}
		case "--brokers":
			if i+1 < len(args) {
				cfg.Brokers = strings.Split(args[i+1], ",")
				// 清理空格
				for j, broker := range cfg.Brokers {
					cfg.Brokers[j] = strings.TrimSpace(broker)
				}
				i++
			}
		case "--topic":
			if i+1 < len(args) {
				cfg.TopicConfigs[0].Name = args[i+1]
				i++
			}
		case "--partitions":
			if i+1 < len(args) {
				if partitions, err := strconv.Atoi(args[i+1]); err == nil {
					cfg.TopicConfigs[0].Partitions = partitions
				}
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
		case "--duration":
			if i+1 < len(args) {
				// Duration 字段不存在，暂时注释
				// if duration, err := time.ParseDuration(args[i+1]); err == nil {
				//     cfg.Benchmark.Duration = duration
				// }
				log.Printf("Duration field not implemented yet: %s", args[i+1])
				i++
			}
		case "--batch-size":
			if i+1 < len(args) {
				if batchSize, err := strconv.Atoi(args[i+1]); err == nil {
					cfg.Producer.BatchSize = batchSize
				}
				i++
			}
		case "--batch-timeout":
			if i+1 < len(args) {
				if timeout, err := time.ParseDuration(args[i+1]); err == nil {
					cfg.Producer.BatchTimeout = timeout
				}
				i++
			}
		case "--compression":
			if i+1 < len(args) {
				cfg.Producer.Compression = args[i+1]
				i++
			}
		case "--acks":
			if i+1 < len(args) {
				switch args[i+1] {
				case "0":
					cfg.Producer.RequiredAcks = 0
				case "1":
					cfg.Producer.RequiredAcks = 1
				case "all", "-1":
					cfg.Producer.RequiredAcks = -1
				}
				i++
			}
		case "--retries":
			if i+1 < len(args) {
				if retries, err := strconv.Atoi(args[i+1]); err == nil {
					cfg.Producer.RetryMax = retries
				}
				i++
			}
		case "--group-id":
			if i+1 < len(args) {
				cfg.Consumer.GroupID = args[i+1]
				i++
			}
		case "--offset-reset":
			if i+1 < len(args) {
				cfg.Consumer.AutoOffsetReset = args[i+1]
				i++
			}
		case "--commit-interval":
			if i+1 < len(args) {
				if interval, err := time.ParseDuration(args[i+1]); err == nil {
					cfg.Consumer.CommitInterval = interval
				}
				i++
			}
		case "--session-timeout":
			if i+1 < len(args) {
				if timeout, err := time.ParseDuration(args[i+1]); err == nil {
					cfg.Consumer.SessionTimeout = timeout
				}
				i++
			}
		case "--connection-pool":
			if i+1 < len(args) {
				if poolSize, err := strconv.Atoi(args[i+1]); err == nil {
					cfg.Performance.ConnectionPoolSize = poolSize
				}
				i++
			}
		case "--producer-pool":
			if i+1 < len(args) {
				if poolSize, err := strconv.Atoi(args[i+1]); err == nil {
					cfg.Performance.ProducerPoolSize = poolSize
				}
				i++
			}
		case "--consumer-pool":
			if i+1 < len(args) {
				if poolSize, err := strconv.Atoi(args[i+1]); err == nil {
					cfg.Performance.ConsumerPoolSize = poolSize
				}
				i++
			}
		}
	}

	return cfg
}

// connectKafka 连接Kafka
func (k *KafkaSimpleHandler) connectKafka(ctx context.Context) error {
	cfg := k.configManager.GetConfig()

	if kafkaCfg, ok := cfg.(*kafkaConfig.KafkaAdapterConfig); ok {
		log.Printf("Connecting to Kafka brokers: %v", kafkaCfg.Brokers)
		if len(kafkaCfg.TopicConfigs) > 0 {
			log.Printf("Target topic: %s", kafkaCfg.TopicConfigs[0].Name)
		}
	} else {
		log.Println("Connecting to Kafka...")
	}

	if err := k.adapter.Connect(ctx, cfg); err != nil {
		return err
	}

	log.Println("Kafka connection established successfully")
	return nil
}

// registerOperations 注册操作
func (k *KafkaSimpleHandler) registerOperations() {
	// Kafka操作注册 - 简化实现
	// TODO: 实现具体的Kafka操作注册
	log.Println("Kafka operations registry not fully implemented yet")
}

// printResults 打印结果
func (k *KafkaSimpleHandler) printResults(metrics *interfaces.Metrics) {
	cfg := k.configManager.GetConfig()
	var kafkaCfg *kafkaConfig.KafkaAdapterConfig
	if kcfg, ok := cfg.(*kafkaConfig.KafkaAdapterConfig); ok {
		kafkaCfg = kcfg
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("KAFKA PERFORMANCE TEST RESULTS")
	fmt.Println(strings.Repeat("=", 60))

	// 基本信息
	if kafkaCfg != nil {
		fmt.Printf("Kafka Brokers: %v\n", kafkaCfg.Brokers)
		if len(kafkaCfg.TopicConfigs) > 0 {
			fmt.Printf("Topic: %s\n", kafkaCfg.TopicConfigs[0].Name)
			fmt.Printf("Partitions: %d\n", kafkaCfg.TopicConfigs[0].Partitions)
		}
		fmt.Printf("Test Type: %s\n", kafkaCfg.Benchmark.TestType)
		fmt.Printf("Total Messages: %d\n", kafkaCfg.Benchmark.Total)
		fmt.Printf("Parallel Workers: %d\n", kafkaCfg.Benchmark.Parallels)
		fmt.Printf("Message Size: %d bytes\n", kafkaCfg.Benchmark.MessageSize)

		// Duration 字段不存在，暂时注释
		// if kafkaCfg.Benchmark.Duration > 0 {
		//     fmt.Printf("Test Duration: %v\n", kafkaCfg.Benchmark.Duration)
		// }
	}

	fmt.Println(strings.Repeat("-", 60))

	// 性能指标
	fmt.Printf("Messages/sec: %d\n", metrics.RPS)
	fmt.Printf("Success Rate: %.2f%%\n", 100.0-metrics.ErrorRate)
	fmt.Printf("Total Messages: %d\n", metrics.TotalOps)

	if kafkaCfg != nil && kafkaCfg.Benchmark.TestType == "produce_consume" {
		fmt.Printf("Produced Messages: %d\n", metrics.WriteOps)
		fmt.Printf("Consumed Messages: %d\n", metrics.ReadOps)
	}

	if metrics.FailedOps > 0 {
		fmt.Printf("Total Errors: %d\n", metrics.FailedOps)
	}

	fmt.Println(strings.Repeat("-", 60))

	// 延迟统计
	fmt.Printf("Avg Latency: %.3f ms\n", float64(metrics.AvgLatency)/float64(time.Millisecond))
	fmt.Printf("P90 Latency: %.3f ms\n", float64(metrics.P90Latency)/float64(time.Millisecond))
	fmt.Printf("P95 Latency: %.3f ms\n", float64(metrics.P95Latency)/float64(time.Millisecond))
	fmt.Printf("P99 Latency: %.3f ms\n", float64(metrics.P99Latency)/float64(time.Millisecond))
	fmt.Printf("Max Latency: %.3f ms\n", float64(metrics.MaxLatency)/float64(time.Millisecond))

	// 吞吐量统计
	if kafkaCfg != nil {
		msgPerSec := float64(metrics.TotalOps) / (float64(metrics.Duration) / float64(time.Second))
		mbPerSec := (msgPerSec * float64(kafkaCfg.Benchmark.MessageSize)) / (1024 * 1024)
		fmt.Printf("Throughput: %.2f MB/s\n", mbPerSec)
	}

	// Kafka特定指标
	if kafkaMetrics := k.getKafkaMetrics(); kafkaMetrics != nil {
		fmt.Println(strings.Repeat("-", 60))
		fmt.Println("Kafka Specific Metrics:")
		if producerMetrics, exists := kafkaMetrics["producer"]; exists {
			fmt.Printf("  Producer Metrics: %v\n", producerMetrics)
		}
		if consumerMetrics, exists := kafkaMetrics["consumer"]; exists {
			fmt.Printf("  Consumer Metrics: %v\n", consumerMetrics)
		}
		if partitionMetrics, exists := kafkaMetrics["partitions"]; exists {
			fmt.Printf("  Partition Distribution: %v\n", partitionMetrics)
		}
		if compressionRatio, exists := kafkaMetrics["compression_ratio"]; exists {
			fmt.Printf("  Compression Ratio: %.2f:1\n", compressionRatio)
		}
	}

	// 生产者/消费者特定信息
	if kafkaCfg != nil {
		if kafkaCfg.Benchmark.TestType == "produce" || kafkaCfg.Benchmark.TestType == "produce_consume" {
			fmt.Println(strings.Repeat("-", 60))
			fmt.Println("Producer Configuration:")
			fmt.Printf("  Batch Size: %d\n", kafkaCfg.Producer.BatchSize)
			fmt.Printf("  Batch Timeout: %v\n", kafkaCfg.Producer.BatchTimeout)
			fmt.Printf("  Required Acks: %v\n", kafkaCfg.Producer.RequiredAcks)
			fmt.Printf("  Compression: %v\n", kafkaCfg.Producer.Compression)
			fmt.Printf("  Max Retries: %v\n", kafkaCfg.Producer.RetryMax)
		}

		if kafkaCfg.Benchmark.TestType == "consume" || kafkaCfg.Benchmark.TestType == "produce_consume" {
			fmt.Println(strings.Repeat("-", 60))
			fmt.Println("Consumer Configuration:")
			fmt.Printf("  Group ID: %s\n", kafkaCfg.Consumer.GroupID)
			fmt.Printf("  Auto Offset Reset: %s\n", kafkaCfg.Consumer.AutoOffsetReset)
			fmt.Printf("  Commit Interval: %v\n", kafkaCfg.Consumer.CommitInterval)
			fmt.Printf("  Session Timeout: %v\n", kafkaCfg.Consumer.SessionTimeout)
		}
	}

	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("KAFKA PERFORMANCE TEST COMPLETED")
	fmt.Println(strings.Repeat("=", 60))
}

// getKafkaMetrics 获取Kafka特定指标
func (k *KafkaSimpleHandler) getKafkaMetrics() map[string]interface{} {
	if k.metricsCollector == nil {
		return nil
	}
	
	// 这里应该从Kafka适配器获取特定指标
	// 为了保持兼容性，先返回空
	return nil
}

// validateArgs 验证参数
func (k *KafkaSimpleHandler) validateArgs(args []string) error {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--broker":
			if i+1 >= len(args) {
				return fmt.Errorf("--broker requires a broker address")
			}
			i++
		case "--test-type":
			if i+1 >= len(args) {
				return fmt.Errorf("--test-type requires a value")
			}
			testType := args[i+1]
			validTypes := []string{"produce", "consume", "produce_consume"}
			valid := false
			for _, vt := range validTypes {
				if testType == vt {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("invalid test type: %s (valid: produce, consume, produce_consume)", testType)
			}
			i++
		case "--compression":
			if i+1 >= len(args) {
				return fmt.Errorf("--compression requires a value")
			}
			compression := args[i+1]
			validCompressions := []string{"none", "gzip", "snappy", "lz4", "zstd"}
			valid := false
			for _, vc := range validCompressions {
				if compression == vc {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("invalid compression type: %s", compression)
			}
			i++
		case "--acks":
			if i+1 >= len(args) {
				return fmt.Errorf("--acks requires a value")
			}
			acks := args[i+1]
			if acks != "0" && acks != "1" && acks != "all" && acks != "-1" {
				return fmt.Errorf("invalid acks value: %s (valid: 0, 1, all)", acks)
			}
			i++
		case "--offset-reset":
			if i+1 >= len(args) {
				return fmt.Errorf("--offset-reset requires a value")
			}
			offsetReset := args[i+1]
			if offsetReset != "earliest" && offsetReset != "latest" {
				return fmt.Errorf("invalid offset reset: %s (valid: earliest, latest)", offsetReset)
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
		}
	}
	return nil
}