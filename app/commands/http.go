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
	"redis-runner/app/core/config"
	"redis-runner/app/core/interfaces"
	"redis-runner/app/core/runner"
	"redis-runner/app/core/utils"
)

// HttpSimpleHandler 简化的HTTP命令处理器
type HttpSimpleHandler struct {
	adapter           *http.HttpAdapter
	configManager     *config.ConfigManager
	operationRegistry *utils.OperationRegistry
	keyGenerator      *utils.DefaultKeyGenerator
	metricsCollector  interfaces.MetricsCollector
	runner            *runner.EnhancedRunner
}

// NewHttpCommandHandler 创建HTTP命令处理器（统一接口）
func NewHttpCommandHandler() *HttpSimpleHandler {
	return &HttpSimpleHandler{
		adapter:           http.NewHttpAdapter(),
		configManager:     config.NewConfigManager(),
		operationRegistry: utils.NewOperationRegistry(),
		keyGenerator:      utils.NewDefaultKeyGenerator(),
	}
}

// Execute 执行HTTP命令
func (h *HttpSimpleHandler) Execute(ctx context.Context, args []string) error {
	// 检查是否请求帮助
	for _, arg := range args {
		if arg == "--help" || arg == "-h" || arg == "help" {
			fmt.Println(h.GetHelp())
			return nil
		}
	}

	log.Println("Starting HTTP load test...")

	// 1. 加载配置
	if err := h.loadConfiguration(args); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// 2. 连接HTTP服务
	if err := h.connectHttp(ctx); err != nil {
		return fmt.Errorf("failed to connect to HTTP service: %w", err)
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

	log.Println("HTTP load test completed successfully")
	return nil
}

// GetHelp 获取帮助信息
func (h *HttpSimpleHandler) GetHelp() string {
	return `Usage: redis-runner http [options]

HTTP Load Testing Tool

Options:
  --url <url>           Target URL (default: http://localhost:8080)
  --method <method>     HTTP method: GET, POST, PUT, DELETE (default: GET)
  --path <path>         Request path (default: /)
  -n <requests>         Total number of requests (default: 1000)
  -c <connections>      Number of parallel connections (default: 10)
  --timeout <duration>  Request timeout (default: 30s)
  --duration <time>     Test duration (e.g. 30s, 5m) - overrides -n
  --body <data>         Request body for POST/PUT
  --header <key:value>  Add HTTP header (can be used multiple times)
  --content-type <type> Content type header (default: application/json)
  --config <file>       Configuration file path

Configuration File:
  --config conf/http.yaml

Load Patterns:
  --ramp-up <duration>  Ramp-up time to reach target connections
  --keep-alive          Use HTTP keep-alive connections (default: true)
  --follow-redirects    Follow HTTP redirects (default: false)

Examples:
  # Basic GET test
  redis-runner http --url http://localhost:8080/api/users -n 10000 -c 50

  # POST test with JSON body
  redis-runner http --url http://api.example.com/users --method POST \\
    --body '{"name":"test","email":"test@example.com"}' \\
    --content-type application/json -n 1000 -c 20

  # Duration-based test with custom headers
  redis-runner http --url https://api.example.com/health \\
    --duration 60s -c 100 \\
    --header "Authorization:Bearer token123" \\
    --header "X-API-Key:secret"

  # Load test with configuration file
  redis-runner http --config conf/http.yaml

  # Stress test with ramp-up
  redis-runner http --url http://localhost:8080 \\
    --duration 5m -c 200 --ramp-up 30s

For more information: https://docs.redis-runner.com/http`
}

// loadConfiguration 加载配置
func (h *HttpSimpleHandler) loadConfiguration(args []string) error {
	// 检查是否使用配置文件
	if h.hasConfigFlag(args) {
		log.Println("Loading HTTP configuration from file...")
		// 使用多源配置加载器
		sources := config.CreateHttpConfigSources("conf/http.yaml", nil)
		return h.configManager.LoadConfiguration(sources...)
	}

	// 使用命令行参数创建配置
	log.Println("Loading HTTP configuration from command line...")
	httpCfg := h.createConfigFromArgs(args)
	h.configManager.SetConfig(httpCfg)
	return nil
}

// hasConfigFlag 检查是否有config标志
func (h *HttpSimpleHandler) hasConfigFlag(args []string) bool {
	for _, arg := range args {
		if arg == "-config" || arg == "--config" {
			return true
		}
	}
	return false
}

// createConfigFromArgs 从命令行参数创建配置
func (h *HttpSimpleHandler) createConfigFromArgs(args []string) *httpConfig.HttpAdapterConfig {
	// 默认配置
	cfg := &httpConfig.HttpAdapterConfig{
		Protocol: "http",
		Connection: httpConfig.HttpConnectionConfig{
			BaseURL:         "http://localhost:8080",
			Timeout:         30 * time.Second,
			MaxConnsPerHost: 10,
			MaxIdleConns:    10,
			IdleConnTimeout: 90 * time.Second,
			KeepAlive:       90 * time.Second, // 使用正确的time.Duration类型
		},
		// 添加默认的请求配置以满足验证要求
		Requests: []httpConfig.HttpRequestConfig{
			{
				Method:      "GET",
				Path:        "/",
				Headers:     map[string]string{"Accept": "application/json"},
				ContentType: "application/json",
				Weight:      100,
			},
		},
		// 添加默认的认证配置
		Auth: httpConfig.HttpAuthConfig{
			Type: "none",
		},
		Benchmark: httpConfig.HttpBenchmarkConfig{
			Total:           1000,
			Parallels:       10,
			Method:          "GET",
			Path:            "/",
			Headers:         make(map[string]string),
			QueryParams:     make(map[string]string),
			FollowRedirects: false,
		},
	}

	headers := make(map[string]string)
	// 获取默认请求配置用于更新
	defaultRequest := &cfg.Requests[0]

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
				method := strings.ToUpper(args[i+1])
				cfg.Benchmark.Method = method
				defaultRequest.Method = method
				i++
			}
		case "--path":
			if i+1 < len(args) {
				path := args[i+1]
				cfg.Benchmark.Path = path
				defaultRequest.Path = path
				i++
			}
		case "--timeout":
			if i+1 < len(args) {
				if timeout, err := time.ParseDuration(args[i+1]); err == nil {
					cfg.Connection.Timeout = timeout
				}
				i++
			}
		case "--duration":
			if i+1 < len(args) {
				if duration, err := time.ParseDuration(args[i+1]); err == nil {
					cfg.Benchmark.Duration = duration
				}
				i++
			}
		case "--body":
			if i+1 < len(args) {
				// Body 字段不存在，暂时注释
				// cfg.Benchmark.Body = args[i+1]
				log.Printf("Body field not implemented yet: %s", args[i+1])
				i++
			}
		case "--header":
			if i+1 < len(args) {
				headerParts := strings.SplitN(args[i+1], ":", 2)
				if len(headerParts) == 2 {
					headers[strings.TrimSpace(headerParts[0])] = strings.TrimSpace(headerParts[1])
				}
				i++
			}
		case "--content-type":
			if i+1 < len(args) {
				headers["Content-Type"] = args[i+1]
				i++
			}
		case "--ramp-up":
			if i+1 < len(args) {
				// RampUpTime 字段不存在，暂时注释
				// if rampUp, err := time.ParseDuration(args[i+1]); err == nil {
				//     cfg.Benchmark.RampUpTime = rampUp
				// }
				log.Printf("Ramp-up time field not implemented yet: %s", args[i+1])
				i++
			}
		case "--keep-alive":
			// KeepAlive 字段可能是 bool 类型
			// cfg.Connection.KeepAlive = true
		case "--no-keep-alive":
			// cfg.Connection.KeepAlive = 0
		case "--follow-redirects":
			cfg.Benchmark.FollowRedirects = true
		}
	}

	// 设置收集到的headers
	if len(headers) > 0 {
		cfg.Benchmark.Headers = headers
		// 同时更新请求配置中的headers
		for k, v := range headers {
			defaultRequest.Headers[k] = v
		}
	}

	return cfg
}

