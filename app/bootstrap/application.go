package bootstrap

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"abc-runner/app/bootstrap/discovery"
	"abc-runner/app/bootstrap/registry"
)

// Application 应用启动器
type Application struct {
	registry *registry.ProtocolRegistry
	builder  *discovery.AutoDIBuilder
	router   *discovery.CommandRouter
	config   *BootstrapConfig
}

// BootstrapConfig 启动配置
type BootstrapConfig struct {
	AutoDiscovery      bool     `json:"auto_discovery"`
	ProtocolScanPaths  []string `json:"protocol_scan_paths"`
	LoggingEnabled     bool     `json:"logging_enabled"`
	LogDirectory       string   `json:"log_directory"`
}

// DefaultBootstrapConfig 默认启动配置
func DefaultBootstrapConfig() *BootstrapConfig {
	return &BootstrapConfig{
		AutoDiscovery:     true,
		ProtocolScanPaths: []string{"app/adapters/*"},
		LoggingEnabled:    true,
		LogDirectory:      "logs",
	}
}

// NewApplication 创建应用实例
func NewApplication() *Application {
	config := DefaultBootstrapConfig()
	builder := discovery.NewAutoDIBuilder()
	
	return &Application{
		registry: registry.NewProtocolRegistry(),
		builder:  builder,
		router:   discovery.NewCommandRouter(builder),
		config:   config,
	}
}

// Run 运行应用
func (app *Application) Run() error {
	// 初始化日志
	if app.config.LoggingEnabled {
		if err := app.initLogging(); err != nil {
			return fmt.Errorf("failed to initialize logging: %w", err)
		}
		defer app.closeLogging()
	}

	// 自动发现协议
	if app.config.AutoDiscovery {
		if err := app.autoDiscoverProtocols(); err != nil {
			return fmt.Errorf("failed to auto-discover protocols: %w", err)
		}
	}

	// 处理命令行参数
	return app.handleCommand()
}

// Shutdown 优雅关闭
func (app *Application) Shutdown(ctx context.Context) error {
	log.Println("Application shutting down...")
	// TODO: 清理资源
	return nil
}

// initLogging 初始化日志
func (app *Application) initLogging() error {
	// 创建日志目录
	if err := os.MkdirAll(app.config.LogDirectory, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// 生成日志文件名
	timestamp := time.Now().Format("20060102")
	base := fmt.Sprintf("%s/abc-runner_%s", app.config.LogDirectory, timestamp)
	logFileName := base + "_1.log"
	seq := 1

	// 检查文件是否存在，如果存在则递增序号
	for {
		if _, err := os.Stat(logFileName); os.IsNotExist(err) {
			break
		}
		seq++
		logFileName = fmt.Sprintf("%s_%d.log", base, seq)
	}

	// 打开日志文件
	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// 设置日志输出
	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("=== ABC-Runner Bootstrap Started ===")

	return nil
}

// autoDiscoverProtocols 自动发现协议
func (app *Application) autoDiscoverProtocols() error {
	log.Println("Auto-discovering protocols and building dependencies...")
	
	// 使用AutoDIBuilder构建所有组件
	if err := app.builder.Build(); err != nil {
		return fmt.Errorf("auto DI build failed: %w", err)
	}
	
	// 自动注册命令
	if err := app.router.AutoRegister(); err != nil {
		return fmt.Errorf("command auto-registration failed: %w", err)
	}
	
	log.Println("Protocol discovery and DI setup completed")
	return nil
}

// handleCommand 处理命令
func (app *Application) handleCommand() error {
	// 处理全局标志
	help := flag.Bool("help", false, "show help information")
	version := flag.Bool("version", false, "show version information")
	flag.Parse()

	if *help {
		app.showGlobalHelp()
		return nil
	}

	if *version {
		app.showVersion()
		return nil
	}

	// 如果没有命令参数，显示帮助
	if flag.NArg() < 1 {
		fmt.Println("Please specify a command")
		app.showGlobalHelp()
		return nil
	}

	// 执行命令
	command := flag.Arg(0)
	args := flag.Args()[1:]
	
	// 创建执行上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	
	// 使用命令路由器执行
	return app.router.Execute(ctx, command, args)
}

// showGlobalHelp 显示全局帮助信息
func (app *Application) showGlobalHelp() {
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
func (app *Application) showVersion() {
	fmt.Println("abc-runner v3.0.0")
	fmt.Println("Build with extreme simplicity and performance in mind")
	fmt.Printf("Release date: %s\n", time.Now().Format("2006-01-02"))
}

// closeLogging 关闭日志
func (app *Application) closeLogging() {
	log.Println("=== ABC-Runner Bootstrap Shutdown ===")
}