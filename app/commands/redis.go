package commands

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"abc-runner/app/adapters/redis"
	redisconfig "abc-runner/app/adapters/redis/config"
	"abc-runner/app/core/config"
	"abc-runner/app/core/interfaces"
	"abc-runner/app/core/reports"
	"abc-runner/app/core/runner"
	"abc-runner/app/core/utils"
)

// RedisSimpleHandler 简化的Redis命令处理器
type RedisSimpleHandler struct {
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

// NewRedisCommandHandler 创建Redis命令处理器（统一接口）
func NewRedisCommandHandler(adapterFactory interfaces.AdapterFactory) *RedisSimpleHandler {
	handler := &RedisSimpleHandler{
		adapterFactory:    adapterFactory,
		configManager:     config.NewConfigManager(nil), // TODO: 注入配置源工厂
		operationRegistry: utils.NewOperationRegistry(),
		keyGenerator:      utils.NewDefaultKeyGenerator(),
	}

	// 注册Redis操作工厂
	redis.RegisterRedisOperations(handler.operationRegistry)

	return handler
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

	// 2. 验证参数
	if err := h.validateArgs(args); err != nil {
		return fmt.Errorf("argument validation failed: %w", err)
	}

	// 3. 加载配置
	if err := h.loadConfiguration(args); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// 4. 初始化适配器
	if h.adapterFactory != nil {
		h.adapter = h.adapterFactory.CreateRedisAdapter()
	} else {
		// 向后兼容：如果未提供工厂，则直接创建适配器
		h.adapter = redis.NewRedisAdapter(nil) // TODO: 注入指标收集器
	}

	// 5. 连接适配器
	if err := h.adapter.Connect(ctx, h.configManager.GetConfig()); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}
	defer h.adapter.Close()

	// 4. 初始化指标收集器
	h.metricsCollector = h.adapter.GetMetricsCollector()

	// 7. 初始化报告管理器
	h.initializeReportManager()

	// 8. 初始化运行器
	h.runner = runner.NewEnhancedRunner(h.adapter, h.configManager.GetConfig(), h.metricsCollector, h.keyGenerator, h.operationRegistry)

	// 9. 运行测试
	log.Println("Running Redis benchmark...")
	_, err = h.runner.RunBenchmark(ctx)
	if err != nil {
		return fmt.Errorf("benchmark execution failed: %w", err)
	}

	// 10. 生成报告
	log.Println("Generating reports...")
	if err := h.reportManager.GenerateReports(); err != nil {
		return fmt.Errorf("report generation failed: %w", err)
	}

	log.Println("Redis benchmark completed successfully")
	return nil
}

// GetHelp 获取帮助信息
func (h *RedisSimpleHandler) GetHelp() string {
	baseHelp := `Usage: abc-runner redis [OPTIONS]

Redis Performance Testing Tool

Options:
  -h, --host HOST          Server hostname (default: 127.0.0.1)
  -p, --port PORT          Server port (default: 6379)
  -a, --auth PASSWORD      Password for authentication
  --mode MODE              Redis mode: standalone, cluster, sentinel (default: standalone)
  -n, --requests COUNT     Total number of requests (default: 100000)
  -c, --concurrency COUNT  Number of parallel connections (default: 50)
  -d, --data-size BYTES    Data size in bytes (default: 3)
  --ttl SECONDS            Key expiration time in seconds (default: 120)
  -r, --random-keys RANGE  Random key range (0 for sequential, >0 for random)
  -R, --read-ratio PERCENT Read operation percentage (default: 50)
  --case CASE_TYPE         Test case type (default: get)
  --config FILE            Configuration file path
  --core-config FILE       Core configuration file path (default: config/core.yaml)

Examples:
  # Basic benchmark test
  abc-runner redis -h localhost -p 6379 -n 10000 -c 10

  # Cluster mode test with large dataset
  abc-runner redis --mode cluster --host 192.168.1.10 --port 6379 \\
    -n 100000 -c 100 -d 1024 -R 70

  # Sentinel mode with authentication
  abc-runner redis --mode sentinel --host 192.168.1.10 --port 26379 \\
    -a mypassword --master mymaster -n 50000 -c 50

  # Configuration file test
  abc-runner redis --config config/redis.yaml

  # Test with core configuration
  abc-runner redis --config config/redis.yaml --core-config config/core.yaml

For more information: https://docs.abc-runner.com/redis`

	return reports.AddReportArgsToHelp(baseHelp)
}

