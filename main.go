package main

import (
	"abc-runner/app/adapters/http"
	"abc-runner/app/adapters/kafka"
	"abc-runner/app/adapters/redis"
	"abc-runner/app/commands"
	"abc-runner/app/core/di"
	"abc-runner/app/core/interfaces"
	"abc-runner/app/core/metrics"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// 全局变量
var (
	commandRouter *SimpleCommandRouter
	logFile       *os.File
	container     *di.Container
)

// MetricsCollectorAdapter 新指标系统适配器
// 将新的泛型指标收集器适配到旧的 interfaces.MetricsCollector 接口
type MetricsCollectorAdapter struct {
	baseCollector *metrics.BaseCollector[map[string]interface{}]
	mutex         sync.RWMutex
}

// NewMetricsCollectorAdapter 创建指标收集器适配器
func NewMetricsCollectorAdapter() *MetricsCollectorAdapter {
	config := metrics.DefaultMetricsConfig()

	// 创建协议无关的基础收集器
	protocolData := map[string]interface{}{
		"application": "abc-runner",
		"version":     "0.2.0",
	}

	baseCollector := metrics.NewBaseCollector(config, protocolData)

	return &MetricsCollectorAdapter{
		baseCollector: baseCollector,
	}
}

// RecordOperation 实现 interfaces.MetricsCollector 接口
func (m *MetricsCollectorAdapter) RecordOperation(result *interfaces.OperationResult) {
	m.baseCollector.Record(result)
}

// GetMetrics 实现 interfaces.MetricsCollector 接口
func (m *MetricsCollectorAdapter) GetMetrics() *interfaces.Metrics {
	snapshot := m.baseCollector.Snapshot()

	// 转换新指标格式到旧格式
	return &interfaces.Metrics{
		TotalOps:   snapshot.Core.Operations.Total,
		SuccessOps: snapshot.Core.Operations.Success,
		FailedOps:  snapshot.Core.Operations.Failed,
		ReadOps:    snapshot.Core.Operations.Read,
		WriteOps:   snapshot.Core.Operations.Write,
		AvgLatency: snapshot.Core.Latency.Average,
		MinLatency: snapshot.Core.Latency.Min,
		MaxLatency: snapshot.Core.Latency.Max,
		P90Latency: snapshot.Core.Latency.P90,
		P95Latency: snapshot.Core.Latency.P95,
		P99Latency: snapshot.Core.Latency.P99,
		ErrorRate:  float64(snapshot.Core.Operations.Failed) / float64(snapshot.Core.Operations.Total) * 100,
		StartTime:  time.Now().Add(-snapshot.Core.Duration),
		EndTime:    time.Now(),
		Duration:   snapshot.Core.Duration,
		RPS:        int32(snapshot.Core.Throughput.RPS),
	}
}

// Reset 实现 interfaces.MetricsCollector 接口
func (m *MetricsCollectorAdapter) Reset() {
	m.baseCollector.Reset()
}

// Export 实现 interfaces.MetricsCollector 接口
func (m *MetricsCollectorAdapter) Export() map[string]interface{} {
	snapshot := m.baseCollector.Snapshot()

	// 根据项目记忆，确保延迟字段为 int64 类型（纳秒）
	return map[string]interface{}{
		"total_ops":    snapshot.Core.Operations.Total,
		"success_ops":  snapshot.Core.Operations.Success,
		"failed_ops":   snapshot.Core.Operations.Failed,
		"read_ops":     snapshot.Core.Operations.Read,
		"write_ops":    snapshot.Core.Operations.Write,
		"success_rate": snapshot.Core.Operations.Rate,
		"rps":          snapshot.Core.Throughput.RPS,
		"avg_latency":  int64(snapshot.Core.Latency.Average),
		"min_latency":  int64(snapshot.Core.Latency.Min),
		"max_latency":  int64(snapshot.Core.Latency.Max),
		"p90_latency":  int64(snapshot.Core.Latency.P90),
		"p95_latency":  int64(snapshot.Core.Latency.P95),
		"p99_latency":  int64(snapshot.Core.Latency.P99),
		"error_rate":   float64(snapshot.Core.Operations.Failed) / float64(snapshot.Core.Operations.Total) * 100,
		"start_time":   time.Now().Add(-snapshot.Core.Duration),
		"end_time":     time.Now(),
		"duration":     int64(snapshot.Core.Duration),
		// 系统指标
		"memory_usage":    snapshot.System.Memory.Usage,
		"goroutine_count": snapshot.System.Goroutine.Active,
		"gc_count":        snapshot.System.GC.NumGC,
		// 协议数据
		"protocol_data": snapshot.Protocol,
	}
}

// GetSnapshot 获取新指标快照（额外方法，供新系统使用）
func (m *MetricsCollectorAdapter) GetSnapshot() *metrics.MetricsSnapshot[map[string]interface{}] {
	return m.baseCollector.Snapshot()
}

// Stop 停止收集器
func (m *MetricsCollectorAdapter) Stop() {
	m.baseCollector.Stop()
}

// CustomAdapterFactory 自定义适配器工厂
type CustomAdapterFactory struct {
	metricsCollector interfaces.MetricsCollector
}

// CreateRedisAdapter 创建Redis适配器
func (f *CustomAdapterFactory) CreateRedisAdapter() interfaces.ProtocolAdapter {
	return redis.NewRedisAdapter(f.metricsCollector)
}

// CreateHttpAdapter 创建HTTP适配器
func (f *CustomAdapterFactory) CreateHttpAdapter() interfaces.ProtocolAdapter {
	return http.NewHttpAdapter(f.metricsCollector)
}

// CreateKafkaAdapter 创建Kafka适配器
func (f *CustomAdapterFactory) CreateKafkaAdapter() interfaces.ProtocolAdapter {
	return kafka.NewKafkaAdapter(f.metricsCollector)
}

func main() {
	initLogging()
	defer closeLogFile()

	// 初始化依赖注入容器
	container = di.NewContainer()

	// 初始化简化命令系统
	if err := initializeCommandSystem(); err != nil {
		log.Fatalf("Failed to initialize command system: %v", err)
	}

	// 执行命令
	if err := executeCommand(); err != nil {
		log.Fatalf("Command execution failed: %v", err)
	}
}

// initLogging 初始化日志配置
func initLogging() {
	// 创建日志目录
	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Printf("Warning: failed to create logs directory: %v", err)
		return
	}

	// 生成日志文件名
	timestamp := time.Now().Format("20060102")
	base := fmt.Sprintf("logs/abc-runner_%s", timestamp)
	logFileName := base + "_1.log"
	seq := 1

	// 检查文件是否存在，如果存在则递增序号
	for {
		if _, err := os.Stat(logFileName); os.IsNotExist(err) {
			break
		}
		logFileName = fmt.Sprintf("%s_%d.log", base, seq)
		seq++
	}

	// 打开日志文件
	var err error
	logFile, err = os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("Warning: failed to open log file: %v", err)
		return
	}

	// 设置日志输出到文件
	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("=== Application started ===")
}

