package commands

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"abc-runner/app/adapters/http"
	httpConfig "abc-runner/app/adapters/http/config"
	"abc-runner/app/adapters/http/operations"
	"abc-runner/app/core/config"
	"abc-runner/app/core/interfaces"
	"abc-runner/app/core/reports"
	"abc-runner/app/core/runner"
	"abc-runner/app/core/utils"
)

// HttpSimpleHandler 简化的HTTP命令处理器
type HttpSimpleHandler struct {
	adapter           *http.HttpAdapter
	configManager     *config.ConfigManager
	operationRegistry *utils.OperationRegistry
	keyGenerator      *utils.DefaultKeyGenerator
	metricsCollector  interfaces.MetricsCollector
	runner            *runner.EnhancedRunner
	reportManager     *reports.ReportManager
	reportArgs        *reports.ReportArgs
}

// NewHttpCommandHandler 创建HTTP命令处理器（统一接口）
func NewHttpCommandHandler() *HttpSimpleHandler {
	handler := &HttpSimpleHandler{
		adapter:           http.NewHttpAdapter(),
		configManager:     config.NewConfigManager(),
		operationRegistry: utils.NewOperationRegistry(),
		keyGenerator:      utils.NewDefaultKeyGenerator(),
	}

	// 注册HTTP操作工厂
	// 使用默认配置创建HTTP操作工厂
	httpCfg := httpConfig.DefaultHTTPConfig()
	httpOpsFactory := operations.NewHttpOperationFactory(httpCfg)
	handler.operationRegistry.Register("http_get", httpOpsFactory)
	handler.operationRegistry.Register("http_post", httpOpsFactory)
	handler.operationRegistry.Register("http_put", httpOpsFactory)
	handler.operationRegistry.Register("http_delete", httpOpsFactory)
	handler.operationRegistry.Register("http_head", httpOpsFactory)
	handler.operationRegistry.Register("http_options", httpOpsFactory)

	return handler
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

	// 1. 解析报告参数
	var err error
	h.reportArgs, err = reports.ParseReportArgs(args)
	if err != nil {
		return fmt.Errorf("failed to parse report arguments: %w", err)
	}

	// 2. 验证参数
	if err := h.validateArgs(args); err != nil {
		return fmt.Errorf("argument validation failed: %w", err)
	}

	// 3. 加载配置
	if err := h.loadConfiguration(args); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// 4. 初始化指标收集器
	h.metricsCollector = h.adapter.GetMetricsCollector()

	// 5. 初始化报告管理器
	h.initializeReportManager()

	// 6. 初始化运行器
	h.runner = runner.NewEnhancedRunner(h.adapter, h.configManager.GetConfig(), h.metricsCollector, h.keyGenerator, h.operationRegistry)

	// 7. 运行测试
	log.Println("Running HTTP load test...")
	_, err = h.runner.RunBenchmark(ctx)
	if err != nil {
		return fmt.Errorf("load test execution failed: %w", err)
	}

	// 8. 生成报告
	log.Println("Generating reports...")
	if err := h.reportManager.GenerateReports(); err != nil {
		return fmt.Errorf("report generation failed: %w", err)
	}

	log.Println("HTTP load test completed successfully")
	return nil
}

// GetHelp 获取帮助信息
func (h *HttpSimpleHandler) GetHelp() string {
	baseHelp := `Usage: abc-runner http [OPTIONS]

HTTP Load Testing Tool

Options:
  --url URL                Target URL
  --method METHOD          HTTP method (default: GET)
  --body BODY              Request body
  --content-type TYPE      Content type
  --header HEADER          Request header (can be used multiple times)
  -n, --requests COUNT     Total number of requests (default: 10000)
  -c, --concurrency COUNT  Number of parallel connections (default: 50)
  --duration DURATION      Test duration (e.g., 30s, 5m)
  --timeout DURATION       Request timeout (default: 30s)
  --ramp-up DURATION       Ramp-up time (default: 0s)
  --config FILE            Configuration file path
  --core-config FILE       Core configuration file path (default: config/core.yaml)

Examples:
  # Basic GET test
  abc-runner http --url http://localhost:8080/api/users -n 10000 -c 50

  # POST test with JSON body
  abc-runner http --url http://api.example.com/users --method POST \\
    --body '{"name":"test","email":"test@example.com"}' \\
    --content-type application/json -n 1000 -c 20

  # Duration-based test with custom headers
  abc-runner http --url https://api.example.com/health \\
    --duration 60s -c 100 \\
    --header "Authorization:Bearer token123" \\
    --header "X-API-Key:secret"

  # Load test with configuration file
  abc-runner http --config config/http.yaml

  # Load test with core configuration
  abc-runner http --config config/http.yaml --core-config config/core.yaml

  # Stress test with ramp-up
  abc-runner http --url http://localhost:8080 \\
    --duration 5m -c 200 --ramp-up 30s

For more information: https://docs.abc-runner.com/http`

	return reports.AddReportArgsToHelp(baseHelp)
}

