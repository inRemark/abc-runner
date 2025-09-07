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
	"redis-runner/app/core/config"
	"redis-runner/app/core/interfaces"
	"redis-runner/app/core/runner"
	"redis-runner/app/core/utils"
)

// RedisSimpleHandler 简化的Redis命令处理器
type RedisSimpleHandler struct {
	adapter           *redis.RedisAdapter
	configManager     *config.ConfigManager
	operationRegistry *utils.OperationRegistry
	keyGenerator      *utils.DefaultKeyGenerator
	metricsCollector  interfaces.MetricsCollector
	runner            *runner.EnhancedRunner
}

// NewRedisCommandHandler 创建Redis命令处理器（统一接口）
func NewRedisCommandHandler() *RedisSimpleHandler {
	return &RedisSimpleHandler{
		adapter:           redis.NewRedisAdapter(),
		configManager:     config.NewConfigManager(),
		operationRegistry: utils.NewOperationRegistry(),
		keyGenerator:      utils.NewDefaultKeyGenerator(),
	}
}

// Execute 执行Redis命令
func (h *RedisSimpleHandler) Execute(ctx context.Context, args []string) error {
	// 检查是否请求帮助
	for _, arg := range args {
		if arg == "--help" || arg == "help" {
			fmt.Println(h.GetHelp())
			return nil
		}
	}

	log.Println("Starting Redis benchmark...")

	// 1. 加载配置
	if err := h.loadConfiguration(args); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// 2. 连接Redis
	if err := h.connectRedis(ctx); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}
	defer h.adapter.Close()

	// 3. 设置指标收集器
	h.metricsCollector = h.adapter.GetMetricsCollector()

	// 4. 注册操作
	h.registerOperations()

	// 5. 创建运行引擎
	h.runner = runner.NewEnhancedRunner(
		h.adapter,
		h.configManager.GetConfig(),
		h.metricsCollector,
		h.keyGenerator,
		h.operationRegistry,
	)

	// 6. 执行基准测试
	metrics, err := h.runner.RunBenchmark(ctx)
	if err != nil {
		return fmt.Errorf("benchmark execution failed: %w", err)
	}

	// 7. 输出结果
	h.printResults(metrics)

	log.Println("Redis benchmark completed successfully")
	return nil
}

// GetHelp 获取帮助信息
func (h *RedisSimpleHandler) GetHelp() string {
	return `Usage: redis-runner redis [options]

Redis Performance Testing Tool

Options:
  -h <hostname>         Redis server hostname (default: 127.0.0.1)
  -p <port>             Redis server port (default: 6379)
  -a <password>         Redis server password
  -n <requests>         Total number of requests (default: 1000)
  -c <connections>      Number of parallel connections (default: 10)
  -t <test>             Test case to run (default: set_get_random)
  -d <size>             Data size of values in bytes (default: 64)
  --mode <mode>         Redis mode: standalone/cluster/sentinel (default: standalone)
  --config <file>       Configuration file path
  --duration <time>     Test duration (e.g. 30s, 5m) - overrides -n
  --read-ratio <ratio>  Read/write ratio (0-100, default: 50)

Configuration File:
  --config conf/redis.yaml

Test Cases:
  set_get_random       Random SET and GET operations (default)
  set_only             Only SET operations
  get_only             Only GET operations
  lpush_lpop           List operations (LPUSH/LPOP)
  sadd_smembers        Set operations (SADD/SMEMBERS)
  zadd_zrange          Sorted set operations (ZADD/ZRANGE)
  hset_hget            Hash operations (HSET/HGET)
  incr                 Counter operations (INCR)
  append               String append operations

Examples:
  # Basic test
  redis-runner redis -h 127.0.0.1 -p 6379 -n 10000 -c 50

  # Duration-based test
  redis-runner redis -h localhost -p 6379 --duration 60s -c 100

  # Cluster test with configuration file
  redis-runner redis --config conf/redis.yaml

  # Custom test case with read-heavy workload
  redis-runner redis -t set_get_random -n 50000 -c 100 --read-ratio 80

  # Set operations only
  redis-runner redis -t set_only -n 10000 -c 50 -d 1024

For more information: https://docs.redis-runner.com/redis`
}

// loadConfiguration 加载配置
func (h *RedisSimpleHandler) loadConfiguration(args []string) error {
	// 检查是否使用配置文件
	if h.hasConfigFlag(args) {
		log.Println("Loading Redis configuration from file...")
		config, err := redisconfig.LoadRedisConfigFromFile("conf/redis.yaml")
		if err != nil {
			return err
		}
		h.configManager.SetConfig(config)
		return nil
	}

	// 使用命令行参数
	log.Println("Loading Redis configuration from command line...")
	config, err := redisconfig.LoadRedisConfigFromArgs(args)
	if err != nil {
		return err
	}
	h.configManager.SetConfig(config)
	return nil
}

// hasConfigFlag 检查是否有config标志
func (h *RedisSimpleHandler) hasConfigFlag(args []string) bool {
	for _, arg := range args {
		if arg == "-config" || arg == "--config" {
			return true
		}
	}
	return false
}

