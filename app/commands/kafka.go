package commands

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"abc-runner/app/adapters/kafka"
	kafkaConfig "abc-runner/app/adapters/kafka/config"
	"abc-runner/app/core/config"
	"abc-runner/app/core/interfaces"
	"abc-runner/app/core/reports"
	"abc-runner/app/core/runner"
	"abc-runner/app/core/utils"
)

// KafkaSimpleHandler 简化的Kafka命令处理器
type KafkaSimpleHandler struct {
	adapterFactory    interfaces.AdapterFactory
	adapter           interfaces.ProtocolAdapter
	configManager     *config.ConfigManager
	operationRegistry *utils.OperationRegistry
	keyGenerator      *utils.DefaultKeyGenerator
	metricsCollector  interfaces.MetricsCollector
	runner            *runner.EnhancedRunner
	reportManager     *reports.ReportManager
	reportArgs        *reports.ReportArgs
}

// NewKafkaCommandHandler 创建Kafka命令处理器（统一接口）
func NewKafkaCommandHandler(adapterFactory interfaces.AdapterFactory) *KafkaSimpleHandler {
	if adapterFactory == nil {
		panic("adapterFactory cannot be nil - dependency injection required")
	}

	handler := &KafkaSimpleHandler{
		adapterFactory:    adapterFactory,
		configManager:     config.NewConfigManager(nil),
		operationRegistry: utils.NewOperationRegistry(),
		keyGenerator:      utils.NewDefaultKeyGenerator(),
	}

	// 注册Kafka操作工厂
	kafka.RegisterKafkaOperations(handler.operationRegistry)

	return handler
}

// Execute 执行Kafka命令
func (k *KafkaSimpleHandler) Execute(ctx context.Context, args []string) error {
	// 检查是否请求帮助
	for _, arg := range args {
		if arg == "--help" || arg == "-h" || arg == "help" {
			fmt.Println(k.GetHelp())
			return nil
		}
	}

	log.Println("Starting Kafka performance test...")

	// 1. 解析报告参数
	var err error
	k.reportArgs, err = reports.ParseReportArgs(args)
	if err != nil {
		return fmt.Errorf("failed to parse report arguments: %w", err)
	}

	// 2. 验证参数
	if err := k.validateArgs(args); err != nil {
		return fmt.Errorf("argument validation failed: %w", err)
	}

	// 3. 加载配置
	if err := k.loadConfiguration(args); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// 4. 初始化适配器（使用DI工厂）
	k.adapter = k.adapterFactory.CreateKafkaAdapter()

	// 5. 连接适配器
	if err := k.adapter.Connect(ctx, k.configManager.GetConfig()); err != nil {
		return fmt.Errorf("failed to connect to Kafka: %w", err)
	}
	defer k.adapter.Close()

	// 6. 初始化指标收集器
	k.metricsCollector = k.adapter.GetMetricsCollector()

	// 7. 初始化报告管理器
	k.initializeReportManager()

	// 8. 初始化运行器
	k.runner = runner.NewEnhancedRunner(k.adapter, k.configManager.GetConfig(), k.metricsCollector, k.keyGenerator, k.operationRegistry)

	// 9. 运行测试
	log.Println("Running Kafka performance test...")
	_, err = k.runner.RunBenchmark(ctx)
	if err != nil {
		return fmt.Errorf("performance test execution failed: %w", err)
	}

	// 10. 生成报告
	log.Println("Generating reports...")
	if err := k.reportManager.GenerateReports(); err != nil {
		return fmt.Errorf("report generation failed: %w", err)
	}

	log.Println("Kafka performance test completed successfully")
	return nil
}

