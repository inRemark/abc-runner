package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"abc-runner/servers/internal/logging"
	"abc-runner/servers/internal/monitoring"
	"abc-runner/servers/pkg/tcp"
)

const (
	defaultConfigFile = "config/servers/tcp-server.yaml"
	defaultHost       = "localhost"
	defaultPort       = 9090
)

func main() {
	var (
		configFile = flag.String("config", defaultConfigFile, "Configuration file path")
		host       = flag.String("host", "", "Server host (overrides config)")
		port       = flag.Int("port", 0, "Server port (overrides config)")
		logLevel   = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
		help       = flag.Bool("help", false, "Show help information")
		version    = flag.Bool("version", false, "Show version information")
	)
	
	flag.Parse()
	
	if *help {
		showHelp()
		return
	}
	
	if *version {
		showVersion()
		return
	}
	
	// 初始化日志
	logger := logging.NewLogger(*logLevel)
	logger.Info("Starting TCP test server", map[string]interface{}{
		"config_file": *configFile,
		"log_level":   *logLevel,
	})
	
	// 加载配置
	serverConfig, err := loadConfig(*configFile, *host, *port)
	if err != nil {
		logger.Fatal("Failed to load configuration", err)
		os.Exit(1)
	}
	
	// 验证配置
	if err := serverConfig.Validate(); err != nil {
		logger.Fatal("Configuration validation failed", err)
		os.Exit(1)
	}
	
	logger.Info("Configuration loaded successfully", map[string]interface{}{
		"address":         serverConfig.GetAddress(),
		"max_connections": serverConfig.MaxConnections,
		"echo_mode":       serverConfig.EchoMode,
	})
	
	// 创建指标收集器
	metricsCollector := monitoring.NewMetricsCollector()
	
	// 创建TCP服务端
	server := tcp.NewTCPServer(serverConfig, logger, metricsCollector)
	
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// 启动服务端
	if err := server.Start(ctx); err != nil {
		logger.Fatal("Failed to start TCP server", err)
		os.Exit(1)
	}
	
	logger.Info("TCP server started successfully", map[string]interface{}{
		"address": serverConfig.GetAddress(),
		"pid":     os.Getpid(),
	})
	
	// 等待中断信号
	waitForShutdown(ctx, cancel, server, logger)
}

// loadConfig 加载配置
func loadConfig(configFile, host string, port int) (*tcp.TCPServerConfig, error) {
	// 使用默认配置
	serverConfig := tcp.NewTCPServerConfig()
	
	// 应用命令行覆盖
	if host != "" {
		serverConfig.BaseConfig.Host = host
	}
	
	if port > 0 {
		serverConfig.BaseConfig.Port = port
	}
	
	return serverConfig, nil
}

// waitForShutdown 等待关闭信号
func waitForShutdown(ctx context.Context, cancel context.CancelFunc, server *tcp.TCPServer, logger *logging.Logger) {
	// 创建信号通道
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	// 等待信号
	select {
	case sig := <-sigChan:
		logger.Info("Received shutdown signal", map[string]interface{}{
			"signal": sig.String(),
		})
	case <-ctx.Done():
		logger.Info("Context cancelled, shutting down")
	}
	
	// 开始优雅关闭
	logger.Info("Initiating graceful shutdown...")
	
	// 创建关闭超时上下文
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	
	// 停止服务端
	if err := server.Stop(shutdownCtx); err != nil {
		logger.Error("Error during server shutdown", err)
	} else {
		logger.Info("Server shutdown completed successfully")
	}
	
	cancel()
}

// showHelp 显示帮助信息
func showHelp() {
	fmt.Printf(`TCP Test Server for abc-runner

USAGE:
    tcp-server [OPTIONS]

OPTIONS:
    -config <file>      Configuration file path (default: %s)
    -host <host>        Server host (overrides config file)
    -port <port>        Server port (overrides config file)
    -log-level <level>  Log level: debug, info, warn, error (default: info)
    -help               Show this help message
    -version            Show version information

EXAMPLES:
    # Start with default configuration
    tcp-server

    # Start with custom config file
    tcp-server -config /path/to/config.yaml

    # Start with custom host and port
    tcp-server -host 0.0.0.0 -port 9999

    # Start with debug logging
    tcp-server -log-level debug

FEATURES:
    - Echo server: Receives data and sends it back to the client
    - Connection management: Handles multiple concurrent connections
    - Configurable timeouts: Read, write, and connection timeouts
    - Connection limits: Maximum number of concurrent connections
    - Metrics collection: Real-time connection and performance metrics
    - Graceful shutdown: Properly closes all connections on exit

PROTOCOL:
    The server uses a simple length-prefixed protocol:
    - First 4 bytes: Message length (big-endian)
    - Following bytes: Message data
    
    For simple testing, you can use tools like telnet or nc:
    echo "Hello" | nc localhost 9090

CONFIGURATION:
    Configuration file should be in YAML format. See examples in config/examples/
    directory for detailed configuration options.

SIGNALS:
    SIGINT, SIGTERM  - Graceful shutdown
`, defaultConfigFile)
}

// showVersion 显示版本信息
func showVersion() {
	fmt.Println("TCP Test Server")
	fmt.Println("Version: 1.0.0")
	fmt.Println("Built for: abc-runner performance testing framework")
	fmt.Println("Protocol: TCP with length-prefixed messages")
	
	// 显示构建信息（如果可用）
	if buildDate := os.Getenv("BUILD_DATE"); buildDate != "" {
		fmt.Printf("Build Date: %s\n", buildDate)
	}
	
	if gitCommit := os.Getenv("GIT_COMMIT"); gitCommit != "" {
		fmt.Printf("Git Commit: %s\n", gitCommit)
	}
}