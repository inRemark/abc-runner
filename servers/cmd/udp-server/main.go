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
	"abc-runner/servers/pkg/udp"
)

const (
	defaultConfigFile = "config/servers/udp-server.yaml"
	defaultHost       = "localhost"
	defaultPort       = 9091
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
	logger.Info("Starting UDP test server", map[string]interface{}{
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
		"address":           serverConfig.GetAddress(),
		"echo_mode":         serverConfig.EchoMode,
		"packet_loss_rate":  serverConfig.PacketLossRate,
		"enable_multicast":  serverConfig.EnableMulticast,
		"enable_broadcast":  serverConfig.EnableBroadcast,
	})
	
	// 创建指标收集器
	metricsCollector := monitoring.NewMetricsCollector()
	
	// 创建UDP服务端
	server := udp.NewUDPServer(serverConfig, logger, metricsCollector)
	
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// 启动服务端
	if err := server.Start(ctx); err != nil {
		logger.Fatal("Failed to start UDP server", err)
		os.Exit(1)
	}
	
	logger.Info("UDP server started successfully", map[string]interface{}{
		"address": serverConfig.GetAddress(),
		"pid":     os.Getpid(),
	})
	
	// 等待中断信号
	waitForShutdown(ctx, cancel, server, logger)
}

// loadConfig 加载配置
func loadConfig(configFile, host string, port int) (*udp.UDPServerConfig, error) {
	// 使用默认配置
	serverConfig := udp.NewUDPServerConfig()
	
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
func waitForShutdown(ctx context.Context, cancel context.CancelFunc, server *udp.UDPServer, logger *logging.Logger) {
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
	fmt.Printf(`UDP Test Server for abc-runner

USAGE:
    udp-server [OPTIONS]

OPTIONS:
    -config <file>      Configuration file path (default: %s)
    -host <host>        Server host (overrides config file)
    -port <port>        Server port (overrides config file)
    -log-level <level>  Log level: debug, info, warn, error (default: info)
    -help               Show this help message
    -version            Show version information

EXAMPLES:
    # Start with default configuration
    udp-server

    # Start with custom config file
    udp-server -config /path/to/config.yaml

    # Start with custom host and port
    udp-server -host 0.0.0.0 -port 8888

    # Start with debug logging
    udp-server -log-level debug

FEATURES:
    - Echo server: Receives UDP packets and sends them back
    - Packet loss simulation: Configurable packet drop rate
    - Multicast support: Join multicast groups (configurable)
    - Broadcast support: Handle broadcast packets (configurable)
    - Metrics collection: Real-time packet and performance metrics
    - Graceful shutdown: Properly closes UDP socket on exit

PROTOCOL:
    UDP is connectionless, so packets are processed independently.
    The server can handle:
    - Unicast packets (point-to-point)
    - Multicast packets (one-to-many)
    - Broadcast packets (one-to-all in network)
    
    For testing, you can use tools like nc (netcat) or custom UDP clients:
    echo "Hello UDP" | nc -u localhost 9091

CONFIGURATION:
    Configuration file should be in YAML format. Key options:
    - echo_mode: Whether to echo received packets back
    - packet_loss_rate: Simulate packet loss (0.0-1.0)
    - max_packet_size: Maximum UDP packet size
    - enable_multicast: Enable multicast support
    - enable_broadcast: Enable broadcast support

SIGNALS:
    SIGINT, SIGTERM  - Graceful shutdown
`, defaultConfigFile)
}

// showVersion 显示版本信息
func showVersion() {
	fmt.Println("UDP Test Server")
	fmt.Println("Version: 1.0.0")
	fmt.Println("Built for: abc-runner performance testing framework")
	fmt.Println("Protocol: UDP (User Datagram Protocol)")
	
	// 显示构建信息（如果可用）
	if buildDate := os.Getenv("BUILD_DATE"); buildDate != "" {
		fmt.Printf("Build Date: %s\n", buildDate)
	}
	
	if gitCommit := os.Getenv("GIT_COMMIT"); gitCommit != "" {
		fmt.Printf("Git Commit: %s\n", gitCommit)
	}
}