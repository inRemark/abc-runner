package commands

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"redis-runner/app/adapters/redis"
	redisconfig "redis-runner/app/adapters/redis/config"
	"redis-runner/app/core/config"
	"redis-runner/app/core/interfaces"
	"redis-runner/app/core/reports"
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
	reportManager     *reports.ReportManager
	reportArgs        *reports.ReportArgs
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

	// 1. 解析报告参数
	var err error
	h.reportArgs, err = reports.ParseReportArgs(args)
	if err != nil {
		return fmt.Errorf("failed to parse report arguments: %w", err)
	}

	// 移除报告参数，只保留业务参数
	businessArgs := reports.RemoveReportArgs(args)

	// 2. 加载配置
	if err := h.loadConfiguration(businessArgs); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// 3. 连接Redis
	if err := h.connectRedis(ctx); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}
	defer h.adapter.Close()

	// 4. 设置指标收集器
	h.metricsCollector = h.adapter.GetMetricsCollector()

	// 5. 初始化报告管理器
	h.initializeReportManager()

	// 6. 注册操作
	h.registerOperations()

	// 7. 创建运行引擎
	h.runner = runner.NewEnhancedRunner(
		h.adapter,
		h.configManager.GetConfig(),
		h.metricsCollector,
		h.keyGenerator,
		h.operationRegistry,
	)

	// 8. 执行基准测试
	metrics, err := h.runner.RunBenchmark(ctx)
	if err != nil {
		return fmt.Errorf("benchmark execution failed: %w", err)
	}

	// 9. 输出结果
	h.printResults(metrics)

	log.Println("Redis benchmark completed successfully")
	return nil
}