// connectHttp 连接HTTP服务
func (h *HttpSimpleHandler) connectHttp(ctx context.Context) error {
	cfg := h.configManager.GetConfig()

	if httpCfg, ok := cfg.(*httpConfig.HttpAdapterConfig); ok {
		log.Printf("Connecting to HTTP service: %s", httpCfg.Connection.BaseURL)
	} else {
		log.Println("Connecting to HTTP service...")
	}

	if err := h.adapter.Connect(ctx, cfg); err != nil {
		return err
	}

	log.Println("HTTP connection established successfully")
	return nil
}

// registerOperations 注册操作
func (h *HttpSimpleHandler) registerOperations() {
	// HTTP操作注册 - 简化实现
	// TODO: 实现具体的HTTP操作注册
	log.Println("HTTP operations registry not fully implemented yet")
}

// printResults 打印结果
func (h *HttpSimpleHandler) printResults(metrics *interfaces.Metrics) {
	cfg := h.configManager.GetConfig()
	var httpCfg *httpConfig.HttpAdapterConfig
	if hcfg, ok := cfg.(*httpConfig.HttpAdapterConfig); ok {
		httpCfg = hcfg
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("HTTP LOAD TEST RESULTS")
	fmt.Println(strings.Repeat("=", 60))

	// 基本信息
	if httpCfg != nil {
		fmt.Printf("Target URL: %s%s\n", httpCfg.Connection.BaseURL, httpCfg.Benchmark.Path)
		fmt.Printf("HTTP Method: %s\n", httpCfg.Benchmark.Method)
		fmt.Printf("Total Requests: %d\n", httpCfg.Benchmark.Total)
		fmt.Printf("Parallel Connections: %d\n", httpCfg.Benchmark.Parallels)

		if httpCfg.Benchmark.Duration > 0 {
			fmt.Printf("Test Duration: %v\n", httpCfg.Benchmark.Duration)
		}
		// RampUpTime 字段不存在，暂时注释
		// if httpCfg.Benchmark.RampUpTime > 0 {
		//     fmt.Printf("Ramp-up Time: %v\n", httpCfg.Benchmark.RampUpTime)
		// }
	}

	fmt.Println(strings.Repeat("-", 60))

	// 性能指标
	fmt.Printf("RPS: %d\n", metrics.RPS)
	fmt.Printf("Success Rate: %.2f%%\n", 100.0-metrics.ErrorRate)
	fmt.Printf("Total Operations: %d\n", metrics.TotalOps)

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

	// HTTP特定指标
	if httpMetrics := h.getHttpMetrics(); httpMetrics != nil {
		fmt.Println(strings.Repeat("-", 60))
		fmt.Println("HTTP Specific Metrics:")
		if statusCodes, exists := httpMetrics["status_codes"]; exists {
			fmt.Printf("  Status Code Distribution: %v\n", statusCodes)
		}
		if dataTransferred, exists := httpMetrics["data_transferred"]; exists {
			fmt.Printf("  Data Transferred: %v MB\n", dataTransferred)
		}
		if avgResponseSize, exists := httpMetrics["avg_response_size"]; exists {
			fmt.Printf("  Avg Response Size: %v bytes\n", avgResponseSize)
		}
		if connectTime, exists := httpMetrics["avg_connect_time"]; exists {
			fmt.Printf("  Avg Connect Time: %.3f ms\n", connectTime.(float64))
		}
		if dnsTime, exists := httpMetrics["avg_dns_time"]; exists {
			fmt.Printf("  Avg DNS Lookup Time: %.3f ms\n", dnsTime.(float64))
		}
	}

	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("HTTP LOAD TEST COMPLETED")
	fmt.Println(strings.Repeat("=", 60))
}

// getHttpMetrics 获取HTTP特定指标
func (h *HttpSimpleHandler) getHttpMetrics() map[string]interface{} {
	if h.metricsCollector == nil {
		return nil
	}

	// TODO: 这里应该从HTTP适配器获取特定指标, 为了保持兼容性，先返回空
	return nil
}

// validateArgs 验证参数
func (h *HttpSimpleHandler) validateArgs(args []string) error {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--url":
			if i+1 >= len(args) {
				return fmt.Errorf("--url requires a URL")
			}
			url := args[i+1]
			if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
				return fmt.Errorf("invalid URL: %s (must start with http:// or https://)", url)
			}
			i++
		case "--method":
			if i+1 >= len(args) {
				return fmt.Errorf("--method requires a HTTP method")
			}
			method := strings.ToUpper(args[i+1])
			validMethods := []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS", "PATCH"}
			valid := false
			for _, vm := range validMethods {
				if method == vm {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("invalid HTTP method: %s", method)
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
		case "--timeout", "--duration", "--ramp-up":
			if i+1 >= len(args) {
				return fmt.Errorf("%s requires a duration value", args[i])
			}
			if _, err := time.ParseDuration(args[i+1]); err != nil {
				return fmt.Errorf("invalid duration for %s: %s", args[i], args[i+1])
			}
			i++
		}
	}
	return nil
}
