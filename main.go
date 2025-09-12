package main

import (
	"abc-runner/app/commands"
	"abc-runner/app/core/di"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

// 全局变量
var (
	commandRouter *SimpleCommandRouter
	logFile       *os.File
	container     *di.Container
)

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
	// 注册Redis命令
	redisHandler := commands.NewRedisCommandHandler(nil) // TODO: 注入适配器工厂
	commandRouter.RegisterCommand("redis", redisHandler)
	commandRouter.RegisterAlias("r", "redis")

	// 注册HTTP命令
	httpHandler := commands.NewHttpCommandHandler(nil) // TODO: 注入适配器工厂
	commandRouter.RegisterCommand("http", httpHandler)
	commandRouter.RegisterAlias("h", "http")

	// 注册Kafka命令
	kafkaHandler := commands.NewKafkaCommandHandler(nil) // TODO: 注入适配器工厂
	commandRouter.RegisterCommand("kafka", kafkaHandler)
	commandRouter.RegisterAlias("k", "kafka")

	return nil
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
