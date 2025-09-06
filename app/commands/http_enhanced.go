package commands

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"redis-runner/app/adapters/http"
	httpConfig "redis-runner/app/adapters/http/config"
	"redis-runner/app/core/command"
	"redis-runner/app/core/config"
	"redis-runner/app/core/interfaces"
	"redis-runner/app/core/runner"
	"redis-runner/app/core/utils"
)

// HttpCommandHandler HTTP增强版命令处理器
type HttpCommandHandler struct {
	*command.BaseCommandHandler
	adapter           *http.HttpAdapter
	operationRegistry *utils.OperationRegistry
	keyGenerator      *utils.DefaultKeyGenerator
	metricsCollector  interfaces.MetricsCollector
	runner            *runner.EnhancedRunner
}

// NewHttpCommandHandler 创建HTTP命令处理器
func NewHttpCommandHandler() *HttpCommandHandler {
	configManager := config.NewConfigManager()
	adapter := http.NewHttpAdapter()
	
	baseHandler := command.NewBaseCommandHandler(
		"http-enhanced",
		"HTTP load testing with enterprise features",
		command.Enhanced,
		false, // 不是弃用的
		adapter,
		configManager,
	)

	return &HttpCommandHandler{
		BaseCommandHandler: baseHandler,
		adapter:            adapter,
		operationRegistry:  utils.NewOperationRegistry(),
		keyGenerator:       utils.NewDefaultKeyGenerator(),
		metricsCollector:   adapter.GetMetricsCollector(),
	}
}

// ExecuteCommand 执行HTTP命令
func (h *HttpCommandHandler) ExecuteCommand(ctx context.Context, args []string) error {
	log.Println("Starting HTTP Enhanced benchmark...")

	// 1. 加载配置
	if err := h.loadConfiguration(args); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// 2. 连接HTTP服务
	if err := h.connectHttp(ctx); err != nil {
		return fmt.Errorf("failed to connect to HTTP service: %w", err)
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
func (h *HttpCommandHandler) loadConfiguration(args []string) error {
	configManager := h.GetConfigManager()

	// 检查是否使用配置文件
	if h.hasConfigFlag(args) {
		log.Println("Loading HTTP configuration from file...")
		sources := config.CreateHttpConfigSources("conf/http.yaml", nil)
		return configManager.LoadConfiguration(sources...)
	}

	// 使用命令行参数创建配置
	log.Println("Loading HTTP configuration from command line...")
	httpConfig := h.createConfigFromArgs(args)
	configManager.SetConfig(httpConfig)
	return nil
}

// hasConfigFlag 检查是否有config标志
func (h *HttpCommandHandler) hasConfigFlag(args []string) bool {
	for _, arg := range args {
		if arg == "-config" || arg == "--config" {
			return true
		}
	}
	return false
}

// createConfigFromArgs 从命令行参数创建配置
func (h *HttpCommandHandler) createConfigFromArgs(args []string) *httpConfig.HttpAdapterConfig {
	// 默认配置
	cfg := &httpConfig.HttpAdapterConfig{
		Protocol: "http",
		Connection: httpConfig.HttpConnectionConfig{
			BaseURL:        "http://localhost:8080",
			Timeout:        30 * time.Second,
			MaxConnsPerHost: 10,
			MaxIdleConns:   10,
			IdleConnTimeout: 90 * time.Second,
		},
		Benchmark: httpConfig.HttpBenchmarkConfig{
			Total:       1000,
			Parallels:   10,
			Method:      "GET",
			Path:        "/",
			Headers:     make(map[string]string),
			QueryParams: make(map[string]string),
		},
	}

	// 解析命令行参数
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--url":
			if i+1 < len(args) {
				cfg.Connection.BaseURL = args[i+1]
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
					cfg.Connection.MaxConnsPerHost = parallels
				}
				i++
			}
		case "--method":
			if i+1 < len(args) {
				cfg.Benchmark.Method = strings.ToUpper(args[i+1])
				i++
			}
		case "--path":
			if i+1 < len(args) {
				cfg.Benchmark.Path = args[i+1]
				i++
			}
		case "--timeout":
			if i+1 < len(args) {
				if timeout, err := time.ParseDuration(args[i+1]); err == nil {
					cfg.Connection.Timeout = timeout
				}
				i++
			}
		}
	}

	return cfg
}