// GetHelp 获取帮助信息
func (k *KafkaSimpleHandler) GetHelp() string {
	baseHelp := `Usage: abc-runner kafka [OPTIONS]

Kafka Performance Testing Tool

Options:
  --broker HOST:PORT       Kafka broker address (can be used multiple times)
  --brokers HOSTS          Comma-separated list of Kafka brokers
  --topic TOPIC            Topic name (default: test-topic)
  --test-type TYPE         Test type: produce, consume, produce_consume (default: produce)
  --group-id GROUP         Consumer group ID (default: test-group)
  --message-size BYTES     Message size in bytes (default: 1024)
  --batch-size BYTES       Batch size in bytes (default: 16384)
  --compression TYPE       Compression type: none, gzip, snappy, lz4, zstd (default: snappy)
  --acks ACKS              Number of acks: 0, 1, all (default: 1)
  -n, --requests COUNT     Total number of messages (default: 1000)
  -c, --concurrency COUNT  Number of parallel connections (default: 3)
  --duration DURATION      Test duration (e.g., 30s, 5m)
  --config FILE            Configuration file path
  --core-config FILE       Core configuration file path (default: config/core.yaml)

Examples:
  # Basic producer test
  abc-runner kafka --broker localhost:9092 --topic test-topic -n 10000 -c 10

  # Consumer test with custom group
  abc-runner kafka --broker localhost:9092 --topic test-topic \\
    --test-type consume --group-id my-group -n 1000

  # High-throughput test with larger messages
  abc-runner kafka --brokers localhost:9092,localhost:9093 \\
    --topic high-throughput --message-size 4096 \\
    --batch-size 65536 -n 100000 -c 10

  # Duration-based mixed workload
  abc-runner kafka --broker localhost:9092 --topic mixed-workload \\
    --test-type produce_consume --duration 60s -c 8

  # Load test with configuration file
  abc-runner kafka --config config/kafka.yaml

  # Load test with core configuration
  abc-runner kafka --config config/kafka.yaml --core-config config/core.yaml

  # Performance test with compression
  abc-runner kafka --broker localhost:9092 --topic perf-test \\
    --compression lz4 --acks all --batch-size 32768 -n 50000

For more information: https://docs.abc-runner.com/kafka`

	return reports.AddReportArgsToHelp(baseHelp)
}

// loadConfiguration 加载配置
func (k *KafkaSimpleHandler) loadConfiguration(args []string) error {
	// 检查是否使用核心配置文件
	coreConfigPath := k.getCoreConfigFlag(args)
	if coreConfigPath != "" {
		log.Printf("Loading core configuration from %s...", coreConfigPath)
		if err := k.configManager.LoadCoreConfiguration(coreConfigPath); err != nil {
			return fmt.Errorf("failed to load core configuration: %w", err)
		}
	}

	// 使用统一配置加载器
	loader := kafkaConfig.NewUnifiedKafkaConfigLoader()

	var configPath string
	if k.hasConfigFlag(args) {
		configPath = k.getConfigFlagValue(args)
		log.Printf("Loading Kafka configuration from file: %s", configPath)
	} else {
		configPath = "" // 让加载器使用默认查找机制
	}

	// 加载配置
	cfg, err := loader.LoadConfig(configPath, args)
	if err != nil {
		return fmt.Errorf("failed to load Kafka configuration: %w", err)
	}

	k.configManager.SetConfig(cfg)
	return nil
}

// hasConfigFlag 检查是否有config标志
func (k *KafkaSimpleHandler) hasConfigFlag(args []string) bool {
	for _, arg := range args {
		if arg == "--config" || arg == "-C" {
			return true
		}
		if strings.HasPrefix(arg, "--config=") {
			return true
		}
	}
	return false
}

// getConfigFlagValue 获取配置文件路径
func (k *KafkaSimpleHandler) getConfigFlagValue(args []string) string {
	for i, arg := range args {
		if (arg == "--config" || arg == "-C") && i+1 < len(args) {
			return args[i+1]
		}
		if strings.HasPrefix(arg, "--config=") {
			return strings.TrimPrefix(arg, "--config=")
		}
	}

	// 使用统一的配置文件查找机制
	foundPath := utils.FindConfigFile("kafka")
	if foundPath != "" {
		return foundPath
	}

	// 回退到默认路径
	return "config/kafka.yaml"
}

