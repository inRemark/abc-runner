package commands

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"redis-runner/app/adapters/redis"
	"redis-runner/app/core/base"
	"redis-runner/app/core/config"
	"redis-runner/app/core/interfaces"
	"redis-runner/app/core/runner"
	"redis-runner/app/core/utils"
)

// RedisCommandHandler 新的Redis命令处理器
type RedisCommandHandler struct {
	configManager     *config.ConfigManager
	adapter           *redis.RedisAdapter
	operationRegistry *utils.OperationRegistry
	keyGenerator      *utils.DefaultKeyGenerator
	metricsCollector  interfaces.MetricsCollector
	runner            *runner.EnhancedRunner
}

// NewRedisCommandHandler 创建Redis命令处理器
func NewRedisCommandHandler() *RedisCommandHandler {
	return &RedisCommandHandler{
		configManager:     config.NewConfigManager(),
		adapter:           redis.NewRedisAdapter(),
		operationRegistry: utils.NewOperationRegistry(),
		keyGenerator:      utils.NewDefaultKeyGenerator(),
		metricsCollector:  base.NewDefaultMetricsCollector(),
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
		h.configManager.GetConfig(),
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
	// 检查是否使用配置文件
	if h.hasConfigFlag(args) {
		log.Println("Loading configuration from file...")
		sources := config.CreateDefaultSources("", nil) // 只使用文件源
		return h.configManager.LoadConfiguration(sources...)
	}

	// 使用命令行参数
	log.Println("Loading configuration from command line...")
	sources := config.CreateDefaultSources("", args)
	return h.configManager.LoadConfiguration(sources...)
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
	cfg := h.configManager.GetConfig()

	log.Printf("Connecting to Redis in %s mode...", cfg.(*config.RedisConfig).Mode)

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
	cfg := h.configManager.GetConfig()
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

// ExecuteFromConfig 从配置文件执行
func (h *RedisCommandHandler) ExecuteFromConfig() error {
	ctx := context.Background()
	return h.ExecuteCommand(ctx, []string{"-config"})
}