// connectHttp 连接HTTP服务
func (h *HttpCommandHandler) connectHttp(ctx context.Context) error {
	cfg := h.GetConfigManager().GetConfig()

	log.Printf("Connecting to HTTP service: %s", cfg.(*httpConfig.HttpAdapterConfig).Connection.BaseURL)

	if err := h.adapter.Connect(ctx, cfg); err != nil {
		return err
	}

	log.Println("HTTP connection established successfully")
	return nil
}

// registerOperations 注册操作
func (h *HttpCommandHandler) registerOperations() {
	http.RegisterHttpOperations(h.operationRegistry)
}

// printResults 打印结果
func (h *HttpCommandHandler) printResults(metrics *interfaces.Metrics) {
	cfg := h.GetConfigManager().GetConfig().(*httpConfig.HttpAdapterConfig)

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("HTTP BENCHMARK RESULTS")
	fmt.Println(strings.Repeat("=", 60))

	// 基本信息
	fmt.Printf("Target URL: %s\n", cfg.Connection.BaseURL)
	fmt.Printf("HTTP Method: %s\n", cfg.Benchmark.Method)
	fmt.Printf("Request Path: %s\n", cfg.Benchmark.Path)
	fmt.Printf("Total Requests: %d\n", cfg.Benchmark.Total)
	fmt.Printf("Parallel Connections: %d\n", cfg.Benchmark.Parallels)
	fmt.Printf("RPS: %d\n", metrics.RPS)
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

	// HTTP特定指标
	httpMetrics := h.adapter.GetProtocolMetrics()
	if statusCodes, exists := httpMetrics["status_codes"]; exists {
		if codes, ok := statusCodes.(map[string]int); ok {
			fmt.Println("HTTP Status Codes:")
			for code, count := range codes {
				fmt.Printf("  %s: %d requests\n", code, count)
			}
		}
	}

	// 连接池统计
	if poolStats := h.adapter.GetConnectionPoolStats(); poolStats != nil {
		fmt.Println("\nConnection Pool Statistics:")
		for key, value := range poolStats {
			fmt.Printf("  %s: %v\n", key, value)
		}
	}

	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("HTTP BENCHMARK COMPLETED")
	fmt.Println(strings.Repeat("=", 60))
}

// GetUsage 获取使用说明
func (h *HttpCommandHandler) GetUsage() string {
	return `Usage: redis-runner http-enhanced [options]

Enhanced HTTP Load Testing Tool

Options:
  --url <url>           Target URL (default: http://localhost:8080)
  --method <method>     HTTP method (default: GET)
  --path <path>         Request path (default: /)
  -n <requests>         Total number of requests (default: 1000)
  -c <connections>      Number of parallel connections (default: 10)
  --timeout <duration>  Request timeout (default: 30s)
  --config <file>       Configuration file path

Configuration File:
  --config conf/http.yaml

Examples:
  # Basic GET test
  redis-runner http-enhanced --url http://api.example.com -n 10000 -c 50

  # POST test with configuration file
  redis-runner http-enhanced --config conf/http.yaml

  # Custom method and path
  redis-runner http-enhanced --url http://localhost:8080 --method POST --path /api/users -n 5000

For more information: https://docs.redis-runner.com/http-enhanced`
}

// ValidateArgs 验证参数
func (h *HttpCommandHandler) ValidateArgs(args []string) error {
	// 基本参数验证
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--url":
			if i+1 >= len(args) {
				return fmt.Errorf("--url requires a value")
			}
			url := args[i+1]
			if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
				return fmt.Errorf("URL must start with http:// or https://")
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
		case "--method":
			if i+1 >= len(args) {
				return fmt.Errorf("--method requires a value")
			}
			method := strings.ToUpper(args[i+1])
			validMethods := []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS", "PATCH"}
			found := false
			for _, valid := range validMethods {
				if method == valid {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("invalid HTTP method: %s", args[i+1])
			}
			i++
		case "--timeout":
			if i+1 >= len(args) {
				return fmt.Errorf("--timeout requires a value")
			}
			if _, err := time.ParseDuration(args[i+1]); err != nil {
				return fmt.Errorf("invalid timeout duration: %s", args[i+1])
			}
			i++
		}
	}

	return nil
}