// getCoreConfigFlag 获取核心配置文件路径
func (k *KafkaSimpleHandler) getCoreConfigFlag(args []string) string {
	for i, arg := range args {
		if arg == "--core-config" && i+1 < len(args) {
			return args[i+1]
		}
		if strings.HasPrefix(arg, "--core-config=") {
			return strings.TrimPrefix(arg, "--core-config=")
		}
	}
	return "" // 返回空字符串表示未指定核心配置文件
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
			Total:        1000,
			Parallels:    3,
			MessageSize:  1024,
			TestType:     "produce",
			DefaultTopic: "test-topic",
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
				i++
			}
		case "--topic":
			if i+1 < len(args) {
				if len(cfg.TopicConfigs) > 0 {
					cfg.TopicConfigs[0].Name = args[i+1]
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
		case "--message-size":
			if i+1 < len(args) {
				if size, err := strconv.Atoi(args[i+1]); err == nil {
					cfg.Benchmark.MessageSize = size
				}
				i++
			}
		case "--batch-size":
			if i+1 < len(args) {
				if size, err := strconv.Atoi(args[i+1]); err == nil {
					cfg.Producer.BatchSize = size
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
				if acks, err := strconv.Atoi(args[i+1]); err == nil {
					cfg.Producer.RequiredAcks = acks
				}
				i++
			}
		case "-n", "--requests":
			if i+1 < len(args) {
				if n, err := strconv.Atoi(args[i+1]); err == nil {
					cfg.Benchmark.Total = n
				}
				i++
			}
		case "-c", "--concurrency":
			if i+1 < len(args) {
				if c, err := strconv.Atoi(args[i+1]); err == nil {
					cfg.Benchmark.Parallels = c
				}
				i++
			}
		case "--duration":
			if i+1 < len(args) {
				if d, err := time.ParseDuration(args[i+1]); err == nil {
					cfg.Benchmark.Timeout = d
				}
				i++
			}
		}
	}

	return cfg
}

// initializeReportManager 初始化报告管理器
func (k *KafkaSimpleHandler) initializeReportManager() {
	if k.reportArgs == nil {
		k.reportArgs = reports.DefaultReportArgs()
	}

	reportConfig := k.reportArgs.ToReportConfig("kafka")

	// 如果加载了核心配置，使用核心配置中的报告设置作为默认值
	coreConfig := k.configManager.GetCoreConfig()
	if coreConfig != nil {
		// 合并核心配置和命令行参数
		if reportConfig.OutputDirectory == "" {
			reportConfig.OutputDirectory = coreConfig.Core.Reports.OutputDir
		}
		if reportConfig.FilePrefix == "" {
			reportConfig.FilePrefix = coreConfig.Core.Reports.FilePrefix
		}
		if len(reportConfig.Formats) == 0 {
			// 转换核心配置中的格式
			formats := make([]reports.ReportFormat, len(coreConfig.Core.Reports.Formats))
			for i, format := range coreConfig.Core.Reports.Formats {
				formats[i] = reports.ReportFormat(format)
			}
			reportConfig.Formats = formats
		}
	}

	k.reportManager = reports.NewReportManager("kafka", k.metricsCollector, reportConfig)
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
				return fmt.Errorf("invalid acks value: %s (valid: 0, 1, all, -1)", acks)
			}
			i++
		case "--total":
			if i+1 >= len(args) {
				return fmt.Errorf("--total requires a value")
			}
			if _, err := strconv.Atoi(args[i+1]); err != nil {
				return fmt.Errorf("invalid value for --total: %s", args[i+1])
			}
			i++
		case "--parallels":
			if i+1 >= len(args) {
				return fmt.Errorf("--parallels requires a value")
			}
			if parallels, err := strconv.Atoi(args[i+1]); err != nil {
				return fmt.Errorf("invalid value for --parallels: %s", args[i+1])
			} else if parallels <= 0 {
				return fmt.Errorf("--parallels must be greater than 0")
			}
			i++
		case "--duration":
			if i+1 >= len(args) {
				return fmt.Errorf("--duration requires a duration value")
			}
			if _, err := time.ParseDuration(args[i+1]); err != nil {
				return fmt.Errorf("invalid duration for --duration: %s", args[i+1])
			}
			i++
		}
	}
	return nil
}