// loadConfiguration 加载配置
func (h *HttpSimpleHandler) loadConfiguration(args []string) error {
	// 检查是否使用核心配置文件
	coreConfigPath := h.getCoreConfigFlag(args)
	if coreConfigPath != "" {
		log.Printf("Loading core configuration from %s...", coreConfigPath)
		if err := h.configManager.LoadCoreConfiguration(coreConfigPath); err != nil {
			return fmt.Errorf("failed to load core configuration: %w", err)
		}
	}

	// 检查是否使用配置文件
	if h.hasConfigFlag(args) {
		configPath := h.getConfigFlagValue(args)
		log.Println("Loading HTTP configuration from file...")
		// 使用多源配置加载器
		sources := config.CreateHttpConfigSources(configPath, nil)
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
func (h *HttpSimpleHandler) getConfigFlagValue(args []string) string {
	for i, arg := range args {
		if (arg == "--config" || arg == "-C") && i+1 < len(args) {
			return args[i+1]
		}
		if strings.HasPrefix(arg, "--config=") {
			return strings.TrimPrefix(arg, "--config=")
		}
	}

	// 使用统一的配置文件查找机制
	foundPath := utils.FindConfigFile("http")
	if foundPath != "" {
		return foundPath
	}

	// 回退到默认路径
	return "./config/http.yaml"
}

// getCoreConfigFlag 获取核心配置文件路径
func (h *HttpSimpleHandler) getCoreConfigFlag(args []string) string {
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
func (h *HttpSimpleHandler) createConfigFromArgs(args []string) *httpConfig.HttpAdapterConfig {
	// 默认配置
	cfg := httpConfig.DefaultHTTPConfig()

	// 设置默认测试用例
	cfg.Benchmark.TestCase = "http_get"

	// 解析命令行参数
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--url":
			if i+1 < len(args) {
				cfg.Connection.BaseURL = args[i+1]
				i++
			}
		case "--method":
			if i+1 < len(args) {
				if len(cfg.Requests) > 0 {
					cfg.Requests[0].Method = strings.ToUpper(args[i+1])
					// 更新测试用例以匹配方法
					cfg.Benchmark.TestCase = fmt.Sprintf("http_%s", strings.ToLower(args[i+1]))
				}
				i++
			}
		case "--body":
			if i+1 < len(args) {
				if len(cfg.Requests) > 0 {
					cfg.Requests[0].Body = args[i+1]
				}
				i++
			}
		case "--content-type":
			if i+1 < len(args) {
				if len(cfg.Requests) > 0 {
					if cfg.Requests[0].Headers == nil {
						cfg.Requests[0].Headers = make(map[string]string)
					}
					cfg.Requests[0].Headers["Content-Type"] = args[i+1]
				}
				i++
			}
		case "--header":
			if i+1 < len(args) {
				if len(cfg.Requests) > 0 {
					parts := strings.SplitN(args[i+1], ":", 2)
					if len(parts) == 2 {
						if cfg.Requests[0].Headers == nil {
							cfg.Requests[0].Headers = make(map[string]string)
						}
						cfg.Requests[0].Headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
					}
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
					cfg.Benchmark.Duration = d
				}
				i++
			}
		case "--timeout":
			if i+1 < len(args) {
				if t, err := time.ParseDuration(args[i+1]); err == nil {
					cfg.Connection.Timeout = t
				}
				i++
			}
		case "--ramp-up":
			if i+1 < len(args) {
				if r, err := time.ParseDuration(args[i+1]); err == nil {
					cfg.Benchmark.RampUp = r
				}
				i++
			}
		}
	}

	return cfg
}

// initializeReportManager 初始化报告管理器
func (h *HttpSimpleHandler) initializeReportManager() {
	if h.reportArgs == nil {
		h.reportArgs = reports.DefaultReportArgs()
	}

	reportConfig := h.reportArgs.ToReportConfig("http")

	// 如果加载了核心配置，使用核心配置中的报告设置作为默认值
	coreConfig := h.configManager.GetCoreConfig()
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

	h.reportManager = reports.NewReportManager("http", h.metricsCollector, reportConfig)
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