// initializeCommandSystem 初始化简化命令系统
func initializeCommandSystem() error {
	log.Println("Initializing simplified command system...")

	commandRouter = NewSimpleCommandRouter()

	// 注册基础命令处理器
	if err := registerCommandHandlers(); err != nil {
		return fmt.Errorf("failed to register command handlers: %w", err)
	}

	log.Println("Command system initialized successfully")
	return nil
}

// registerCommandHandlers 注册命令处理器
func registerCommandHandlers() error {
	// 使用新的泛型指标系统适配器
	metricsCollector := NewMetricsCollectorAdapter()

	// 创建自定义适配器工厂，注入新的指标收集器
	adapterFactory := &CustomAdapterFactory{
		metricsCollector: metricsCollector,
	}

	// 向DI容器注册具体实现（运行时注册）
	container.Provide(func() interfaces.MetricsCollector {
		return metricsCollector
	})
	container.Provide(func() interfaces.AdapterFactory {
		return adapterFactory
	})

	// 使用DI容器获取依赖并注册命令
	return container.Invoke(func(factory interfaces.AdapterFactory) {
		// 注册Redis命令（使用DI工厂）
		redisHandler := commands.NewRedisCommandHandler(factory)
		commandRouter.RegisterCommand("redis", redisHandler)
		commandRouter.RegisterAlias("r", "redis")

		// 注册HTTP命令（使用DI工厂）
		httpHandler := commands.NewHttpCommandHandler(factory)
		commandRouter.RegisterCommand("http", httpHandler)
		commandRouter.RegisterAlias("h", "http")

		// 注册Kafka命令（使用DI工厂）
		kafkaHandler := commands.NewKafkaCommandHandler(factory)
		commandRouter.RegisterCommand("kafka", kafkaHandler)
		commandRouter.RegisterAlias("k", "kafka")
	})
}