// loadConfiguration 加载配置
func (h *RedisSimpleHandler) loadConfiguration(args []string) error {
	// 检查是否使用配置文件
	coreConfigPath := h.getCoreConfigFlag(args)
	if coreConfigPath != "" {
		log.Printf("Loading core configuration from %s...", coreConfigPath)
		if err := h.configManager.LoadCoreConfiguration(coreConfigPath); err != nil {
			return fmt.Errorf("failed to load core configuration: %w", err)
		}
	}

	// 使用统一配置加载器
	loader := redisconfig.NewUnifiedRedisConfigLoader()

	var configPath string
	if h.hasConfigFlag(args) {
		configPath = h.getConfigFlagValue(args)
		log.Printf("Loading Redis configuration from file: %s", configPath)
	} else {
		configPath = "" // 让加载器使用默认查找机制
	}

	// 加载配置
	cfg, err := loader.LoadConfig(configPath, args)
	if err != nil {
		return fmt.Errorf("failed to load Redis configuration: %w", err)
	}

	h.configManager.SetConfig(cfg)
	return nil
}

// hasConfigFlag 检查是否有config标志
func (h *RedisSimpleHandler) hasConfigFlag(args []string) bool {
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
func (h *RedisSimpleHandler) getConfigFlagValue(args []string) string {
	for i, arg := range args {
		if (arg == "--config" || arg == "-C") && i+1 < len(args) {
			return args[i+1]
		}
		if strings.HasPrefix(arg, "--config=") {
			return strings.TrimPrefix(arg, "--config=")
		}
	}

	// 使用统一的配置文件查找机制
	foundPath := utils.FindConfigFile("redis")
	if foundPath != "" {
		return foundPath
	}

	// 回退到默认路径
	return "config/redis.yaml"
}

// getCoreConfigFlag 获取核心配置文件路径
func (h *RedisSimpleHandler) getCoreConfigFlag(args []string) string {
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
func (h *RedisSimpleHandler) createConfigFromArgs(args []string) *redisconfig.RedisConfig {
	// 默认配置
	cfg := redisconfig.NewDefaultRedisConfig()

	// 设置默认测试用例
	cfg.BenchMark.Case = "get"

	// 解析命令行参数
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-h", "--host":
			if i+1 < len(args) {
				cfg.Standalone.Addr = args[i+1] + ":" + strings.Split(cfg.Standalone.Addr, ":")[1]
				i++
			}
		case "-p", "--port":
			if i+1 < len(args) {
				parts := strings.Split(cfg.Standalone.Addr, ":")
				cfg.Standalone.Addr = parts[0] + ":" + args[i+1]
				i++
			}
		case "-a", "--auth":
			if i+1 < len(args) {
				cfg.Standalone.Password = args[i+1]
				i++
			}
		case "--mode":
			if i+1 < len(args) {
				cfg.Mode = args[i+1]
				i++
			}
		case "-n", "--requests":
			if i+1 < len(args) {
				if n, err := strconv.Atoi(args[i+1]); err == nil {
					cfg.BenchMark.Total = n
				}
				i++
			}
		case "-c", "--concurrency":
			if i+1 < len(args) {
				if c, err := strconv.Atoi(args[i+1]); err == nil {
					cfg.BenchMark.Parallels = c
				}
				i++
			}
		case "-d", "--data-size":
			if i+1 < len(args) {
				if d, err := strconv.Atoi(args[i+1]); err == nil {
					cfg.BenchMark.DataSize = d
				}
				i++
			}
		case "--ttl":
			if i+1 < len(args) {
				if ttl, err := strconv.Atoi(args[i+1]); err == nil {
					cfg.BenchMark.TTL = ttl
				}
				i++
			}
		case "-r", "--random-keys":
			if i+1 < len(args) {
				if r, err := strconv.Atoi(args[i+1]); err == nil {
					cfg.BenchMark.RandomKeys = r
				}
				i++
			}
		case "-R", "--read-ratio":
			if i+1 < len(args) {
				if r, err := strconv.Atoi(args[i+1]); err == nil {
					cfg.BenchMark.ReadPercent = r
				}
				i++
			}
		case "--case":
			if i+1 < len(args) {
				cfg.BenchMark.Case = args[i+1]
				i++
			}
		}
	}

	return cfg
}

