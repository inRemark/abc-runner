package commands

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"redis-runner/app/adapters/redis"
	redisconfig "redis-runner/app/adapters/redis/config"
	"redis-runner/app/core/command"
	"redis-runner/app/core/config"
	"redis-runner/app/core/interfaces"
	"redis-runner/app/core/runner"
	"redis-runner/app/core/utils"
)

// RedisCommandHandler Redis增强版命令处理器
type RedisCommandHandler struct {
	*command.BaseCommandHandler
	adapter           *redis.RedisAdapter
	operationRegistry *utils.OperationRegistry
	keyGenerator      *utils.DefaultKeyGenerator
	metricsCollector  interfaces.MetricsCollector
	runner            *runner.EnhancedRunner
}

// NewRedisCommandHandler 创建Redis命令处理器
func NewRedisCommandHandler() *RedisCommandHandler {
	configManager := config.NewConfigManager()
	adapter := redis.NewRedisAdapter()
	
	baseHandler := command.NewBaseCommandHandler(
		"redis-enhanced",
		"Redis performance testing with advanced features",
		command.Enhanced,
		false, // 不是弃用的
		adapter,
		configManager,
	)

	return &RedisCommandHandler{
		BaseCommandHandler: baseHandler,
		adapter:           adapter,
		operationRegistry: utils.NewOperationRegistry(),
		keyGenerator:      utils.NewDefaultKeyGenerator(),
		metricsCollector:  adapter.GetMetricsCollector(),
	}
}

// ExecuteCommand 执行Redis命令（新架构入口）
func (h *RedisCommandHandler) ExecuteCommand(ctx context.Context, args []string) error {
	// 1. 加载配置
	if err := h.loadConfiguration(args); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// 2. 连接Redis
	if err := h.connectRedis(ctx); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}
	defer h.adapter.Close()

	// 3. 注册操作
	h.registerOperations()

	// 4. 创建运行引擎
	h.runner = runner.NewEnhancedRunner(
		h.adapter,
		h.GetConfigManager().GetConfig(),
		h.metricsCollector,
		h.keyGenerator,
		h.operationRegistry,
	)

	// 5. 执行基准测试
	metrics, err := h.runner.RunBenchmark(ctx)
	if err != nil {
		return fmt.Errorf("benchmark execution failed: %w", err)
	}

	// 6. 输出结果
	h.printResults(metrics)

	return nil
}

// loadConfiguration 加载配置
func (h *RedisCommandHandler) loadConfiguration(args []string) error {
	configManager := h.GetConfigManager()
	
	// 检查是否使用配置文件
	if h.hasConfigFlag(args) {
		log.Println("Loading Redis configuration from file...")
		// 使用新的Redis配置加载器
		config, err := redisconfig.LoadRedisConfigFromFile("conf/redis.yaml")
		if err != nil {
			return err
		}
		configManager.SetConfig(config)
		return nil
	}

	// 使用命令行参数
	log.Println("Loading Redis configuration from command line...")
	// 使用新的Redis配置加载器
	config, err := redisconfig.LoadRedisConfigFromArgs(args)
	if err != nil {
		return err
	}
	configManager.SetConfig(config)
	return nil
}

// hasConfigFlag 检查是否有config标志
func (h *RedisCommandHandler) hasConfigFlag(args []string) bool {
	for _, arg := range args {
		if arg == "-config" || arg == "--config" {
			return true
		}
	}
	return false
}

// connectRedis 连接Redis
func (h *RedisCommandHandler) connectRedis(ctx context.Context) error {
	cfg := h.GetConfigManager().GetConfig()

	// 提取Redis配置
	var redisConfig *redisconfig.RedisConfig
	if adapter, ok := cfg.(*redisconfig.RedisConfigAdapter); ok {
		redisConfig = adapter.GetRedisConfig()
	} else {
		// 如果不是适配器，尝试转换
		var err error
		redisConfig, err = redisconfig.ExtractRedisConfig(cfg)
		if err != nil {
			return fmt.Errorf("failed to extract Redis config: %w", err)
		}
	}

	log.Printf("Connecting to Redis in %s mode...", redisConfig.GetMode())

	if err := h.adapter.Connect(ctx, cfg); err != nil {
		return err
	}

	log.Println("Redis connection established successfully")
	return nil
}