// executeCommand 执行命令
func executeCommand() error {
	// 处理全局标志
	if handled := handleGlobalFlags(); handled {
		return nil
	}

	// 解析命令和参数
	subCmd, args := parseCommandArgs()
	if subCmd == "" {
		fmt.Println("Please specify a command")
		showGlobalHelp()
		return nil
	}

	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// 使用简化路由器执行命令
	return commandRouter.Execute(ctx, subCmd, args)
}

// handleGlobalFlags 处理全局标志
func handleGlobalFlags() bool {
	help := flag.Bool("help", false, "show help information")
	version := flag.Bool("version", false, "show version information")
	flag.Parse()

	if *help {
		showGlobalHelp()
		return true
	}

	if *version {
		showVersion()
		return true
	}

	return false
}

// parseCommandArgs 解析命令和参数
func parseCommandArgs() (string, []string) {
	if flag.NArg() < 1 {
		return "", nil
	}

	subCmd := flag.Arg(0)
	args := flag.Args()[1:]

	// 检查是否是帮助命令
	if subCmd == "help" || subCmd == "-h" || subCmd == "--help" {
		showGlobalHelp()
		return "", nil
	}

	return subCmd, args
}

// showGlobalHelp 显示全局帮助信息
func showGlobalHelp() {
	fmt.Println("abc-runner - Unified Performance Testing Tool")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  abc-runner <command> [options]")
	fmt.Println()
	fmt.Println("COMMANDS:")
	fmt.Println("  redis, r         Redis performance testing")
	fmt.Println("  http, h          HTTP load testing")
	fmt.Println("  kafka, k         Kafka performance testing")
	fmt.Println()
	fmt.Println("GLOBAL OPTIONS:")
	fmt.Println("  --help, -h       Show help information")
	fmt.Println("  --version, -v    Show version information")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  abc-runner redis --config config/redis.yaml")
	fmt.Println("  abc-runner r -n 1000 -c 10")
	fmt.Println("  abc-runner http --url http://localhost:8080")
	fmt.Println("  abc-runner kafka --brokers localhost:9092")
	fmt.Println()
	fmt.Println("Use \"abc-runner <command> --help\" for more information about a command.")
}

// showVersion 显示版本信息
func showVersion() {
	version := "0.2.0"
	releaseDate := "2025-09-8"
	fmt.Printf("Version: %s\n", version)
	fmt.Printf("Release date: %s\n", releaseDate)
}

// SimpleCommandRouter 简化的命令路由器
type SimpleCommandRouter struct {
	commands map[string]CommandHandler
	aliases  map[string]string
}

// CommandHandler 简化的命令处理器接口
type CommandHandler interface {
	Execute(ctx context.Context, args []string) error
	GetHelp() string
}

// NewSimpleCommandRouter 创建简化命令路由器
func NewSimpleCommandRouter() *SimpleCommandRouter {
	return &SimpleCommandRouter{
		commands: make(map[string]CommandHandler),
		aliases:  make(map[string]string),
	}
}

// RegisterCommand 注册命令
func (r *SimpleCommandRouter) RegisterCommand(name string, handler CommandHandler) {
	r.commands[name] = handler
	log.Printf("Registered command: %s", name)
}

// RegisterAlias 注册别名
func (r *SimpleCommandRouter) RegisterAlias(alias, command string) {
	r.aliases[alias] = command
	log.Printf("Registered alias: %s -> %s", alias, command)
}

// Execute 执行命令
func (r *SimpleCommandRouter) Execute(ctx context.Context, command string, args []string) error {
	// 解析别名
	if target, exists := r.aliases[command]; exists {
		command = target
	}

	// 查找命令处理器
	handler, exists := r.commands[command]
	if !exists {
		return fmt.Errorf("unknown command: %s", command)
	}

	// 执行命令
	return handler.Execute(ctx, args)
}

// closeLogFile 关闭日志文件
func closeLogFile() {
	if logFile != nil {
		log.Println("=== Application shutdown ===")
		err := logFile.Close()
		if err != nil {
			fmt.Printf("failed to close log file: %v\n", err)
		}
	}
}
