package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"abc-runner/servers/internal/config"
	"abc-runner/servers/internal/logging"
	"abc-runner/servers/internal/monitoring"
	"abc-runner/servers/pkg/http"
)

const (
	defaultConfigFile = "config/servers/http-server.yaml"
	defaultHost       = "localhost"
	defaultPort       = 8080
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
	logger.Info("Starting HTTP test server", map[string]interface{}{
		"config_file": *configFile,
		"log_level":   *logLevel,
	})

	// 加载配置
	configLoader := config.NewHTTPConfigLoader()
	serverConfig, err := loadConfig(configLoader, *configFile, *host, *port)
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
		"address": serverConfig.GetAddress(),
		"tls":     serverConfig.TLS.Enabled,
	})

	// 创建指标收集器
	metricsCollector := monitoring.NewMetricsCollector()

	// 创建HTTP服务端
	server := http.NewHTTPServer(serverConfig, logger, metricsCollector)

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动服务端
	if err := server.Start(ctx); err != nil {
		logger.Fatal("Failed to start HTTP server", err)
		os.Exit(1)
	}

	logger.Info("HTTP server started successfully", map[string]interface{}{
		"address": serverConfig.GetAddress(),
		"pid":     os.Getpid(),
	})

	// 等待中断信号
	waitForShutdown(ctx, cancel, server, logger)
}

// loadConfig 加载配置
func loadConfig(loader *config.HTTPConfigLoader, configFile, host string, port int) (*http.HTTPServerConfig, error) {
	var serverConfig *http.HTTPServerConfig
	var err error

	// 检查配置文件是否存在
	if _, checkErr := os.Stat(configFile); checkErr == nil {
		// 从文件加载配置
		serverConfig, err = loader.LoadFromFile(configFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load config from file: %w", err)
		}
	} else {
		// 使用默认配置
		log.Printf("Config file not found, using default configuration")
		serverConfig = http.NewHTTPServerConfig()
	}

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
func waitForShutdown(ctx context.Context, cancel context.CancelFunc, server *http.HTTPServer, logger *logging.Logger) {
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
	fmt.Printf(`HTTP Test Server for abc-runner

USAGE:
    http-server [OPTIONS]

OPTIONS:
    -config <file>      Configuration file path (default: %s)
    -host <host>        Server host (overrides config file)
    -port <port>        Server port (overrides config file)
    -log-level <level>  Log level: debug, info, warn, error (default: info)
    -help               Show this help message
    -version            Show version information

EXAMPLES:
    # Start with default configuration
    http-server

    # Start with custom config file
    http-server -config /path/to/config.yaml

    # Start with custom host and port
    http-server -host 0.0.0.0 -port 9090

    # Start with debug logging
    http-server -log-level debug

ENDPOINTS:
    /              - Root endpoint with server information
    /health        - Health check endpoint
    /metrics       - Metrics endpoint
    /echo          - Echo request body and headers
    /delay         - Test endpoint with configurable delay (?delay=1s)
    /status        - Test endpoint with configurable status code (?code=404)
    /data          - Test endpoint with configurable response size (?size=1024)

CONFIGURATION:
    Configuration file should be in YAML format. See examples in config/examples/
    directory for detailed configuration options.

SIGNALS:
    SIGINT, SIGTERM  - Graceful shutdown
`, defaultConfigFile)
}

// showVersion 显示版本信息
func showVersion() {
	fmt.Println("HTTP Test Server")
	fmt.Println("Version: 1.0.0")
	fmt.Println("Built for: abc-runner performance testing framework")
	fmt.Println("Protocol: HTTP/1.1")

	// 显示构建信息（如果可用）
	if buildDate := os.Getenv("BUILD_DATE"); buildDate != "" {
		fmt.Printf("Build Date: %s\n", buildDate)
	}

	if gitCommit := os.Getenv("GIT_COMMIT"); gitCommit != "" {
		fmt.Printf("Git Commit: %s\n", gitCommit)
	}
}

// getConfigFilePath 获取配置文件路径
func getConfigFilePath(configFile string) string {
	// 如果是绝对路径，直接返回
	if filepath.IsAbs(configFile) {
		return configFile
	}

	// 尝试相对于当前目录
	if _, err := os.Stat(configFile); err == nil {
		return configFile
	}

	// 尝试相对于可执行文件目录
	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)
		configPath := filepath.Join(execDir, configFile)
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
	}

	// 尝试相对于工作目录的上级目录（通常用于开发环境）
	if wd, err := os.Getwd(); err == nil {
		parentDir := filepath.Dir(wd)
		configPath := filepath.Join(parentDir, configFile)
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
	}

	// 返回原始路径
	return configFile
}