// registerOperations 注册操作
func (h *RedisCommandHandler) registerOperations() {
	redis.RegisterRedisOperations(h.operationRegistry)
}

// printResults 打印结果
func (h *RedisCommandHandler) printResults(metrics *interfaces.Metrics) {
	cfg := h.GetConfigManager().GetConfig()
	benchmarkConfig := cfg.GetBenchmark()

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("REDIS BENCHMARK RESULTS")
	fmt.Println(strings.Repeat("=", 60))

	// 基本信息
	fmt.Printf("Test Case: %s\n", benchmarkConfig.GetTestCase())
	fmt.Printf("Total Requests: %d\n", benchmarkConfig.GetTotal())
	fmt.Printf("Parallel Connections: %d\n", benchmarkConfig.GetParallels())
	fmt.Printf("RPS: %d\n", metrics.RPS)
	fmt.Printf("Success Rate: %.2f%%\n", 100.0-metrics.ErrorRate)

	fmt.Println(strings.Repeat("-", 60))

	// 延迟统计
	fmt.Printf("Avg Latency: %.3f ms\n", float64(metrics.AvgLatency)/float64(time.Millisecond))
	fmt.Printf("P95 Latency: %.3f ms\n", float64(metrics.P95Latency)/float64(time.Millisecond))
	fmt.Printf("P99 Latency: %.3f ms\n", float64(metrics.P99Latency)/float64(time.Millisecond))

	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("BENCHMARK COMPLETED")
	fmt.Println(strings.Repeat("=", 60))
}

// 向后兼容的函数

// RedisCommand 兼容原有的RedisCommand函数
func RedisCommand(args []string) {
	handler := NewRedisCommandHandler()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	if err := handler.ExecuteCommand(ctx, args); err != nil {
		log.Fatalf("Redis command execution failed: %v", err)
	}
}

// Start 兼容原有的Start函数
func Start() {
	handler := NewRedisCommandHandler()
	if err := handler.ExecuteFromConfig(); err != nil {
		log.Fatalf("Redis benchmark execution failed: %v", err)
	}
}

// GetUsage 获取使用说明
func (h *RedisCommandHandler) GetUsage() string {
	return `Usage: redis-runner redis-enhanced [options]

Enhanced Redis Performance Testing Tool

Options:
  -h <hostname>         Redis server hostname (default: 127.0.0.1)
  -p <port>             Redis server port (default: 6379)
  -a <password>         Redis server password
  -n <requests>         Total number of requests (default: 1000)
  -c <connections>      Number of parallel connections (default: 10)
  -t <test>             Test case to run (default: set_get_random)
  --mode <mode>         Redis mode: standalone/cluster (default: standalone)
  --config <file>       Configuration file path

Configuration File:
  --config conf/redis.yaml

Examples:
  # Basic test
  redis-runner redis-enhanced -h 127.0.0.1 -p 6379 -n 10000 -c 50

  # Cluster test with configuration file
  redis-runner redis-enhanced --config conf/redis.yaml

  # Custom test case
  redis-runner redis-enhanced -t set_get_random -n 50000 -c 100

For more information: https://docs.redis-runner.com/redis-enhanced`
}

// ValidateArgs 验证参数
func (h *RedisCommandHandler) ValidateArgs(args []string) error {
	// 基本参数验证
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-h":
			if i+1 >= len(args) {
				return fmt.Errorf("-h requires a hostname")
			}
			i++
		case "-p":
			if i+1 >= len(args) {
				return fmt.Errorf("-p requires a port number")
			}
			if _, err := strconv.Atoi(args[i+1]); err != nil {
				return fmt.Errorf("invalid port number: %s", args[i+1])
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
		case "--mode":
			if i+1 >= len(args) {
				return fmt.Errorf("--mode requires a value")
			}
			mode := args[i+1]
			if mode != "standalone" && mode != "cluster" {
				return fmt.Errorf("invalid mode: %s (valid: standalone, cluster)", mode)
			}
			i++
		}
	}

	return nil
}

// ExecuteFromConfig 从配置文件执行
func (h *RedisCommandHandler) ExecuteFromConfig() error {
	ctx := context.Background()
	return h.ExecuteCommand(ctx, []string{"-config"})
}