// initializeReportManager 初始化报告管理器
func (h *RedisSimpleHandler) initializeReportManager() {
	if h.reportArgs == nil {
		h.reportArgs = reports.DefaultReportArgs()
	}

	reportConfig := h.reportArgs.ToReportConfig("redis")

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

	h.reportManager = reports.NewReportManager("redis", h.metricsCollector, reportConfig)
}

// validateArgs 验证参数
func (h *RedisSimpleHandler) validateArgs(args []string) error {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--addr":
			if i+1 >= len(args) {
				return fmt.Errorf("--addr requires a address")
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
		case "--mode":
			if i+1 >= len(args) {
				return fmt.Errorf("--mode requires a value")
			}
			mode := args[i+1]
			if mode != "standalone" && mode != "cluster" && mode != "sentinel" {
				return fmt.Errorf("invalid mode: %s (valid: standalone, cluster, sentinel)", mode)
			}
			i++
		case "--read-percent":
			if i+1 >= len(args) {
				return fmt.Errorf("--read-percent requires a value")
			}
			if ratio, err := strconv.ParseFloat(args[i+1], 64); err != nil {
				return fmt.Errorf("invalid read percent: %s", args[i+1])
			} else if ratio < 0 || ratio > 100 {
				return fmt.Errorf("read percent must be between 0 and 100")
			}
			i++
		case "--data-size":
			if i+1 >= len(args) {
				return fmt.Errorf("--data-size requires a value")
			}
			if _, err := strconv.Atoi(args[i+1]); err != nil {
				return fmt.Errorf("invalid value for --data-size: %s", args[i+1])
			}
			i++
		case "--ttl":
			if i+1 >= len(args) {
				return fmt.Errorf("--ttl requires a value")
			}
			if _, err := strconv.Atoi(args[i+1]); err != nil {
				return fmt.Errorf("invalid value for --ttl: %s", args[i+1])
			}
			i++
		case "--random-keys":
			if i+1 >= len(args) {
				return fmt.Errorf("--random-keys requires a value")
			}
			if _, err := strconv.Atoi(args[i+1]); err != nil {
				return fmt.Errorf("invalid value for --random-keys: %s", args[i+1])
			}
			i++
		case "--case":
			if i+1 >= len(args) {
				return fmt.Errorf("--case requires a value")
			}
			i++
		case "--password":
			if i+1 >= len(args) {
				return fmt.Errorf("--password requires a value")
			}
			i++
		case "--db":
			if i+1 >= len(args) {
				return fmt.Errorf("--db requires a value")
			}
			if _, err := strconv.Atoi(args[i+1]); err != nil {
				return fmt.Errorf("invalid value for --db: %s", args[i+1])
			}
			i++
		case "--sentinel-master-name":
			if i+1 >= len(args) {
				return fmt.Errorf("--sentinel-master-name requires a value")
			}
			i++
		case "--sentinel-addrs":
			if i+1 >= len(args) {
				return fmt.Errorf("--sentinel-addrs requires a value")
			}
			i++
		case "--sentinel-password":
			if i+1 >= len(args) {
				return fmt.Errorf("--sentinel-password requires a value")
			}
			i++
		case "--sentinel-db":
			if i+1 >= len(args) {
				return fmt.Errorf("--sentinel-db requires a value")
			}
			if _, err := strconv.Atoi(args[i+1]); err != nil {
				return fmt.Errorf("invalid value for --sentinel-db: %s", args[i+1])
			}
			i++
		case "--cluster-addrs":
			if i+1 >= len(args) {
				return fmt.Errorf("--cluster-addrs requires a value")
			}
			i++
		case "--cluster-password":
			if i+1 >= len(args) {
				return fmt.Errorf("--cluster-password requires a value")
			}
			i++
		}
	}
	return nil
}