// connectRedis 连接Redis
func (h *RedisSimpleHandler) connectRedis(ctx context.Context) error {
	cfg := h.configManager.GetConfig()

	// 提取Redis配置
	var redisConfig *redisconfig.RedisConfig
	if adapter, ok := cfg.(*redisconfig.RedisConfigAdapter); ok {
		redisConfig = adapter.GetRedisConfig()
	} else {
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
func (h *RedisSimpleHandler) registerOperations() {
	redis.RegisterRedisOperations(h.operationRegistry)
}

// printResults 打印结果
func (h *RedisSimpleHandler) printResults(metrics *interfaces.Metrics) {
	cfg := h.configManager.GetConfig()
	benchmarkConfig := cfg.GetBenchmark()

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("REDIS BENCHMARK RESULTS")
	fmt.Println(strings.Repeat("=", 60))

	// 基本信息
	fmt.Printf("Test Case: %s\n", benchmarkConfig.GetTestCase())
	fmt.Printf("Total Requests: %d\n", benchmarkConfig.GetTotal())
	fmt.Printf("Parallel Connections: %d\n", benchmarkConfig.GetParallels())
	fmt.Printf("Data Size: %d bytes\n", cfg.GetBenchmark().GetDataSize())

	// 测试模式信息
	// 一些字段可能不存在，暂时注释
	// if benchmarkConfig.GetDuration() > 0 {
	//     fmt.Printf("Test Duration: %v\n", benchmarkConfig.GetDuration())
	// }
	// fmt.Printf("Read Ratio: %.0f%%\n", benchmarkConfig.GetReadRatio()*100)

	fmt.Println(strings.Repeat("-", 60))

	// 性能指标
	fmt.Printf("RPS: %d\n", metrics.RPS)
	fmt.Printf("Success Rate: %.2f%%\n", 100.0-metrics.ErrorRate)
	fmt.Printf("Total Operations: %d\n", metrics.TotalOps)
	fmt.Printf("Read Operations: %d\n", metrics.ReadOps)
	fmt.Printf("Write Operations: %d\n", metrics.WriteOps)

	fmt.Println(strings.Repeat("-", 60))

	// 延迟统计
	fmt.Printf("Avg Latency: %.3f ms\n", float64(metrics.AvgLatency)/float64(time.Millisecond))
	fmt.Printf("P90 Latency: %.3f ms\n", float64(metrics.P90Latency)/float64(time.Millisecond))
	fmt.Printf("P95 Latency: %.3f ms\n", float64(metrics.P95Latency)/float64(time.Millisecond))
	fmt.Printf("P99 Latency: %.3f ms\n", float64(metrics.P99Latency)/float64(time.Millisecond))
	fmt.Printf("Max Latency: %.3f ms\n", float64(metrics.MaxLatency)/float64(time.Millisecond))

	// Redis特定指标
	if redisMetrics := h.getRedisMetrics(); redisMetrics != nil {
		fmt.Println(strings.Repeat("-", 60))
		fmt.Println("Redis Specific Metrics:")
		if hitRate, exists := redisMetrics["hit_rate"]; exists {
			fmt.Printf("  Cache Hit Rate: %.2f%%\n", hitRate.(float64)*100)
		}
		if keyspace, exists := redisMetrics["keyspace_hits"]; exists {
			fmt.Printf("  Keyspace Hits: %v\n", keyspace)
		}
		if evictions, exists := redisMetrics["evicted_keys"]; exists {
			fmt.Printf("  Evicted Keys: %v\n", evictions)
		}
		if memory, exists := redisMetrics["memory_usage"]; exists {
			fmt.Printf("  Memory Usage: %v MB\n", memory)
		}
		if connections, exists := redisMetrics["connected_clients"]; exists {
			fmt.Printf("  Connected Clients: %v\n", connections)
		}
	}

	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("REDIS BENCHMARK COMPLETED")
	fmt.Println(strings.Repeat("=", 60))
}

// getRedisMetrics 获取Redis特定指标
func (h *RedisSimpleHandler) getRedisMetrics() map[string]interface{} {
	if h.metricsCollector == nil {
		return nil
	}

	// 这里应该从Redis适配器获取特定指标
	// 为了保持兼容性，先返回空
	return nil
}

// validateArgs 验证参数
func (h *RedisSimpleHandler) validateArgs(args []string) error {
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
			if mode != "standalone" && mode != "cluster" && mode != "sentinel" {
				return fmt.Errorf("invalid mode: %s (valid: standalone, cluster, sentinel)", mode)
			}
			i++
		case "--read-ratio":
			if i+1 >= len(args) {
				return fmt.Errorf("--read-ratio requires a value")
			}
			if ratio, err := strconv.ParseFloat(args[i+1], 64); err != nil {
				return fmt.Errorf("invalid read ratio: %s", args[i+1])
			} else if ratio < 0 || ratio > 100 {
				return fmt.Errorf("read ratio must be between 0 and 100")
			}
			i++
		}
	}
	return nil
}