// GetHelp 获取帮助信息
func (h *RedisSimpleHandler) GetHelp() string {
	baseHelp := `Usage: redis-runner redis [options]

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

	return reports.AddReportArgsToHelp(baseHelp)
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

// initializeReportManager 初始化报告管理器
func (h *RedisSimpleHandler) initializeReportManager() {
	if h.reportArgs == nil {
		h.reportArgs = reports.DefaultReportArgs()
	}

	reportConfig := h.reportArgs.ToReportConfig("redis")
	h.reportManager = reports.NewReportManager("redis", h.metricsCollector, reportConfig)
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

	// 生成详细报告
	h.generateDetailedReports(metrics)

	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("REDIS BENCHMARK COMPLETED")
	fmt.Println(strings.Repeat("=", 60))
}

// getRedisMetrics 获取Redis特定指标
func (h *RedisSimpleHandler) getRedisMetrics() map[string]interface{} {
	if h.metricsCollector == nil {
		return nil
	}

	// 从适配器获取Redis特定指标
	redisAdapter := h.adapter
	if redisAdapter == nil {
		return nil
	}

	// 获取Redis客户端的状态信息
	redisInfo := make(map[string]interface{})

	// 获取连接池状态
	if poolStats := h.getConnectionPoolStats(redisAdapter); poolStats != nil {
		redisInfo["connection_pool"] = poolStats
	}

	// 获取Redis的INFO信息（如果可能）
	if serverInfo := h.getRedisServerInfo(redisAdapter); serverInfo != nil {
		redisInfo["server_info"] = serverInfo
	}

	// 获取操作统计
	if exportedMetrics := h.metricsCollector.Export(); exportedMetrics != nil {
		redisInfo["detailed_metrics"] = exportedMetrics
	}

	return redisInfo
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

// getConnectionPoolStats 获取连接池统计
func (h *RedisSimpleHandler) getConnectionPoolStats(adapter *redis.RedisAdapter) map[string]interface{} {
	poolStats := make(map[string]interface{})

	// 这里可以添加具体的连接池统计代码
	// 由于RedisAdapter的内部结构，这里使用适配器的公共方法
	poolStats["mode"] = adapter.GetMode()
	poolStats["connected"] = adapter.IsConnected()

	return poolStats
}

// getRedisServerInfo 获取Redis服务器信息
func (h *RedisSimpleHandler) getRedisServerInfo(adapter *redis.RedisAdapter) map[string]interface{} {
	serverInfo := make(map[string]interface{})

	// 添加基本服务器信息
	serverInfo["protocol"] = "redis"
	serverInfo["adapter_type"] = "redis"

	return serverInfo
}

// generateDetailedReports 生成详细报告
func (h *RedisSimpleHandler) generateDetailedReports(metrics *interfaces.Metrics) {
	// 使用统一报告管理器
	if h.reportManager != nil {
		// 设置Redis特定指标
		h.reportManager.SetProtocolMetrics(h.getRedisMetrics())

		// 生成所有报告
		if err := h.reportManager.GenerateReports(); err != nil {
			fmt.Printf("Warning: Failed to generate reports: %v\n", err)
		}
		return
	}

	// 备用方案：使用原有的报告生成逻辑
	// 检查指标收集器是否支持报告生成
	if h.metricsCollector == nil {
		fmt.Println("\nWARNING: Metrics collector not available for detailed reporting")
		return
	}

	fmt.Println("\n" + strings.Repeat("-", 60))
	fmt.Println("DETAILED REDIS PERFORMANCE REPORT")
	fmt.Println(strings.Repeat("-", 60))

	// 生成控制台详细报告
	h.generateConsoleReport()

	// 生成JSON报告文件
	h.generateJSONReport()

	// 生成CSV报告文件
	h.generateCSVReport()

	fmt.Println("\n" + strings.Repeat("-", 60))
	fmt.Println("Report Generation Completed")
	fmt.Println(strings.Repeat("-", 60))
}

// generateConsoleReport 生成控制台报告
func (h *RedisSimpleHandler) generateConsoleReport() {
	// 使用Redis指标报告器生成详细控制台报告
	if exportedMetrics := h.metricsCollector.Export(); exportedMetrics != nil {
		fmt.Println("\n=== Console Detailed Report ===")

		// 显示详细指标
		for key, value := range exportedMetrics {
			switch key {
			case "rps":
				fmt.Printf("Requests per Second: %v\n", value)
			case "avg_latency":
				if latency, ok := value.(int64); ok {
					fmt.Printf("Average Latency: %.3f ms\n", float64(latency)/float64(time.Millisecond))
				}
			case "p95_latency":
				if latency, ok := value.(int64); ok {
					fmt.Printf("P95 Latency: %.3f ms\n", float64(latency)/float64(time.Millisecond))
				}
			case "p99_latency":
				if latency, ok := value.(int64); ok {
					fmt.Printf("P99 Latency: %.3f ms\n", float64(latency)/float64(time.Millisecond))
				}
			case "error_rate":
				fmt.Printf("Error Rate: %.2f%%\n", value)
			case "duration":
				if duration, ok := value.(int64); ok {
					fmt.Printf("Total Duration: %.3f seconds\n", float64(duration)/float64(time.Second))
				}
			}
		}
	}
}

// generateJSONReport 生成JSON报告
func (h *RedisSimpleHandler) generateJSONReport() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Warning: Failed to generate JSON report: %v\n", r)
		}
	}()

	filename := fmt.Sprintf("redis_benchmark_%s.json", time.Now().Format("20060102_150405"))

	// 构建报告数据
	reportData := map[string]interface{}{
		"timestamp":     time.Now().Format(time.RFC3339),
		"protocol":      "redis",
		"base_metrics":  h.metricsCollector.Export(),
		"redis_metrics": h.getRedisMetrics(),
	}

	// 将数据序列化为JSON
	jsonData, err := json.MarshalIndent(reportData, "", "  ")
	if err != nil {
		fmt.Printf("Warning: Failed to marshal JSON report: %v\n", err)
		return
	}

	// 写入文件
	err = ioutil.WriteFile(filename, jsonData, 0644)
	if err != nil {
		fmt.Printf("Warning: Failed to write JSON report to %s: %v\n", filename, err)
		return
	}

	fmt.Printf("JSON report saved to: %s\n", filename)
}

// generateCSVReport 生成CSV报告
func (h *RedisSimpleHandler) generateCSVReport() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Warning: Failed to generate CSV report: %v\n", r)
		}
	}()

	filename := fmt.Sprintf("redis_benchmark_%s.csv", time.Now().Format("20060102_150405"))

	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Warning: Failed to create CSV report file %s: %v\n", filename, err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入CSV头部
	header := []string{"timestamp", "total_ops", "success_ops", "failed_ops", "rps", "avg_latency_ms", "p95_latency_ms", "p99_latency_ms", "error_rate"}
	if err := writer.Write(header); err != nil {
		fmt.Printf("Warning: Failed to write CSV header: %v\n", err)
		return
	}

	// 获取指标数据
	metrics := h.metricsCollector.Export()

	// 写入数据行（安全处理类型断言）
	record := []string{
		time.Now().Format(time.RFC3339),
		fmt.Sprintf("%v", h.getMetricValue(metrics, "total_ops")),
		fmt.Sprintf("%v", h.getMetricValue(metrics, "success_ops")),
		fmt.Sprintf("%v", h.getMetricValue(metrics, "failed_ops")),
		fmt.Sprintf("%v", h.getMetricValue(metrics, "rps")),
		fmt.Sprintf("%.3f", h.getLatencyInMs(metrics, "avg_latency")),
		fmt.Sprintf("%.3f", h.getLatencyInMs(metrics, "p95_latency")),
		fmt.Sprintf("%.3f", h.getLatencyInMs(metrics, "p99_latency")),
		fmt.Sprintf("%.2f", h.getMetricValue(metrics, "error_rate")),
	}

	if err := writer.Write(record); err != nil {
		fmt.Printf("Warning: Failed to write CSV record: %v\n", err)
		return
	}

	fmt.Printf("CSV report saved to: %s\n", filename)
}

// getMetricValue 安全获取指标值
func (h *RedisSimpleHandler) getMetricValue(metrics map[string]interface{}, key string) interface{} {
	if value, exists := metrics[key]; exists {
		return value
	}
	return 0
}

// getLatencyInMs 获取延迟值（毫秒）
func (h *RedisSimpleHandler) getLatencyInMs(metrics map[string]interface{}, key string) float64 {
	if value, exists := metrics[key]; exists {
		if latency, ok := value.(int64); ok {
			return float64(latency) / float64(time.Millisecond)
		}
	}
	return 0.0
}
